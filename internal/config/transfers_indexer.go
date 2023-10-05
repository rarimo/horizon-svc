package config

import (
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/lorem"
)

type TransfersIndexerConfig struct {
	RunnerName           string              `fig:"runner_name"`
	FetchOperationsLimit uint64              `fig:"fetch_operations_limit"`
	TransfersConsumer    msgs.ConsumerConfig `fig:"-"`
}

func (c *config) TransfersIndexer() TransfersIndexerConfig {
	return c.transfersIndexer.Do(func() interface{} {
		result := TransfersIndexerConfig{
			RunnerName:           "horizon-transfers-indexer",
			FetchOperationsLimit: 100,
		}

		yamlName := "transfers_indexer"

		err := figure.
			Out(&result).
			From(kv.MustGetStringMap(c.getter, yamlName)).
			Please()

		if err != nil {
			panic(errors.Wrap(err, "failed to figure out "+yamlName))
		}

		if isEmptyOrRandom(result.RunnerName) {
			result.RunnerName = yamlName + "_" + lorem.RandomName()
		}

		result.TransfersConsumer = msgs.NewConsumerer(newPathGetter(c.getter), yamlName+".transfers_consumer", msgs.ConsumererOpts{
			RedisClient: c.RedisClient(),
		}).Consumer()

		return result
	}).(TransfersIndexerConfig)
}
