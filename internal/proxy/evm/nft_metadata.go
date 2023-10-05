package evm

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/metadata_fetcher"
	"github.com/rarimo/horizon-svc/internal/proxy/evm/generated/erc1155"
	"github.com/rarimo/horizon-svc/internal/proxy/evm/generated/erc721"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"math/big"
)

func (e *evmProxy) NftMetadata(ctx context.Context, opts *types.NftMetadataOpts) (*data.NftMetadata, error) {
	metadataUri, err := e.getNftMetadataUri(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token uri")
	}

	metadata, err := e.fetcher.GetMetadata(ctx, metadataUri, opts.TokenID)
	if err != nil {
		if errors.Cause(err) == metadata_fetcher.ErrNonRetriable {
			return nil, types.ErrorNotFound
		}
		return nil, errors.Wrap(err, "failed to get metadata")
	}

	return metadata, nil
}

func (e *evmProxy) getNftMetadataUri(opts *types.NftMetadataOpts) (string, error) {
	tokenID, ok := big.NewInt(0).SetString(opts.TokenID, 10)
	if !ok {
		return "", errors.New("failed to parse nft id")
	}

	var metadataUri string
	var err error
	switch opts.TokenType {
	case tokenmanager.Type_ERC721:
		metadataUri, err = e.getErc721TokenUri(opts.TokenAddress, tokenID)
	case tokenmanager.Type_ERC1155:
		metadataUri, err = e.getErc1155TokenUri(opts.TokenAddress, tokenID)
	default:
		return "", errors.New("unsupported token type")
	}
	if err != nil {
		return "nil", errors.Wrap(err, "failed to get token uri")
	}

	return metadataUri, nil
}

func (e *evmProxy) getErc721TokenUri(tokenAddress string, tokenID *big.Int) (string, error) {
	token, err := erc721.NewERC721Caller(common.HexToAddress(tokenAddress), e.cli)
	if err != nil {
		return "", errors.Wrap(err, "failed to create erc721 token caller")
	}

	uri, err := token.TokenURI(&bind.CallOpts{}, tokenID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get token uri")
	}

	return uri, nil

}

func (e *evmProxy) getErc1155TokenUri(tokenAddress string, tokenID *big.Int) (string, error) {
	token, err := erc1155.NewERC1155Caller(common.HexToAddress(tokenAddress), e.cli)
	if err != nil {
		return "", errors.Wrap(err, "failed to create erc1155 token")
	}

	uri, err := token.Uri(&bind.CallOpts{}, tokenID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get token uri")
	}

	return uri, nil
}
