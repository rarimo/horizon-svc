package evm

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func logToWithdrawal(ctx context.Context, cli *ethclient.Client, log types.Log, chain string) (*msgs.WithdrawalMsg, error) {
	receipt, err := cli.TransactionReceipt(ctx, log.TxHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get receipt")
	}

	receiptBytes, err := json.Marshal(receipt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal receipt")
	}

	return &msgs.WithdrawalMsg{
		Hash:        data.FormatWithdrawalID(chain, log.TxHash.String()),
		BlockHeight: int64(log.BlockNumber),
		TxResult:    receiptBytes,
		Success:     receipt.Status == types.ReceiptStatusSuccessful,
	}, nil
}
