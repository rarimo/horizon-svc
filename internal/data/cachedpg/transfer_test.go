//go:build manual_test
// +build manual_test

package cachedpg

import (
	"database/sql"
	"math/big"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack"

	"github.com/rarimo/horizon-svc/internal/data"
)

func TestEncodeDecodeTransfer(t *testing.T) {
	transfer := data.Transfer{
		ID:        111,
		Index:     []byte("foobar"),
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Creator: sql.NullString{
			Valid:  true,
			String: "foobar",
		},
		RarimoTx:          []byte("foobar"),
		RarimoTxTimestamp: time.Now(),
		Origin:            "foobar",
		Tx:                []byte("foobar"),
		EventID:           1,
		FromChain:         "foobar",
		ToChain:           "foobar",
		Receiver:          "foobar",
		Amount: data.Int256{
			big.NewInt(1000),
		},
		BundleData: nil,
		BundleSalt: nil,
		TokenIndex: "foobar",
	}

	encoded, err := msgpack.Marshal(transfer)
	if !assert.NoError(t, err) {
		return
	}

	var decoded data.Transfer
	err = msgpack.Unmarshal(encoded, &decoded)
	if !assert.NoError(t, err) {
		return
	}

	spew.Dump(decoded)
}
