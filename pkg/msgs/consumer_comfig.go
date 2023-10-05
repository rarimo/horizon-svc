package msgs

import (
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/lorem"
)

type ConsumerConfig struct {
	Name                 string        `fig:"name"`
	Queue                string        `fig:"queue,required"`
	PrefetchLimit        int64         `fig:"prefetch_limit"`
	PollDuration         time.Duration `fig:"poll_duration"`
	MinRetryPeriod       time.Duration `fig:"min_retry_period"`
	MaxRetryPeriod       time.Duration `fig:"max_retry_period"`
	RetryConsumeAttempts uint64        `fig:"retry_consume_attempts"`
	RedisClient          *redis.Client `fig:"-"`
}

type Consumerer interface {
	Consumer() ConsumerConfig
}

type consumerer struct {
	getter kv.Getter
	once   comfig.Once

	opts ConsumererOpts
	name string
}

type ConsumererOpts struct {
	RedisClient *redis.Client
}

func NewConsumerer(getter kv.Getter, name string, opts ConsumererOpts) Consumerer {
	if opts.RedisClient == nil {
		panic("redis client is required for consumerer")
	}

	return &consumerer{
		getter: getter,
		name:   name,
		opts:   opts,
	}
}

func (c *consumerer) Consumer() ConsumerConfig {
	return c.once.Do(func() interface{} {
		cfg := ConsumerConfig{
			PrefetchLimit:        100,
			PollDuration:         1 * time.Second,
			MinRetryPeriod:       1 * time.Second,
			MaxRetryPeriod:       1 * time.Minute,
			RetryConsumeAttempts: 5,
			RedisClient:          c.opts.RedisClient,
		}

		err := figure.
			Out(&cfg).
			With(figure.BaseHooks).
			From(kv.MustGetStringMap(c.getter, c.name)).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out consumer"))
		}

		if isEmptyOrRandom(cfg.Name) {
			cfg.Name = lorem.RandomName()
		}

		return cfg
	}).(ConsumerConfig)
}

func isEmptyOrRandom(s string) bool {
	return s == "" || s == "RANDOM"
}
