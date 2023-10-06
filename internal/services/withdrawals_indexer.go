package services

import (
	"context"
	"database/sql"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"github.com/rarimo/xo/types/xo"
	"gitlab.com/distributed_lab/logan/v3"
	"time"
)

func RunWithdrawalsIndexer(ctx context.Context, cfg config.Config) {
	windexer := &withdrawalsIndexer{
		log: cfg.Log().WithField("who", cfg.WithdrawalsIndexer().RunnerName),
	}

	msgs.NewConsumer(
		cfg.Log(),
		cfg.WithdrawalsIndexer().WithdrawalsConsumer,
		windexer,
	).Run(ctx)
}

type withdrawalsIndexer struct {
	log     *logan.Entry
	storage data.Storage
}

func (p *withdrawalsIndexer) Handle(ctx context.Context, msgs []msgs.Message) error {
	withdrawals := make([]data.Withdrawal, 0, len(msgs))

	for _, msg := range msgs {
		wmsg := msg.MustWithdrawalMessage()
		withdrawals = append(withdrawals, data.Withdrawal{
			Hash: wmsg.Hash.Bytes(),
			BlockHeight: sql.NullInt64{
				Int64: wmsg.BlockHeight,
				Valid: true,
			},
			TxResult: xo.NullJsonb{
				Jsonb: wmsg.TxResult,
				Valid: true,
			},
			Success: sql.NullBool{
				Bool:  wmsg.Success,
				Valid: true,
			},
			CreatedAt: time.Now().UTC(),
		})
	}

	return p.storage.WithdrawalQ().InsertBatchCtx(ctx, withdrawals...)
}
