package cachedpg

import (
	"context"
	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/go-redis/redis/v8"
	"github.com/rarimo/horizon-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type WithdrawalQ struct {
	log   *logan.Entry
	raw   data.WithdrawalQ
	cache *marshaler.Marshaler
}

func (q *WithdrawalQ) InsertBatchCtx(ctx context.Context, withdrawals ...data.Withdrawal) error {
	return q.raw.InsertBatchCtx(ctx, withdrawals...)
}

func (q *WithdrawalQ) WithdrawalByHashCtx(ctx context.Context, hash []byte, _ bool) (*data.Withdrawal, error) {
	var withdrawal *data.Withdrawal

	err := tryGetFromCache(ctx, q.cache, makeWithdrawalCacheHashKey(string(hash)), withdrawal)
	if err == nil {
		q.log.Debug("hit")
		return withdrawal, nil
	}

	q.log.Debug("miss")
	if errors.Cause(err) != redis.Nil {
		q.log.WithError(err).Error("failed to get withdrawal from cache")
	}

	withdrawal, err = q.raw.WithdrawalByHashCtx(ctx, hash, false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select withdrawal")
	}

	if err := q.cacheEveryWithdrawal(ctx, []data.Withdrawal{*withdrawal}); err != nil {
		q.log.WithError(err).Error("failed to set withdrawal to cache")
	}

	return withdrawal, nil
}

func (q *WithdrawalQ) cacheEveryWithdrawal(ctx context.Context, withdrawals []data.Withdrawal) error {
	for _, withdrawal := range withdrawals {
		opts := store.WithTags(withdrawalTags([]data.Withdrawal{withdrawal}))

		err := q.cache.Set(ctx, makeWithdrawalCacheHashKey(string(withdrawal.Hash)), withdrawal, opts)
		if err != nil {
			return errors.Wrap(err, "failed to set withdrawal to cache by hash", logan.F{
				"withdrawal_hash": withdrawal.Hash,
			})
		}
	}

	return nil
}

func withdrawalTags(withdrawals []data.Withdrawal) []string {
	tags := make([]string, len(withdrawals))

	for i, withdrawal := range withdrawals {
		tags[i] = makeWithdrawalCacheHashKey(string(withdrawal.Hash))
	}

	return tags
}

func makeWithdrawalCacheHashKey(withdrawalHash string) string {
	return "withdrawal:" + withdrawalHash
}
