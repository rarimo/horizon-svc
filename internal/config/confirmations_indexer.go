package config

import (
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/lorem"
)

type ConfirmationsIndexerCfg struct {
	RunnerName            string              `fig:"runner_name"`
	ConfirmationsConsumer msgs.ConsumerConfig `fig:"confirmations_consumer"`
}

func (c *config) ConfirmationsIndexer() ConfirmationsIndexerCfg {
	return c.confirmationsIndexer.Do(func() interface{} {
		var result ConfirmationsIndexerCfg
		yamlName := "confirmations_indexer"

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

		result.ConfirmationsConsumer = msgs.NewConsumerer(newPathGetter(c.getter), yamlName+".confirmations_consumer", msgs.ConsumererOpts{
			RedisClient: c.RedisClient(),
		}).Consumer()

		return result
	}).(ConfirmationsIndexerCfg)
}
