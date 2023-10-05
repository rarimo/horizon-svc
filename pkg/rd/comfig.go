package rd

import (
	"github.com/adjust/rmq/v5"

	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/redis/go-redis/v9"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type Rediser interface {
	RedisClient() *redis.Client
	RMQ() rmq.Connection
}

type rediser struct {
	getter kv.Getter
	once   comfig.Once
}

func NewRediser(getter kv.Getter) Rediser {
	return &rediser{
		getter: getter,
	}
}

type cfg struct {
	Addr     string `fig:"addr,required"`
	DB       int    `fig:"db"`
	Username string `fig:"username"`
	Password string `fig:"password"`
}

func (d *rediser) readConfig() cfg {
	var config cfg
	err := figure.Out(&config).
		From(kv.MustGetStringMap(d.getter, "redis")).
		Please()
	if err != nil {
		panic(errors.Wrap(err, "failed to figure out"))
	}

	return config
}

func (d *rediser) RedisClient() *redis.Client {
	return d.once.Do(func() interface{} {
		cfg := d.readConfig()

		client := redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			DB:       cfg.DB, // FIXME divide DBs for cache and for queues
			Username: cfg.Username,
			Password: cfg.Password,
		})

		return client
	}).(*redis.Client)
}

func (d *rediser) RMQ() rmq.Connection {
	return d.once.Do(func() interface{} {
		config := d.readConfig()

		connection, err := rmq.OpenConnection("horizon", "tcp", config.Addr, 1, nil)
		if err != nil {
			panic(errors.Wrap(err, "failed to open connection for queues"))
		}

		return connection
	}).(rmq.Connection)
}
