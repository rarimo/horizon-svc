//go:build manual_test
// +build manual_test

package soltx

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/olegfomenko/solana-go/rpc"

	"github.com/rarimo/horizon-svc/resources"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/olegfomenko/solana-go"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/stretchr/testify/assert"
)

func TestBuildTx(t *testing.T) {
	recentBlockHash, err := solana.HashFromBase58("4FkXvsqDihRtDH4hBevZdLjCq2b82tC7XSY51qeTnchh")
	if !assert.NoError(t, err, "expected to decode recent block hash") {
		return
	}

	blockHasherMock := newMockBlockHasher(t)
	blockHasherMock.
		On("GetRecentBlockhash", mock.Anything, mock.Anything).
		Return(&rpc.GetRecentBlockhashResult{
			Value: &rpc.BlockhashResult{
				Blockhash: recentBlockHash,
			},
		}, nil)

	programID, err := solana.PublicKeyFromBase58("GexDbBi7B2UrJDi9JkrWH9fFVhmysN7u5C9zT2HkC6yZ")
	if !assert.NoError(t, err, "expected to decode program id") {
		return
	}
	bridgeAdmin, err := solana.PublicKeyFromBase58("FCpFKSEboCUGg1Qs8NFwH2suMAHYWvFUUiVWk8cKwNqf")
	if !assert.NoError(t, err, "expected to decode bridge admin") {
		return
	}

	builder := NewBuilder(blockHasherMock, programID, bridgeAdmin)

	ctx := context.Background()

	req, txdata, sender := makeData(t)

	tx, err := builder.BuildTx(ctx, req, txdata)
	if !assert.NoError(t, err, "expected to build tx") {
		return
	}

	decodedFrom64, err := base64.StdEncoding.DecodeString(tx.Attributes.Envelope)
	if !assert.NoError(t, err, "expected to decode base64") {
		return
	}

	decodedTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(decodedFrom64))
	if !assert.NoError(t, err, "expected to decode transaction") {
		return
	}

	if !assert.Equal(t, 1, len(decodedTx.Signatures), "expected to have 1 signature") {
		return
	}

	if !assert.True(t, decodedTx.HasAccount(sender), "expected to have sender") {
		return
	}

	if !assert.True(t, decodedTx.Message.IsWritable(sender), "expected to have sender as writable") {
		return
	}
}

func makeData(t *testing.T) (*resources.BuildTx, resources.SolanaTxData, solana.PublicKey) {
	sender := "EXauvdsp15K8DyxXXnRgafmrfNCiLuER2Eh2NjrzSWY2"
	hexedSender := hex.EncodeToString([]byte(sender))
	if !strings.HasPrefix(hexedSender, "0x") {
		hexedSender = "0x" + hexedSender
	}

	t.Log("hexedSender:", hexedSender)

	receiver := "DVSiXFTkToMD7MSF1pscuE9gpNmV9g5tPfwn1hv1cJNy"
	hexedReceiver := hex.EncodeToString([]byte(receiver))
	if !strings.HasPrefix(hexedReceiver, "0x") {
		hexedReceiver = "0x" + hexedReceiver
	}

	t.Log("hexedReceiver:", hexedReceiver)

	amount := "10000000000"

	txdata := resources.SolanaTxData{
		Amount:        &amount,
		Receiver:      "Solana:" + hexedReceiver,
		TargetNetwork: "Solana",
	}

	req := resources.BuildTx{
		Key: resources.Key{
			Type: "build-tx-requests",
		},
		Attributes: resources.BuildTxAttributes{
			Network: "Solana",
			TxType:  resources.TxTypeDepositNative,
		},
		Relationships: resources.BuildTxRelationships{
			Creator: *resources.
				NewStringKey("rarimo1c8rul78jzn0pg69u3hfdmvduyk8jd9g7ctsae9", resources.ACCOUNTS).
				AsRelation(),
			CreatorAccount: *resources.
				NewStringKey(
					"Solana:"+hexedSender,
					resources.ACCOUNT_EXTERNAL_IDS).
				AsRelation(),
		},
	}

	return &req, txdata, solana.MustPublicKeyFromBase58(sender)
}

func TestRandomSolanaKeypair(t *testing.T) {
	private, err := solana.NewRandomPrivateKey()
	if !assert.NoError(t, err, "expected to generate private") {
		return
	}

	fmt.Println(private.String())
	fmt.Println(private.PublicKey().String())
}

func TestSolanaAccountFromHex(t *testing.T) {
	raw := "FCpFKSEboCUGg1Qs8NFwH2suMAHYWvFUUiVWk8cKwNqf"

	hexed := hex.EncodeToString([]byte(raw))

	fmt.Println(hexed)

	fromHex, err := hex.DecodeString(hexed)
	if !assert.NoError(t, err, "expected to decode hex") {
		return
	}

	account, err := solana.PublicKeyFromBase58(string(fromHex))
	if !assert.NoError(t, err, "expected to decode base58") {
		return
	}

	fmt.Println(account.String())
}

func TestSolanaKeyFromHex(t *testing.T) {
	raw := "Solana:464370464b5345626f43554767315173384e4677483273754d41485957764655556956576b38634b774e7166"

	net, addr, err := data.DecodeAccountID(data.AccountID(raw))
	if !assert.NoError(t, err, "expected to decode account id") {
		return
	}

	account, err := solana.PublicKeyFromBase58(string(addr))
	if !assert.NoError(t, err, "expected to decode base58") {
		return
	}

	fmt.Println(net, account.String())
}

func TestSolanaKeypair(t *testing.T) {
	rawPrivate := `fill me in`

	private, err := solana.PrivateKeyFromBase58(rawPrivate)
	if !assert.NoError(t, err, "expected to decode private key") {
		return
	}

	fmt.Println(private.PublicKey().String())

}

func TestDecodeSolanaTx(t *testing.T) {
	raw := ``

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if !assert.NoError(t, err, "expected to decode base64") {
		return
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(decoded))
	if !assert.NoError(t, err, "expected to decode transaction") {
		return
	}

	spew.Dump(tx)
}
