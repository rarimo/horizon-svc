package solana

import (
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/rarimo/horizon-svc/internal/chain_gateway"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/metadata_fetcher"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
)

type solanaProxy struct {
	cli     *rpc.Client
	fetcher metadata_fetcher.Client
	gateway chain_gateway.ChainGateway
	chain   string
}

func New(chain data.Chain, fetcher metadata_fetcher.Client, gateway chain_gateway.ChainGateway) types.Proxy {
	return &solanaProxy{
		chain:   chain.Name,
		cli:     rpc.New(chain.Rpc),
		fetcher: fetcher,
		gateway: gateway,
	}
}
