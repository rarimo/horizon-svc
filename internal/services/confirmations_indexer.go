package services

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/core"
	"time"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	rarimocore "github.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/distributed_lab/logan/v3"
)

func RunConfirmationsIndexer(ctx context.Context, cfg config.Config) {
	cindexer := &confirmationsIndexer{
		log:        cfg.Log().WithField("who", cfg.ConfirmationsIndexer().RunnerName),
		rarimocore: cfg.Core().Rarimocore(),
		storage:    cfg.CachedStorage().Clone(),
	}

	msgs.NewConsumer(
		cfg.Log(),
		cfg.ConfirmationsIndexer().ConfirmationsConsumer,
		cindexer,
	).Run(ctx)
}

type confirmationsIndexer struct {
	log        *logan.Entry
	rarimocore core.Rarimocore
	storage    data.Storage
}

func (p *confirmationsIndexer) Handle(ctx context.Context, msgs []msgs.Message) error {
	confirmations := make([]data.Confirmation, 0, 10*len(msgs))
	confirmedTransferIDs := make([]string, 0, 10*len(msgs))

	for _, msg := range msgs {
		cmsg := msg.MustConfirmationOpMessage()

		confirmation, err := p.rarimocore.GetConfirmation(ctx, cmsg.ConfirmationID)
		if err != nil {
			return errors.Wrap(err, "failed to get confirmation")
		}

		for _, transferIndex := range confirmation.Indexes {
			confirmedTransferIDs = append(confirmedTransferIDs, transferIndex)
			confirmations = append(confirmations, data.Confirmation{
				TransferIndex:     []byte(transferIndex),
				RarimoTransaction: data.MustDBHash(cmsg.TransactionHash),
				CreatedAt:         time.Now().UTC(),
			})
		}
	}

	return p.storage.Transaction(func() error {
		err := p.storage.TransferQ().SetStatusByIndexCtx(ctx, int(rarimocore.OpStatus_SIGNED), confirmedTransferIDs...)
		if err != nil {
			return errors.Wrap(err, "failed to set status for confirmed transfers", logan.F{
				"transfer_ids": confirmedTransferIDs,
			})
		}

		err = p.storage.ConfirmationQ().InsertBatchCtx(ctx, confirmations...)
		return errors.Wrap(err, "failed to insert confirmations")
	})
}
