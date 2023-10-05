package near

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/utils"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	"github.com/rarimo/horizon-svc/pkg/s3"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/near-bridge-go/pkg/client"
	block "gitlab.com/rarimo/near-bridge-go/pkg/client/block"
	near "gitlab.com/rarimo/near-bridge-go/pkg/types"
	"strconv"
	"strings"
)

var ErrNoShardsAvailable = errors.New("No shards available")

type nearProducer struct {
	log       *logan.Entry
	chain     data.Chain
	kv        *redis.KeyValueProvider
	s3        s3.Connector
	cfg       *config.BridgeProducerChainConfig
	publisher services.QPublisher
	cursorer  types.Cursorer
	contract  near.AccountID
}

func New(
	cfg *config.BridgeProducerChainConfig,
	log *logan.Entry,
	chain data.Chain,
	kv *redis.KeyValueProvider,
	publisher services.QPublisher,
	s3 s3.Connector,
	cursorKey,
	contractAddress string,
) types.Producer {
	f := logan.F{
		"chain": chain.Name,
		"rpc":   chain.Rpc,
	}

	cli, err := client.NewClient(chain.Rpc)
	if err != nil {
		panic(errors.Wrap(err, "failed to connect to near node", f))
	}

	initialCursor := producers.DefaultInitialCursor
	if cfg != nil && cfg.SkipCatchup {
		lastBlock, err := cli.BlockDetails(context.Background(), block.FinalityFinal())
		if err != nil {
			panic(errors.Wrap(err, "failed to get last block", f))
		}

		initialCursor = strconv.FormatUint(lastBlock.Header.Height, 10)
	}

	return &nearProducer{
		log.WithField("who", chain.Name+"_near_bridge_events_producer"),
		chain,
		kv,
		s3,
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

	for {
		if err = ctx.Err(); err != nil {
			return errors.Wrap(err, "died by context")
		}

		f := logan.F{
			"cursor": start,
		}

		p.log.WithFields(f).Info("Starting iteration")

		blocks, err := p.listBlocks(ctx, uint64(p.cfg.BatchSize), uint64(start))
		if err != nil {
			return errors.Wrap(err, "failed to fetch blocks", f)
		}

		if len(blocks) == 0 {
			p.log.WithFields(f).Debug("No blocks to process")
			return nil
		}

		f = f.Merge(logan.F{"batch_size": len(blocks)})

		p.log.WithFields(f).Debug("Got blocks batch")
	}

	return nil
}

// listBlocks - queries the list of the objects in the bucket, grouped by "/" delimiter.
// Returns the list of blocks that can be fetched
// See more about data structure https://github.com/near/near-lake#data-structure
func (p *nearProducer) listBlocks(ctx context.Context, limit uint64, fromBlock near.BlockHeight) ([]types.BlockHeight, error) {
	res, err := p.s3.ListObjects(ctx, int64(limit), normalizeBlockHeight(fromBlock))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get list blocks")
	}

	blocks := make([]near.BlockHeight, 0)

	for _, block := range res.CommonPrefixes {
		if block.Prefix == nil {
			continue
		}

		parsed, err := strconv.ParseUint(strings.TrimSuffix(*block.Prefix, s3.Delimiter), 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse block height into unit64")
		}

		blocks = append(blocks, parsed)
	}

	p.log.WithFields(logan.F{
		"start_from_block": fromBlock,
		"batch_size":       len(blocks),
	}).Debug("Listed blocks")

	return blocks, nil
}

func (p *nearProducer) processBlocks(ctx context.Context, blocks []near.BlockHeight) (bool, error) {
	for _, block := range blocks {
		msg, err := p.getMessage(ctx, block)
		if err != nil {
			return false, errors.Wrap(err, "failed to fetch message", logan.F{
				"block": block,
			})
		}

		err = p.processShards(ctx, block, msg.Shards)
		if err != nil {
			return false, errors.Wrap(err, "failed to extract events from shards")
		}

		p.fromBlock = block + 1
	}

	return false, nil
}

