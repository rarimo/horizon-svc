package config

import (
	"reflect"

	"gitlab.com/distributed_lab/kit/kv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type evmConfig struct {
	BridgeContract common.Address    `fig:"bridge_contract,required"`
	RPCClient      *ethclient.Client `fig:"rpc,required"`
}

func (c *config) EVM() *evmConfig {
	return c.evm.Do(func() interface{} {
		var evm evmConfig

		err := figure.
			Out(&evm).
			From(kv.MustGetStringMap(c.getter, "evm")).
			With(figure.BaseHooks, figure.EthereumHooks).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out evm"))
		}

		return &evm
	}).(*evmConfig)
}

var evmHooks = figure.Hooks{
	"common.Address": func(raw interface{}) (reflect.Value, error) {
		v, err := cast.ToStringE(raw)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "expected string")
		}

		return reflect.ValueOf(common.HexToAddress(v)), nil
	},
	"*ethclient.Client": func(raw interface{}) (reflect.Value, error) {
		v, err := cast.ToStringE(raw)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "expected string")
		}

		client, err := ethclient.Dial(v)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "failed to dial eth rpc")
		}

		return reflect.ValueOf(client), nil
	},
}
