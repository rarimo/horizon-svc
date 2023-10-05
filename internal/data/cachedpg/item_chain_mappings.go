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
	itemChainMapping = "item_chain_mapping"
)

type ItemChainMappingQ struct {
	log   *logan.Entry
	raw   data.ItemChainMappingQ
	cache *marshaler.Marshaler
}

func (q *ItemChainMappingQ) InsertBatchCtx(ctx context.Context, chainMappings ...data.ItemChainMapping) error {
	tags := make([]string, 0, 3*len(chainMappings))
	for _, cm := range chainMappings {
		tags = append(tags, icmCacheTags(cm.Item, cm.Network)...)
	}

	err := q.cache.Invalidate(ctx, store.WithInvalidateTags(tags))
	if err != nil {
		q.log.WithError(err).Error("failed to invalidate cache")
	}

	return q.raw.InsertBatchCtx(ctx, chainMappings...)
}

func (q *ItemChainMappingQ) InsertCtx(ctx context.Context, icm *data.ItemChainMapping) error {
	err := q.cache.Invalidate(ctx, store.WithInvalidateTags(icmCacheTags(icm.Item, icm.Network)))
	if err != nil {
		q.log.WithError(err).Error("failed to invalidate cache")
	}

	return q.raw.InsertCtx(ctx, icm)
}

func (q *ItemChainMappingQ) ItemChainMappingsByNetworkCtx(ctx context.Context, network int, isForUpdate bool) ([]data.ItemChainMapping, error) {
	if !isForUpdate {
		var result []data.ItemChainMapping

		err := tryGetFromCache(ctx, q.cache, icmNetworkCacheKey(network), &result)
		if err == nil {
			q.log.Debug("hit")
			return result, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get item chain mapping form cache")
		}
	}

	result, err := q.raw.ItemChainMappingsByNetworkCtx(ctx, network, isForUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get item chain mapping by network", logan.F{
			"network": network,
		})
	}

	if len(result) != 0 {
		err = q.cache.Set(ctx,
			icmNetworkCacheKey(network),
			result,
			store.WithTags([]string{icmNetworkCacheKey(network)}))
		if err != nil {
			q.log.WithError(err).Error("failed to set item chain mapping to cache")
		}
	}

	return result, nil
}

func (q *ItemChainMappingQ) ItemChainMappingByItemNetworkCtx(ctx context.Context, item int64, network int, isForUpdate bool) (*data.ItemChainMapping, error) {
	if !isForUpdate {
		var icm data.ItemChainMapping

		err := tryGetFromCache(ctx, q.cache, itemChainMappingCacheKey(item, network), &icm)
		if err == nil {
			q.log.Debug("hit")
			return &icm, nil
		}

		q.log.Debug("miss")
		if errors.Cause(err) != redis.Nil {
			q.log.WithError(err).Error("failed to get item chain mapping form cache")
		}
	}

	icm, err := q.raw.ItemChainMappingByItemNetworkCtx(ctx, item, network, isForUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get item chain mapping by item and network", logan.F{
			"item":    item,
			"network": network,
		})
	}

	if icm != nil {
		err = q.cache.Set(ctx,
			itemChainMappingCacheKey(item, network),
			icm,
			store.WithTags(icmCacheTags(item, network)))
		if err != nil {
			q.log.WithError(err).Error("failed to set item chain mapping to cache")
		}
	}

	return icm, nil
}

func (q *ItemChainMappingQ) DeleteByItemCtx(ctx context.Context, item int64) error {
	err := q.cache.Invalidate(ctx,
		store.WithInvalidateTags([]string{icmItemCacheKey(item)}))
	if err != nil {
		q.log.WithError(err).Error("failed to invalidate item chain mapping cache")
	}

	return q.raw.DeleteByItemCtx(ctx, item)
}

func (q *ItemChainMappingQ) DeleteCtx(ctx context.Context, icm *data.ItemChainMapping) error {
	err := q.cache.Invalidate(ctx,
		store.WithInvalidateTags([]string{
			itemChainMappingCacheKey(icm.Item, icm.Network),
		}))
	if err != nil {
		q.log.WithError(err).Error("failed to invalidate item chain mapping cache")
	}
	return q.raw.DeleteCtx(ctx, icm)
}

func itemChainMappingCacheKey(item int64, network int) string {
	return fmt.Sprintf("%s:%d:%d", itemChainMapping, item, network)
}

func icmCacheTags(item int64, network int) []string {
	return []string{
		itemChainMappingCacheKey(item, network),
		icmItemCacheKey(item),       // to invalidate all mappings for this item
		icmNetworkCacheKey(network), // to invalidate all mappings for this network
	}
}

func icmItemCacheKey(item int64) string {
	return fmt.Sprintf("%s:%d", itemChainMapping, item)

}

func icmNetworkCacheKey(network int) string {
	return fmt.Sprintf("%s:%d", itemChainMapping, network)
}
