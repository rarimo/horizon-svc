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

type ApprovalQ struct {
	log   *logan.Entry
	raw   data.ApprovalQ
	cache *marshaler.Marshaler
}

func (q *ApprovalQ) InsertBatchCtx(ctx context.Context, approvals ...data.Approval) error {
	return q.raw.InsertBatchCtx(ctx, approvals...)
}

func (q *ApprovalQ) ApprovalsByTransferIndexCtx(ctx context.Context, transferIndex []byte, isForUpdate bool) ([]data.Approval, error) {
	if !isForUpdate {
		var approvals []data.Approval
		err := tryGetFromCache(ctx, q.cache, makeApprovalsByTransferCacheKey(string(transferIndex)), &approvals)
		if err == nil {
			q.log.Debug("hit")
			return approvals, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get approvals form cache")
		}
	}

	approvals, err := q.raw.ApprovalsByTransferIndexCtx(ctx, transferIndex, isForUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get approvals by transfer idx", logan.F{
			"transfer":   string(transferIndex),
			"for_update": isForUpdate,
		})
	}

	if len(approvals) != 0 {
		err = q.cache.Set(ctx, makeApprovalsByTransferCacheKey(string(transferIndex)), approvals) // approvals are immutable, so we don't need to invalidate them in any way
		if err != nil {
			q.log.WithError(err).Error("failed to cache approvals batch")
			return approvals, nil
		}

		err = q.cacheEveryApproval(ctx, approvals)
		if err != nil {
			q.log.WithError(err).Error("failed to cache approvals individually")
		}
	}

	return approvals, nil
}

func (q *ApprovalQ) cacheEveryApproval(ctx context.Context, approvals []data.Approval) error {
	for _, approval := range approvals {
		err := q.cache.Set(ctx, makeApprovalCacheKey(approval.ID), approval)
		if err != nil {
			return errors.Wrap(err, "failed to cache approval", logan.F{
				"id": approval.ID,
			})
		}
	}

	return nil
}

func makeApprovalCacheKey(id int64) string {
	return "approval:" + strconv.FormatInt(id, 10)
}

func makeApprovalsByTransferCacheKey(transferIndex string) string {
	return "approvals_by_transfer:" + transferIndex
}
