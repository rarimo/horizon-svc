package utils

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/chain_gateway/types"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/rpc"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func LoadNfts(ctx context.Context, cli *rpc.Client, opts *types.NftListOpts, requestTokenAddress, handlerTokenAddress *string, requestLimit uint64) ([]rpc.Nft, *data.Cursor, error) {
	var (
		pageNumber   = opts.PageNumber
		itemIndex    = opts.ItemIndex
		resp         *rpc.FetchNftResult
		err          error
		nfts         = make([]rpc.Nft, 0)
		responseNfts []rpc.Nft
	)
	for {
		resp, err = cli.FetchNfts(ctx, opts.AccountAddress, requestTokenAddress, requestLimit, pageNumber)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to get nfts")
		}

		limit := opts.Limit - uint64(len(nfts))
		responseNfts, pageNumber, itemIndex = handleNfts(resp.Nfts, handlerTokenAddress, limit, pageNumber, itemIndex)

		if pageNumber == nil && itemIndex == nil {
			return nfts, nil, nil
		}

		nfts = append(nfts, responseNfts...)

		if len(nfts) == int(opts.Limit) {
			return nfts, data.NewCursor(*pageNumber, *itemIndex), nil
		}
	}
}

func handleNfts(input []rpc.Nft, tokenAddress *string, limit uint64, page, itemIndex *uint64) ([]rpc.Nft, *uint64, *uint64) {
	if len(input) == 0 {
		return []rpc.Nft{}, nil, nil
	}

	startIndex := 0
	if itemIndex != nil {
		startIndex = int(*itemIndex)
	}

	pageNumber := 1
	if page != nil {
		pageNumber = int(*page)
	}

	result := make([]rpc.Nft, 0)

	for i, nft := range input {
		if startIndex > i {
			continue
		}
		if tokenAddress != nil && *tokenAddress != "" && nft.CollectionAddress != *tokenAddress {
			continue
		}

		result = append(result, nft)
		startIndex = i + 1

		if len(result) != int(limit) {
			continue
		}

		break
	}

	if startIndex == len(input) || uint64(startIndex) == rpc.MaxFetchNftLimit {
		startIndex = 0
		pageNumber++
	}

	i := uint64(startIndex)
	p := uint64(pageNumber)

	return result, &p, &i
}
