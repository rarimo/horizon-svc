package config

import (
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/lorem"
)

type ApprovalsIndexerConfig struct {
	RunnerName        string              `fig:"runner_name"`
	ApprovalsConsumer msgs.ConsumerConfig `fig:"-"`
}

func (c *config) ApprovalsIndexer() ApprovalsIndexerConfig {
	return c.approvalsIndexer.Do(func() interface{} {
		var cfg ApprovalsIndexerConfig
		yamlName := "approvals_indexer"

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

		cfg.ApprovalsConsumer = msgs.NewConsumerer(newPathGetter(c.getter), yamlName+".approvals_consumer", msgs.ConsumererOpts{
			RedisClient: c.RedisClient(),
		}).Consumer()

		return cfg
	}).(ApprovalsIndexerConfig)
}

func isEmptyOrRandom(s string) bool {
	return s == "" || s == "RANDOM"
}
