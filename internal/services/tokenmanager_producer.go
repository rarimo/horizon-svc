package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	abcitypes "github.com/tendermint/tendermint/abci/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/rarimo/horizon-svc/internal/data/redis"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"

	tokentypes "github.com/rarimo/rarimo-core/x/tokenmanager/types"

	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func RunTokenManagerEventsProducer(ctx context.Context, cfg config.Config) {
	log := cfg.Log().WithField("who", cfg.TokenManagerProducer().RunnerName)

	itemEventsPublisher, err := msgs.NewPublisher(cfg.Log(), cfg.RedisClient(),
		cfg.TokenManagerProducer().RunnerName+"_item_events_publisher",
		cfg.TokenManagerProducer().ItemsQueueName)
	if err != nil {
		panic(errors.Wrap(err, "failed to create item events publisher"))
	}

	collectionEventsPublisher, err := msgs.NewPublisher(cfg.Log(), cfg.RedisClient(),
		cfg.TokenManagerProducer().RunnerName+"_collection_events_publisher",
		cfg.TokenManagerProducer().CollectionsQueueName)
	if err != nil {
		panic(errors.Wrap(err, "failed to create collection events publisher"))
	}

	producer := &tokenManagerOpProducer{
		log:                       log,
		thttp:                     cfg.Tendermint(),
		itemEventsPublisher:       itemEventsPublisher,
		collectionEventsPublisher: collectionEventsPublisher,
		kv:                        redis.NewKeyValueProvider(cfg),
		txQ:                       cfg.CachedStorage().TransactionQ(),
	}

	msgs.NewConsumer(
		log,
		cfg.TokenManagerProducer().BlockRangeConsumer,
		producer,
	).Run(ctx)

}

type tokenManagerOpProducer struct {
	log   *logan.Entry
	thttp *http.HTTP

	itemEventsPublisher       QPublisher
	collectionEventsPublisher QPublisher

	kv  data.KeyValueQ
	txQ data.TransactionQ
}

func (p *tokenManagerOpProducer) Handle(ctx context.Context, msgs []msgs.Message) error {
	for _, msg := range msgs {
		brMsg := msg.MustBlockRangeMessage()
		p.log.WithFields(logan.F{
			"start": brMsg.Start,
			"end":   brMsg.End,
		}).Info("received block range message")

		if err := p.produceMsgs(ctx, brMsg); err != nil {
			return errors.Wrap(err, "failed to produce op msgs", logan.F{
				"start": brMsg.Start,
				"end":   brMsg.End,
			})
		}
	}

	return nil
}

func (p *tokenManagerOpProducer) produceMsgs(ctx context.Context, blockRange msgs.BlockRangeMessage) error {
	if err := p.publishBlockEvents(ctx, blockRange.Start, blockRange.End); err != nil {
		return errors.Wrap(err, "failed to publish block events", logan.F{
			"start": blockRange.Start,
			"end":   blockRange.End,
		})
	}

	if err := p.publishTxEvents(ctx, blockRange.Start, blockRange.End); err != nil {
		return errors.Wrap(err, "failed to publish tx events", logan.F{
			"start": blockRange.Start,
			"end":   blockRange.End,
		})
	}

	return nil
}

