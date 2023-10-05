package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rarimo/horizon-svc/internal/config"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/redis/go-redis/v9"
)

type KeyValueProvider struct {
	cli *redis.Client
	log *logan.Entry
}

func NewKeyValueProvider(cfg config.Config) *KeyValueProvider {
	return &KeyValueProvider{
		cli: cfg.RedisClient(),
		log: cfg.Log().WithField("who", "key-value-provider"),
	}
}

func (p *KeyValueProvider) Upsert(ctx context.Context, kv data.KeyValue) error {
	raw, err := json.Marshal(kv)
	if err != nil {
		return errors.Wrap(err, "failed to marshal", logan.F{
			"key": kv.Key,
		})
	}

	if err := p.cli.Set(ctx, kv.Key, raw, 0).Err(); err != nil {
		return errors.Wrap(err, "failed to set", logan.F{
			"key": kv.Key,
		})
	}

	return nil
}

func (p *KeyValueProvider) Get(ctx context.Context, key string) (*data.KeyValue, error) {
	raw, err := p.cli.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to get", logan.F{
			"key": key,
		})
	}

	var kv data.KeyValue
	if err := json.Unmarshal([]byte(raw), &kv); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal", logan.F{
			"key": key,
			"raw": raw,
		})
	}

	return &kv, nil
}

func (p *KeyValueProvider) Remove(ctx context.Context, key string) error {
	return p.cli.Del(ctx, key).Err()
}

func (p *KeyValueProvider) InsertNX(ctx context.Context, key string, value string, expiry time.Duration) (bool, error) {
	ok, err := p.cli.SetNX(ctx, key, value, expiry).Result()
	if err != nil {
		return false, errors.Wrap(err, "failed to setnx", logan.F{
			"key": key,
		})
	}

	return ok, nil
}
