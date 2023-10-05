package config

import (
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/lorem"
)

type RarimoCoreProducerConfig struct {
	RunnerName string `fig:"runner_name"`

	TransfersQueueName     string `fig:"transfers_queue_name,required"`
	ConfirmationsQueueName string `fig:"confirmations_queue_name,required"`
	ApprovalsQueueName     string `fig:"approvals_queue_name,required"`
	RejectionsQueueName    string `fig:"rejections_queue_name,required"`
	VotesQueueName         string `fig:"votes_queue_name,required"`

	BlockRangeConsumer msgs.ConsumerConfig `fig:"-"`
}

func (c *config) RarimoCoreProducer() *RarimoCoreProducerConfig {
	return c.rarimocoreProducer.Do(func() interface{} {
		var result RarimoCoreProducerConfig
		serviceName := "rarimocore_producer"

		rarimocoreProducerData := kv.MustGetStringMap(c.getter, serviceName)
		err := figure.
			Out(&result).
			From(rarimocoreProducerData).
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
	}).(*RarimoCoreProducerConfig)
}
