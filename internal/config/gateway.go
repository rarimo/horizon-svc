package config

import (
	"reflect"

	"github.com/rarimo/horizon-svc/internal/data/mem"
	"github.com/rarimo/horizon-svc/internal/rpc"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type ChainGateway struct {
	Gateways []ChainGatewayCfg `fig:"gateways,required"`
}

type ChainGatewayCfg struct {
	Name string                   `fig:"name,required"`
	Type tokenmanager.NetworkType `fig:"type,required"`
	Rpc  *rpc.Client              `fig:"rpc,required"`
}

func (c *config) ChainGateway() *ChainGateway {
	return c.chainGateway.Do(func() interface{} {
		var config ChainGateway

		err := figure.
			Out(&config).
			With(figure.BaseHooks, chainGatewayHooks, mem.Hooks).
			From(kv.MustGetStringMap(c.getter, "chain_gateway")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure chain gateway"))
		}

		return &config
	}).(*ChainGateway)
}

var chainGatewayHooks = figure.Hooks{
	"*rpc.Client": func(raw interface{}) (reflect.Value, error) {
		v, err := cast.ToStringE(raw)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "expected string")
		}

		client, err := rpc.NewClient(v)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "failed to create client")
		}

		return reflect.ValueOf(client), nil
	},
}
