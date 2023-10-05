package services

import (
	"context"
	"strconv"
	"time"

	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
)

type QPublisher interface {
	PublishMsgs(ctx context.Context, msgs ...msgs.Message) error
}

func RunBlockRangeProducer(ctx context.Context, cfg config.Config) {
	log := cfg.Log().WithField("who", cfg.BlockRangeProducer().RunnerName)

	blockRangePublisher, err := msgs.NewPublisher(cfg.Log(),
		cfg.RedisClient(),
		cfg.BlockRangeProducer().RunnerName+"_qpublisher",
		cfg.BlockRangeProducer().QueueName)
	if err != nil {
		panic(errors.Wrap(err, "failed to create block range publisher"))
	}

	producer := blockRangeProducer{
		log:             log,
		kv:              redis.NewKeyValueProvider(cfg),
		thttp:           cfg.Tendermint(),
		publisher:       blockRangePublisher,
		cursorKey:       cfg.BlockRangeProducer().CursorKey,
		blockRangeLimit: cfg.BlockRangeProducer().BlockRangeLimit,
	}

	if len(cfg.BlockRangeProducer().SpecialCaseBlocks) != 0 {
		err = producer.produceSpecialCase(ctx, cfg.BlockRangeProducer().SpecialCaseBlocks)
		if err != nil {
			panic(errors.Wrap(err, "failed to produce special case blocks"))
		}
	}

	running.WithBackOff(ctx,
		log,
		cfg.BlockRangeProducer().RunnerName,
		producer.produceOnce,
		cfg.BlockRangeProducer().BlockTime,
		time.Second*10,
		time.Minute)
}

type blockRangeProducer struct {
	log       *logan.Entry
	kv        data.KeyValueQ
	thttp     *http.HTTP
	publisher QPublisher

	cursorKey       string
	blockRangeLimit int64
}

func (p *blockRangeProducer) produceOnce(ctx context.Context) error {
	startCursorKV, err := p.kv.Get(ctx, p.cursorKey)
	if err != nil {
		return errors.Wrap(err, "failed to get cursor", logan.F{
			"cursor_name": p.cursorKey,
		})
	}

	if startCursorKV == nil {
		now := time.Now()
		startCursorKV = &data.KeyValue{
			Key:       p.cursorKey,
			Value:     "1",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	start, err := strconv.ParseInt(startCursorKV.Value, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse cursor", logan.F{
			"cursor_name": p.cursorKey,
			"raw":         startCursorKV.Value,
		})
	}

	for {
		if err := ctx.Err(); err != nil {
			return errors.Wrap(err, "died by context")
		}

		end := start + p.blockRangeLimit

		latest, err := p.thttp.Block(ctx, nil)
		if err != nil {
			return errors.Wrap(err, "failed to get latest block")
		}

		if latest == nil || latest.Block == nil {
			// looks like core didn't start up yet
			p.log.Warn("nil latest block fetched from core, waiting")
			return nil
		}

		if end > latest.Block.Height {
			end = latest.Block.Height
		}

		if start > end {
			p.log.Debug("start > end waiting for next block")
			return nil
		}

		p.log.WithFields(logan.F{
			"start": start,
			"end":   end,
		}).Info("Producing block range")

		msg := msgs.BlockRangeMessage{
			Start: start,
			End:   end,
		}

		if err := p.publisher.PublishMsgs(ctx, msg.Message()); err != nil {
			return errors.Wrap(err, "failed to publish block range", logan.F{
				"start": start,
				"end":   end,
			})
		}

		start = end + 1

		err = p.kv.Upsert(ctx, data.KeyValue{
			Key:       startCursorKV.Key,
			Value:     strconv.FormatInt(start, 10),
			CreatedAt: startCursorKV.CreatedAt,
			UpdatedAt: time.Now(),
		})
		if err != nil {
			p.log.WithFields(logan.F{
				"cursor_key":   startCursorKV.Key,
				"cursor_value": start,
			}).Warn("failed to save cursor")
		}
	}
}

func (p *blockRangeProducer) produceSpecialCase(ctx context.Context, blocks []config.SpecialCaseBlockRange) error {
	messages := make([]msgs.Message, len(blocks))
	for i, b := range blocks {
		p.log.WithFields(logan.F{
			"start": b.From,
			"end":   b.To,
		}).Debug("got special case block range")

		m := msgs.BlockRangeMessage{
			Start: b.From,
			End:   b.To,
		}
		messages[i] = m.Message()
	}

	p.log.WithField("sc_blocks", len(messages)).Debug("producing special case block ranges")

	return p.publisher.PublishMsgs(ctx, messages...)
}
