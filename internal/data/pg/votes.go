package pg

import (
	"context"

	"github.com/Masterminds/squirrel"

	"github.com/rarimo/horizon-svc/internal/data"
)

func (q VoteQ) InsertBatchCtx(ctx context.Context, votes ...data.Vote) error {
	stmt := squirrel.Insert("public.votes").
		Columns("transfer_index", "choice", "rarimo_transaction", "created_at")

	for _, vote := range votes {
		stmt = stmt.Values(
			vote.TransferIndex, vote.Choice,
			vote.RarimoTransaction, vote.CreatedAt)
	}

	return q.db.ExecContext(ctx, stmt)
}
