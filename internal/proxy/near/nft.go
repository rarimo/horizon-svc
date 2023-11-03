package near

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rarimo/near-go/common"
	"github.com/rarimo/near-go/nearclient"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (e *nearProxy) getNfts(ctx context.Context, chain, tokenAddress, accountAddress string, limit, itemIndex *uint64) ([]common.NftView, error) {
	fields := logan.F{
		"account":       accountAddress,
		"chain":         chain,
		"token_address": tokenAddress,
	}

	tokenAddr, err := hexutil.Decode(tokenAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode token address", fields)
	}

	argsMap := map[string]interface{}{
		"account_id": accountAddress,
	}
	if limit != nil && *limit != 0 {
		argsMap["limit"] = *limit
	}
	if itemIndex != nil && *itemIndex != 0 {
		argsMap["from_index"] = fmt.Sprintf("%d", *itemIndex)
	}
	args, err := argsToBase64(argsMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal create args", fields)
	}

	resp, err := e.cli.ContractViewCallFunction(
		ctx,
		string(tokenAddr),
		common.ContractNftTokensForOwner,
		*args,
		nearclient.FinalityFinal(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get nft balance", fields)
	}
	if resp.Result == nil {
		return []common.NftView{}, nil
	}

	var result []common.NftView
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal nft view", fields)
	}

	return result, nil
}
