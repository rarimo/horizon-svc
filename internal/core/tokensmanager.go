package core

import (
	"context"
	"math"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/types/query"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

// FIXME rm this
type TokenManager interface {
	GetItem(ctx context.Context, infoIndex string) (*tokenmanager.Item, error)
	GetItems(ctx context.Context) ([]tokenmanager.Item, error)
	GetCollections(ctx context.Context) ([]tokenmanager.Collection, error)
	GetParams(ctx context.Context) (*tokenmanager.Params, error)
	CollectionData(ctx context.Context, chain, address string) (*tokenmanager.CollectionData, error)
}

type tokenManager struct {
	tokenmanager tokenmanager.QueryClient
}

func (c *tokenManager) CollectionData(ctx context.Context, chain, address string) (*tokenmanager.CollectionData, error) {
	resp, err := c.tokenmanager.CollectionData(ctx, &tokenmanager.QueryGetCollectionDataRequest{
		Chain:   chain,
		Address: address,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "error getting collection entry")
	}

	return &resp.Data, nil
}

func (c *tokenManager) GetItem(ctx context.Context, infoIndex string) (*tokenmanager.Item, error) {
	resp, err := c.tokenmanager.Item(ctx, &tokenmanager.QueryGetItemRequest{
		Index: infoIndex,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "error getting item entry")
	}

	return &resp.Item, nil
}

func (c *tokenManager) GetItems(ctx context.Context) ([]tokenmanager.Item, error) {
	items := make([]tokenmanager.Item, 0, 100) // TODO make this configurable or just keep correlating with the genesis
	page := query.PageRequest{Limit: math.MaxUint64}

	for {
		req := &tokenmanager.QueryAllItemRequest{
			Pagination: &page,
		}

		resp, err := c.tokenmanager.ItemAll(ctx, req)
		if err != nil {
			return nil, errors.Wrap(err, "error getting all items")
		}

		items = append(items, resp.Item...)

		if len(resp.Pagination.NextKey) == 0 {
			return items[:], nil
		}

		page.Key = resp.Pagination.NextKey
	}
}

func (c *tokenManager) GetCollections(ctx context.Context) ([]tokenmanager.Collection, error) {
	collections := make([]tokenmanager.Collection, 0, 100) // TODO make this configurable or just keep correlating with the genesis
	page := query.PageRequest{Limit: math.MaxUint64}

	for {
		req := &tokenmanager.QueryAllCollectionRequest{
			Pagination: &page,
		}

		resp, err := c.tokenmanager.CollectionAll(ctx, req)
		if err != nil {
			return nil, errors.Wrap(err, "error getting all collections")
		}

		collections = append(collections, resp.Collection...)

		if len(resp.Pagination.NextKey) == 0 {
			return collections[:], nil
		}

		page.Key = resp.Pagination.NextKey
	}
}

func (c *tokenManager) GetParams(ctx context.Context) (*tokenmanager.Params, error) {
	resp, err := c.tokenmanager.Params(ctx, &tokenmanager.QueryParamsRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "error getting params entry")
	}

	return &resp.Params, nil
}
