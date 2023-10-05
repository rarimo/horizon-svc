package config

import (
	"reflect"

	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type solanaConfig struct {
	RPCClient            *rpc.Client      `fig:"rpc,required"`
	ProgramID            solana.PublicKey `fig:"program_id,required"`              // config value should be base58 encoded
	BridgeAdminPublicKey solana.PublicKey `fig:"bridge_admin_public_key,required"` // config value should be base58 encoded
}

func (c *config) Solana() *solanaConfig {
	return c.solana.Do(func() interface{} {
		var solana solanaConfig

		err := figure.
			Out(&solana).
			From(kv.MustGetStringMap(c.getter, "solana")).
			With(figure.BaseHooks, solanaHooks).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out solana"))
		}

		return &solana
	}).(*solanaConfig)
}

var solanaHooks = figure.Hooks{
	"*rpc.Client": func(raw interface{}) (reflect.Value, error) {
		v, err := cast.ToStringE(raw)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "expected string")
		}

		return reflect.ValueOf(rpc.New(v)), nil
	},
	"solana.PublicKey": func(raw interface{}) (reflect.Value, error) {
		v, err := cast.ToStringE(raw)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "expected string")
		}

		key, err := solana.PublicKeyFromBase58(v)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "failed to parse public key from base58 formatted string")
		}

		return reflect.ValueOf(key), nil
	},
}
