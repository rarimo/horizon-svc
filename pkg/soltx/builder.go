package soltx

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/olegfomenko/solana-go"
	"gitlab.com/distributed_lab/logan/v3"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/olegfomenko/solana-go/rpc"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/resources"
	"github.com/rarimo/solana-program-go/contracts/bridge"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

//go:generate mockery --name blockHasher --case underscore --inpackage
type blockHasher interface {
	GetRecentBlockhash(context.Context, rpc.CommitmentType) (*rpc.GetRecentBlockhashResult, error)
}

type Builder struct {
	blockhasher blockHasher

	programID     solana.PublicKey
	bridgeAdminPK solana.PublicKey
}

func NewBuilder(blockhasher blockHasher, programID solana.PublicKey, bridgeAdminPK solana.PublicKey) *Builder {
	return &Builder{blockhasher: blockhasher, programID: programID, bridgeAdminPK: bridgeAdminPK}
}

func (b *Builder) BuildTx(ctx context.Context, req *resources.BuildTx, rawTxData interface{}) (*resources.UnsubmittedTx, error) {
	txData, ok := rawTxData.(resources.SolanaTxData)
	if !ok {
		return nil, errors.From(errors.New("invalid tx_data"), logan.F{
			"expected": "resources.SolanaTxData",
			"actual":   fmt.Sprintf("%T", rawTxData),
		})
	}

	blockhash, err := b.blockhasher.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get recent blockhash")
	}

	_, publicKey, err := data.DecodePublicKey(data.PublicKey(req.Relationships.CreatorAccount.Data.ID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode account id", logan.F{
			"raw": req.Relationships.CreatorAccount.Data.ID,
		})
	}

	ownerPK, err := solana.PublicKeyFromBase58(string(publicKey))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse owner public key", logan.F{
			"raw_address": string(publicKey),
		})
	}

	_, receiverAddr, err := data.DecodeAccountID(data.AccountID(txData.Receiver)) // TODO check network supported
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode receiver address", logan.F{
			"raw": txData.Receiver,
		})
	}

	var bundleSeed [32]byte
	copy(bundleSeed[:], txData.BundleSeed[:])
	bundleData := []byte(txData.BundleData)

	switch req.Attributes.TxType {
	case resources.TxTypeDepositNative:
		if txData.Amount == nil {
			return nil, errors.New("amount is required")
		}

		amount, err := strconv.ParseUint(*txData.Amount, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse amount", logan.F{
				"raw": txData.Amount,
			})
		}

		args := bridge.DepositNativeArgs{
			Amount:          amount,
			NetworkTo:       txData.TargetNetwork,
			ReceiverAddress: hex.EncodeToString(receiverAddr),
			BundleData:      &bundleData,
			BundleSeed:      &bundleSeed,
		}

		instruction, err := bridge.DepositNativeInstruction(b.programID, b.bridgeAdminPK, ownerPK, args)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create deposit native instruction")
		}

		return makeTx(b.programID, instruction, blockhash.Value.Blockhash, ownerPK)
	case resources.TxTypeDepositFT:
		if txData.Amount == nil {
			return nil, errors.New("amount is required")
		}

		amount, err := strconv.ParseUint(*txData.Amount, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse amount", logan.F{
				"raw": txData.Amount,
			})
		}

		args := bridge.DepositFTArgs{
			Amount:          amount,
			NetworkTo:       txData.TargetNetwork,
			ReceiverAddress: hex.EncodeToString(receiverAddr),
			BundleData:      &bundleData,
			BundleSeed:      &bundleSeed,
		}

		if txData.TokenAddr == nil {
			return nil, errors.New("token address is required")
		}

		tokenPK, err := solana.PublicKeyFromBase58(*txData.TokenAddr)
		if err != nil {
			panic(err)
		}

		instruction, err := bridge.DepositFTInstruction(b.programID, b.bridgeAdminPK, tokenPK, ownerPK, args)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create deposit FT instruction")
		}

		return makeTx(b.programID, instruction, blockhash.Value.Blockhash, ownerPK)
	case resources.TxTypeDepositNFT:
		args := bridge.DepositNFTArgs{
			NetworkTo:       txData.TargetNetwork,
			ReceiverAddress: hex.EncodeToString(receiverAddr),
			BundleData:      &bundleData,
			BundleSeed:      &bundleSeed,
		}

		if txData.TokenAddr == nil {
			return nil, errors.New("token address is required")
		}

		tokenPK, err := solana.PublicKeyFromBase58(*txData.TokenAddr) // aka mint
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse token public key", logan.F{
				"raw_address": txData.TokenAddr,
			})
		}

		instruction, err := bridge.DepositNFTInstruction(b.programID, b.bridgeAdminPK, tokenPK, ownerPK, args)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create deposit NFT instruction")
		}

		return makeTx(b.programID, instruction, blockhash.Value.Blockhash, ownerPK)
	default:
		return nil, errors.From(errors.New("unsupported tx type"), logan.F{
			"tx_type": req.Attributes.TxType,
			"allowed": resources.SupportedTxTypesSolana(),
		})
	}

}

func makeTx(
	programID solana.PublicKey,
	instruction solana.Instruction,
	blockhash solana.Hash,
	ownerPK solana.PublicKey,
) (*resources.UnsubmittedTx, error) {
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			instruction,
		},
		blockhash,
		solana.TransactionPayer(ownerPK),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transaction")
	}

	envelopeBin, err := tx.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal transaction")
	}

	return &resources.UnsubmittedTx{
		Key: resources.Key{
			ID:   "solana-tx",
			Type: resources.UNSUBMITTED_TRANSACTIONS,
		},
		Attributes: resources.UnsubmittedTxAttributes{
			ContractAddr: programID.String(),
			Envelope:     hexutil.Encode(envelopeBin),
			GeneratedAt:  time.Now().UTC().String(),
		},
	}, nil
}
