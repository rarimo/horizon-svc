//go:build manual_test
// +build manual_test

package ethtx

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/rarimo/horizon-svc/resources"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rarimo/evm-bridge-contracts/bindings/contracts/bridge/bridge"
	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/ethclient"
)

func TestBuildTx(t *testing.T) {
	ctx := context.Background()

	evmRPCAddr := "wss://goerli.infura.io/ws/v3/..."

	client, err := ethclient.Dial(evmRPCAddr)
	if !assert.NoError(t, err, "expected to connect to EVM RPC") {
		return
	}

	abi, err := bridge.BridgeMetaData.GetAbi()
	if !assert.NoError(t, err, "expected abi") {
		return
	}

	builder := NewBuilder(client, abi, common.HexToAddress("0x4B9Bd5452a741f991AE25377C18d50e68323A40E"))

	req, txData, _ := makeErc20Data(t)

	tx, err := builder.BuildTx(ctx, req, txData)
	if !assert.NoError(t, err, "expected to build tx") {
		return
	}

	spew.Dump(tx)
}

func makeErc20Data(t *testing.T) (*resources.BuildTx, resources.EthTxData, common.Address) {
	sender := "0xF65F3f18D9087c4E35BAC5b9746492082e186872"

	amount := "10000000000"
	wrapped := false
	tokenAddr := common.HexToAddress("0x40b5ea58c9Ec7521d6Ba42d90Af730ab55Ec720c")

	txData := resources.EthTxData{
		Amount:        &amount,
		BundleData:    "0x",
		BundleSalt:    "0x0000000000000000000000000000000000000000000000000000000000000000",
		IsWrapped:     &wrapped,
		Receiver:      "Goerli:0xA256C2e695B7a9070228f9E340CC8739356A7066",
		TargetNetwork: "Goerli",
		TokenAddr:     &tokenAddr,
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
					"Goerli:"+sender,
					resources.ACCOUNT_EXTERNAL_IDS).
				AsRelation(),
		},
	}

	return &req, txData, common.HexToAddress(sender)
}
