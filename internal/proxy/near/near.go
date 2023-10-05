package near

import (
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/metadata_fetcher"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"github.com/rarimo/near-go/nearclient"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type nearProxy struct {
	cli     *nearclient.Client
	fetcher metadata_fetcher.Client
}

func New(chain data.Chain, fetcher metadata_fetcher.Client) types.Proxy {
	cli, err := nearclient.New(chain.Rpc)
	if err != nil {
		panic(errors.Wrap(err, "failed to create near client", logan.F{
			"chain": chain.Name,
			"rpc":   chain.Rpc,
		}))
	}
	return &nearProxy{
		cli:     cli,
		fetcher: fetcher,
	}
}
