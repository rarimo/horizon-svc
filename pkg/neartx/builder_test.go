//go:build manual_test
// +build manual_test

package neartx

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/rarimo/near-go/common"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/rarimo/horizon-svc/resources"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

func TestBuildTx(t *testing.T) {
	nearInfoMock := NewMockNearInfoer(t)

	nearInfoMock.
		On("AccessKeyView", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(common.AccessKeyView{
			AccessKey: common.AccessKey{
				Nonce: 223,
			},
		}, nil)

	nearInfoMock.
		On("BlockDetails", mock.Anything, mock.Anything).
		Return(common.BlockView{
			Header: common.BlockHeaderView{
				Hash: common.MustCryptoHashFromBase58("8BwwsZ2nd73V7cDCirQs4sxiZ8n8HKMksWr7bb9Cxdcr"),
			},
		}, nil)

	bridgeAddr := common.AccountID("bridge.rarimo.testnet")

	builder := NewBuilder(nearInfoMock, bridgeAddr)

	ctx := context.Background()
	req, txdata := makeData(t)

	tx, err := builder.BuildTx(ctx, req, txdata)
	if !assert.NoError(t, err, "expected to build tx") {
		return
	}

	spew.Dump(tx)
}

func makeData(t *testing.T) (*resources.BuildTx, resources.NearTxData) {
	req := resources.BuildTx{
		Key: resources.Key{
			Type: "build-tx-requests",
		},
		Attributes: resources.BuildTxAttributes{
			Network: "Near",
			TxType:  resources.TxTypeDepositNative,
		},
		Relationships: resources.BuildTxRelationships{
			Creator: *resources.
				NewStringKey("rarimo1c8rul78jzn0pg69u3hfdmvduyk8jd9g7ctsae9", resources.ACCOUNTS).
				AsRelation(),
			CreatorAccount: *resources.
				NewStringKey(
					"Near:0x636c69656e742e636861696e6c696e6b2e746573746e6574",
					resources.ACCOUNT_EXTERNAL_IDS).
				AsRelation(),
		},
	}

	strp := func(s string) *string {
		return &s
	}

	txdata := resources.NearTxData{
		SenderPublicKey: "Near:0x00eff85923feec05711f9ee4de85d010ece899bef01bab4871604cfd1af16d2119",
		Receiver:        "Solana:0x445653695846546b546f4d44374d53463170736375453967704e6d56396735745066776e31687631634a4e79",
		TargetNetwork:   "Solana",
		Amount:          strp("10000000000"),
	}

	return &req, txdata
}

func TestPredefinedPubKey(t *testing.T) {
	pk := "ed25519:H9k5eiU4xXS3M4z8HzKJSLaZdqGdGwBG49o7orNC4eZW"

	pk58, err := common.NewBase58PublicKey(pk)
	if !assert.NoError(t, err, "expected to parse public key from base58") {
		return
	}

	hexed := hex.EncodeToString(toBytes(pk58.ToPublicKey()))

	fmt.Println("0x" + hexed)
}

func TestAccountID(t *testing.T) {
	accountID := "client.chainlink.testnet"

	hexed := hex.EncodeToString([]byte(accountID))

	fmt.Println("0x" + hexed)

	decoded, err := hex.DecodeString(hexed)
	if !assert.NoError(t, err, "expected to decode hexed account id") {
		return
	}

	fmt.Println(common.AccountID(decoded))
}

func TestGenKeyPair(t *testing.T) {
	keypair, err := common.GenerateKeyPair(common.KeyTypeED25519, rand.Reader)
	if !assert.NoError(t, err, "expected keypair to be generated") {
		return
	}

	fmt.Println(string(keypair.PrivateKey))
	fmt.Println(keypair.PublicKey.String())
	fmt.Println(keypair.PublicKey.ToPublicKey().String())

	hexed := hex.EncodeToString(toBytes(keypair.PublicKey.ToPublicKey()))

	fmt.Println("0x" + hexed)

	decoded, err := hex.DecodeString(hexed)
	if !assert.NoError(t, err, "expected to decode hexed public key") {
		return
	}

	decodedPublic, err := common.PublicKeyFromBytes(decoded)
	if !assert.NoError(t, err, "expected to parse public key from bytes") {
		return
	}

	fmt.Println(decodedPublic.String())
}

func toBytes(publicKey common.PublicKey) []byte {
	return append([]byte{publicKey.TypeByte()}, publicKey.Value()...)
}

func TestKeyPair(t *testing.T) {
	private := `fill your private key here`
	keyPair, err := common.NewBase58KeyPair(private)

	if !assert.NoError(t, err, "expected keypair to be created") {
		return
	}

	public64 := base64.StdEncoding.EncodeToString([]byte(keyPair.PublicKey.ToPublicKey().String()))

	rawPublic, err := base64.StdEncoding.DecodeString(public64)
	if !assert.NoError(t, err, "expected to decode public key from base64") {
		return
	}

	public, err := common.NewBase58PublicKey(string(rawPublic))
	if !assert.NoError(t, err, "expected to parse public key from base58") {
		return
	}

	if !assert.Equal(t, keyPair.PublicKey, public, "expected public keys to match") {
		return
	}
}
