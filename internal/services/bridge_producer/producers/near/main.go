package near

import (
	"context"
	"encoding/json"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/utils"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"github.com/rarimo/near-go/common"
	"github.com/rarimo/near-go/nearprovider"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"strconv"
	"strings"
)

var ErrNoShardsAvailable = errors.New("No shards available")

const (
	EventPrefix      = "EVENT_JSON:"
	DefaultBatchSize = uint64(200)
)

type nearProducer struct {
	log       *logan.Entry
	chain     data.Chain
	near      nearprovider.Provider
	cfg       *config.BridgeProducerChainConfig
	publisher services.QPublisher
	cursorer  types.Cursorer
	contract  common.AccountID
}

func New(
	cfg *config.BridgeProducerChainConfig,
	log *logan.Entry,
	chain data.Chain,
	kv *redis.KeyValueProvider,
	publisher services.QPublisher,
	near nearprovider.Provider,
	cursorKey,
	contractAddress string,
) types.Producer {
	f := logan.F{
		"chain": chain.Name,
		"rpc":   chain.Rpc,
	}

	initialCursor := producers.DefaultInitialCursor
	if cfg != nil && cfg.SkipCatchup {
		lastBlock, err := near.GetLastKnownBlockHeight(context.Background())
		if err != nil {
			panic(errors.Wrap(err, "failed to get last block", f))
		}

		initialCursor = strconv.FormatUint(*lastBlock, 10)
	}

	return &nearProducer{
		log.WithField("who", chain.Name+"_near_bridge_events_producer"),
		chain,
		near,
		cfg,
		publisher,
		utils.NewCursorer(log, kv, cursorKey+"_near", initialCursor),
		contractAddress,
	}
}

func (p *nearProducer) Run(ctx context.Context) error {
	start, err := p.cursorer.GetStartCursor(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get start cursor")
	}

	batchSize := DefaultBatchSize
	if p.cfg != nil && p.cfg.BatchSize > 0 {
		batchSize = uint64(p.cfg.BatchSize)
	}

	for {
		if err = ctx.Err(); err != nil {
			return errors.Wrap(err, "died by context")
		}

		f := logan.F{
			"cursor": start,
		}

		p.log = p.log.WithFields(f)
		p.log.Info("Starting iteration")

		blocks, err := p.near.ListBlocks(ctx, batchSize, uint64(start))
		if err != nil {
			return errors.Wrap(err, "failed to fetch blocks", f)
		}

		if len(blocks) == 0 {
			p.log.Info("No blocks to process")
			return nil
		}

		f = f.Merge(logan.F{"batch_size": len(blocks)})
		p.log = p.log.WithFields(f)
		p.log.Debug("Got blocks batch")

		start, err = p.processBlocks(ctx, blocks, start)
		if err != nil {
			return errors.Wrap(err, "iteration failed", f)
		}

		p.log.Infof("Iteration finished")
	}
}

func (p *nearProducer) processBlocks(ctx context.Context, blocks []common.BlockHeight, cursor int64) (int64, error) {
	fromBlock := cursor

	for _, block := range blocks {
		f := logan.F{
			"height": block,
		}

		msg, err := p.near.GetMessage(ctx, block)
		if err != nil {
			return cursor, errors.Wrap(err, "failed to fetch message", f)
		}

		err = p.processShards(ctx, block, msg.Shards)
		if err != nil {
			return cursor, errors.Wrap(err, "failed to extract events from shards", f)
		}

		fromBlock = int64(block) + 1

		err = p.cursorer.SetStartCursor(ctx, fromBlock)
		if err != nil {
			return cursor, errors.Wrap(err, "failed to set cursor")
		}
	}

	return fromBlock, nil
}

func (p *nearProducer) processShards(ctx context.Context, block common.BlockHeight, shards []*common.ShardView) error {
	f := logan.F{"height": block}
	if len(shards) == 0 {
		p.log.WithFields(f).Debug("No shards in block")
		return nil
	}

	for _, shard := range shards {
		if shard.Chunk == nil {
			continue
		}

		for _, transactionView := range shard.Chunk.Transactions {
			err := p.processTx(ctx, transactionView, block, shard.ShardID)
			if err != nil {
				return errors.Wrap(err, "failed to process tx", f.Merge(logan.F{
					"tx_hash":    transactionView.Transaction.Hash.String(),
					"shard_id":   shard.ShardID,
					"chunk_hash": shard.Chunk.Header.ChunkHash.String(),
				}))
			}
		}
	}

	return nil
}

func (p *nearProducer) processTx(ctx context.Context, transaction common.ShardChunkTransactionView, block common.BlockHeight, shardID common.ShardID) error {
	f := logan.F{
		"shard_id": shardID,
		"tx_hash":  transaction.Transaction.Hash.String(),
		"sender":   transaction.Transaction.SignerID,
	}

	tx, err := p.near.GetTransaction(ctx, transaction.Transaction.Hash, transaction.Transaction.SignerID)
	if err != nil {
		return errors.Wrap(err, "failed to get transaction", f)
	}

	err = p.processReceiptsOutcomes(ctx, block, tx)
	return errors.Wrap(err, "failed to process receipts", f)
}

func (p *nearProducer) processReceiptsOutcomes(ctx context.Context, height common.BlockHeight, tx *common.FinalExecutionOutcomeWithReceiptView) error {
	messages := make([]msgs.Message, 0)

	for _, receiptOutcome := range tx.FinalExecutionOutcomeView.ReceiptsOutcome {
		if receiptOutcome.Outcome.ExecutorID != p.contract {
			continue
		}

		for _, log := range receiptOutcome.Outcome.Logs {
			event := getEventFromLog(log)
			if event == nil {
				continue
			}

			txResult, err := json.Marshal(tx)
			if err != nil {
				return errors.Wrap(err, "failed to marshal tx result")
			}

			messages = append(messages, msgs.WithdrawalMsg{
				Hash:        data.FormatWithdrawalID(p.chain.Name, tx.Transaction.Hash.String()),
				BlockHeight: int64(height),
				TxResult:    txResult,
				Success:     tx.Status.Failure == nil,
			}.Message())
		}
	}

	err := p.publisher.PublishMsgs(ctx, messages...)
	if err != nil {
		return errors.Wrap(err, "failed to publish messages")
	}

	return nil
}

func getEventFromLog(log string) *common.BridgeWithdrawnEvent {
	if !strings.HasPrefix(log, EventPrefix) {
		return nil
	}

	eventRaw := strings.TrimPrefix(log, EventPrefix)
	var event common.BridgeWithdrawnEvent

	err := json.Unmarshal([]byte(eventRaw), &event)
	if err != nil {
		// Skipping event if it's not valid NEP-297 event https://nomicon.io/Standards/EventsFormat
		return nil
	}

	if !isWithdrawnEventValid(event) {
		return nil
	}

	return &event
}

func isWithdrawnEventValid(event common.BridgeWithdrawnEvent) bool {
	switch common.BridgeEventType(event.Event) {
	case common.EventTypeNFTWithdrawn, common.EventTypeFTWithdrawn, common.EventTypeNativeWithdrawn:
		return true
	default:
		return false
	}
}
