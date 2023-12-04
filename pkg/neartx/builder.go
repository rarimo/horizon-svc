package neartx

import (
	"context"
	"fmt"
	"github.com/rarimo/near-go/nearclient"
	"gitlab.com/distributed_lab/logan/v3"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/rarimo/horizon-svc/internal/data"

	"github.com/rarimo/horizon-svc/resources"
	"github.com/rarimo/near-go/common"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

//go:generate mockery --name NearInfoer --case underscore --inpackage
type NearInfoer interface {
	AccessKeyView(context.Context, common.AccountID, common.Base58PublicKey, nearclient.BlockCharacteristic) (common.AccessKeyView, error)
	BlockDetails(context.Context, nearclient.BlockCharacteristic) (common.BlockView, error)
}

type Builder struct {
	nearInfo   NearInfoer
	bridgeAddr common.AccountID
}

func NewBuilder(nearInfo NearInfoer, bridgeAddr common.AccountID) *Builder {
	return &Builder{nearInfo: nearInfo, bridgeAddr: bridgeAddr}
}

func (b *Builder) BuildTx(ctx context.Context, req *resources.BuildTx, rawTxData interface{}) (*resources.UnsubmittedTx, error) {
	txData, ok := rawTxData.(resources.NearTxData)
	if !ok {
		return nil, errors.From(errors.New("invalid tx_data"), logan.F{
			"expected": "resources.NearTxData",
			"actual":   fmt.Sprintf("%T", rawTxData),
		})
	}

	_, creatorAccount, err := data.DecodeAccountID(data.AccountID(req.Relationships.CreatorAccount.Data.ID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode creator account id")
	}

	accountID := common.AccountID(creatorAccount)

	_, senderPubKeyBB, err := data.DecodePublicKey(data.PublicKey(txData.SenderPublicKey))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode sender public common")
	}

	senderPK, err := common.PublicKeyFromBytes(senderPubKeyBB)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse sender public common from bytes")
	}

	accessKey, err := b.nearInfo.AccessKeyView(ctx,
		accountID,
		senderPK.ToBase58PublicKey(),
		nearclient.FinalityFinal(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get nonce")
	}

	latestBlock, err := b.nearInfo.BlockDetails(ctx, nearclient.FinalityFinal())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get latest block")
	}

	switch req.Attributes.TxType {
	case resources.TxTypeDepositNative:
		if err := validateNativeDepositTxData(txData); err != nil {
			return nil, err
		}

		amount, err := common.BalanceFromString(*txData.Amount)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse amount", logan.F{
				"raw_amount": txData.Amount,
			})
		}

		actions := []common.Action{
			common.NewNativeDepositCall(common.NativeDepositArgs{
				ReceiverId: txData.Receiver,
				Chain:      txData.TargetNetwork,
			}, common.DefaultFunctionCallGas, amount),
		}

		return makeTx(
			req.Relationships.CreatorAccount.Data.ID,
			b.bridgeAddr,
			senderPK.ToBase58PublicKey(),
			accessKey.Nonce,
			latestBlock.Header.Hash,
			actions,
		)
	case resources.TxTypeDepositFT:
		if err := validateFTDepositTxData(txData); err != nil {
			return nil, err
		}

		amount, err := common.BalanceFromString(*txData.Amount)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse amount", logan.F{
				"raw_amount": txData.Amount,
			})
		}

		msg, err := common.NewTransferArgs(*txData.TokenAddr, req.Relationships.CreatorAccount.Data.ID, txData.Receiver, txData.TargetNetwork, *txData.IsWrapped).String()
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal transfer args")
		}

		actions := []common.Action{
			common.NewFtTransferCall(common.FtTransferArgs{
				ReceiverId: b.bridgeAddr,
				Amount:     amount,
				Msg:        msg,
			}, common.DefaultFunctionCallGas),
		}

		return makeTx(
			req.Relationships.CreatorAccount.Data.ID,
			*txData.TokenAddr, // todo this should be of form {network}:{token_addr}
			senderPK.ToBase58PublicKey(),
			accessKey.Nonce,
			latestBlock.Header.Hash,
			actions,
		)
	case resources.TxTypeDepositNFT:
		if err := validateNFTDepositTxData(txData); err != nil {
			return nil, err
		}

		msg, err := common.NewTransferArgs(*txData.TokenAddr, req.Relationships.CreatorAccount.Data.ID, txData.Receiver, txData.TargetNetwork, *txData.IsWrapped).String()
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal transfer args")
		}

		actions := []common.Action{
			common.NewNftTransferCall(common.NftTransferArgs{
				ReceiverId: b.bridgeAddr,
				TokenID:    *txData.TokenId,
				Msg:        msg,
			}, common.DefaultFunctionCallGas/2),
		}

		return makeTx(
			req.Relationships.CreatorAccount.Data.ID,
			*txData.TokenAddr,
			senderPK.ToBase58PublicKey(),
			accessKey.Nonce,
			latestBlock.Header.Hash,
			actions,
		)
	default:
		return nil, errors.From(errors.New("unsupported tx type"), logan.F{
			"tx_type": req.Attributes.TxType,
			"allowed": resources.SupportedTxTypesNear(),
		})
	}
}

func makeTx(sender, receiver common.AccountID, senderPK common.Base58PublicKey, nonce common.Nonce, blockHash common.Hash, actions []common.Action, opts ...nearclient.TransactionOpt) (*resources.UnsubmittedTx, error) {
	tx := common.Transaction{
		SignerID:   sender,
		PublicKey:  senderPK.ToPublicKey(),
		Nonce:      nonce,
		ReceiverID: receiver,
		BlockHash:  blockHash,
		Actions:    actions,
	}

	txh, serialized, err := tx.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize tx")
	}

	return &resources.UnsubmittedTx{
		Key: resources.Key{
			ID:   txh.String(),
			Type: resources.UNSUBMITTED_TRANSACTIONS,
		},
		Attributes: resources.UnsubmittedTxAttributes{
			Envelope:    hexutil.Encode(serialized),
			GeneratedAt: time.Now().UTC().String(),
		},
	}, nil
}

func validateNativeDepositTxData(data resources.NearTxData) error {
	if data.Amount == nil {
		return errors.New("amount is required")
	}

	return nil
}

func validateFTDepositTxData(data resources.NearTxData) error {
	if data.TokenAddr == nil {
		return errors.New("token address is required")
	}

	if data.Amount == nil {
		return errors.New("amount is required")
	}

	if data.IsWrapped == nil {
		return errors.New("is_wrapped is required")
	}

	return nil
}

func validateNFTDepositTxData(data resources.NearTxData) error {
	if data.TokenAddr == nil {
		return errors.New("token address is required")
	}

	if data.TokenId == nil {
		return errors.New("token id is required")
	}

	if data.IsWrapped == nil {
		return errors.New("is_wrapped is required")
	}

	return nil
}
