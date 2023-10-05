package evm

import (
	"context"
	gateway "github.com/rarimo/horizon-svc/internal/chain_gateway/types"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (e *evmProxy) NftList(ctx context.Context, opts *types.NftsOpts) ([]*data.Nft, *data.Cursor, error) {
	nfts, cursor, err := e.gateway.Get(opts.Chain).NftList(ctx, &gateway.NftListOpts{
		Chain:          opts.Chain,
		AccountAddress: opts.AccountAddress,
		TokenAddress:   opts.TokenAddress,
		Limit:          opts.Limit,
		PageNumber:     opts.PageNumber,
		ItemIndex:      opts.ItemIndex,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get nfts from chain gateway", logan.F{
			"chain":           opts.Chain,
			"token_address":   opts.TokenAddress,
			"account_address": opts.AccountAddress,
		})
	}

	return nfts, cursor, nil
}
