package config

import (
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"time"
)

type RateLimiterConfig struct {
	Disabled    bool          `fig:"disabled"`
	Period      time.Duration `fig:"period"`
	Limit       int64         `fig:"limit"`
	RedisPrefix string        `fig:"redis_prefix"`
}

func (c *config) RateLimiter() *RateLimiterConfig {
	return c.rateLimiter.Do(func() interface{} {
		var result RateLimiterConfig

		err := figure.
			Out(&result).
			From(kv.MustGetStringMap(c.getter, "rate_limits")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out rate limiter config"))
		}

		return &result
	}).(*RateLimiterConfig)
}