func (p *nearProducer) processShards(ctx context.Context, block near.BlockHeight, shards []*client.ShardView) error {
	if len(shards) == 0 {
		p.log.WithFields(logan.F{"block": block}).Debug("No shards in block")
		return nil
	}

	for _, shard := range shards {
		if shard.Chunk == nil {
			continue
		}

		for _, transactionView := range shard.Chunk.Transactions {
			err := p.processTx(ctx, transactionView, block, shard.ShardID)
			if err != nil {
				return errors.Wrap(err, "failed to process tx")
			}
		}
	}

	return nil
}

func (s *nearProducer) processTx(ctx context.Context, transaction client.ShardChunkTransactionView, block neartypes.BlockHeight, shardID neartypes.ShardID) error {
	f := logan.F{
		"block_height": block,
		"shard_id":     shardID,
		"tx_hash":      transaction.Transaction.Hash.String(),
		"sender":       transaction.Transaction.SignerID,
	}

	tx, err := s.GetTx(ctx, transaction.Transaction.Hash, transaction.Transaction.SignerID)
	if err != nil {
		return errors.Wrap(err, "failed to get transaction", f)
	}

	err = s.processReceiptsOutcomes(ctx, tx.FinalExecutionOutcomeView.ReceiptsOutcome, tx.Transaction.Hash.String())
	return errors.Wrap(err, "failed to process receipts", f)
}

type Message struct {
	Block  *client.BlockView   `json:"block"`
	Shards []*client.ShardView `json:"shards"`
}

// getMessage by the given block height gets the objects:
// - block.json
// - shard_N.json
// Returns the result as `Message` struct
func (p *nearProducer) getMessage(ctx context.Context, blockHeight near.BlockHeight) (*Message, error) {
	block, err := p.getBlock(ctx, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get block")
	}

	shards, err := p.getShards(ctx, blockHeight, uint64(len(block.Chunks)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shards")
	}

	return &Message{
		Block:  block,
		Shards: shards,
	}, nil
}

func (p *nearProducer) getBlock(ctx context.Context, blockHeight near.BlockHeight) (*client.BlockView, error) {
	res, err := p.s3.GetObject(ctx, fmt.Sprintf("%s%sblock.json", normalizeBlockHeight(blockHeight), s3.Delimiter))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get block")
	}

	var block client.BlockView

	err = json.NewDecoder(res.Body).Decode(&block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse block response")
	}

	p.log.WithFields(logan.F{
		"block_height": blockHeight,
	}).Debug("Got block")

	return &block, nil
}

func (p *nearProducer) getShard(ctx context.Context, blockHeight near.BlockHeight, shardID uint64) (*client.ShardView, error) {
	res, err := p.s3.GetObject(ctx, fmt.Sprintf("%s%sshard_%d.json", normalizeBlockHeight(blockHeight), s3.Delimiter, shardID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shard")
	}

	var shard client.ShardView
	err = json.NewDecoder(res.Body).Decode(&shard)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse shard response")
	}

	p.log.WithFields(logan.F{
		"shard_id":     shardID,
		"block_height": blockHeight,
	}).Debug("Got shard")

	return &shard, nil
}

func (p *nearProducer) getShards(ctx context.Context, blockHeight near.BlockHeight, numberOfShards uint64) ([]*client.ShardView, error) {
	if numberOfShards == 0 {
		return nil, ErrNoShardsAvailable
	}

	shards := make([]*client.ShardView, numberOfShards)

	for i := uint64(0); i < numberOfShards; i++ {
		shard, err := p.getShard(ctx, blockHeight, i)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get shard")
		}
		shards[i] = shard
	}

	p.log.WithFields(logan.F{
		"block_height": blockHeight,
		"shards_count": len(shards),
	}).Debug("Got shards")

	return shards, nil
}

func normalizeBlockHeight(blockHeight near.BlockHeight) string {
	return fmt.Sprintf("%012d", blockHeight)
}
