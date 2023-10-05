package near

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rarimo/horizon-svc/internal/amount"
	proxyTypes "github.com/rarimo/horizon-svc/internal/proxy/types"
	"github.com/rarimo/horizon-svc/internal/proxy/utils"
	"github.com/rarimo/near-go/common"
	"github.com/rarimo/near-go/nearclient"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"math/big"
)

func (e *nearProxy) BalanceOf(ctx context.Context, opts *proxyTypes.BalanceOfOpts) (*amount.Amount, error) {
	if opts == nil {
		return nil, errors.New("opts is nil")
	}

	switch opts.TokenType {
	case tokenmanager.Type_NATIVE:
		return e.getNativeBalance(ctx, opts)
	case tokenmanager.Type_NEAR_FT:
		return e.getFtBalance(ctx, opts)
	case tokenmanager.Type_NEAR_NFT:
		return e.getNftBalance(ctx, opts)
	default:
		return nil, errors.From(errors.New("unsupported token type"), logan.F{
			"token_type": opts.TokenType,
			"chain":      opts.Chain,
			"account_id": opts.AccountAddress,
		})
	}
}

func (e *nearProxy) getNativeBalance(ctx context.Context, opts *proxyTypes.BalanceOfOpts) (*amount.Amount, error) {
	acc, err := e.cli.AccountView(ctx, opts.AccountAddress, nearclient.FinalityFinal())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get native account balance", logan.F{
			"account": opts.AccountAddress,
			"chain":   opts.Chain,
		})
	}

	return utils.AmountFromBig(acc.Amount.Big(), opts.Decimals), nil
}

func (e *nearProxy) getFtBalance(ctx context.Context, opts *proxyTypes.BalanceOfOpts) (*amount.Amount, error) {
	fields := logan.F{
		"account":       opts.AccountAddress,
		"chain":         opts.Chain,
		"token_address": opts.TokenAddress,
	}

	tokenAddr, err := hexutil.Decode(opts.TokenAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode token address", fields)
	}

	args, err := argsToBase64(map[string]interface{}{"account_id": opts.AccountAddress})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal create args", fields)
	}

	resp, err := e.cli.ContractViewCallFunction(
		ctx,
		string(tokenAddr),
		common.ContractFtBalanceOf,
		*args,
		nearclient.FinalityFinal(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ft balance", fields)
	}

	var result string
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal ft balance", fields)
	}

	bigAmnt, ok := big.NewInt(0).SetString(result, 10)
	if !ok {
		return nil, errors.From(errors.New("failed to parse ft balance"), fields)
	}

	return utils.AmountFromBig(bigAmnt, opts.Decimals), nil
}

func (e *nearProxy) getNftBalance(ctx context.Context, opts *proxyTypes.BalanceOfOpts) (*amount.Amount, error) {
	nfts, err := e.getNfts(ctx, opts.Chain, opts.TokenAddress, opts.AccountAddress, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get nft balance", logan.F{
			"account":       opts.AccountAddress,
			"chain":         opts.Chain,
			"token_address": opts.TokenAddress,
		})
	}

	return utils.AmountFromBig(big.NewInt(int64(len(nfts))), opts.Decimals), nil
}
