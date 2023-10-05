package pg

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/rarimo/horizon-svc/internal/data"
)

func (q ItemChainMappingQ) InsertBatchCtx(ctx context.Context, chainMappings ...data.ItemChainMapping) error {
	stmt := squirrel.Insert("public.item_chain_mappings").
		Columns("item", "network", "address", "token_id", "created_at", "updated_at")

	for _, m := range chainMappings {
		stmt = stmt.
			Values(m.Item, m.Network, m.Address, m.TokenID, m.CreatedAt, m.UpdatedAt)
	}

	return q.db.ExecContext(ctx, stmt)
}

func (q ItemChainMappingQ) DeleteByItemCtx(ctx context.Context, item int64) error {
	return q.db.ExecContext(ctx,
		squirrel.
			Delete("public.item_chain_mappings").
			Where(squirrel.Eq{"item": item}))
}
