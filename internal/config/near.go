package config

import (
	"github.com/rarimo/near-go/common"
	"gitlab.com/distributed_lab/figure/v3"
	"reflect"

	"github.com/rarimo/near-go/nearclient"
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type nearConfig struct {
	Client        *nearclient.Client `fig:"addr,required"`
	BridgeAddress common.AccountID   `fig:"bridge_address,required"`
}

func (c *config) Near() *nearConfig {
	return c.near.Do(func() interface{} {
		var cfg nearConfig

		err := figure.
			Out(&cfg).
			From(kv.MustGetStringMap(c.getter, "near")).
			With(figure.BaseHooks, nearHooks).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out near"))
		}

		return &cfg
	}).(*nearConfig)
}

var nearHooks = figure.Hooks{
	"*nearclient.Client": func(raw interface{}) (reflect.Value, error) {
		v, err := cast.ToStringE(raw)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "expected string")
		}

		client, err := nearclient.New(v)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "failed to create near client")
		}

		return reflect.ValueOf(client), nil
	},
	"types.AccountID": func(raw interface{}) (reflect.Value, error) {
		v, err := cast.ToStringE(raw)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "expected string")
		}

		return reflect.ValueOf(v), nil
	},
}
