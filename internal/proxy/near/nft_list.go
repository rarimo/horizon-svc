package near

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

// near nft contract method `nft_tokens_for_owner` doesn't support pagination through
// page number but through token index.
var defaultPageNumber uint64 = 0

func (e *nearProxy) NftList(ctx context.Context, opts *types.NftsOpts) ([]*data.Nft, *data.Cursor, error) {
	fields := logan.F{
		"account":       opts.AccountAddress,
		"chain":         opts.Chain,
		"token_address": opts.TokenAddress,
	}

	collection, err := e.getCollectionMetadata(ctx, opts.TokenAddress)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get nft collection metadata", fields)
	}

	nfts, err := e.getNfts(ctx, opts.Chain, opts.TokenAddress, opts.AccountAddress, &opts.Limit, opts.ItemIndex)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get nft balance", fields)
	}

	if len(nfts) == 0 {
		return []*data.Nft{}, nil, nil
	}

	result := make([]*data.Nft, len(nfts))

	for i, nft := range nfts {
		metadata, err := e.getNftMetadata(ctx, nft.Metadata, nft.TokenID)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to get nft metadata", fields)
		}

		item := &data.Nft{
			ID:             nft.TokenID,
			Name:           nft.Metadata.Title,
			Description:    &nft.Metadata.Description,
			ImageURL:       metadata.ImageURL,
			CollectionName: collection.Name,
			Attributes:     metadata.Attributes,
		}

		result[i] = item
	}

	index := uint64(0)
	if opts.ItemIndex != nil {
		index = *opts.ItemIndex
	}

	if len(result) == int(opts.Limit) {
		return result, data.NewCursor(defaultPageNumber, index+opts.Limit), nil
	}

	return result, nil, nil
}
