package solana

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/programs/token"
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/rarimo/horizon-svc/internal/amount"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"github.com/rarimo/horizon-svc/internal/proxy/utils"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"math/big"
)

func (e *solanaProxy) BalanceOf(ctx context.Context, opts *types.BalanceOfOpts) (*amount.Amount, error) {
	if opts == nil {
		return nil, errors.New("opts is nil")
	}

	switch opts.TokenType {
	case tokenmanager.Type_NATIVE:
		return e.getNativeBalance(ctx, opts)
	case tokenmanager.Type_METAPLEX_FT, tokenmanager.Type_METAPLEX_NFT:
		return e.getBalance(ctx, opts)
	default:
		return nil, errors.From(errors.New("unsupported token type"), logan.F{
			"token_type": opts.TokenType,
			"chain":      opts.Chain,
			"account_id": opts.AccountAddress,
		})

	}
}

func (e *solanaProxy) getNativeBalance(ctx context.Context, opts *types.BalanceOfOpts) (*amount.Amount, error) {
	fields := logan.F{
		"account": opts.AccountAddress,
		"chain":   opts.Chain,
	}

	pubKey, err := solana.PublicKeyFromBase58(opts.AccountAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse account address", fields)
	}
	bal, err := e.cli.GetBalance(ctx, pubKey, rpc.CommitmentFinalized)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get native account balance", fields)
	}

	return utils.AmountFromBig(big.NewInt(0).SetInt64(int64(bal.Value)), opts.Decimals), nil
}

func (e *solanaProxy) getBalance(ctx context.Context, opts *types.BalanceOfOpts) (*amount.Amount, error) {
	fields := logan.F{
		"account":       opts.AccountAddress,
		"token_address": opts.TokenAddress,
		"chain":         opts.Chain,
	}

	pubKey, err := solana.PublicKeyFromBase58(opts.AccountAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse account address", fields)
	}

	tokenAddrBytes, err := hexutil.Decode(opts.TokenAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode token address", fields)
	}

	mintKey := solana.PublicKeyFromBytes(tokenAddrBytes)
	tokenKey, _, err := solana.FindAssociatedTokenAddress(pubKey, mintKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find associated token address", fields)
	}

	data, err := e.cli.GetAccountInfo(ctx, tokenKey)
	if err != nil {
		if errors.Cause(err) == rpc.ErrNotFound {
			return utils.AmountFromBig(big.NewInt(0), opts.Decimals), nil
		}
		return nil, errors.Wrap(err, "failed to get account info", fields)
	}

	var account token.Account
	err = borsh.Deserialize(&account, data.Value.Data.GetBinary())
	if err != nil {
		return nil, errors.Wrap(err, "failed to deserialize account", fields)
	}

	return utils.AmountFromBig(big.NewInt(0).SetUint64(account.Amount), opts.Decimals), nil
}
