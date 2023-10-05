package config

import (
	"github.com/rarimo/horizon-svc/pkg/msgs"
	figure "gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/lorem"
)

type VotesIndexerCfg struct {
	RunnerName    string              `fig:"runner_name"`
	VotesConsumer msgs.ConsumerConfig `fig:"-"`
}

func (c *config) VotesIndexer() VotesIndexerCfg {
	return c.votesIndexer.Do(func() interface{} {
		var cfg VotesIndexerCfg
		yamlName := "votes_indexer"

		err := figure.
			Out(&cfg).
			From(kv.MustGetStringMap(c.getter, yamlName)).
			Please()

		if err != nil {
			panic(errors.Wrap(err, "failed to figure out "+yamlName))
		}

		if isEmptyOrRandom(cfg.RunnerName) {
			cfg.RunnerName = yamlName + "_" + lorem.RandomName()
		}

		cfg.VotesConsumer = msgs.NewConsumerer(newPathGetter(c.getter), yamlName+".votes_consumer", msgs.ConsumererOpts{
			RedisClient: c.RedisClient(),
		}).Consumer()

		return cfg
	}).(VotesIndexerCfg)
}
