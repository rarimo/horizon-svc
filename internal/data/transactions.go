package data

import (
	"encoding/hex"
	"strings"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func MustDBHash(in string) []byte {
	hashBB, err := hex.DecodeString(strings.ToLower(in)) // bytes.HexBytes.String() does strings.ToUpper(hex.EncodeToString)
	if err != nil {
		panic(errors.Wrap(err, "failed to decode hex hash", logan.F{
			"raw": in,
		}))
	}
	return hashBB
}
