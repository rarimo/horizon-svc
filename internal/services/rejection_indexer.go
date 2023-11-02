package services

import (
	"context"
	"time"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/rarimo/horizon-svc/pkg/msgs"

	"github.com/rarimo/horizon-svc/internal/data"
	rarimocore "github.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/rarimo/horizon-svc/internal/config"
)

func RunRejectionIndexer(ctx context.Context, cfg config.Config) {
	rindexer := &rejectionIndexer{
		log:        cfg.Log().WithField("who", cfg.RejectionsIndexer().RunnerName),
		rarimocore: rarimocore.NewQueryClient(cfg.Cosmos()),
		storage:    cfg.CachedStorage().Clone(),
	}

	msgs.NewConsumer(
		cfg.Log(),
		cfg.RejectionsIndexer().RejectionsConsumer,
		rindexer,
	).Run(ctx)
}

type rejectionIndexer struct {
	log        *logan.Entry
	rarimocore rarimocore.QueryClient
	storage    data.Storage
}

func (p *rejectionIndexer) Handle(ctx context.Context, msgs []msgs.Message) error {
	rejectedTransferIndices := make([]string, 0, len(msgs))
	rejections := make([]data.Rejection, 0, len(msgs))

	for _, msg := range msgs {
		rmsg := msg.MustRejectionOpMessage()
		rejectedTransferIndices = append(rejectedTransferIndices, rmsg.OperationID)
		rejections = append(rejections, data.Rejection{
			TransferIndex:     []byte(rmsg.OperationID),
			RarimoTransaction: data.MustDBHash(rmsg.TransactionHash),
			CreatedAt:         time.Now().UTC(),
		})
	}

	return p.storage.Transaction(func() error {
		err := p.storage.
			TransferQ().
			SetStatusByIndexCtx(ctx, int(rarimocore.OpStatus_NOT_APPROVED), rejectedTransferIndices...)
		if err != nil {
			return errors.Wrap(err, "failed to set status by index", logan.F{
				"status":  int(rarimocore.OpStatus_NOT_APPROVED),
				"indices": rejectedTransferIndices,
			})
		}

		err = p.storage.
			RejectionQ().
			InsertBatchCtx(ctx, rejections...)
		if err != nil {
			return errors.Wrap(err, "failed to insert rejections")
		}

		return nil
	})
}
