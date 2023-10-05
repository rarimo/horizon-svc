package pg

import (
	"context"

	"github.com/Masterminds/squirrel"

	"github.com/rarimo/horizon-svc/internal/data"
)

func (q TransactionQ) InsertBatchCtx(ctx context.Context, transfers ...data.Transaction) error {
	stmt := squirrel.Insert("public.transactions").
		Columns("hash", "block_height", "index", "raw_tx", "tx_result", "tx_timestamp", "created_at")

	for _, transfer := range transfers {
		stmt = stmt.Values(
			transfer.Hash,
			transfer.BlockHeight,
			transfer.Index,
			transfer.RawTx,
			transfer.TxResult,
			transfer.TxTimestamp,
			transfer.CreatedAt)
	}

	stmt = stmt.Suffix("ON CONFLICT(hash) DO NOTHING") // transactions are immutable

	return q.db.ExecContext(ctx, stmt)
}
