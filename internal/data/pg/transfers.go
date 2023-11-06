package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"gitlab.com/distributed_lab/kit/pgdb"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/Masterminds/squirrel"

	"github.com/rarimo/horizon-svc/internal/data"
)

func (q TransferQ) UpsertBatchCtx(ctx context.Context, transfers ...data.Transfer) error {
	stmt := squirrel.Insert("public.transfers").
		Columns("index", "status",
			"created_at", "updated_at", "creator",
			"rarimo_tx", "rarimo_tx_timestamp", "origin",
			"tx", "event_id", "from_chain",
			"to_chain", "receiver", "amount",
			"bundle_data", "bundle_salt", "item_index")

	for _, transfer := range transfers {
		stmt = stmt.Values(
			transfer.Index, transfer.Status,
			transfer.CreatedAt, transfer.UpdatedAt, transfer.Creator,
			transfer.RarimoTx, transfer.RarimoTxTimestamp, transfer.Origin,
			transfer.Tx, transfer.EventID, transfer.FromChain,
			transfer.ToChain, transfer.Receiver, transfer.Amount,
			transfer.BundleData, transfer.BundleSalt, transfer.ItemIndex)
	}

	// mitigating conflict on index problems in case transfer gets re-submitted
	stmt = stmt.Suffix(
		`ON CONFLICT(index) DO ` +
			`UPDATE SET ` +
			`status = EXCLUDED.status, updated_at = EXCLUDED.updated_at, creator = EXCLUDED.creator, rarimo_tx = EXCLUDED.rarimo_tx, rarimo_tx_timestamp = EXCLUDED.rarimo_tx_timestamp, origin = EXCLUDED.origin, tx = EXCLUDED.tx, event_id = EXCLUDED.event_id, from_chain = EXCLUDED.from_chain, to_chain = EXCLUDED.to_chain, receiver = EXCLUDED.receiver, amount = EXCLUDED.amount, bundle_data = EXCLUDED.bundle_data, bundle_salt = EXCLUDED.bundle_salt, item_index = EXCLUDED.item_index `)

	return q.db.ExecContext(ctx, stmt)
}

func (q TransferQ) SetStatusByIndexCtx(ctx context.Context, status int, indexes ...string) error {
	stmt := squirrel.
		Update("public.transfers").
		Set("status", status).
		Where(squirrel.Eq{"index": indexes})

	return q.db.ExecContext(ctx, stmt)
}

func (q TransferQ) SelectCtx(ctx context.Context, selector data.TransferSelector) ([]data.Transfer, error) {
	stmt := applyTransfersSelector(
		squirrel.Select("*").From("public.transfers"),
		selector)

	var transfers []data.Transfer

	if err := q.db.SelectContext(ctx, &transfers, stmt); err != nil {
		return nil, errors.Wrap(err, "failed to select transfers")
	}

	return transfers, nil
}

func applyTransfersSelector(stmt squirrel.SelectBuilder, selector data.TransferSelector) squirrel.SelectBuilder {
	if selector.Origin != nil {
		stmt = stmt.Where(squirrel.Eq{"origin": selector.Origin})
	}

	if selector.RarimoTx != nil {
		stmt = stmt.Where(squirrel.Eq{"rarimo_tx": data.MustDBHash(string(hexutil.MustDecode(*selector.RarimoTx)))})
	}

	if selector.ChainTx != nil {
		stmt = stmt.Where(squirrel.Eq{"tx": []byte(*selector.ChainTx)})
	}

	if selector.SourceChain != nil {
		stmt = stmt.Where(squirrel.Eq{"from_chain": selector.SourceChain})
	}

	if selector.DestinationChain != nil {
		stmt = stmt.Where(squirrel.Eq{"to_chain": selector.DestinationChain})
	}

	if selector.Receiver != nil {
		stmt = stmt.Where(squirrel.Eq{"receiver": selector.Receiver})
	}

	if selector.Status != nil {
		stmt = stmt.Where(squirrel.Eq{"status": selector.Status})
	}

	if selector.Creator != nil {
		stmt = stmt.Where(squirrel.Eq{"creator": selector.Creator})
	}

	if selector.Before != nil {
		stmt = stmt.Where(squirrel.Lt{"rarimo_tx_timestamp": selector.Before})
	}

	if selector.After != nil {
		stmt = stmt.Where(squirrel.Gt{"rarimo_tx_timestamp": selector.After})
	}

	if selector.ItemIndex != nil {
		stmt = stmt.Where(squirrel.Eq{"item_index": selector.ItemIndex})
	}

	stmt = applyTransfersPagination(stmt, selector.Sort, selector.PageCursor, selector.PageSize)

	return stmt
}

func applyTransfersPagination(stmt squirrel.SelectBuilder, sorts pgdb.Sorts, cursor, limit uint64) squirrel.SelectBuilder {
	if limit != 0 {
		stmt = stmt.Limit(limit)
	}

	if len(sorts) == 0 {
		sorts = pgdb.Sorts{"-time"}
	}

	stmt = sorts.ApplyTo(stmt, map[string]string{
		"id":   "id",
		"time": "rarimo_tx_timestamp",
	})

	if cursor != 0 {
		comp := ">" // default to ascending order
		if sortDesc := strings.HasPrefix(string(sorts[0]), "-"); sortDesc {
			comp = "<"
		}

		stmt = stmt.Where(fmt.Sprintf("id %s ?", comp), cursor)
	}

	return stmt
}
