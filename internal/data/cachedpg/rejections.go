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

type RejectionQ struct {
	log   *logan.Entry
	raw   data.RejectionQ
	cache *marshaler.Marshaler
}

func (q *RejectionQ) InsertBatchCtx(ctx context.Context, rejections ...data.Rejection) error {
	return q.raw.InsertBatchCtx(ctx, rejections...)
}

func (q *RejectionQ) RejectionsByTransferIndexCtx(ctx context.Context, transferIndex []byte, isForUpdate bool) ([]data.Rejection, error) {
	if !isForUpdate {
		var rejections []data.Rejection
		err := tryGetFromCache(ctx, q.cache, makeRejectionsByTransferCacheKey(string(transferIndex)), &rejections)
		if err == nil {
			q.log.Debug("hit")
			return rejections, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get rejections form cache")
		}
	}

	rejections, err := q.raw.RejectionsByTransferIndexCtx(ctx, transferIndex, isForUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rejections by transfer idx", logan.F{
			"transfer":   string(transferIndex),
			"for_update": isForUpdate,
		})
	}

	if len(rejections) != 0 {
		err = q.cache.Set(ctx, makeRejectionsByTransferCacheKey(string(transferIndex)), rejections) // rejections are immutable, so we don't need to invalidate them in any way
		if err != nil {
			q.log.WithError(err).Error("failed to cache rejections batch")
			return rejections, nil
		}

		err = q.cacheEveryRejection(ctx, rejections)
		if err != nil {
			q.log.WithError(err).Error("failed to cache rejections individually")
		}
	}

	return rejections, nil
}

func (q *RejectionQ) cacheEveryRejection(ctx context.Context, rejections []data.Rejection) error {
	for _, rejection := range rejections {
		err := q.cache.Set(ctx, makeRejectionCacheKey(rejection.ID), rejection)
		if err != nil {
			return errors.Wrap(err, "failed to cache rejection", logan.F{
				"id": rejection.ID,
			})
		}
	}

	return nil
}

func makeRejectionCacheKey(id int64) string {
	return "rejection:" + strconv.FormatInt(id, 10)
}

func makeRejectionsByTransferCacheKey(transferIndex string) string {
	return "rejections_by_transfer:" + transferIndex
}
