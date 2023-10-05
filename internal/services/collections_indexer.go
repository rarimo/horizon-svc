package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/rarimo/horizon-svc/internal/core"

	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func RunCollectionsIndexer(ctx context.Context, cfg config.Config) {
	log := cfg.Log().WithField("who", cfg.CollectionsIndexer().RunnerName)

	handler := &collectionIndexer{
		log:          log,
		tokenmanager: tokenmanager.NewQueryClient(cfg.Cosmos()),
		storage:      cfg.CachedStorage(),
		chains:       cfg.ChainsQ(),
	}

	if err := handler.loadGenesisCollections(ctx, cfg.Genesis()); err != nil {
		panic(errors.Wrap(err, "failed to catchup collections"))
	}

	msgs.NewConsumer(
		cfg.Log(),
		cfg.CollectionsIndexer().CollectionEventsConsumer,
		handler,
	).Run(ctx)
}

type collectionIndexer struct {
	log                *logan.Entry
	tokenmanager       tokenmanager.QueryClient
	tokenmanagerGetter core.TokenManager
	storage            data.Storage
	chains             data.ChainsQ
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

func (p *collectionIndexer) loadGenesisCollections(ctx context.Context, cfg config.GenesisCfg) error {
	if cfg.Disabled {
		p.log.Info("skipping genesis collections loading (disabled in config)")
		return nil
	}

	for _, col := range cfg.GenesisState.Collections {
		exising, err := p.storage.CollectionQ().CollectionByIndexCtx(ctx, []byte(col.Index), false)
		if err != nil {
			return errors.Wrap(err, "failed to get collection from storage", logan.F{
				"index": col.Index,
			})
		}

		if exising != nil {
			continue
		}

		err = p.storage.Transaction(func() error {
			savedCollection, err := p.saveCollection(ctx, col)
			if err != nil {
				return errors.Wrap(err, "failed to save collection", logan.F{
					"index": col.Index,
				})
			}

			cdataToSave := make([]*tokenmanager.CollectionData, 0, len(col.Data))
			for _, genesisData := range cfg.GenesisState.Datas {
				if genesisData.Collection != col.Index {
					continue
				}

				cdataToSave = append(cdataToSave, &tokenmanager.CollectionData{
					Index: &tokenmanager.CollectionDataIndex{
						Chain:   genesisData.Index.Chain,
						Address: genesisData.Index.Address,
					},
					Collection: genesisData.Collection,
					TokenType:  genesisData.TokenType,
					Wrapped:    genesisData.Wrapped,
					Decimals:   genesisData.Decimals,
				})
			}

			err = p.saveCollectionData(ctx, savedCollection.ID, cdataToSave)
			if err != nil {
				return errors.Wrap(err, "failed  to save collection data", logan.F{
					"index": col.Index,
				})
			}

			return nil
		})

		if err != nil {
			return errors.Wrap(err, "failed to save collection", logan.F{
				"index": col.Index,
			})
		}
	}

	return nil
}

func (p *collectionIndexer) handleCollectionCreated(ctx context.Context, msg msgs.CollectionCreatedMessage) error {
	collectionResp, err := p.tokenmanager.Collection(ctx, &tokenmanager.QueryGetCollectionRequest{
		Index: msg.Index,
	})

	if err != nil {
		return errors.Wrap(err, "failed to get collection from core", logan.F{
			"index": msg.Index,
		})
	}

	_, err = p.saveCollection(ctx, collectionResp.Collection)
	return err
}

func (p *collectionIndexer) saveCollection(ctx context.Context, coreCollection tokenmanager.Collection) (*data.Collection, error) {
	now := time.Now()
	col := data.Collection{
		Index:     []byte(coreCollection.Index),
		Metadata:  []byte("{}"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	var err error

	col.Metadata, err = json.Marshal(coreCollection.Meta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal collection metadata", logan.F{
			"index": coreCollection.Index,
		})
	}

	if err := p.storage.CollectionQ().InsertCtx(ctx, &col); err != nil {
		return nil, errors.Wrap(err, "failed to insert collection", logan.F{
			"index": coreCollection.Index,
		})
	}

	return &col, nil
}

func (p *collectionIndexer) saveCollectionData(ctx context.Context, collectionID int64, coreCollectionData []*tokenmanager.CollectionData) error {
	now := time.Now()

	collectionDataBatch := make([]data.CollectionChainMapping, len(coreCollectionData))
	for i, collectionData := range coreCollectionData {
		network := p.chains.Get(collectionData.Index.Chain)
		if network == nil {
			p.log.WithField("core_chain", collectionData.Index.Chain).Warn("chain not supported on horizon")
			continue
		}

		collectionDataBatch[i] = data.CollectionChainMapping{
			Collection: collectionID,
			Network:    network.ID,
			Address:    []byte(collectionData.Index.Address),
			TokenType: sql.NullInt64{
				Int64: int64(collectionData.TokenType),
				Valid: true,
			},
			Wrapped: sql.NullBool{
				Bool:  collectionData.Wrapped,
				Valid: true,
			},
			Decimals: sql.NullInt64{
				Int64: int64(collectionData.Decimals),
				Valid: true,
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	err := p.storage.CollectionChainMappingQ().InsertBatchCtx(ctx, collectionDataBatch...)
	return errors.Wrap(err, "failed to insert collection data", logan.F{
		"collection": collectionID,
	})
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

	collectionDataResp, err := p.tokenmanager.CollectionDataByCollectionForChain(ctx,
		&tokenmanager.QueryGetCollectionDataByCollectionForChainRequest{
			CollectionIndex: string(collection.Index),
			Chain:           network.Name,
		})
	if err != nil {
		// FIXME(hp): handle not found and precondition failed ?
		return errors.Wrap(err, "failed to get collection data from core", logan.F{
			"index": msg.CollectionIndex,
			"chain": network.Name,
		})
	}

	now := time.Now()
	err = p.storage.CollectionChainMappingQ().InsertCtx(ctx, &data.CollectionChainMapping{
		Collection: collection.ID,
		Network:    network.ID,
		Address:    []byte(collectionDataResp.Data.Index.Address),
		TokenType: sql.NullInt64{
			Int64: int64(collectionDataResp.Data.TokenType),
			Valid: true,
		},
		Wrapped: sql.NullBool{
			Bool:  collectionDataResp.Data.Wrapped,
			Valid: true,
		},
		Decimals: sql.NullInt64{
			Int64: int64(collectionDataResp.Data.Decimals),
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

	collectionData, err := p.tokenmanager.CollectionDataByCollectionForChain(ctx,
		&tokenmanager.QueryGetCollectionDataByCollectionForChainRequest{
			CollectionIndex: string(collection.Index),
			Chain:           network.Name,
		})
	if err != nil {
		// FIXME(hp): handle not found and precondition failed ?
		return errors.Wrap(err, "failed to get collection data from core", logan.F{
			"index": msg.CollectionIndex,
			"chain": network.Name,
		})
	}

	ccm.TokenType = sql.NullInt64{
		Int64: int64(collectionData.Data.TokenType),
		Valid: true,
	}
	ccm.Wrapped = sql.NullBool{
		Bool:  collectionData.Data.Wrapped,
		Valid: true,
	}
	ccm.Decimals = sql.NullInt64{
		Int64: int64(collectionData.Data.Decimals),
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
