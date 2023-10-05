package cachedpg

import (
	"context"
	"fmt"

	"github.com/eko/gocache/lib/v4/marshaler"

	"github.com/redis/go-redis/v9"
	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/rarimo/horizon-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
)

const (
	itemKeyBase = "item"
)

type ItemQ struct {
	log   *logan.Entry
	raw   data.ItemQ
	cache *marshaler.Marshaler
}

func (q *ItemQ) InsertCtx(ctx context.Context, i *data.Item) error {
	return q.raw.InsertCtx(ctx, i)
}

func (q *ItemQ) ItemByIDCtx(ctx context.Context, id int64, isForUpdate bool) (*data.Item, error) {
	if !isForUpdate {
		var item data.Item

		err := tryGetFromCache(ctx, q.cache, itemIDCacheKey(id), &item)
		if err == nil {
			q.log.Debug("hit")
			return &item, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get item form cache")
		}
	}

	item, err := q.raw.ItemByIDCtx(ctx, id, isForUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get item by id", logan.F{
			"id": id,
		})
	}

	if item != nil {
		if err := q.cacheItem(ctx, item); err != nil {
			q.log.WithError(err).Error("failed to cache item")
		}
	}

	return item, nil

}

func (q *ItemQ) ItemByIndexCtx(ctx context.Context, index []byte, isForUpdate bool) (*data.Item, error) {
	if !isForUpdate {
		var item data.Item

		err := tryGetFromCache(ctx, q.cache, itemIndexCacheKey(index), &item)
		if err == nil {
			q.log.Debug("hit")
			return &item, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get item form cache")
		}
	}

	item, err := q.raw.ItemByIndexCtx(ctx, index, isForUpdate)
	if err != nil {
		return nil, err
	}

	if item != nil {
		if err := q.cacheItem(ctx, item); err != nil {
			q.log.WithError(err).Error("failed to cache item")
		}
	}

	return item, nil
}

func (q *ItemQ) UpdateCtx(ctx context.Context, i *data.Item) error {
	err := q.cache.Invalidate(ctx,
		store.WithInvalidateTags([]string{itemIDCacheKey(i.ID)})) // since we are updating by id
	if err != nil {
		q.log.WithError(err).Error("failed to invalidate item cache")
	}

	return q.raw.UpdateCtx(ctx, i)
}

func (q *ItemQ) DeleteCtx(ctx context.Context, i *data.Item) error {
	err := q.cache.Invalidate(ctx,
		store.WithInvalidateTags([]string{itemIDCacheKey(i.ID)})) // since we are deleting by id
	if err != nil {
		q.log.WithError(err).Error("failed to invalidate item cache")
	}

	return q.raw.DeleteCtx(ctx, i)
}

func (q *ItemQ) cacheItem(ctx context.Context, i *data.Item) error {
	err := q.cache.Set(ctx, itemIndexCacheKey(i.Index), i, store.WithTags(itemCacheTags(i)))
	if err != nil {
		return errors.Wrap(err, "failed to cache item by index")
	}

	err = q.cache.Set(ctx, itemIDCacheKey(i.ID), i, store.WithTags(itemCacheTags(i)))
	if err != nil {
		return errors.Wrap(err, "failed to cache item by id")
	}

	return nil
}

func itemIndexCacheKey(index []byte) string {
	return fmt.Sprintf("%s:%s", itemKeyBase, string(index))
}

func itemIDCacheKey(id int64) string {
	return fmt.Sprintf("%s:%d", itemKeyBase, id)
}

func itemCacheTags(i *data.Item) []string {
	return []string{
		itemIndexCacheKey(i.Index),
		itemIDCacheKey(i.ID),
	}
}
