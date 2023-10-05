package utils

import (
	"github.com/rarimo/horizon-svc/internal/amount"
	"math/big"
)

func AmountFromBig(input *big.Int, precision uint32) *amount.Amount {
	amnt := amount.NewFromIntWithPrecision(input, int(precision))
	return &amnt
}
