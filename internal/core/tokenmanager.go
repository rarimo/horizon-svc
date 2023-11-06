package core

import (
	"context"
	"gitlab.com/distributed_lab/logan/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/types/query"
	tokenmanagertypes "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const PageLimit = 100

type tokenmanager struct {
	tm tokenmanagertypes.QueryClient
}

var _ Tokenmanager = &tokenmanager{}

func (c *tokenmanager) GetCollectionDataByCollectionForChain(ctx context.Context, index, chain string) (*tokenmanagertypes.CollectionData, error) {
	collectionDataResp, err := c.tm.CollectionDataByCollectionForChain(ctx,
		&tokenmanagertypes.QueryGetCollectionDataByCollectionForChainRequest{
			CollectionIndex: index,
			Chain:           chain,
		})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get collection data from core", logan.F{
			"index": index,
			"chain": chain,
		})
	}

	return &collectionDataResp.Data, nil
}

func (c *tokenmanager) GetCollection(ctx context.Context, index string) (*tokenmanagertypes.Collection, error) {
	resp, err := c.tm.Collection(ctx, &tokenmanagertypes.QueryGetCollectionRequest{
		Index: index,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get collection from core", logan.F{
			"index": index,
		})
	}

	return &resp.Collection, nil
}

func (c *tokenmanager) GetCollectionData(ctx context.Context, chain, address string) (*tokenmanagertypes.CollectionData, error) {
	resp, err := c.tm.CollectionData(ctx, &tokenmanagertypes.QueryGetCollectionDataRequest{
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

func (c *tokenmanager) GetItem(ctx context.Context, infoIndex string) (*tokenmanagertypes.Item, error) {
	resp, err := c.tm.Item(ctx, &tokenmanagertypes.QueryGetItemRequest{
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

func (c *tokenmanager) GetItems(ctx context.Context) ([]tokenmanagertypes.Item, error) {
	var items []tokenmanagertypes.Item
	page := query.PageRequest{Limit: PageLimit}

	for {
		req := &tokenmanagertypes.QueryAllItemRequest{
			Pagination: &page,
		}

		resp, err := c.tm.ItemAll(ctx, req)
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

func (c *tokenmanager) GetCollections(ctx context.Context) ([]tokenmanagertypes.Collection, error) {
	var collections []tokenmanagertypes.Collection
	page := query.PageRequest{Limit: PageLimit}

	for {
		req := &tokenmanagertypes.QueryAllCollectionRequest{
			Pagination: &page,
		}

		resp, err := c.tm.CollectionAll(ctx, req)
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

func (c *tokenmanager) GetParams(ctx context.Context) (*tokenmanagertypes.Params, error) {
	resp, err := c.tm.Params(ctx, &tokenmanagertypes.QueryParamsRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "error getting params entry")
	}

	return &resp.Params, nil
}

func (c *tokenmanager) GetOnChainItemByItemForChain(ctx context.Context, itemIndex, chain string) (*tokenmanagertypes.OnChainItem, error) {
	resp, err := c.tm.OnChainItemByItemForChain(ctx, &tokenmanagertypes.QueryGetOnChainItemByItemForChainRequest{
		ItemIndex: itemIndex,
		Chain:     chain,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "error getting on-chain item entry")
	}

	return &resp.Item, nil
}

func (c *tokenmanager) GetOnChainItem(ctx context.Context, chain, address, tokenID string) (*tokenmanagertypes.OnChainItem, error) {
	onChainItemResp, err := c.tm.OnChainItem(ctx, &tokenmanagertypes.QueryGetOnChainItemRequest{
		Chain:   chain,
		Address: address,
		TokenID: tokenID,
	})

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get item", logan.F{
			"chain":    chain,
			"address":  address,
			"token_id": tokenID,
		})
	}

	return &onChainItemResp.Item, nil
}
