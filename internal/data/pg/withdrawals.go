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
			withdrawal.Origin,
			withdrawal.Hash,
			withdrawal.Success,
			withdrawal.CreatedAt,
		)
	}

	stmt = stmt.Suffix("ON CONFLICT(origin) DO NOTHING") // withdrawals are immutable

	return q.db.ExecContext(ctx, stmt)
}
