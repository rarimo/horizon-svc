package config

import (
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/lorem"
)

type TokenManagerProducerConfig struct {
	RunnerName           string              `fig:"runner_name"`
	ItemsQueueName       string              `fig:"items_queue_name,required"`
	CollectionsQueueName string              `fig:"collections_queue_name,required"`
	BlockRangeConsumer   msgs.ConsumerConfig `fig:"-"`
}

func (c *config) TokenManagerProducer() *TokenManagerProducerConfig {
	return c.tokenmanagerProducer.Do(func() interface{} {
		var result TokenManagerProducerConfig
		serviceName := "tokenmanager_producer"

		tokenmanagerProducerData := kv.MustGetStringMap(c.getter, serviceName)
		err := figure.
			Out(&result).
			From(tokenmanagerProducerData).
			Please()
		if err != nil {
			panic(err)
		}

		if isEmptyOrRandom(result.RunnerName) {
			result.RunnerName = serviceName + "_" + lorem.RandomName()
		}

		result.BlockRangeConsumer = msgs.NewConsumerer(newPathGetter(c.getter), serviceName+".block_range_consumer", msgs.ConsumererOpts{
			RedisClient: c.RedisClient(),
		}).Consumer()

		return &result

	}).(*TokenManagerProducerConfig)
}
