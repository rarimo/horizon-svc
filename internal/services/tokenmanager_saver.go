package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/core"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"strconv"
	"time"
)

const (
	IsGenesisParsedKey = "is_genesis_parsed"
)

type TokenmanagerSaver struct {
	log          *logan.Entry
	storage      data.Storage
	chains       data.ChainsQ
	kv           data.KeyValueQ
	genesis      config.GenesisConfig
	tokenmanager core.Tokenmanager
}

func NewTokenmanagerSaver(cfg config.Config) *TokenmanagerSaver {
	return &TokenmanagerSaver{
		cfg.Log().WithField("who", "tokenmanager_saver"),
		cfg.CachedStorage().Clone(),
		cfg.ChainsQ(),
		redis.NewKeyValueProvider(cfg),
		cfg.Genesis(),
		cfg.Core().Tokenmanager(),
	}
}

func (s *TokenmanagerSaver) ParseAndSaveGenesis(ctx context.Context) error {
	if s.genesis.Disabled {
		s.log.Info("skipping genesis items loading (disabled in config)")
	}

	isParsed, err := s.getIsParsed(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get genesis parsed key")
	}

	if isParsed {
		s.log.Info("genesis already parsed")
		return nil
	}

	if err := s.saveGenesisCollections(ctx); err != nil {
		return errors.Wrap(err, "failed to save collections from genesis")
	}

	if err := s.saveGenesisItems(ctx); err != nil {
		return errors.Wrap(err, "failed to save items from genesis")
	}

	return nil
}

func (s *TokenmanagerSaver) setIsParsed(ctx context.Context, value bool) error {
	now := time.Now().UTC()
	err := s.kv.Upsert(ctx, data.KeyValue{
		Key:       IsGenesisParsedKey,
		Value:     strconv.FormatBool(value),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return errors.Wrap(err, "failed to upsert genesis parsed key")
	}

	return nil
}

func (s *TokenmanagerSaver) getIsParsed(ctx context.Context) (bool, error) {
	kv, err := s.kv.Get(ctx, IsGenesisParsedKey)
	if err != nil {
		return false, errors.Wrap(err, "failed to get genesis parsed key")
	}

	if kv == nil {
		err = s.setIsParsed(ctx, false)
		if err != nil {
			return false, errors.Wrap(err, "failed to upsert genesis parsed key")
		}

		return false, nil
	}

	return strconv.ParseBool(kv.Value)
}

func (s *TokenmanagerSaver) SaveCollection(ctx context.Context, coreCollection tokenmanager.Collection) (*data.Collection, error) {
	now := time.Now().UTC()
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

	if err := s.storage.CollectionQ().InsertCtx(ctx, &col); err != nil {
		return nil, errors.Wrap(err, "failed to insert collection", logan.F{
			"index": coreCollection.Index,
		})
	}

	return &col, nil
}

func (s *TokenmanagerSaver) saveGenesisCollections(ctx context.Context) error {
	for _, col := range s.genesis.GenesisState.Collections {
		exising, err := s.storage.CollectionQ().CollectionByIndexCtx(ctx, []byte(col.Index), false)
		if err != nil {
			return errors.Wrap(err, "failed to get collection from storage", logan.F{
				"index": col.Index,
			})
		}

		if exising != nil {
			continue
		}

		err = s.storage.Transaction(func() error {
			savedCollection, err := s.SaveCollection(ctx, col)
			if err != nil {
				return errors.Wrap(err, "failed to save collection", logan.F{
					"index": col.Index,
				})
			}

			cdataToSave := make([]*tokenmanager.CollectionData, 0, len(col.Data))
			for _, genesisData := range s.genesis.GenesisState.Datas {
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

			err = s.saveCollectionData(ctx, savedCollection.ID, cdataToSave)
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

func (s *TokenmanagerSaver) saveCollectionData(ctx context.Context, collectionID int64, coreCollectionData []*tokenmanager.CollectionData) error {
	now := time.Now().UTC()

	collectionDataBatch := make([]data.CollectionChainMapping, len(coreCollectionData))
	for i, collectionData := range coreCollectionData {
		network := s.chains.Get(collectionData.Index.Chain)
		if network == nil {
			s.log.WithField("core_chain", collectionData.Index.Chain).Warn("chain not supported on horizon")
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

	if err := s.storage.CollectionChainMappingQ().InsertBatchCtx(ctx, collectionDataBatch...); err != nil {
		return errors.Wrap(err, "failed to insert collection data", logan.F{
			"collection": collectionID,
		})
	}

	return nil
}

func (s *TokenmanagerSaver) saveGenesisItems(ctx context.Context) error {
	onChainItems := make(map[int64][]*tokenmanager.OnChainItemIndex)

	for _, item := range s.genesis.GenesisState.Items {
		existing, err := s.storage.ItemQ().ItemByIndexCtx(ctx, []byte(item.Index), false)
		if err != nil {
			return errors.Wrap(err, "failed to get item from storage", logan.F{
				"index": item.Index,
			})
		}

		if existing != nil {
			continue
		}

		dataItem, err := s.SaveItem(ctx, item)
		if err != nil {
			return errors.Wrap(err, "failed to save genesis item", logan.F{
				"index": item.Index,
			})
		}

		onChainItems[dataItem.ID] = item.OnChain
	}

	for itemID, onChain := range onChainItems {
		if err := s.saveOnChainItems(ctx, itemID, onChain); err != nil {
			return errors.Wrap(err, "failed to save genesis item on-chain data")
		}
	}

	return nil
}

func (s *TokenmanagerSaver) SaveItem(ctx context.Context, coreItem tokenmanager.Item) (*data.Item, error) {
	collection, err := s.storage.CollectionQ().CollectionByIndexCtx(ctx, []byte(coreItem.Collection), false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get collection from storage", logan.F{
			"index": coreItem.Collection,
		})
	}

	if collection == nil {
		coreCollection, err := s.tokenmanager.GetCollection(ctx, coreItem.Collection)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get collection")
		}

		collection, err = s.SaveCollection(ctx, *coreCollection)
		if err != nil {
			return nil, errors.Wrap(err, "failed to save collection", logan.F{
				"index": coreItem.Collection,
			})
		}
	}

	now := time.Now().UTC()
	item := data.Item{
		Index: []byte(coreItem.Index),
		Collection: sql.NullInt64{
			Int64: collection.ID,
			Valid: true,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	item.Metadata, err = json.Marshal(coreItem.Meta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal item metadata", logan.F{
			"index": coreItem.Index,
		})
	}

	if err := s.storage.ItemQ().InsertCtx(ctx, &item); err != nil {
		return nil, errors.Wrap(err, "failed to insert item", logan.F{
			"index": item.Index,
		})
	}

	return &item, nil
}

func (s *TokenmanagerSaver) saveOnChainItems(ctx context.Context, itemID int64, onChain []*tokenmanager.OnChainItemIndex) error {
	now := time.Now().UTC()
	onChainItemBatch := make([]data.ItemChainMapping, len(onChain))
	for i, onChainItem := range onChain {
		network := s.chains.Get(onChainItem.Chain)
		if network == nil {
			s.log.WithField("core_chain", onChainItem.Chain).Warn("chain not supported on horizon")
			continue
		}

		onChainItemBatch[i] = data.ItemChainMapping{
			Item:      itemID,
			Network:   network.ID,
			Address:   []byte(onChainItem.Address),
			TokenID:   []byte(onChainItem.TokenID),
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	return s.storage.ItemChainMappingQ().InsertBatchCtx(ctx, onChainItemBatch...)
}
