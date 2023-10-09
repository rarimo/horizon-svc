package solana

import (
	"context"
	"encoding/json"
	"fmt"
	bin "github.com/gagliardetto/binary"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/cursorer"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	msgs2 "github.com/rarimo/horizon-svc/pkg/msgs"
	"github.com/rarimo/solana-program-go/contracts/bridge"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const (
	DataInstructionCodeIndex = 0
)

type solanaProducer struct {
	log       *logan.Entry
	chain     data.Chain
	cli       *rpc.Client
	cursorer  types.Cursorer
	publisher services.QPublisher
	programId solana.PublicKey
}

func New(cfg *config.BridgeProducerChainConfig, log *logan.Entry, chain data.Chain, kv *redis.KeyValueProvider, publisher services.QPublisher, cursorKey string) types.Producer {
	f := logan.F{
		"chain": chain.Name,
		"rpc":   chain.Rpc,
	}

	cli := rpc.New(chain.Rpc)
	programId := solana.MustPublicKeyFromBase58(chain.BridgeContract)

	initialCursor := producers.DefaultInitialCursor
	if cfg != nil && cfg.SkipCatchup {
		signatures, err := cli.GetSignaturesForAddress(context.Background(), programId)
		if err != nil {
			panic(errors.Wrap(err, "failed to get last signatures", f))
		}

		initialCursor = signatures[len(signatures)-1].Signature.String()
	}

	return &solanaProducer{
		log,
		chain,
		cli,
		cursorer.NewCursorer(log, kv, cursorKey, initialCursor),
		publisher,
		programId,
	}
}

func (p *solanaProducer) Run(ctx context.Context) error {
	start, err := p.cursorer.GetStartCursor(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get start cursor")
	}

	p.log.Info(fmt.Sprintf("Catchupping history from %s", start.Value))

	for {
		if err = ctx.Err(); err != nil {
			return errors.Wrap(err, "died by context")
		}

		signatures, err := p.cli.GetSignaturesForAddressWithOpts(ctx, p.programId, &rpc.GetSignaturesForAddressOpts{
			Before: start.Signature(),
		})
		if err != nil {
			return errors.Wrap(err, "failed to get signatures")
		}

		for _, sig := range signatures {
			if err = p.processTransaction(ctx, sig.Signature); err != nil {
				return errors.Wrap(err, "failed to process transaction")
			}
		}

		if len(signatures) == 0 {
			return nil
		}

		if err = p.cursorer.SetStartCursor(ctx, start.SetSignature(signatures[len(signatures)-1].Signature)); err != nil {
			return errors.Wrap(err, "failed to set start cursor")
		}
	}
}

func (p *solanaProducer) processTransaction(ctx context.Context, sig solana.Signature) error {
	out, err := p.cli.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Encoding: solana.EncodingBase64,
	})
	if err != nil {
		return errors.Wrap(err, "error getting transaction from solana")
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
	if err != nil {
		return errors.Wrap(err, "error decoding transaction")
	}

	result, err := json.Marshal(tx)
	if err != nil {
		return errors.Wrap(err, "error marshaling transaction")
	}

	block, err := p.cli.GetBlockWithOpts(ctx, out.Slot, &rpc.GetBlockOpts{
		TransactionDetails: rpc.TransactionDetailsNone,
		Commitment:         rpc.CommitmentFinalized,
	})
	if err != nil {
		return errors.Wrap(err, "error getting block")
	}

	height := int64(*block.BlockHeight)
	accounts := tx.Message.AccountKeys
	msgs := make([]msgs2.Message, 0)

	for _, instruction := range tx.Message.Instructions {
		if accounts[instruction.ProgramIDIndex] == p.programId {
			switch bridge.Instruction(instruction.Data[DataInstructionCodeIndex]) {
			case bridge.InstructionWithdrawNative, bridge.InstructionWithdrawFT, bridge.InstructionWithdrawNFT:
				msgs = append(msgs, msgs2.WithdrawalMsg{
					Hash:        data.FormatWithdrawalID(p.chain.Name, sig.String()),
					BlockHeight: height,
					TxResult:    result,
					Success:     out.Meta.Err == nil,
				}.Message())
			default:
				continue
			}
		}
	}

	if err = p.publisher.PublishMsgs(ctx, msgs...); err != nil {
		return errors.Wrap(err, "error publishing messages")
	}

	return nil

}
