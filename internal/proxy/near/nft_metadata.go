package near

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/metadata_fetcher"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"github.com/rarimo/near-go/common"
	"github.com/rarimo/near-go/nearclient"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (e *nearProxy) NftMetadata(ctx context.Context, opts *types.NftMetadataOpts) (*data.NftMetadata, error) {
	fields := logan.F{
		"token_id":   opts.TokenID,
		"token_type": opts.TokenType,
		"chain":      opts.Chain,
	}

	tokenAddr, err := hexutil.Decode(opts.TokenAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode token address", fields)
	}

	args, err := argsToBase64(map[string]interface{}{"token_id": opts.TokenID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal create args", fields)
	}

	resp, err := e.cli.ContractViewCallFunction(
		ctx,
		string(tokenAddr),
		common.ContractNftGet,
		*args,
		nearclient.FinalityFinal(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get an NFT", fields)
	}

	var result *common.NftView
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal an NFT", fields)
	}

	return e.getNftMetadata(ctx, result.Metadata, opts.TokenID)
}

func (e *nearProxy) getNftMetadata(ctx context.Context, metadata *common.NftMetadataView, tokenId string) (*data.NftMetadata, error) {
	chainMetadata := &data.NftMetadata{
		MetadataUrl: &metadata.Reference,
		Name:        metadata.Title,
		ImageURL:    metadata.Media,
		Description: &metadata.Description,
	}

	if metadata.Reference == "" {
		return chainMetadata, nil
	}

	result, err := e.fetcher.GetMetadata(ctx, metadata.Reference, tokenId)
	if err != nil {
		if errors.Cause(err) == metadata_fetcher.ErrNonRetriable {
			return nil, types.ErrorNotFound
		}
		return nil, errors.Wrap(err, "failed to get metadata")
	}

	chainMetadata.Merge(result)

	return chainMetadata, nil
}
