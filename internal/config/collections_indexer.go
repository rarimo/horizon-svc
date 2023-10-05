package config

import (
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/lorem"
)

type CollectionsIndexerConfig struct {
	RunnerName               string              `fig:"runner_name"`
	CollectionEventsConsumer msgs.ConsumerConfig `fig:"-"`
}

func (c *config) CollectionsIndexer() CollectionsIndexerConfig {
	return c.collectionsIndexer.Do(func() interface{} {
		var result CollectionsIndexerConfig
		serviceName := "collections_indexer"

		err := figure.
			Out(&result).
			From(kv.MustGetStringMap(c.getter, serviceName)).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out "+serviceName))
		}

		if isEmptyOrRandom(result.RunnerName) {
			result.RunnerName = serviceName + "_" + lorem.RandomName()
		}

		result.CollectionEventsConsumer = msgs.NewConsumerer(newPathGetter(c.getter), serviceName+".collection_events_consumer", msgs.ConsumererOpts{
			RedisClient: c.RedisClient(),
		}).Consumer()

		return result
	}).(CollectionsIndexerConfig)
}
