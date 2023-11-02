package config

import (
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/lorem"
)

type WithdrawalsIndexerConfig struct {
	RunnerName          string              `fig:"runner_name"`
	WithdrawalsConsumer msgs.ConsumerConfig `fig:"-"`
}

func (c *config) WithdrawalsIndexer() *WithdrawalsIndexerConfig {
	return c.withdrawalsIndexer.Do(func() interface{} {
		var cfg WithdrawalsIndexerConfig
		yamlName := "withdrawals_indexer"

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

		cfg.WithdrawalsConsumer = msgs.NewConsumerer(newPathGetter(c.getter), yamlName+".withdrawals_consumer", msgs.ConsumererOpts{
			RedisClient: c.RedisClient(),
		}).Consumer()

		return &cfg
	}).(*WithdrawalsIndexerConfig)
}
