package services

import (
	"context"
	"database/sql"
	"math/big"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	rarimocore "github.com/rarimo/rarimo-core/x/rarimocore/types"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func RunTransfersIndexer(ctx context.Context, cfg config.Config) {
	tindexer := &transfersIndexer{
		log:          cfg.Log().WithField("who", cfg.TransfersIndexer().RunnerName),
		rarimocore:   rarimocore.NewQueryClient(cfg.Cosmos()),
		tokenmanager: tokenmanager.NewQueryClient(cfg.Cosmos()),
		storage:      cfg.CachedStorage().Clone(),
	}

	msgs.NewConsumer(
		cfg.Log(),
		cfg.TransfersIndexer().TransfersConsumer,
		tindexer,
	).Run(ctx)
}

type transfersIndexer struct {
	log          *logan.Entry
	rarimocore   rarimocore.QueryClient
	tokenmanager tokenmanager.QueryClient

	storage data.Storage
}

func (p *transfersIndexer) Handle(ctx context.Context, msgs []msgs.Message) error {
	transfers := make([]data.Transfer, len(msgs))

	p.log.WithField("messages", len(msgs)).Debug("starting handling messages")

	for i, msg := range msgs {
		tmsg := msg.MustTransferOpMessage()

		p.log.WithFields(logan.F{
			"transfer_id": tmsg.TransferID,
			"tx_hash":     tmsg.TransactionHash,
		}).Debug("handling transfer")

		resp, err := p.rarimocore.Operation(ctx, &rarimocore.QueryGetOperationRequest{
			Index: tmsg.TransferID,
		})
		if err != nil {
			return errors.Wrap(err, "failed to get operation", logan.F{
				"index": tmsg.TransferID,
			})
		}

		p.log.WithFields(logan.F{
			"index": resp.Operation.Index,
			"type":  resp.Operation.OperationType.String(),
		}).Debug("got operation")

		var operationDetails rarimocore.Transfer
		if err := proto.Unmarshal(resp.Operation.Details.Value, &operationDetails); err != nil {
			return errors.Wrap(err, "failed to unmarshal operation details", logan.F{
				"index": resp.Operation.Index,
			})
		}

		onChainItemResp, err := p.tokenmanager.OnChainItem(ctx, &tokenmanager.QueryGetOnChainItemRequest{
			Chain:   operationDetails.From.Chain,
			Address: operationDetails.From.Address,
			TokenID: operationDetails.From.TokenID,
		})

		if err != nil {
			return errors.Wrap(err, "failed to get item", logan.F{
				"chain":    operationDetails.From.Chain,
				"address":  operationDetails.From.Address,
				"token_id": operationDetails.From.TokenID,
			})
		}

		transferData, err := p.makeTransfer(ctx, tmsg.TransactionHash, resp.Operation, operationDetails, onChainItemResp.Item.Item)
		if err != nil {
			return errors.Wrap(err, "failed to make transfer", logan.F{
				"index": resp.Operation.Index,
			})
		}

		p.log.WithFields(logan.F{
			"transfer": string(transferData.Index),
			"status":   transferData.Status,
			"item":     transferData.TokenIndex,
		}).Debug("made transfer")

		transfers[i] = *transferData
	}

	return p.storage.TransferQ().UpsertBatchCtx(ctx, transfers...)
}

func (p *transfersIndexer) getTransferStatus(ctx context.Context, transfer rarimocore.Operation) (rarimocore.OpStatus, error) {
	// FIXME
	//  slight race condition may occur here, should be fixed on database level
	//  by inserting with status based on a content of confirmations/approvals/rejections tables

	confirmations, err := p.storage.ConfirmationQ().ConfirmationsByTransferIndexCtx(ctx, []byte(transfer.Index), false)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get confirmations", logan.F{
			"transfer_index": transfer.Index,
		})
	}

	if len(confirmations) != 0 {
		return rarimocore.OpStatus_SIGNED, nil
	}

	approvals, err := p.storage.ApprovalQ().ApprovalsByTransferIndexCtx(ctx, []byte(transfer.Index), false)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get approvals", logan.F{
			"transfer_index": transfer.Index,
		})
	}

	if len(approvals) != 0 {
		return rarimocore.OpStatus_APPROVED, nil
	}

	rejections, err := p.storage.RejectionQ().RejectionsByTransferIndexCtx(ctx, []byte(transfer.Index), false)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get rejections", logan.F{
			"transfer_index": transfer.Index,
		})
	}

	if len(rejections) != 0 {
		return rarimocore.OpStatus_NOT_APPROVED, nil
	}

	return transfer.Status, nil
}

func (p *transfersIndexer) makeTransfer(ctx context.Context,
	txHash string, operation rarimocore.Operation,
	operationDetails rarimocore.Transfer, itemIndex string,
) (*data.Transfer, error) {
	transferStatus, err := p.getTransferStatus(ctx, operation)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transfer status", logan.F{
			"transfer_index": operation.Index,
		})
	}

	eventID, err := strconv.ParseInt(operationDetails.EventId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse event_id", logan.F{
			"raw": operationDetails.EventId,
		})
	}

	amount, ok := big.NewInt(0).SetString(operationDetails.Amount, 10)
	if !ok {
		return nil, errors.From(errors.New("failed to parse amount"), logan.F{
			"raw": operationDetails.Amount,
		})
	}

	now := time.Now().UTC()
	return &data.Transfer{
		Index:             []byte(operation.Index),
		Status:            int(transferStatus),
		CreatedAt:         now,
		UpdatedAt:         now,
		RarimoTxTimestamp: time.Unix(int64(operation.Timestamp), 0),
		Creator: sql.NullString{
			String: operation.Creator,
			Valid:  true,
		},
		RarimoTx:   data.MustDBHash(txHash),
		Origin:     operationDetails.Origin,
		Tx:         []byte(operationDetails.Tx), // TODO what tx is this ? make consistent form (hex?/base64?)
		EventID:    eventID,
		FromChain:  operationDetails.From.Chain,
		ToChain:    operationDetails.To.Chain,
		Receiver:   operationDetails.Receiver,
		Amount:     data.Int256{amount},
		BundleData: []byte(operationDetails.BundleData),
		BundleSalt: []byte(operationDetails.BundleSalt),
		TokenIndex: itemIndex,
	}, nil
}
