package evm

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/metadata_fetcher"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type evmProxy struct {
	cli     *ethclient.Client
	fetcher metadata_fetcher.Client
}

func New(chain data.Chain, fetcher metadata_fetcher.Client) types.Proxy {
	cli, err := ethclient.Dial(chain.Rpc)
	if err != nil {
		panic(errors.Wrap(err, "failed to connect to ethereum node", logan.F{
			"chain": chain.Name,
			"rpc":   chain.Rpc,
		}))
	}
	return &evmProxy{
		cli:     cli,
		fetcher: fetcher,
	}
}
