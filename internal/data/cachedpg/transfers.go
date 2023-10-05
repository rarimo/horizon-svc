package cachedpg

import (
	"context"
	"strconv"

	"github.com/eko/gocache/lib/v4/marshaler"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/redis/go-redis/v9"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type TransfersQ struct {
	log   *logan.Entry
	raw   data.TransferQ
	cache *marshaler.Marshaler
}

func (q *TransfersQ) SelectCtx(ctx context.Context, selector data.TransferSelector) ([]data.Transfer, error) {
	{
		var transfers []data.Transfer

		err := tryGetFromCache(ctx, q.cache, selector.MustCacheKey(), &transfers)
		if err == nil {
			q.log.Debug("hit")
			return transfers, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get transfers from cache")
		}
	}

	transfers, err := q.raw.SelectCtx(ctx, selector)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select transfers")
	}

	if len(transfers) != 0 {
		err = q.cache.Set(ctx, selector.MustCacheKey(), transfers, store.WithTags(transferTags(transfers)))
		if err != nil {
			q.log.WithError(err).Error("failed to set transfers to cache")
			return transfers, nil
		}

		err = q.cacheEveryTransfer(ctx, transfers)
		if err != nil {
			q.log.WithError(err).Error("failed to set transfers to cache individually")
		}
	}

	return transfers, nil
}

func (q *TransfersQ) UpsertBatchCtx(ctx context.Context, transfers ...data.Transfer) error {
	err := q.raw.UpsertBatchCtx(ctx, transfers...)
	if err != nil {
		return err
	}

	opts := store.WithInvalidateTags(transferIndexTags(transfers))
	if err := q.cache.Invalidate(ctx, opts); err != nil {
		q.log.WithError(err).Error("failed to invalidate transfers cache by indices")
	}

	return nil
}

func (q *TransfersQ) TransferByIndexCtx(ctx context.Context, index []byte, isForUpdate bool) (*data.Transfer, error) {
	if !isForUpdate {
		var transfer data.Transfer

		err := tryGetFromCache(ctx, q.cache, makeTransferCacheIdxKey(string(index)), &transfer)
		if err == nil {
			q.log.Debug("hit")
			return &transfer, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get transfer from cache")
		}
	}

	transfer, err := q.raw.TransferByIndexCtx(ctx, index, isForUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select transfer")
	}

	if err := q.cacheEveryTransfer(ctx, []data.Transfer{*transfer}); err != nil {
		q.log.WithError(err).Error("failed to set transfer to cache")
	}

	return transfer, nil
}

func (q *TransfersQ) SetStatusByIndexCtx(ctx context.Context, status int, indexes ...string) error {
	err := q.raw.SetStatusByIndexCtx(ctx, status, indexes...)
	if err != nil {
		return errors.Wrap(err, "failed to set status")
	}

	opts := store.WithInvalidateTags(transferIndiciesTags(indexes))
	if err := q.cache.Invalidate(ctx, opts); err != nil {
		q.log.WithError(err).Error("failed to invalidate transfers cache")
	}

	return nil
}

func (q *TransfersQ) cacheEveryTransfer(ctx context.Context, transfers []data.Transfer) error {
	for _, transfer := range transfers {
		opts := store.WithTags(transferTags([]data.Transfer{transfer}))

		err := q.cache.Set(ctx, makeTransferCacheKey(transfer.ID), transfer, opts)
		if err != nil {
			return errors.Wrap(err, "failed to set transfer to cache by id", logan.F{
				"transfer_id":  transfer.ID,
				"transfer_idx": string(transfer.Index),
			})
		}

		err = q.cache.Set(ctx, makeTransferCacheIdxKey(string(transfer.Index)), transfer, opts)
		if err != nil {
			return errors.Wrap(err, "failed to set transfer to cache by index", logan.F{
				"transfer_id":  transfer.ID,
				"transfer_idx": string(transfer.Index),
			})
		}
	}

	return nil
}

func transferIndiciesTags(indexes []string) []string {
	tags := make([]string, len(indexes))

	for i, index := range indexes {
		tags[i] = "transfer:" + index
	}

	return tags
}

func transferTags(transfers []data.Transfer) []string {
	tags := make([]string, 2*len(transfers))

	for i, transfer := range transfers {
		tags[i] = "transfer:" + strconv.FormatInt(transfer.ID, 10)
		tags[i+1] = "transfer:" + string(transfer.Index)
	}

	return tags
}

func transferIndexTags(transfers []data.Transfer) []string {
	tags := make([]string, len(transfers))

	for i, transfer := range transfers {
		tags[i] = "transfer:" + string(transfer.Index)
	}

	return tags
}

func makeTransferCacheKey(transferID int64) string {
	return "transfer:" + strconv.FormatInt(transferID, 10)
}

func makeTransferCacheIdxKey(transferIdx string) string {
	return "transfer:" + transferIdx
}
