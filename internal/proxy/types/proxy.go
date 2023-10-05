package types

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/amount"
	"github.com/rarimo/horizon-svc/internal/data"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

var ErrorNotFound = errors.New("not found")

type Proxy interface {
	BalanceOf(ctx context.Context, opts *BalanceOfOpts) (*amount.Amount, error)
	NftMetadata(ctx context.Context, opts *NftMetadataOpts) (*data.NftMetadata, error)
	NftList(ctx context.Context, opts *NftsOpts) ([]*data.Nft, *data.Cursor, error)
}

type BalanceOfOpts struct {
	AccountAddress string            `json:"account_addr"`
	Chain          string            `json:"chain"` // chain name
	Decimals       uint32            `json:"decimals"`
	TokenType      tokenmanager.Type `json:"token_type"`
	TokenAddress   string            `json:"token_address"`      // hex-encoded
	TokenID        string            `json:"token_id,omitempty"` // hex-encoded
}

type NftMetadataOpts struct {
	Chain        string            `json:"chain"` // chain name
	TokenType    tokenmanager.Type `json:"token_type"`
	TokenAddress string            `json:"token_address"`      // hex-encoded
	TokenID      string            `json:"token_id,omitempty"` // hex-encoded
}

type NftsOpts struct {
	AccountAddress string  `json:"account_addr"`
	Chain          string  `json:"chain"`         // chain name
	TokenAddress   string  `json:"token_address"` // hex-encoded
	Limit          uint64  `json:"limit"`
	PageNumber     *uint64 `json:"page_number"`
	ItemIndex      *uint64 `json:"item_index"`
}
