package rpc

import (
	"context"
	"encoding/json"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/near-go/nearclient/jsonrpc"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

var MaxFetchNftLimit uint64 = 40
var FirstPageNumber uint64 = 1

// Client defines typed wrappers for the Quicknode RPC API.
type Client struct {
	c *jsonrpc.JsonRpcClient
}

func NewClient(rawUrl string) (*Client, error) {
	cli, err := jsonrpc.New(rawUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create rpc client")
	}
	return &Client{c: cli}, nil
}

func (r *Client) FetchNfts(ctx context.Context, walletAddress string, contractAddress *string, limit uint64, page *uint64) (*FetchNftResult, error) {
	res, err := r.c.CallRPC(ctx, "qn_fetchNFTs", prepareArgs(walletAddress, contractAddress, limit, page))
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch nfts")
	}

	var result FetchNftResult
	err = json.Unmarshal(res.Result, &result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal result")
	}
	return &result, nil
}

func prepareArgs(walletAddress string, contractAddress *string, limit uint64, page *uint64) map[string]interface{} {
	args := map[string]interface{}{
		"wallet":     walletAddress,
		"perPage":    limit,
		"omitFields": []string{"provenance", "creators"},
		"page":       data.GetOrDefaultUint64Ptr(page, &FirstPageNumber),
	}
	if limit > MaxFetchNftLimit {
		args["perPage"] = MaxFetchNftLimit
	}
	if data.NotNilStrPtr(contractAddress) {
		args["contracts"] = []string{*contractAddress}
	}

	return args
}
