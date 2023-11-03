package solana

import (
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/metadata_fetcher"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
)

type solanaProxy struct {
	cli     *rpc.Client
	fetcher metadata_fetcher.Client
	chain   string
}

func New(chain data.Chain, fetcher metadata_fetcher.Client) types.Proxy {
	return &solanaProxy{
		chain:   chain.Name,
		cli:     rpc.New(chain.Rpc),
		fetcher: fetcher,
	}
}
