package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func RunItemsIndexer(ctx context.Context, cfg config.Config) {
	log := cfg.Log().WithField("who", cfg.ItemsIndexer().RunnerName+"_items_indexer")

	handler := &itemsIndexer{
		log:          log,
		tokenmanager: tokenmanager.NewQueryClient(cfg.Cosmos()),
		storage:      cfg.CachedStorage(),
		chains:       cfg.ChainsQ(),
		saver:        NewTokenmanagerSaver(cfg),
	}

	msgs.NewConsumer(
		cfg.Log(),
		cfg.ItemsIndexer().ItemEventsConsumer,
		handler,
	).Run(ctx)
}

type itemsIndexer struct {
	log          *logan.Entry
	tokenmanager tokenmanager.QueryClient
	storage      data.Storage
	chains       data.ChainsQ
	saver        *TokenmanagerSaver
}

func (p *itemsIndexer) Handle(ctx context.Context, batch []msgs.Message) error {
	p.log.WithFields(logan.F{
		"batch_size": len(batch),
		"batch":      batch,
	}).Debug("got batch")

	for _, raw := range batch {
		if err := p.handle(ctx, raw); err != nil {
			return errors.Wrap(err, "failed to handle message", logan.F{
				"message": raw,
			})
		}
	}

	return nil
}

func (p *itemsIndexer) handle(ctx context.Context, raw msgs.Message) error {
	switch raw.Type() {
	case tokenmanager.EventTypeItemCreated:
		return p.handleItemCreated(ctx, raw.MustItemCreatedMessage())
	case tokenmanager.EventTypeItemRemoved:
		return p.handleItemRemoved(ctx, raw.MustItemRemovedMessage())
	case tokenmanager.EventTypeOnChainItemCreated:
		return p.handleOnChainItemCreated(ctx, raw.MustItemOnChainDataCreatedMessage())
	case tokenmanager.EventTypeOnChainItemRemoved:
		return p.handleOnChainItemRemoved(ctx, raw.MustItemOnChainDataRemovedMessage())
	case tokenmanager.EventTypeSeedCreated:
		return p.handleSeedCreated(ctx, raw.MustSeedCreatedMessage())
	case tokenmanager.EventTypeSeedRemoved:
		return p.handleSeedRemoved(ctx, raw.MustSeedRemovedMessage())
	default:
		p.log.WithField("type", raw.Type()).Warn("unknown message type")
		return nil
	}
}

func (p *itemsIndexer) handleItemCreated(ctx context.Context, msg msgs.ItemCreatedMessage) error {
	p.log.
		WithField("index", msg.Index).
		Debug("handling item created")

	itemResp, err := p.tokenmanager.Item(ctx, &tokenmanager.QueryGetItemRequest{Index: msg.Index})
	if err != nil {
		return errors.Wrap(err, "failed to get item from core", logan.F{
			"index": msg.Index,
		})
	}

	_, err = p.saver.SaveItem(ctx, itemResp.Item)
	return errors.Wrap(err, "failed to save item", logan.F{
		"index": msg.Index,
	})
}

func (p *itemsIndexer) handleItemRemoved(ctx context.Context, msg msgs.ItemRemovedMessage) error {
	p.log.
		WithField("index", msg.Index).
		Debug("handling item removed")

	return p.storage.Transaction(func() error {
		item, err := p.storage.ItemQ().ItemByIndexCtx(ctx, []byte(msg.Index), true)
		if err != nil {
			return errors.Wrap(err, "failed to get item from storage", logan.F{
				"index": msg.Index,
			})
		}

		err = p.storage.ItemChainMappingQ().DeleteByItemCtx(ctx, item.ID)
		if err != nil {
			return errors.Wrap(err, "failed to delete item chain mappings", logan.F{
				"index": msg.Index,
			})
		}

		// TODO add state and set status `removed` instead of deleting (to avoid problems with restarting from genesis)
		err = p.storage.ItemQ().DeleteCtx(ctx, item)
		if err != nil {
			return errors.Wrap(err, "failed to delete item", logan.F{
				"index": msg.Index,
			})
		}

		return nil
	})
}

func (p *itemsIndexer) handleOnChainItemCreated(ctx context.Context, msg msgs.ItemOnChainDataCreatedMessage) error {
	p.log.
		WithFields(logan.F{
			"index": msg.ItemIndex,
			"chain": msg.Chain,
		}).
		Debug("handling on chain item created")

	network := p.chains.Get(msg.Chain)
	if network == nil {
		p.log.WithField("core_chain", msg.Chain).Warn("chain not supported on horizon")
		return nil
	}

	item, err := p.storage.ItemQ().ItemByIndexCtx(ctx, []byte(msg.ItemIndex), false)
	if err != nil {
		return errors.Wrap(err, "failed to get item from storage", logan.F{
			"index": msg.ItemIndex,
		})
	}

	if item == nil {
		return errors.From(errors.New("item not found in storage"), logan.F{
			"index": msg.ItemIndex,
		})
	}

	onChainItemResp, err := p.tokenmanager.OnChainItemByItemForChain(ctx,
		&tokenmanager.QueryGetOnChainItemByItemForChainRequest{
			ItemIndex: string(item.Index),
			Chain:     network.Name,
		})
	if err != nil {
		// FIXME(hp): handle not found and precondition failed ?
		return errors.Wrap(err, "failed to get on-chain item from core", logan.F{
			"index": msg.ItemIndex,
		})
	}

	now := time.Now().UTC()
	err = p.storage.ItemChainMappingQ().InsertCtx(ctx, &data.ItemChainMapping{
		Item:      item.ID,
		Network:   network.ID,
		Address:   []byte(onChainItemResp.Item.Index.Address),
		TokenID:   []byte(onChainItemResp.Item.Index.TokenID),
		CreatedAt: now,
		UpdatedAt: now,
	})

	if err != nil {
		return errors.Wrap(err, "failed to insert item chain mapping")
	}

	return nil
}

