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
	collectionChainMapping = "collection_chain_mapping"
)

type CollectionChainMappingQ struct {
	log   *logan.Entry
	raw   data.CollectionChainMappingQ
	cache *marshaler.Marshaler
}

func (q *CollectionChainMappingQ) InsertBatchCtx(ctx context.Context, chainMappings ...data.CollectionChainMapping) error {
	tags := make([]string, 0, 3*len(chainMappings))
	for _, cm := range chainMappings {
		tags = append(tags, icmCacheTags(cm.Collection, cm.Network)...)
	}

	err := q.cache.Invalidate(ctx, store.WithInvalidateTags(tags))
	if err != nil {
		q.log.WithError(err).Error("failed to invalidate cache")
	}

	return q.raw.InsertBatchCtx(ctx, chainMappings...)
}

func (q *CollectionChainMappingQ) InsertCtx(ctx context.Context, ccm *data.CollectionChainMapping) error {
	err := q.cache.Invalidate(ctx, store.WithInvalidateTags(icmCacheTags(ccm.Collection, ccm.Network)))
	if err != nil {
		q.log.WithError(err).Error("failed to invalidate cache")
	}

	return q.raw.InsertCtx(ctx, ccm)
}

func (q *CollectionChainMappingQ) CollectionChainMappingByCollectionNetworkCtx(ctx context.Context, collection int64, network int, isForUpdate bool) (*data.CollectionChainMapping, error) {
	if !isForUpdate {
		var ccm data.CollectionChainMapping

		err := tryGetFromCache(ctx, q.cache, collectionChainMappingCacheKey(collection, network), &ccm)
		if err == nil {
			q.log.Debug("hit")
			return &ccm, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get collection chain mapping form cache")
		}
	}

	ccm, err := q.raw.CollectionChainMappingByCollectionNetworkCtx(ctx, collection, network, isForUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get collection chain mapping from db", logan.F{
			"collection": collection,
			"network":    network,
		})
	}

	if ccm != nil {
		err = q.cache.Set(ctx,
			collectionChainMappingCacheKey(collection, network),
			ccm,
			store.WithTags([]string{collectionChainMappingCacheKey(collection, network)}),
		)
		if err != nil {
			q.log.WithError(err).Error("failed to set collection chain mapping to cache")
		}
	}

	return ccm, nil
}

func (q *CollectionChainMappingQ) UpsertCtx(ctx context.Context, ccm *data.CollectionChainMapping) error {
	err := q.cache.Invalidate(ctx,
		store.WithInvalidateTags(ccmCacheTags(ccm.Collection, ccm.Network)))
	if err != nil {
		q.log.WithError(err).Error("failed to invalidate cached entry")
	}

	return q.raw.UpsertCtx(ctx, ccm)
}

func (q *CollectionChainMappingQ) DeleteByCollectionCtx(ctx context.Context, collection int64) error {
	err := q.cache.Invalidate(ctx,
		store.WithInvalidateTags(ccmCollectionTags(collection)))
	if err != nil {
		q.log.
			WithFields(logan.F{
				"collection": collection,
			}).
			WithError(err).Error("failed to invalidate cached entry")
	}

	return q.raw.DeleteByCollectionCtx(ctx, collection)
}

func (q *CollectionChainMappingQ) DeleteCtx(ctx context.Context, ccm *data.CollectionChainMapping) error {
	err := q.cache.Invalidate(ctx,
		store.WithInvalidateTags(ccmCacheTags(ccm.Collection, ccm.Network)))
	if err != nil {
		q.log.WithFields(logan.F{
			"collection": ccm.Collection,
			"network":    ccm.Network,
		}).WithError(err).Error("failed to invalidate cached entry")
	}

	return q.raw.DeleteCtx(ctx, ccm)
}

func collectionChainMappingCacheKey(collection int64, network int) string {
	return fmt.Sprintf("%s:%d:%d", collectionChainMapping, collection, network)
}

func ccmCacheTags(collection int64, network int) []string {
	return []string{
		collectionChainMappingCacheKey(collection, network),
		fmt.Sprintf("%s:%d", collectionChainMapping, collection), // to invalidate all mappings for collection
		fmt.Sprintf("%s:%d", collectionChainMapping, network),    // to invalidate all mappings for network
	}
}

func ccmCollectionTags(collection int64) []string {
	return []string{
		fmt.Sprintf("%s:%d", collectionChainMapping, collection),
	}
}

func ccmNetworkTags(network int) []string {
	return []string{
		fmt.Sprintf("%s:%d", collectionChainMapping, network),
	}
}
