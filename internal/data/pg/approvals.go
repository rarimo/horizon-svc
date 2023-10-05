package pg

import (
	"context"

	"github.com/Masterminds/squirrel"

	"github.com/rarimo/horizon-svc/internal/data"
)

func (q ApprovalQ) InsertBatchCtx(ctx context.Context, approvals ...data.Approval) error {
	stmt := squirrel.Insert("public.approvals").
		Columns("transfer_index", "rarimo_transaction", "created_at")

	for _, approval := range approvals {
		stmt = stmt.
			Values(approval.TransferIndex, approval.RarimoTransaction, approval.CreatedAt)
	}

	return q.db.ExecContext(ctx, stmt)
}
