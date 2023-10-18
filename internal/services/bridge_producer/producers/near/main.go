package near

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/cursorer"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"github.com/rarimo/near-go/common"
	"github.com/rarimo/near-go/nearprovider"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"strconv"
	"strings"
)

const (
	EventPrefix      = "EVENT_JSON:"
	DefaultBatchSize = uint64(200)
)

type nearProducer struct {
	cfg            *config.BridgeProducerChainConfig
	log            *logan.Entry
	cursorer       types.Cursorer
	publisher      services.QPublisher
	near           nearprovider.Provider
	bridgeContract common.AccountID
}

func New(
	cfg *config.BridgeProducerChainConfig,
	log *logan.Entry,
	kv *redis.KeyValueProvider,
	publisher services.QPublisher,
	near nearprovider.Provider,
	bridgeContract common.AccountID,
	cursorKey string,
) types.Producer {
	initialCursor := producers.DefaultInitialCursor
	if cfg != nil && cfg.SkipCatchup {
		lastBlock, err := near.GetLastKnownBlockHeight(context.Background())
		if err != nil {
			panic(errors.Wrap(err, "failed to get last block"))
		}

		initialCursor = strconv.FormatUint(*lastBlock, 10)
	}

	return &nearProducer{
		cfg,
		log,
		cursorer.NewCursorer(log, kv, cursorKey, initialCursor),
		publisher,
		near,
		string(hexutil.MustDecode(bridgeContract)),
	}
}

func (p *nearProducer) Run(ctx context.Context) error {
	start, err := p.cursorer.GetStartCursor(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get start cursor")
	}

	p.log.Info(fmt.Sprintf("Catchupping history from %s", start.Value))

	batchSize := DefaultBatchSize
	if p.cfg != nil && p.cfg.BatchSize > 0 {
		batchSize = uint64(p.cfg.BatchSize)
	}

	for {
		if err = ctx.Err(); err != nil {
			return errors.Wrap(err, "died by context")
		}

		f := logan.F{
			"cursor": start.Value,
		}

		p.log = p.log.WithFields(f)
		p.log.Info("Starting iteration")

		blocks, err := p.near.ListBlocks(ctx, batchSize, start.Uint())
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

func (p *nearProducer) processBlocks(ctx context.Context, blocks []common.BlockHeight, cursor *types.Cursor) (*types.Cursor, error) {
	fromBlock := cursor

	for _, block := range blocks {
		f := logan.F{
			"height": block,
		}

		msg, err := p.near.GetMessage(ctx, block)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch message", f)
		}

		err = p.processShards(ctx, block, msg.Shards)
		if err != nil {
			return nil, errors.Wrap(err, "failed to extract events from shards", f)
		}

		fromBlock = fromBlock.SetInt64(int64(block) + 1)

		err = p.cursorer.SetStartCursor(ctx, fromBlock)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set cursor")
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
			err := p.processTx(ctx, transactionView, shard.ShardID)
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

func (p *nearProducer) processTx(ctx context.Context, transaction common.ShardChunkTransactionView, shardID common.ShardID) error {
	f := logan.F{
		"shard_id": shardID,
		"tx_hash":  transaction.Transaction.Hash.String(),
		"sender":   transaction.Transaction.SignerID,
	}

	tx, err := p.near.GetTransaction(ctx, transaction.Transaction.Hash, transaction.Transaction.SignerID)
	if err != nil {
		return errors.Wrap(err, "failed to get transaction", f)
	}

	err = p.processReceiptsOutcomes(ctx, tx)
	return errors.Wrap(err, "failed to process receipts", f)
}

func (p *nearProducer) processReceiptsOutcomes(ctx context.Context, tx *common.FinalExecutionOutcomeWithReceiptView) error {
	messages := make([]msgs.Message, 0)

	for _, receiptOutcome := range tx.FinalExecutionOutcomeView.ReceiptsOutcome {
		if receiptOutcome.Outcome.ExecutorID != p.bridgeContract {
			continue
		}

		for _, log := range receiptOutcome.Outcome.Logs {
			event := getEventFromLog(log)
			if event == nil {
				continue
			}

			messages = append(messages, msgs.WithdrawalMsg{
				Origin:  event.Data[0].Origin,
				Hash:    tx.Transaction.Hash.String(),
				Success: tx.Status.Failure == nil,
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
