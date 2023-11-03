package core

import (
	"context"
	rarimocoretypes "github.com/rarimo/rarimo-core/x/rarimocore/types"
	tokenmanagertypes "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"google.golang.org/grpc"
)

type Core interface {
	Tokenmanager() Tokenmanager
	Rarimocore() Rarimocore
}

type Tokenmanager interface {
	GetItem(ctx context.Context, infoIndex string) (*tokenmanagertypes.Item, error)
	GetItems(ctx context.Context) ([]tokenmanagertypes.Item, error)
	GetCollections(ctx context.Context) ([]tokenmanagertypes.Collection, error)
	GetParams(ctx context.Context) (*tokenmanagertypes.Params, error)
	GetCollectionData(ctx context.Context, chain, address string) (*tokenmanagertypes.CollectionData, error)
	GetCollection(ctx context.Context, index string) (*tokenmanagertypes.Collection, error)
	GetCollectionDataByCollectionForChain(ctx context.Context, index, chain string) (*tokenmanagertypes.CollectionData, error)
	GetOnChainItem(ctx context.Context, chain, address, tokenID string) (*tokenmanagertypes.OnChainItem, error)
	GetOnChainItemByItemForChain(ctx context.Context, itemIndex, chain string) (*tokenmanagertypes.OnChainItem, error)
}

type Rarimocore interface {
	GetConfirmation(ctx context.Context, root string) (*rarimocoretypes.Confirmation, error)
	GetOperation(ctx context.Context, index string) (*rarimocoretypes.Operation, error)
}

type corer struct {
	tm *tokenmanager
	rc *rarimocore
}

func NewCore(cli *grpc.ClientConn) Core {
	return &corer{
		&tokenmanager{tokenmanagertypes.NewQueryClient(cli)},
		&rarimocore{rarimocoretypes.NewQueryClient(cli)},
	}
}

func (c *corer) Tokenmanager() Tokenmanager {
	return c.tm
}

func (c *corer) Rarimocore() Rarimocore {
	return c.rc
}
