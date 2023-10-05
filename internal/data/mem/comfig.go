package mem

import (
	"sort"

	"github.com/rarimo/horizon-svc/internal/data"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func NewChainer(getter kv.Getter) Chainer {
	return &chainer{
		getter: getter,
	}
}

type Chainer interface {
	ChainsQ() data.ChainsQ
}

type chainer struct {
	getter kv.Getter
	once   comfig.Once
}

func (c *chainer) ChainsQ() data.ChainsQ {
	return NewChainsQ(c.readConfig())
}

func (c *chainer) readConfig() []data.Chain {
	return c.once.Do(func() interface{} {
		cfg := struct {
			Chains []data.Chain `fig:"chains,required"`
		}{}

		err := figure.
			Out(&cfg).
			With(figure.BaseHooks, Hooks).
			From(kv.MustGetStringMap(c.getter, "data")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out chains"))
		}

		sort.Slice(cfg.Chains, func(i, j int) bool {
			return cfg.Chains[i].ID < cfg.Chains[j].ID
		})

		return cfg.Chains
	}).([]data.Chain)
}

var Hooks = figure.Hooks{
	"tokenmanager.NetworkType": figure.BaseHooks["int32"],
}
