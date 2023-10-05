package config

import (
	"time"

	"gitlab.com/distributed_lab/lorem"

	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

type SpecialCaseBlockRange struct {
	From int64 `fig:"from,required"`
	To   int64 `fig:"to,required"`
}

type BlockRangesProducerConfig struct {
	RunnerName        string                  `fig:"runner_name"`
	QueueName         string                  `fig:"queue_name,required"`
	CursorKey         string                  `fig:"cursor_key,required"`
	BlockRangeLimit   int64                   `fig:"block_range_limit,required"`
	BlockTime         time.Duration           `fig:"block_time,required"`
	SpecialCaseBlocks []SpecialCaseBlockRange `fig:"special_case_blocks"`
}

func (c *config) BlockRangeProducer() *BlockRangesProducerConfig {
	return c.blockRangeProducer.Do(func() interface{} {
		var result BlockRangesProducerConfig
		serviceName := "block_ranges_producer"

		blockRangesProducerData := kv.MustGetStringMap(c.getter, serviceName)
		err := figure.
			Out(&result).
			From(blockRangesProducerData).
			Please()
		if err != nil {
			panic(err)
		}

		if isEmptyOrRandom(result.RunnerName) {
			result.RunnerName = serviceName + "_" + lorem.RandomName()
		}

		return &result
	}).(*BlockRangesProducerConfig)
}
