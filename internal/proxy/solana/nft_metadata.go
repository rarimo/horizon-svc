package solana

import (
	"bytes"
	"context"
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/metadata_fetcher"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (e *solanaProxy) NftMetadata(ctx context.Context, opts *types.NftMetadataOpts) (*data.NftMetadata, error) {
	if opts == nil {
		return nil, errors.New("item is nil")
	}

	fields := logan.F{
		"token_id":   opts.TokenID,
		"token_type": opts.TokenType,
		"chain":      opts.Chain,
	}
	mint, err := solana.PublicKeyFromBase58(opts.TokenID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse token id", fields)
	}

	metadata, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find token metadata address", fields)
	}

	metadataInfo, err := e.cli.GetAccountInfo(ctx, metadata)
	if err != nil {
		if errors.Cause(err) == rpc.ErrNotFound {
			return nil, types.ErrorNotFound
		}
		return nil, errors.Wrap(err, "failed to get metadata account info", fields)
	}

	var metaplexMetadata types.MetaplexMetadata
	err = borsh.Deserialize(&metaplexMetadata, metadataInfo.Value.Data.GetBinary())
	if err != nil {
		return nil, err
	}

	metadataURL := trimMatches(metaplexMetadata.Data.URI)
	chainMetadata := &data.NftMetadata{
		MetadataUrl: &metadataURL,
		Name:        trimMatches(metaplexMetadata.Data.Name),
	}

	nftMetadata, err := e.fetcher.GetMetadata(ctx, *chainMetadata.MetadataUrl, opts.TokenID)
	if err != nil {
		if errors.Cause(err) == metadata_fetcher.ErrNonRetriable {
			return nil, types.ErrorNotFound
		}
		return nil, errors.Wrap(err, "failed to get metadata")
	}

	nftMetadata.Merge(chainMetadata)

	return nftMetadata, nil
}

func trimMatches(str string) string {
	return string(bytes.Trim([]byte(str), "\x00"))
}
