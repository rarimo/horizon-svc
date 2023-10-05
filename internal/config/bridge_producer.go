package config

import (
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/lorem"
)

func (c *config) BridgeProducer() *BridgeProducerConfig {
	return c.bridgeProducer.Do(func() interface{} {
		var result BridgeProducerConfig
		serviceName := "bridge_producer"

		err := figure.
			Out(&result).
			From(kv.MustGetStringMap(c.getter, serviceName)).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out BridgeProducer config"))
		}

		if isEmptyOrRandom(result.RunnerName) {
			result.RunnerName = serviceName + "_" + lorem.RandomName()
		}

		return &result
	}).(*BridgeProducerConfig)
}

type BridgeProducerChainConfig struct {
	ID          int  `fig:"id,required"`
	SkipCatchup bool `fig:"skip_catchup"`
	BatchSize   int  `fig:"batch_size"`
}

type BridgeProducerConfig struct {
	RunnerName           string                       `fig:"runner_name"`
	WithdrawalsQueueName string                       `fig:"withdrawals_queue_name,required"`
	CursorKey            string                       `fig:"cursor_key,required"`
	Chains               []*BridgeProducerChainConfig `fig:"chains"`
}

func (b *BridgeProducerConfig) ChainConfigByID(ID int) *BridgeProducerChainConfig {
	for _, chain := range b.Chains {
		if chain.ID == ID {
			return chain
		}
	}

	return nil
}
