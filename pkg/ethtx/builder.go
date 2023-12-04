package ethtx

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rarimo/evm-bridge-contracts/bindings/contracts/bridge/bridge"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/resources"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Builder struct {
	ethClient    *ethclient.Client
	bridgeAbi    *abi.ABI // need only abi to pack arguments
	contractAddr common.Address
}

func NewBuilder(ethClient *ethclient.Client, bridgeAbi *abi.ABI, contractAddr common.Address) *Builder {
	return &Builder{
		ethClient:    ethClient,
		bridgeAbi:    bridgeAbi,
		contractAddr: contractAddr,
	}
}

func (b *Builder) BuildTx(ctx context.Context, req *resources.BuildTx, rawTxData interface{}) (*resources.UnsubmittedTx, error) {
	txData, ok := rawTxData.(resources.EthTxData)
	if !ok {
		return nil, errors.From(errors.New("invalid tx_data"), logan.F{
			"expected": "resources.EthTxData",
			"actual":   fmt.Sprintf("%T", rawTxData),
		})
	}

	gp, err := b.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to suggest gas price")
	}

	_, accountID, err := data.DecodeAccountID(data.AccountID(req.Relationships.CreatorAccount.Data.ID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode account id")
	}

	nonce, err := b.ethClient.PendingNonceAt(ctx, common.BytesToAddress(accountID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pending nonce", logan.F{
			"raw_address": req.Relationships.CreatorAccount.Data.ID,
		})
	}

	var bundleSalt [32]byte
	copy(bundleSalt[:], txData.BundleSalt[:])

	bundle := bridge.IBundlerBundle{
		Salt:   bundleSalt,
		Bundle: []byte(txData.BundleData),
	}

	baseTxParams := types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gp,
		Gas:      uint64(21000), // TODO where do i get it?
		To:       &b.contractAddr,
	}

	switch req.Attributes.TxType {
	case resources.TxTypeDepositNative:
		if err := validateNativeDepositTxData(txData); err != nil {
			return nil, err
		}

		baseTxParams.Value, ok = new(big.Int).SetString(*txData.Amount, 10)
		if !ok {
			return nil, errors.From(errors.New("failed to parse amount"), logan.F{
				"raw": txData.Amount,
			})
		}

		baseTxParams.Data, err = b.bridgeAbi.Pack("depositNative",
			bundle,
			txData.TargetNetwork,
			txData.Receiver)
		if err != nil {
			return nil, errors.Wrap(err, "failed to pack depositNative input")
		}

		return makeTx(b.contractAddr, &baseTxParams)
	case resources.TxTypeDepositErc20:
		if err := validateErc20DepositTxData(txData); err != nil {
			return nil, err
		}

		amount, ok := new(big.Int).SetString(*txData.Amount, 10)
		if !ok {
			return nil, errors.From(errors.New("failed to parse amount"), logan.F{
				"raw": txData.Amount,
			})
		}

		baseTxParams.Data, err = b.bridgeAbi.Pack("depositERC20",
			txData.TokenAddr,
			amount,
			bundle,
			txData.TargetNetwork,
			txData.Receiver,
			*txData.IsWrapped)
		if err != nil {
			return nil, errors.Wrap(err, "failed to pack depositERC20 input")
		}

		return makeTx(b.contractAddr, &baseTxParams)
	case resources.TxTypeDepositErc721:
		if err := validateErc721DepositTxData(txData); err != nil {
			return nil, err
		}

		tokenID, ok := new(big.Int).SetString(*txData.TokenId, 10)
		if !ok {
			return nil, errors.From(errors.New("failed to parse token id"), logan.F{
				"raw": txData.TokenId,
			})
		}

		baseTxParams.Data, err = b.bridgeAbi.Pack("depositERC721",
			txData.TokenAddr,
			tokenID,
			bundle,
			txData.TargetNetwork,
			txData.Receiver,
			*txData.IsWrapped)
		if err != nil {
			return nil, errors.Wrap(err, "failed to pack depositERC721 input")
		}

		return makeTx(b.contractAddr, &baseTxParams)
	case resources.TxTypeDepositErc1155:
		if err := validateErc1155DepositTxData(txData); err != nil {
			return nil, err
		}

		tokenID, ok := new(big.Int).SetString(*txData.TokenId, 10)
		if !ok {
			return nil, errors.From(errors.New("failed to parse token id"), logan.F{
				"raw": txData.TokenId,
			})
		}

		amount, ok := new(big.Int).SetString(*txData.Amount, 10)
		if !ok {
			return nil, errors.From(errors.New("failed to parse amount"), logan.F{
				"raw": txData.Amount,
			})
		}

		baseTxParams.Data, err = b.bridgeAbi.Pack("depositERC1155",
			txData.TokenAddr,
			tokenID,
			amount,
			bundle,
			txData.TargetNetwork,
			txData.Receiver,
			*txData.IsWrapped)
		if err != nil {
			return nil, errors.Wrap(err, "failed to pack depositERC1155 input")
		}

		return makeTx(b.contractAddr, &baseTxParams)
	default:
		return nil, errors.From(errors.New("unsupported tx type"), logan.F{
			"tx_type": req.Attributes.TxType,
			"allowed": resources.SupportedTxTypesEth(),
		})
	}
}

func makeTx(contractAddr common.Address, data types.TxData) (*resources.UnsubmittedTx, error) {
	tx := types.NewTx(data)

	envelope, err := tx.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal envelope")
	}

	return &resources.UnsubmittedTx{
		Key: resources.Key{
			ID:   tx.Hash().String(),
			Type: resources.UNSUBMITTED_TRANSACTIONS,
		},
		Attributes: resources.UnsubmittedTxAttributes{
			ContractAddr: contractAddr.String(),
			Envelope:     hexutil.Encode(envelope),
			GeneratedAt:  time.Now().UTC().String(),
		},
	}, nil
}

func validateNativeDepositTxData(data resources.EthTxData) error {
	if data.Amount == nil {
		return errors.New("amount is required")
	}

	return nil
}

func validateErc20DepositTxData(data resources.EthTxData) error {
	if data.Amount == nil {
		return errors.New("amount is required")
	}

	if data.TokenAddr == nil {
		return errors.New("token address is required")
	}

	if data.IsWrapped == nil {
		return errors.New("is wrapped flag is required")
	}

	return nil
}

func validateErc721DepositTxData(data resources.EthTxData) error {
	if data.TokenId == nil {
		return errors.New("token id is required")
	}

	if data.TokenAddr == nil {
		return errors.New("token address is required")
	}

	if data.IsWrapped == nil {
		return errors.New("is wrapped flag is required")
	}

	return nil
}

func validateErc1155DepositTxData(data resources.EthTxData) error {
	if data.Amount == nil {
		return errors.New("amount is required")
	}

	return validateErc721DepositTxData(data)
}
