package types

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/data"
)

type ChainGateway interface {
	NftList(ctx context.Context, opts *NftListOpts) ([]*data.Nft, *data.Cursor, error)
}

type NftListOpts struct {
	Chain          string  `json:"chain"`
	AccountAddress string  `json:"account_address"`
	TokenAddress   string  `json:"token_address"`
	Limit          uint64  `json:"limit"`
	PageNumber     *uint64 `json:"page_number"`
	ItemIndex      *uint64 `json:"item_index"`
}
