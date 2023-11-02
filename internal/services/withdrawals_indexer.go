package services

import (
	"context"
	"database/sql"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/logan/v3"
	"time"
)

func RunWithdrawalsIndexer(ctx context.Context, cfg config.Config) {
	windexer := &withdrawalsIndexer{
		log:     cfg.Log().WithField("who", cfg.WithdrawalsIndexer().RunnerName),
		storage: cfg.NewStorage().Clone(),
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
			Origin: hexutil.MustDecode(wmsg.Origin),
			Hash: sql.NullString{
				String: wmsg.Hash,
				Valid:  true,
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
