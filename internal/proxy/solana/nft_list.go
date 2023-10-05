package solana

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mr-tron/base58"
	gateway "github.com/rarimo/horizon-svc/internal/chain_gateway/types"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (e *solanaProxy) NftList(ctx context.Context, opts *types.NftsOpts) ([]*data.Nft, *data.Cursor, error) {
	fields := logan.F{
		"token_address":   opts.TokenAddress,
		"account_address": opts.AccountAddress,
		"chain":           opts.Chain,
	}

	tokenAddress, err := hexutil.Decode(opts.TokenAddress)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to decode token address", fields)
	}

	nfts, cursor, err := e.gateway.Get(opts.Chain).NftList(ctx, &gateway.NftListOpts{
		Chain:          opts.Chain,
		AccountAddress: opts.AccountAddress,
		TokenAddress:   base58.Encode(tokenAddress),
		Limit:          opts.Limit,
		PageNumber:     opts.PageNumber,
		ItemIndex:      opts.ItemIndex,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get nfts from chain gateway", fields)
	}

	return nfts, cursor, nil
}
