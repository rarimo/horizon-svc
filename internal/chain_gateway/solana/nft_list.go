package solana

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/chain_gateway/types"
	"github.com/rarimo/horizon-svc/internal/chain_gateway/utils"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/rpc"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (c *solanaChainGateway) NftList(ctx context.Context, opts *types.NftListOpts) ([]*data.Nft, *data.Cursor, error) {
	nfts, cursor, err := c.loadNfts(ctx, opts)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to load nfts")
	}

	result := utils.ParseNftList(nfts)
	return result, cursor, nil
}

func (c *solanaChainGateway) loadNfts(ctx context.Context, opts *types.NftListOpts) ([]rpc.Nft, *data.Cursor, error) {
	return utils.LoadNfts(ctx, c.cfg.Rpc, opts, nil, &opts.TokenAddress, rpc.MaxFetchNftLimit)
}