func (p *tokenManagerOpProducer) publishBlockEvents(ctx context.Context, blockFrom, blockTo int64) error {
	cursorKey := tokenmanagerBlockRangeBlocksCursor(blockFrom, blockTo)

	cursorKV, err := p.kv.Get(ctx, cursorKey)
	if err != nil {
		return errors.Wrap(err, "failed to get block range blocks cursor", logan.F{
			"key": cursorKey,
		})
	}

	if cursorKV == nil {
		now := time.Now().UTC()
		cursorKV = &data.KeyValue{
			Key:       cursorKey,
			Value:     strconv.FormatInt(blockFrom, 10),
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	block, err := strconv.ParseInt(cursorKV.Value, 10, 0)
	if err != nil {
		return errors.Wrap(err, "failed to parse block range blocks cursor value", logan.F{
			"key":   cursorKey,
			"value": cursorKV.Value,
		})
	}

	for {
		if block != blockFrom && block >= blockTo {
			return nil
		}

		p.log.WithFields(logan.F{
			"current_block": block,
			"from":          blockFrom,
			"to":            blockTo,
		}).Debug("fetching block results")

		blockResults, err := p.thttp.BlockResults(ctx, &block)
		if err != nil {
			return errors.Wrap(err, "failed to get block results", logan.F{
				"block": block,
			})
		}

		blockEventsSet := extractBlockEndEvents(blockResults)

		f := logan.F{
			"block": block,
		}

		p.log.WithFields(f).Info("publishing messages (blocks)")

		p.log.WithFields(f.Merge(logan.F{
			"events": len(blockEventsSet.itemEvents),
		})).Debug("publishing item events (blocks)")

		if err := p.itemEventsPublisher.PublishMsgs(ctx, blockEventsSet.itemEvents...); err != nil {
			return errors.Wrap(err, "failed to publish item events (blocks)", f)
		}

		p.log.WithFields(f.Merge(logan.F{
			"events": len(blockEventsSet.collectionEvents),
		})).Debug("publishing collection events (blocks)")

		if err := p.collectionEventsPublisher.PublishMsgs(ctx, blockEventsSet.collectionEvents...); err != nil {
			return errors.Wrap(err, "failed to publish collection events (blocks)", f)
		}

		p.log.WithFields(f).Info("published messages (blocks)")

		block += 1
		err = p.kv.Upsert(ctx, data.KeyValue{
			Key:       cursorKV.Key,
			Value:     strconv.FormatInt(block, 10),
			CreatedAt: cursorKV.CreatedAt,
			UpdatedAt: time.Now().UTC(),
		})
		if err != nil {
			return errors.Wrap(err, "failed to update block range cursor", logan.F{
				"key":   cursorKV.Key,
				"value": block,
			})
		}

		// do we save block ?
	}
}

func (p *tokenManagerOpProducer) publishTxEvents(ctx context.Context, blockFrom, blockTo int64) error {
	cursorKey := tokenmanagerBlockRangeTxsCursor(blockFrom, blockTo)

	blockRangeCursor, err := p.kv.Get(ctx, cursorKey)
	if err != nil {
		return errors.Wrap(err, "failed to get block range txs cursor", logan.F{
			"key": cursorKey,
		})
	}

	if blockRangeCursor == nil {
		now := time.Now().UTC()
		blockRangeCursor = &data.KeyValue{
			Key:       cursorKey,
			Value:     "1",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	page64, err := strconv.ParseInt(blockRangeCursor.Value, 10, 0)
	if err != nil {
		return errors.Wrap(err, "failed to parse block range txs cursor value", logan.F{
			"key":   cursorKey,
			"value": blockRangeCursor.Value,
		})
	}

	page := int(page64)

	for {
		p.log.WithFields(logan.F{
			"page": page,
			"from": blockFrom,
			"to":   blockTo,
		}).Debug("fetching page of transactions")

		txs, err := getTransactionsPage(ctx, p.thttp, blockFrom, blockTo, page)
		if err != nil {
			return errors.Wrap(err, "failed to get transactions page", logan.F{
				"start": blockFrom,
				"end":   blockTo,
				"page":  page,
			})
		}

		if len(txs) == 0 {
			if err := p.kv.Remove(ctx, cursorKey); err != nil {
				p.log.WithField("cursor", cursorKey).Warn("failed to remove cursor after finishing block")
			}

			p.log.WithFields(logan.F{
				"start": blockFrom,
				"end":   blockTo,
			}).Info("no txs left, finished processing block range")

			return nil
		}

		p.log.WithFields(logan.F{
			"page": page,
			"from": blockFrom,
			"to":   blockTo,
			"txs":  len(txs),
		}).Debug("got transactions, publishing tx events")

		txsEventsSet := extractTxEvents(p.log, txs)

		f := logan.F{
			"start": blockFrom,
			"end":   blockTo,
		}

		p.log.WithFields(f).Info("publishing messages (txs)")

		p.log.WithFields(f.Merge(logan.F{
			"events": len(txsEventsSet.itemEvents),
		})).Debug("publishing item events (txs)")

		if err := p.itemEventsPublisher.PublishMsgs(ctx, txsEventsSet.itemEvents...); err != nil {
			return errors.Wrap(err, "failed to publish item events", f)
		}

		p.log.WithFields(f.Merge(logan.F{
			"events": len(txsEventsSet.collectionEvents),
		})).Debug("publishing collection events (txs)")

		if err := p.collectionEventsPublisher.PublishMsgs(ctx, txsEventsSet.collectionEvents...); err != nil {
			return errors.Wrap(err, "failed to publish collection events", f)
		}

		p.log.WithFields(f).Info("published messages (txs)")

		page += 1
		err = p.kv.Upsert(ctx, data.KeyValue{
			Key:       blockRangeCursor.Key,
			Value:     strconv.Itoa(page),
			CreatedAt: blockRangeCursor.CreatedAt,
			UpdatedAt: time.Now().UTC(),
		})
		if err != nil {
			return errors.Wrap(err, "failed to update block range cursor", logan.F{
				"key":   blockRangeCursor.Key,
				"value": page,
			})
		}

		if err := p.txQ.InsertBatchCtx(ctx, mustToData(txs)...); err != nil {
			return errors.Wrap(err, "failed to insert transactions", f)
		}
	}
}

type tokenManagerEvents struct {
	itemEvents       []msgs.Message
	collectionEvents []msgs.Message
}

func extractTxEvents(log *logan.Entry, txs []txWithBlockInfo) tokenManagerEvents {
	eventsSet := tokenManagerEvents{
		itemEvents:       make([]msgs.Message, 0, len(txs)),
		collectionEvents: make([]msgs.Message, 0, len(txs)),
	}

	for _, txInfo := range txs {
		log.WithFields(logan.F{
			"tx":  txInfo.tx.Hash.String(),
			"raw": txInfo.tx.TxResult.Log,
		}).Debug("extracting events from transaction")

		txEventsSet := eventsToMsgs(txInfo.tx.Hash.String(), txInfo.tx.TxResult.Events)

		eventsSet.itemEvents = append(eventsSet.itemEvents, txEventsSet.itemEvents...)
		eventsSet.collectionEvents = append(eventsSet.collectionEvents, txEventsSet.collectionEvents...)
	}

	return eventsSet
}

func extractBlockEndEvents(blockResults *coretypes.ResultBlockResults) tokenManagerEvents {
	return eventsToMsgs(fmt.Sprintf("block:%d", blockResults.Height), blockResults.EndBlockEvents)
}

func eventsToMsgs(txHash string, events []abcitypes.Event) tokenManagerEvents {
	eventsSet := tokenManagerEvents{
		itemEvents:       make([]msgs.Message, 0, len(events)),
		collectionEvents: make([]msgs.Message, 0, len(events)),
	}

	for _, event := range events {
		switch event.Type {
		case tokentypes.EventTypeCollectionCreated:
			eventsSet.collectionEvents = append(eventsSet.collectionEvents,
				mustCollectionCreatedMsg(cosmostypes.Event(event)).Message())
		case tokentypes.EventTypeCollectionRemoved:
			eventsSet.collectionEvents = append(eventsSet.collectionEvents,
				mustCollectionRemovedMsg(cosmostypes.Event(event)).Message())
		case tokentypes.EventTypeCollectionDataCreated:
			eventsSet.collectionEvents = append(eventsSet.collectionEvents,
				mustCollectionDataCreatedMsg(cosmostypes.Event(event)).Message())
		case tokentypes.EventTypeCollectionDataRemoved:
			eventsSet.collectionEvents = append(eventsSet.collectionEvents,
				mustCollectionDataRemovedMsg(cosmostypes.Event(event)).Message())
		case tokentypes.EventTypeCollectionDataUpdated:
			eventsSet.collectionEvents = append(eventsSet.collectionEvents,
				mustCollectionDataUpdatedMsg(cosmostypes.Event(event)).Message())
		case tokentypes.EventTypeItemCreated:
			eventsSet.itemEvents = append(eventsSet.itemEvents,
				mustItemCreatedMsg(cosmostypes.Event(event)).Message())
		case tokentypes.EventTypeItemRemoved:
			eventsSet.itemEvents = append(eventsSet.itemEvents,
				mustItemRemovedMsg(cosmostypes.Event(event)).Message())
		case tokentypes.EventTypeOnChainItemCreated:
			eventsSet.itemEvents = append(eventsSet.itemEvents,
				mustOnChainItemCreatedMsg(cosmostypes.Event(event)).Message())
		case tokentypes.EventTypeOnChainItemRemoved:
			eventsSet.itemEvents = append(eventsSet.itemEvents,
				mustOnChainItemRemovedMsg(cosmostypes.Event(event)).Message())
		case tokentypes.EventTypeSeedCreated:
			eventsSet.itemEvents = append(eventsSet.itemEvents,
				mustSeedCreatedMsg(cosmostypes.Event(event)).Message())
		case tokentypes.EventTypeSeedRemoved:
			eventsSet.itemEvents = append(eventsSet.itemEvents,
				mustSeedRemovedMsg(cosmostypes.Event(event)).Message())
		}
	}

	eventsSet.itemEvents = eventsSet.itemEvents[:]
	eventsSet.collectionEvents = eventsSet.collectionEvents[:]

	return eventsSet
}

func mustCollectionCreatedMsg(event cosmostypes.Event) msgs.CollectionCreatedMessage {
	return msgs.CollectionCreatedMessage{
		Index: mustGetCollectionIndex(event),
	}
}

func mustCollectionRemovedMsg(event cosmostypes.Event) msgs.CollectionRemovedMessage {
	return msgs.CollectionRemovedMessage{
		Index: mustGetCollectionIndex(event),
	}
}

func mustGetCollectionIndex(event cosmostypes.Event) string {
	for _, attr := range event.Attributes {
		if string(attr.Key) == tokentypes.AttributeKeyCollectionIndex {
			return string(attr.Value)
		}
	}

	panic(fmt.Sprintf("failed to find collection index in event attributes, event_type=[%s]", event.Type))
}

func mustCollectionDataCreatedMsg(event cosmostypes.Event) msgs.CollectionDataCreatedMessage {
	var msg msgs.CollectionDataCreatedMessage
	msg.CollectionIndex, msg.Chain = mustGetCollectionDataParams(event)
	return msg
}

func mustCollectionDataRemovedMsg(event cosmostypes.Event) msgs.CollectionDataRemovedMessage {
	var msg msgs.CollectionDataRemovedMessage
	msg.CollectionIndex, msg.Chain = mustGetCollectionDataParams(event)
	return msg
}

func mustCollectionDataUpdatedMsg(event cosmostypes.Event) msgs.CollectionDataUpdatedMessage {
	var msg msgs.CollectionDataUpdatedMessage
	msg.CollectionIndex, msg.Chain = mustGetCollectionDataParams(event)
	return msg
}

func mustGetCollectionDataParams(event cosmostypes.Event) (colIndex string, chain string) {
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case tokentypes.AttributeKeyCollectionIndex:
			colIndex = string(attr.Value)
		case tokentypes.AttributeKeyCollectionDataChain:
			chain = string(attr.Value)
		}
	}

	if colIndex == "" {
		panic(fmt.Sprintf("failed to find collection index in event attributes, event_type=[%s]", event.Type))
	}

	if chain == "" {
		panic(fmt.Sprintf("failed to find collection chain in event attributes, event_type=[%s]", event.Type))
	}

	return colIndex, chain
}

func mustItemCreatedMsg(event cosmostypes.Event) msgs.ItemCreatedMessage {
	return msgs.ItemCreatedMessage{
		Index: mustGetItemIndex(event),
	}
}

func mustItemRemovedMsg(event cosmostypes.Event) msgs.ItemRemovedMessage {
	return msgs.ItemRemovedMessage{
		Index: mustGetItemIndex(event),
	}
}

func mustGetItemIndex(event cosmostypes.Event) string {
	for _, attr := range event.Attributes {
		if string(attr.Key) == tokentypes.AttributeKeyItemIndex {
			return string(attr.Value)
		}
	}

	panic(fmt.Sprintf("failed to find item index in event attributes, event_type=[%s]", event.Type))
}

func mustOnChainItemCreatedMsg(event cosmostypes.Event) msgs.ItemOnChainDataCreatedMessage {
	var msg msgs.ItemOnChainDataCreatedMessage
	msg.ItemIndex, msg.Chain = mustGetOnChainItemParams(event)
	return msg
}

func mustOnChainItemRemovedMsg(event cosmostypes.Event) msgs.ItemOnChainDataRemovedMessage {
	var msg msgs.ItemOnChainDataRemovedMessage
	msg.ItemIndex, msg.Chain = mustGetOnChainItemParams(event)
	return msg
}

func mustGetOnChainItemParams(event cosmostypes.Event) (itemIndex string, chain string) {
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case tokentypes.AttributeKeyItemIndex:
			itemIndex = string(attr.Value)
		case tokentypes.AttributeKeyOnChainItemChain:
			chain = string(attr.Value)
		}
	}

	if itemIndex == "" {
		panic(fmt.Sprintf("failed to find item index in event attributes, event_type=[%s]", event.Type))
	}

	if chain == "" {
		panic(fmt.Sprintf("failed to find item chain in event attributes, event_type=[%s]", event.Type))
	}

	return itemIndex, chain
}

