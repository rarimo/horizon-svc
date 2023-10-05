package pg

import (
	"context"

	"github.com/Masterminds/squirrel"

	"github.com/rarimo/horizon-svc/internal/data"
)

func (q RejectionQ) InsertBatchCtx(ctx context.Context, rejections ...data.Rejection) error {
	stmt := squirrel.Insert("public.rejections").
		Columns("transfer_index", "rarimo_transaction", "created_at")

	for _, rejection := range rejections {
		stmt = stmt.
			Values(rejection.TransferIndex, rejection.RarimoTransaction, rejection.CreatedAt)
	}

	return q.db.ExecContext(ctx, stmt)
}
