package services

import (
	"context"
	"time"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	rarimocore "github.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/distributed_lab/logan/v3"
)

func RunApprovalIndexer(ctx context.Context, cfg config.Config) {
	aindexer := &approvalIndexer{
		log:     cfg.Log().WithField("who", cfg.ApprovalsIndexer().RunnerName),
		storage: cfg.CachedStorage().Clone(),
	}

	msgs.NewConsumer(
		cfg.Log(),
		cfg.ApprovalsIndexer().ApprovalsConsumer,
		aindexer,
	).Run(ctx)
}

type approvalIndexer struct {
	log     *logan.Entry
	storage data.Storage
}

func (p *approvalIndexer) Handle(ctx context.Context, msgs []msgs.Message) error {
	approvals := make([]data.Approval, 0, len(msgs))
	approvedTransferIndices := make([]string, 0, len(msgs))

	for _, msg := range msgs {
		amsg := msg.MustApprovalOpMessage()
		approvedTransferIndices = append(approvedTransferIndices, amsg.OperationID)
		approvals = append(approvals, data.Approval{
			TransferIndex:     []byte(amsg.OperationID),
			RarimoTransaction: data.MustDBHash(amsg.TransactionHash),
			CreatedAt:         time.Now().UTC(),
		})
	}

	return p.storage.Transaction(func() error {
		err := p.storage.
			TransferQ().
			SetStatusByIndexCtx(ctx, int(rarimocore.OpStatus_APPROVED), approvedTransferIndices...)
		if err != nil {
			return errors.Wrap(err, "failed to set status by index", logan.F{
				"status":  int(rarimocore.OpStatus_APPROVED),
				"indices": approvedTransferIndices,
			})
		}

		return p.storage.ApprovalQ().InsertBatchCtx(ctx, approvals...)
	})
}