func mustSeedCreatedMsg(event cosmostypes.Event) msgs.SeedCreatedMessage {
	var msg msgs.SeedCreatedMessage
	msg.ItemIndex, msg.Seed = mustGetSeedParams(event)
	return msg
}

func mustSeedRemovedMsg(event cosmostypes.Event) msgs.SeedRemovedMessage {
	var msg msgs.SeedRemovedMessage
	msg.ItemIndex, msg.Seed = mustGetSeedParams(event)
	return msg
}

func mustGetSeedParams(event cosmostypes.Event) (itemIndex string, seed string) {
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case tokentypes.AttributeKeyItemIndex:
			itemIndex = string(attr.Value)
		case tokentypes.AttributeKeySeed:
			seed = string(attr.Value)
		}
	}

	if itemIndex == "" {
		panic(fmt.Sprintf("failed to find item index in event attributes, event_type=[%s]", event.Type))
	}

	if seed == "" {
		panic(fmt.Sprintf("failed to find seed in event attributes, event_type=[%s]", event.Type))
	}

	return itemIndex, seed
}

func tokenmanagerBlockRangeTxsCursor(start, end int64) string {
	return fmt.Sprintf("tokenmanager-producer_block_range_txs_cursor:%d-%d", start, end)
}

func tokenmanagerBlockRangeBlocksCursor(start, end int64) string {
	return fmt.Sprintf("tokenmanager-producer_block_range_blocks_cursor:%d-%d", start, end)
}
