package services

import (
	"context"
	"database/sql"
	"github.com/rarimo/horizon-svc/internal/core"
	"time"

	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func RunCollectionsIndexer(ctx context.Context, cfg config.Config) {
	log := cfg.Log().WithField("who", cfg.CollectionsIndexer().RunnerName)

	handler := &collectionIndexer{
		log:          log,
		tokenmanager: cfg.Core().Tokenmanager(),
		storage:      cfg.CachedStorage().Clone(),
		chains:       cfg.ChainsQ(),
		saver:        NewTokenmanagerSaver(cfg),
	}

	msgs.NewConsumer(
		cfg.Log(),
		cfg.CollectionsIndexer().CollectionEventsConsumer,
		handler,
	).Run(ctx)
}

type collectionIndexer struct {
	log          *logan.Entry
	tokenmanager core.Tokenmanager
	storage      data.Storage
	chains       data.ChainsQ
	saver        *TokenmanagerSaver
}

func (p *collectionIndexer) Handle(ctx context.Context, batch []msgs.Message) error {
	for _, raw := range batch {
		if err := p.handle(ctx, raw); err != nil {
			return errors.Wrap(err, "failed to handle message", logan.F{
				"type": raw.Type(),
			})
		}
	}

	return nil
}

func (p *collectionIndexer) handle(ctx context.Context, raw msgs.Message) error {
	switch raw.Type() {
	case msgs.MessageTypeCollectionCreated:
		return p.handleCollectionCreated(ctx, raw.MustCollectionCreatedMessage())
	case msgs.MessageTypeCollectionRemoved:
		return p.handleCollectionRemoved(ctx, raw.MustCollectionRemovedMessage())
	case msgs.MessageTypeCollectionDataCreated:
		return p.handleCollectionDataCreated(ctx, raw.MustCollectionDataCreatedMessage())
	case msgs.MessageTypeCollectionDataUpdated:
		return p.handleCollectionDataUpdated(ctx, raw.MustCollectionDataUpdatedMessage())
	case msgs.MessageTypeCollectionDataRemoved:
		return p.handleCollectionDataRemoved(ctx, raw.MustCollectionDataRemovedMessage())
	default:
		p.log.WithField("type", raw.Type()).Warn("unknown message type")
		return nil
	}
}

func (p *collectionIndexer) handleCollectionCreated(ctx context.Context, msg msgs.CollectionCreatedMessage) error {
	collection, err := p.tokenmanager.GetCollection(ctx, msg.Index)
	if err != nil {
		return errors.Wrap(err, "failed to get collection")
	}

	_, err = p.saver.SaveCollection(ctx, *collection)
	return err
}

func (p *collectionIndexer) handleCollectionRemoved(ctx context.Context, msg msgs.CollectionRemovedMessage) error {
	return p.storage.Transaction(func() error {
		collection, err := p.storage.CollectionQ().CollectionByIndexCtx(ctx, []byte(msg.Index), true)
		if err != nil {
			return errors.Wrap(err, "failed to get collection from db", logan.F{
				"index": msg.Index,
			})
		}

		err = p.storage.CollectionChainMappingQ().DeleteByCollectionCtx(ctx, collection.ID)
		if err != nil {
			return errors.Wrap(err, "failed to delete collection data for collection", logan.F{
				"collection": collection.ID,
			})
		}

		// TODO add state and set status `removed` instead of deleting (to avoid problems with restarting from genesis)
		err = p.storage.CollectionQ().DeleteCtx(ctx, collection)
		if err != nil {
			return errors.Wrap(err, "failed to delete collection", logan.F{
				"collection": collection.ID,
			})
		}

		return nil
	})
}

func (p *collectionIndexer) handleCollectionDataCreated(ctx context.Context, msg msgs.CollectionDataCreatedMessage) error {
	network := p.chains.Get(msg.Chain)
	if network == nil {
		return errors.From(errors.New("chain not supported on horizon"), logan.F{
			"core_chain": msg.Chain,
		})
	}

	collection, err := p.storage.CollectionQ().CollectionByIndexCtx(ctx, []byte(msg.CollectionIndex), false)
	if err != nil {
		return errors.Wrap(err, "failed to get collection from db", logan.F{
			"index": msg.CollectionIndex,
		})
	}

	if collection == nil {
		return errors.From(errors.New("collection not found in storage"), logan.F{
			"index": msg.CollectionIndex,
		})
	}

	coreCollection, err := p.tokenmanager.GetCollectionDataByCollectionForChain(ctx, string(collection.Index), network.Name)
	if err != nil {
		return errors.Wrap(err, "failed to get collection data")
	}

	now := time.Now().UTC()
	err = p.storage.CollectionChainMappingQ().InsertCtx(ctx, &data.CollectionChainMapping{
		Collection: collection.ID,
		Network:    network.ID,
		Address:    []byte(coreCollection.Index.Address),
		TokenType: sql.NullInt64{
			Int64: int64(coreCollection.TokenType),
			Valid: true,
		},
		Wrapped: sql.NullBool{
			Bool:  coreCollection.Wrapped,
			Valid: true,
		},
		Decimals: sql.NullInt64{
			Int64: int64(coreCollection.Decimals),
			Valid: true,
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return errors.Wrap(err, "failed to insert collection data")
	}

	return nil
}

func (p *collectionIndexer) handleCollectionDataUpdated(ctx context.Context, msg msgs.CollectionDataUpdatedMessage) error {
	collection, err := p.storage.CollectionQ().CollectionByIndexCtx(ctx, []byte(msg.CollectionIndex), false)
	if err != nil {
		return errors.Wrap(err, "failed to get collection from db", logan.F{
			"index": msg.CollectionIndex,
		})
	}

	if collection == nil {
		return errors.From(errors.New("collection not found"), logan.F{
			"index": msg.CollectionIndex,
		})
	}

	network := p.chains.Get(msg.Chain)
	if network == nil {
		return errors.From(errors.New("chain not supported on horizon"), logan.F{
			"core_chain": msg.Chain,
		})
	}

	ccm, err := p.storage.CollectionChainMappingQ().
		CollectionChainMappingByCollectionNetworkCtx(ctx, collection.ID, network.ID, true)
	if err != nil {
		return errors.Wrap(err, "failed to get collection data from db", logan.F{
			"collection": collection.ID,
			"network":    network.ID,
		})
	}

	if ccm == nil {
		return errors.From(errors.New("collection data not found"), logan.F{
			"collection": collection.ID,
			"network":    network.ID,
		})
	}

	collectionData, err := p.tokenmanager.GetCollectionDataByCollectionForChain(ctx, string(collection.Index), network.Name)
	if err != nil {
		return errors.Wrap(err, "failed to get collection data")
	}

	ccm.TokenType = sql.NullInt64{
		Int64: int64(collectionData.TokenType),
		Valid: true,
	}
	ccm.Wrapped = sql.NullBool{
		Bool:  collectionData.Wrapped,
		Valid: true,
	}
	ccm.Decimals = sql.NullInt64{
		Int64: int64(collectionData.Decimals),
		Valid: true,
	}

	if err := p.storage.CollectionChainMappingQ().UpsertCtx(ctx, ccm); err != nil {
		return errors.Wrap(err, "failed to update collection data", logan.F{
			"collection": collection.ID,
			"network":    network.ID,
		})
	}

	return nil
}

func (p *collectionIndexer) handleCollectionDataRemoved(ctx context.Context, msg msgs.CollectionDataRemovedMessage) error {
	collection, err := p.storage.CollectionQ().CollectionByIndexCtx(ctx, []byte(msg.CollectionIndex), false)
	if err != nil {
		return errors.Wrap(err, "failed to get collection from db", logan.F{
			"index": msg.CollectionIndex,
		})
	}

	if collection == nil {
		return errors.From(errors.New("collection not found"), logan.F{
			"index": msg.CollectionIndex,
		})
	}

	network := p.chains.Get(msg.Chain)
	if network == nil {
		return errors.From(errors.New("chain not supported on horizon"), logan.F{
			"core_chain": msg.Chain,
		})
	}

	ccm, err := p.storage.CollectionChainMappingQ().CollectionChainMappingByCollectionNetworkCtx(ctx, collection.ID, network.ID, true)
	if err != nil {
		return errors.Wrap(err, "failed to get collection data from db", logan.F{
			"collection": collection.ID,
			"network":    network.ID,
		})
	}

	if ccm == nil {
		p.log.WithFields(logan.F{
			"collection": msg.CollectionIndex,
			"network":    msg.Chain,
		}).Warn("collection data not found")
		return nil
	}

	if err := p.storage.CollectionChainMappingQ().DeleteCtx(ctx, ccm); err != nil {
		return errors.Wrap(err, "failed to delete collection data", logan.F{
			"collection": collection.ID,
			"network":    network.ID,
		})
	}

	return nil
}
