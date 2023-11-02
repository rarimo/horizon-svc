package solana

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	bin "github.com/gagliardetto/binary"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/producers/cursorer"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"github.com/rarimo/rarimo-core/x/rarimocore/crypto/operation/origin"
	"github.com/rarimo/solana-program-go/contracts/bridge"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const (
	DataInstructionCodeIndex = 0
)

type solanaProducer struct {
	log       *logan.Entry
	cursorer  types.Cursorer
	publisher services.QPublisher
	chain     string
	programId solana.PublicKey
	cli       *rpc.Client
}

func New(
	cfg *config.BridgeProducerChainConfig,
	log *logan.Entry,
	kv *redis.KeyValueProvider,
	publisher services.QPublisher,
	chain *data.Chain,
	bridgeContract,
	cursorKey string,
) types.Producer {
	f := logan.F{
		"chain": chain.Name,
		"rpc":   chain.Rpc,
	}

	cli := rpc.New(chain.Rpc)
	programId := solana.PublicKeyFromBytes(hexutil.MustDecode(bridgeContract))

	initialCursor := solana.Signature{}.String()
	if cfg != nil && cfg.SkipCatchup {
		signatures, err := cli.GetSignaturesForAddress(context.Background(), programId)
		if err != nil {
			panic(errors.Wrap(err, "failed to get last signatures", f))
		}

		initialCursor = signatures[len(signatures)-1].Signature.String()
	}

	return &solanaProducer{
		log,
		cursorer.NewCursorer(log, kv, cursorKey, initialCursor),
		publisher,
		chain.Name,
		programId,
		cli,
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

	accounts := tx.Message.AccountKeys
	messages := make([]msgs.Message, 0)

	for i, instruction := range tx.Message.Instructions {
		if accounts[instruction.ProgramIDIndex] == p.programId {
			switch bridge.Instruction(instruction.Data[DataInstructionCodeIndex]) {
			case bridge.InstructionWithdrawNative, bridge.InstructionWithdrawFT, bridge.InstructionWithdrawNFT:
				hash := origin.NewDefaultOriginBuilder().
					SetTxHash(sig.String()).
					SetOpId(fmt.Sprint(i)).
					SetCurrentNetwork(p.chain).
					Build().
					GetOrigin()
				messages = append(messages, msgs.WithdrawalMsg{
					Origin:  hexutil.Encode(hash[:]),
					Hash:    sig.String(),
					Success: out.Meta.Err == nil,
				}.Message())
			default:
				continue
			}
		}
	}

	if err = p.publisher.PublishMsgs(ctx, messages...); err != nil {
		return errors.Wrap(err, "error publishing messages")
	}

	return nil

}
