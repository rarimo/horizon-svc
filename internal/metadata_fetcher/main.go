package metadata_fetcher

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/ipfs"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

var ErrNonRetriable = errors.New("non retriable")

const maxBodySize = 10 << 20 // 10 MB

type Client interface {
	GetMetadata(ctx context.Context, uri string, nftId string) (*data.NftMetadata, error)
}

func New(httpClient *http.Client, ipfsGateway ipfs.Gateway) Client {
	return &client{
		httpClient:  httpClient,
		ipfsGateway: ipfsGateway,
	}
}

type client struct {
	httpClient  *http.Client
	ipfsGateway ipfs.Gateway
}

func (c *client) GetMetadata(ctx context.Context, uri string, nftId string) (*data.NftMetadata, error) {
	tokenURL, err := parseUri(uri, nftId)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	payload, err := c.getPayload(ctx, tokenURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metadata payload")
	}

	return extractMetadata(payload, tokenURL.String())
}

func (c *client) getPayload(ctx context.Context, parsedURL *url.URL) ([]byte, error) {
	var payload []byte
	var err error

	switch parsedURL.Scheme {
	case "http", "https":
		payload, err = c.getHttpMeta(ctx, *parsedURL)
	case "ipfs":
		payload, err = c.ipfsGateway.Get(ctx, parsedURL.Host+parsedURL.Path)
	case "data":
		payload, err = parseDataContentType(*parsedURL)
	case "":
		return nil, errors.Wrap(ErrNonRetriable, "empty url schema")
	default:
		return nil, errors.New("unexpected schema")
	}

	if err != nil {
		if errors.Cause(err) == ipfs.ErrNotFound {
			return nil, errors.Wrap(ErrNonRetriable, "metadata not found")
		}
		return nil, errors.Wrap(err, "failed to get metadata")
	}

	return sanitizeJSON(payload), nil
}

func (c *client) getHttpMeta(ctx context.Context, metaURL url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metaURL.String(), nil)
	if err != nil {
		panic(errors.Wrap(err, "failed to create request"))
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform request")
	}

	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		reader := io.LimitReader(resp.Body, maxBodySize)
		result, err := io.ReadAll(reader)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read body")
		}

		return result, nil
	case http.StatusForbidden, http.StatusUnauthorized:
		return nil, errors.Wrap(ErrNonRetriable, "not allowed to access metadata", logan.F{
			"url": metaURL.String(),
		})
	case http.StatusNotFound:
		return nil, errors.Wrap(ErrNonRetriable, "metadata not found", logan.F{
			"url": metaURL.String(),
		})
	default:
		return nil, errors.From(errors.New("unexpected status code"), logan.F{
			"status_code": resp.StatusCode,
		})
	}
}