func (p *itemsIndexer) handleOnChainItemRemoved(ctx context.Context, msg msgs.ItemOnChainDataRemovedMessage) error {
	p.log.
		WithFields(logan.F{
			"index": msg.ItemIndex,
			"chain": msg.Chain,
		}).
		Debug("handling on chain item removed")

	item, err := p.storage.ItemQ().ItemByIndexCtx(ctx, []byte(msg.ItemIndex), false)
	if err != nil {
		return errors.Wrap(err, "failed to get item from storage", logan.F{
			"index": msg.ItemIndex,
		})
	}

	if item == nil {
		return errors.From(errors.New("item not found in storage"), logan.F{
			"index": msg.ItemIndex,
		})
	}

	network := p.chains.Get(msg.Chain)
	if network == nil {
		p.log.WithField("core_chain", msg.Chain).Warn("chain not supported on horizon")
		return nil
	}

	icm, err := p.storage.ItemChainMappingQ().ItemChainMappingByItemNetworkCtx(ctx, item.ID, network.ID, true)
	if err != nil {
		return errors.Wrap(err, "failed to get item chain mapping from storage", logan.F{
			"index": msg.ItemIndex,
		})
	}

	if icm == nil {
		p.log.WithFields(logan.F{
			"index": msg.ItemIndex,
			"chain": msg.Chain,
		}).Warn("item chain mapping not found in storage")
		return nil
	}

	if err := p.storage.ItemChainMappingQ().DeleteCtx(ctx, icm); err != nil {
		return errors.Wrap(err, "failed to delete item chain mapping", logan.F{
			"index": msg.ItemIndex,
		})
	}

	return nil
}

func (p *itemsIndexer) handleSeedCreated(ctx context.Context, msg msgs.SeedCreatedMessage) error {
	p.log.
		WithFields(logan.F{
			"index": msg.ItemIndex,
			"seed":  msg.Seed,
		}).
		Debug("handling seed created")

	return p.storage.Transaction(func() error {
		item, err := p.storage.ItemQ().ItemByIndexCtx(ctx, []byte(msg.ItemIndex), true)
		if err != nil {
			return errors.Wrap(err, "failed to get item from storage", logan.F{
				"index": msg.ItemIndex,
			})
		}

		if item == nil {
			return errors.From(errors.New("item not found in storage"), logan.F{
				"index": msg.ItemIndex,
			})
		}

		if json.Valid(item.Metadata) {
			var meta tokenmanager.ItemMetadata
			err := json.Unmarshal(item.Metadata, &meta)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal item metadata", logan.F{
					"index": msg.ItemIndex,
				})
			}

			meta.Seed = msg.Seed

			item.Metadata, err = json.Marshal(meta)
			if err != nil {
				return errors.Wrap(err, "failed to marshal item metadata", logan.F{
					"index": msg.ItemIndex,
				})
			}
		}

		err = p.storage.ItemQ().UpdateCtx(ctx, item)
		if err != nil {
			return errors.Wrap(err, "failed to update item", logan.F{
				"index": msg.ItemIndex,
			})
		}

		return nil
	})
}

func (p *itemsIndexer) handleSeedRemoved(ctx context.Context, msg msgs.SeedRemovedMessage) error {
	p.log.
		WithFields(logan.F{
			"index": msg.ItemIndex,
			"seed":  msg.Seed,
		}).
		Debug("handling seed removed")

	return p.storage.Transaction(func() error {
		item, err := p.storage.ItemQ().ItemByIndexCtx(ctx, []byte(msg.ItemIndex), true)
		if err != nil {
			return errors.Wrap(err, "failed to get item from storage", logan.F{
				"index": msg.ItemIndex,
			})
		}

		if item == nil {
			return errors.From(errors.New("item not found in storage"), logan.F{
				"index": msg.ItemIndex,
			})
		}

		if json.Valid(item.Metadata) {
			var meta tokenmanager.ItemMetadata
			err := json.Unmarshal(item.Metadata, &meta)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal item metadata", logan.F{
					"index": msg.ItemIndex,
				})
			}

			if meta.Seed == msg.Seed { // removing only if seed matches to avoid writes by non-consecutive events TODO should be checked by lastSeenBlock
				meta.Seed = ""
			}

			item.Metadata, err = json.Marshal(meta)
			if err != nil {
				return errors.Wrap(err, "failed to marshal item metadata", logan.F{
					"index": msg.ItemIndex,
				})
			}
		}

		err = p.storage.ItemQ().UpdateCtx(ctx, item)
		if err != nil {
			return errors.Wrap(err, "failed to update item", logan.F{
				"index": msg.ItemIndex,
			})
		}

		return nil
	})
}
