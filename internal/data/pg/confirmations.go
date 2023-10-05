package pg

import (
	"context"

	"github.com/Masterminds/squirrel"

	"github.com/rarimo/horizon-svc/internal/data"
)

func (q ConfirmationQ) InsertBatchCtx(ctx context.Context, confirmations ...data.Confirmation) error {
	stmt := squirrel.Insert("public.confirmations").
		Columns("transfer_index", "rarimo_transaction", "created_at")

	for _, confirmation := range confirmations {
		stmt = stmt.
			Values(confirmation.TransferIndex, confirmation.RarimoTransaction, confirmation.CreatedAt)
	}

	return q.db.ExecContext(ctx, stmt)
}
