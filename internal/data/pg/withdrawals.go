package pg

import (
	"context"

	"github.com/Masterminds/squirrel"

	"github.com/rarimo/horizon-svc/internal/data"
)

func (q WithdrawalQ) InsertBatchCtx(ctx context.Context, withdrawals ...data.Withdrawal) error {
	stmt := squirrel.Insert("public.withdrawals").Columns(colsWithdrawal)

	for _, withdrawal := range withdrawals {
		stmt = stmt.Values(
			withdrawal.Hash,
			withdrawal.BlockHeight,
			withdrawal.TxResult,
			withdrawal.Success,
			withdrawal.CreatedAt,
		)
	}

	stmt = stmt.Suffix("ON CONFLICT(hash) DO NOTHING") // transactions are immutable

	return q.db.ExecContext(ctx, stmt)
}
