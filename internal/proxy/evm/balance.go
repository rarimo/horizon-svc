package evm

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rarimo/horizon-svc/internal/amount"
	"github.com/rarimo/horizon-svc/internal/proxy/evm/generated/erc1155"
	"github.com/rarimo/horizon-svc/internal/proxy/evm/generated/erc20"
	"github.com/rarimo/horizon-svc/internal/proxy/evm/generated/erc721"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"github.com/rarimo/horizon-svc/internal/proxy/utils"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"math/big"
)

func (e *evmProxy) BalanceOf(ctx context.Context, opts *types.BalanceOfOpts) (*amount.Amount, error) {
	if opts == nil {
		return nil, errors.New("opts is nil")
	}

	switch opts.TokenType {
	case tokenmanager.Type_NATIVE:
		return e.getNativeBalance(ctx, opts)
	case tokenmanager.Type_ERC20:
		return e.getErc20Balance(ctx, opts)
	case tokenmanager.Type_ERC721:
		return e.getErc721Balance(ctx, opts)
	case tokenmanager.Type_ERC1155:
		return e.getErc1155Balance(ctx, opts)
	default:
		return nil, errors.From(errors.New("unsupported token type"), logan.F{
			"token_type":      opts.TokenType,
			"chain":           opts.Chain,
			"account_address": opts.AccountAddress,
		})

	}
}

func (e *evmProxy) getNativeBalance(ctx context.Context, opts *types.BalanceOfOpts) (*amount.Amount, error) {
	bal, err := e.cli.BalanceAt(ctx, common.HexToAddress(opts.AccountAddress), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get native account balance", logan.F{
			"account": opts.AccountAddress,
			"chain":   opts.Chain,
		})
	}

	return utils.AmountFromBig(bal, opts.Decimals), nil
}

func (e *evmProxy) getErc20Balance(ctx context.Context, opts *types.BalanceOfOpts) (*amount.Amount, error) {
	fields := logan.F{
		"account_address": opts.AccountAddress,
		"token_address":   opts.TokenAddress,
		"chain":           opts.Chain,
	}

	caller, err := erc20.NewERC20Caller(common.HexToAddress(opts.TokenAddress), e.cli)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create erc20 caller", fields)
	}

	bal, err := caller.BalanceOf(&bind.CallOpts{Context: ctx}, common.HexToAddress(opts.AccountAddress))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get erc20 balance", fields)
	}

	return utils.AmountFromBig(bal, opts.Decimals), nil
}

func (e *evmProxy) getErc721Balance(ctx context.Context, opts *types.BalanceOfOpts) (*amount.Amount, error) {
	fields := logan.F{
		"account_address": opts.AccountAddress,
		"token_address":   opts.TokenAddress,
		"chain":           opts.Chain,
	}

	caller, err := erc721.NewERC721Caller(common.HexToAddress(opts.TokenAddress), e.cli)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create erc721 caller", fields)
	}

	bal, err := caller.BalanceOf(&bind.CallOpts{Context: ctx}, common.HexToAddress(opts.AccountAddress))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get erc721 balance", fields)
	}

	return utils.AmountFromBig(bal, opts.Decimals), nil
}

func (e *evmProxy) getErc1155Balance(ctx context.Context, opts *types.BalanceOfOpts) (*amount.Amount, error) {
	fields := logan.F{
		"account":       opts.AccountAddress,
		"token_address": opts.TokenAddress,
		"chain":         opts.Chain,
	}

	if opts.TokenID == "" {
		return nil, errors.From(errors.New("token id is required for erc1155"), fields)
	}

	caller, err := erc1155.NewERC1155Caller(common.HexToAddress(opts.TokenAddress), e.cli)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create erc1155 caller", fields)
	}

	tokenId, ok := big.NewInt(0).SetString(opts.TokenID, 10)
	if !ok {
		return nil, errors.Wrap(err, "failed to convert token id to big int", fields)
	}

	bal, err := caller.BalanceOf(&bind.CallOpts{Context: ctx}, common.HexToAddress(opts.AccountAddress), tokenId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get erc1155 balance", fields)
	}

	return utils.AmountFromBig(bal, opts.Decimals), nil
}
