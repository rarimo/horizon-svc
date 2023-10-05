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

type VoteQ struct {
	log   *logan.Entry
	raw   data.VoteQ
	cache *marshaler.Marshaler
}

func (q *VoteQ) InsertBatchCtx(ctx context.Context, votes ...data.Vote) error {
	return q.raw.InsertBatchCtx(ctx, votes...)
}

func (q *VoteQ) VotesByTransferIndexCtx(ctx context.Context, transferIndex []byte, isForUpdate bool) ([]data.Vote, error) {
	if !isForUpdate {
		var votes []data.Vote
		err := tryGetFromCache(ctx, q.cache, makeVotesByTransferCacheKey(string(transferIndex)), &votes)
		if err == nil {
			q.log.Debug("hit")
			return votes, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get votes form cache")
		}
	}

	votes, err := q.raw.VotesByTransferIndexCtx(ctx, transferIndex, isForUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get votes by transfer idx", logan.F{
			"transfer":   string(transferIndex),
			"for_update": isForUpdate,
		})
	}

	if len(votes) != 0 {
		err = q.cache.Set(ctx, makeVotesByTransferCacheKey(string(transferIndex)), votes) // votes are immutable, so we don't need to invalidate them in any way
		if err != nil {
			q.log.WithError(err).Error("failed to cache votes batch")
			return votes, nil
		}

		err = q.cacheEveryVote(ctx, votes)
		if err != nil {
			q.log.WithError(err).Error("failed to cache votes individually")
		}
	}

	return votes, nil
}

func (q *VoteQ) cacheEveryVote(ctx context.Context, votes []data.Vote) error {
	for _, vote := range votes {
		err := q.cache.Set(ctx, makeVoteCacheKey(vote.ID), vote)
		if err != nil {
			return errors.Wrap(err, "failed to cache vote", logan.F{
				"id": vote.ID,
			})
		}
	}

	return nil
}

func makeVoteCacheKey(id int64) string {
	return "vote:" + strconv.FormatInt(id, 10)
}

func makeVotesByTransferCacheKey(transferIndex string) string {
	return "votes_by_transfer:" + transferIndex
}
