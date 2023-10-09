package evm

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func logToWithdrawal(ctx context.Context, cli *ethclient.Client, log types.Log, origin [32]byte, chain string) (*msgs.WithdrawalMsg, error) {
	receipt, err := cli.TransactionReceipt(ctx, log.TxHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get receipt")
	}

	return &msgs.WithdrawalMsg{
		Origin:  hexutil.Encode(origin[:]),
		Hash:    log.TxHash.String(),
		Success: receipt.Status == types.ReceiptStatusSuccessful,
	}, nil
}
