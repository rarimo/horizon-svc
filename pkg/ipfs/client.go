package ipfs

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"sync"

	"gitlab.com/distributed_lab/logan/v3"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

func NewIpfsGateway(credentials []IpfsCredentials) Gateway {
	return &ipfsGateway{
		client:      &http.Client{},
		credentials: credentials,
	}
}

type ipfsGateway struct {
	client       *http.Client
	credentials  []IpfsCredentials
	credentialsI int // credentials are used in round-robin fashion
	mu           sync.Mutex
}

func (c *ipfsGateway) GetReader(ctx context.Context, resourceID string) (io.ReadCloser, error) {
	c.mu.Lock()
	credentials := c.credentials[c.credentialsI]
	c.credentialsI = (c.credentialsI + 1) % len(c.credentials)
	c.mu.Unlock()

	gatewayURL := *credentials.GatewayURL
	gatewayURL.Path = path.Join(gatewayURL.Path, "api/v0/cat")
	query := gatewayURL.Query()
	query.Set("arg", resourceID)
	gatewayURL.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gatewayURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.SetBasicAuth(credentials.ProjectID, credentials.ProjectSecret)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform request")
	}

	if resp.StatusCode == http.StatusOK {
		const maxBodySize = 2e7 // ~ 20 mb
		return http.MaxBytesReader(nil, resp.Body, maxBodySize), nil
	}

	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusInternalServerError:
		respErr, err := readInternalServerError(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read internal server error")
		}

		return nil, errors.Wrap(respErr.ToError(), "received internal server error")
	default:
		return nil, errors.From(errors.New("unexpected status code"), logan.F{
			"status_code": resp.StatusCode,
		})
	}
}

func (c *ipfsGateway) Get(ctx context.Context, resourceID string) ([]byte, error) {
	body, err := c.GetReader(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(body)
}

func readInternalServerError(body io.Reader) (*errResponse, error) {
	payload, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	var errResp errResponse
	err = json.Unmarshal(payload, &errResp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal error response", logan.F{
			"payload": string(payload),
		})
	}

	return &errResp, nil
}
