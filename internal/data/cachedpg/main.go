package cachedpg

import (
	"context"

	"github.com/eko/gocache/lib/v4/marshaler"

	"github.com/eko/gocache/lib/v4/cache"
	redisstore "github.com/eko/gocache/store/redis/v4"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/redis/go-redis/v9"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Storage struct {
	log   *logan.Entry
	raw   data.Storage
	cache *marshaler.Marshaler
}

func NewStorage(log *logan.Entry, raw data.Storage, redisClient *redis.Client) *Storage {
	return &Storage{
		log:   log,
		raw:   raw,
		cache: marshaler.New(cache.New[any](redisstore.NewRedis(redisClient))),
	}
}

// Transaction is a wrapper around data.Storage.Transaction and its internals are not cached.
// You should not get data from cache within the transaction (because of possible race conditions),
// use plain pg.Storage's methods instead to get data from DB
func (s *Storage) Transaction(f func() error) error {
	return s.raw.Transaction(f)
}

func (s *Storage) TransferQ() data.TransferQ {
	return &TransfersQ{
		log:   s.log.WithField("who", "transfers-cached-q"),
		raw:   s.raw.TransferQ(),
		cache: s.cache,
	}
}

func (s *Storage) ConfirmationQ() data.ConfirmationQ {
	return &ConfirmationsQ{
		log:   s.log.WithField("who", "confirmations-cached-q"),
		raw:   s.raw.ConfirmationQ(),
		cache: s.cache,
	}
}

func (s *Storage) TransactionQ() data.TransactionQ {
	return s.raw.TransactionQ() // as it does not have cacheable methods TODO implement cached version when methods arrive
}

func (s *Storage) VoteQ() data.VoteQ {
	return &VoteQ{
		log:   s.log.WithField("who", "votes-cached-q"),
		raw:   s.raw.VoteQ(),
		cache: s.cache,
	}
}

func (s *Storage) ApprovalQ() data.ApprovalQ {
	return &ApprovalQ{
		log:   s.log.WithField("who", "approvals-cached-q"),
		raw:   s.raw.ApprovalQ(),
		cache: s.cache,
	}
}

func (s *Storage) RejectionQ() data.RejectionQ {
	return &RejectionQ{
		log:   s.log.WithField("who", "rejections-cached-q"),
		raw:   s.raw.RejectionQ(),
		cache: s.cache,
	}
}

func (s *Storage) CollectionQ() data.CollectionQ {
	return &CollectionQ{
		log:   s.log.WithField("who", "collections-cached-q"),
		raw:   s.raw.CollectionQ(),
		cache: s.cache,
	}
}

func (s *Storage) CollectionChainMappingQ() data.CollectionChainMappingQ {
	return &CollectionChainMappingQ{
		log:   s.log.WithField("who", "collection-chain-mappings-cached-q"),
		raw:   s.raw.CollectionChainMappingQ(),
		cache: s.cache,
	}
}

func (s *Storage) ItemQ() data.ItemQ {
	return &ItemQ{
		log:   s.log.WithField("who", "items-cached-q"),
		raw:   s.raw.ItemQ(),
		cache: s.cache,
	}
}

func (s *Storage) ItemChainMappingQ() data.ItemChainMappingQ {
	return &ItemChainMappingQ{
		log:   s.log.WithField("who", "item-chain-mappings-cached-q"),
		raw:   s.raw.ItemChainMappingQ(),
		cache: s.cache,
	}
}

func tryGetFromCache(ctx context.Context, c *marshaler.Marshaler, key string, v interface{}) error {
	if _, err := c.Get(ctx, key, v); err != nil {
		return errors.Wrap(err, "failed to get from cache")
	}

	return nil
}
