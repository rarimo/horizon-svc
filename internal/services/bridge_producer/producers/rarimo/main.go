package rarimo

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/cursorer"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"github.com/rarimo/rarimo-core/app"
	"github.com/rarimo/rarimo-core/ethermint/encoding"
	bridge "github.com/rarimo/rarimo-core/x/bridge/types"
	thttp "github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"strconv"
	"strings"
)

const (
	InvalidHeightMessagePrefix = "RPC error -32603 - Internal error: height"
)

type rarimoProducer struct {
	log       *logan.Entry
	cursorer  types.Cursorer
	publisher services.QPublisher
	tm        *thttp.HTTP
	txConfig  client.TxConfig
}

func New(
	cfg *config.BridgeProducerChainConfig,
	log *logan.Entry,
	kv *redis.KeyValueProvider,
	publisher services.QPublisher,
	tm *thttp.HTTP,
	cursorKey string,
) types.Producer {
	initialCursor := producers.DefaultInitialCursor
	if cfg != nil && cfg.SkipCatchup {
		netStatus, err := tm.Status(context.Background())
		if err != nil {
			panic(errors.Wrap(err, "failed to get network status"))
		}

		initialCursor = strconv.FormatInt(netStatus.SyncInfo.LatestBlockHeight, 10)
	}

	return &rarimoProducer{
		log,
		cursorer.NewCursorer(log, kv, cursorKey, initialCursor),
		publisher,
		tm,
		encoding.MakeConfig(app.ModuleBasics).TxConfig,
	}
}

func (p *rarimoProducer) Run(ctx context.Context) error {
	start, err := p.cursorer.GetStartCursor(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get start cursor")
	}

	p.log.Info(fmt.Sprintf("Catchupping history from %s", start.Value))

	for {
		if err = ctx.Err(); err != nil {
			return errors.Wrap(err, "died by context")
		}

		f := logan.F{
			"cursor": start.Value,
		}

		log := p.log.WithFields(f)
		h := start.Int()
		resp, err := p.tm.Block(ctx, &h)
		if err != nil {
			if strings.HasPrefix(err.Error(), InvalidHeightMessagePrefix) {
				log.Debug("invalid height, waiting for the next block")
				return nil
			}

			return errors.Wrap(err, "failed to get txs by block height")
		}

		messages := make([]msgs.Message, 0)

		for _, rawTx := range resp.Block.Txs {
			hash := strings.ToUpper(hex.EncodeToString(rawTx.Hash()))

			fields := f.Merge(logan.F{
				"hash": hash,
			})

			tx, err := p.txConfig.TxDecoder()(rawTx)
			if err != nil {
				return errors.Wrap(err, "failed to decode tx", fields)
			}

			for _, message := range tx.GetMsgs() {
				switch cosmosMsg := message.(type) {

				case *bridge.MsgWithdrawNative:
					messages = append(messages, msgs.WithdrawalMsg{
						Origin:  cosmosMsg.Origin,
						Hash:    hash,
						Success: true,
					}.Message())
				default:
					continue
				}
			}
		}

		err = p.publisher.PublishMsgs(ctx, messages...)
		if err != nil {
			return errors.Wrap(err, "failed to publish messages", f)
		}

		start = start.SetInt64(start.Int() + 1)
		err = p.cursorer.SetStartCursor(ctx, start)
		log.Debug("finished processing block")
	}
}
