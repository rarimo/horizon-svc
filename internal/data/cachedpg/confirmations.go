package cachedpg

import (
	"context"
	"strconv"

	"github.com/eko/gocache/lib/v4/marshaler"

	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/redis/go-redis/v9"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type ConfirmationsQ struct {
	log   *logan.Entry
	raw   data.ConfirmationQ
	cache *marshaler.Marshaler
}

func (q *ConfirmationsQ) InsertBatchCtx(ctx context.Context, confirmations ...data.Confirmation) error {
	return q.raw.InsertBatchCtx(ctx, confirmations...)
}

func (q *ConfirmationsQ) ConfirmationsByTransferIndexCtx(ctx context.Context, transferIndex []byte, isForUpdate bool) ([]data.Confirmation, error) {
	if !isForUpdate {
		var confirmations []data.Confirmation
		err := tryGetFromCache(ctx, q.cache, makeConfByTransferCacheKey(string(transferIndex)), &confirmations)
		if err == nil {
			q.log.Debug("hit")
			return confirmations, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get confirmations form cache")
		}
	}

	confirmations, err := q.raw.ConfirmationsByTransferIndexCtx(ctx, transferIndex, isForUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get confirmations from cache", logan.F{
			"for_update":     isForUpdate,
			"transfer_index": string(transferIndex),
		})
	}

	if len(confirmations) != 0 {
		err = q.cache.Set(ctx, makeConfByTransferCacheKey(string(transferIndex)), confirmations) // confirmations are immutable, so we don't need to invalidate them in any way
		if err != nil {
			q.log.WithError(err).Error("failed to cache confirmations batch")
			return confirmations, nil
		}

		err = q.cacheEveryConfirmation(ctx, confirmations)
		if err != nil {
			q.log.WithError(err).Error("failed to cache confirmations individually")
		}
	}

	return confirmations, nil
}

func (q *ConfirmationsQ) cacheEveryConfirmation(ctx context.Context, confirmations []data.Confirmation) error {
	for _, conf := range confirmations {
		err := q.cache.Set(ctx, makeConfCacheKey(conf.ID), conf)
		if err != nil {
			return errors.Wrap(err, "failed to cache confirmation", logan.F{
				"id": conf.ID,
			})
		}
	}

	return nil
}

func makeConfCacheKey(id int64) string {
	return "confirmation:" + strconv.FormatInt(id, 10)
}

func makeConfByTransferCacheKey(transferIndex string) string {
	return "confirmations_by_transfer:" + transferIndex
}
