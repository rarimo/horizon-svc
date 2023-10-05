package pg

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/rarimo/horizon-svc/internal/data"
)

func (q CollectionChainMappingQ) InsertBatchCtx(ctx context.Context, chainMappings ...data.CollectionChainMapping) error {
	stmt := squirrel.Insert("public.collection_chain_mappings").
		Columns("collection", "network", "address",
			"token_type", "wrapped", "decimals",
			"created_at", "updated_at")

	for _, m := range chainMappings {
		stmt = stmt.
			Values(m.Collection, m.Network, m.Address,
				m.TokenType, m.Wrapped, m.Decimals,
				m.CreatedAt, m.UpdatedAt)
	}

	return q.db.ExecContext(ctx, stmt)
}

func (q CollectionChainMappingQ) DeleteByCollectionCtx(ctx context.Context, collection int64) error {
	return q.db.ExecContext(ctx,
		squirrel.
			Delete("public.collection_chain_mappings").
			Where(squirrel.Eq{"collection": collection}))
}
