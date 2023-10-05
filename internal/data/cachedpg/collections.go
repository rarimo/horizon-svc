package cachedpg

import (
	"context"
	"fmt"

	"github.com/eko/gocache/lib/v4/marshaler"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/redis/go-redis/v9"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const (
	collectionKeyBase = "collection"
)

type CollectionQ struct {
	log   *logan.Entry
	raw   data.CollectionQ
	cache *marshaler.Marshaler
}

func (q *CollectionQ) InsertCtx(ctx context.Context, c *data.Collection) error {
	return q.raw.InsertCtx(ctx, c)
}

func (q *CollectionQ) CollectionByIndexCtx(ctx context.Context, index []byte, isForUpdate bool) (*data.Collection, error) {
	if !isForUpdate {
		var collection data.Collection

		err := tryGetFromCache(ctx, q.cache, collectionIndexCacheKey(index), &collection)
		if err == nil {
			q.log.Debug("hit")
			return &collection, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get collection form cache")
		}
	}

	collection, err := q.raw.CollectionByIndexCtx(ctx, index, isForUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to colleciton by index", logan.F{
			"index": index,
		})
	}

	if collection != nil {
		if err := q.cacheCollection(ctx, collection); err != nil {
			q.log.WithError(err).Error("failed to cache collection")
		}
	}

	return collection, nil
}

func (q *CollectionQ) DeleteCtx(ctx context.Context, c *data.Collection) error {
	err := q.cache.Invalidate(ctx,
		store.WithInvalidateTags([]string{collectionIDCacheKey(c.ID)})) // since we are deleting by id
	if err != nil {
		q.log.WithError(err).Error("failed to invalidate collection cache")
	}

	return q.raw.DeleteCtx(ctx, c)
}

func (q *CollectionQ) cacheCollection(ctx context.Context, c *data.Collection) error {
	err := q.cache.Set(ctx, collectionIndexCacheKey(c.Index), c, store.WithTags(collectionCacheTags(c)))
	if err != nil {
		return errors.Wrap(err, "failed to cache item by index")
	}

	err = q.cache.Set(ctx, collectionIDCacheKey(c.ID), c, store.WithTags(collectionCacheTags(c)))
	if err != nil {
		return errors.Wrap(err, "failed to cache item by id")
	}

	return nil
}

func collectionIndexCacheKey(index []byte) string {
	return fmt.Sprintf("%s:%s", collectionKeyBase, string(index))
}

func collectionIDCacheKey(id int64) string {
	return fmt.Sprintf("%s:%d", collectionKeyBase, id)
}

func collectionCacheTags(collection *data.Collection) []string {
	return []string{
		collectionIndexCacheKey(collection.Index),
		collectionIDCacheKey(collection.ID),
	}
}
