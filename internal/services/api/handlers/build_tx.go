package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rarimo/horizon-svc/pkg/txbuild"

	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"golang.org/x/exp/slices"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"gitlab.com/distributed_lab/logan/v3"

	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/rarimo/horizon-svc/resources"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func newBuildTxRequest(r *http.Request) (*resources.BuildTx, interface{}, error) {
	var req resources.BuildTxRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, nil, errors.Wrap(err, "failed to decode body")
	}

	return validateBuildTx(r, req)
}

type TxBuilder interface {
	BuildTx(ctx context.Context, req *resources.BuildTx, txData interface{}) (*resources.UnsubmittedTx, error)
}

func BuildTx(w http.ResponseWriter, r *http.Request) {
	request, txData, err := newBuildTxRequest(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}
	tx, err := Builder(r).BuildTx(r.Context(), request, txData)
	if err != nil {
		if errors.Cause(err) == txbuild.ErrUnsupportedNetworkType {
			Log(r).WithField("network", request.Attributes.Network).Error("no builder for network")
			ape.RenderErr(w, problems.InternalError()) // not user's problem that we are here after validation
			return
		}
		panic(errors.Wrap(err, "failed to build tx", logan.F{
			"request": request,
		}))
	}

	ape.Render(w, resources.UnsubmittedTxResponse{
		Data:     *tx,
		Included: resources.Included{},
	})
}

func validateBuildTx(r *http.Request, req resources.BuildTxRequest) (*resources.BuildTx, interface{}, error) {
	chains := ChainsQ(r).List()

	supported := make([]interface{}, len(chains))

	for i, chain := range chains {
		switch {
		case strings.ToLower(chain.Name) == "rarimo":
			continue // rarimo is not supported yet TODO make it less crutchy
		case !slices.Contains(
			[]tokenmanager.NetworkType{
				tokenmanager.NetworkType_EVM,
				tokenmanager.NetworkType_Solana,
				tokenmanager.NetworkType_Near,
			}, chain.Type):

			continue
		default:
			supported[i] = chain.Name
		}
	}

	errs := validation.Errors{
		"data/attributes/network": validation.Validate(req.Data.Attributes.Network,
			validation.Required, validation.In(supported...)),
		"data/attributes/tx_type": validation.Validate(req.Data.Attributes.TxType,
			validation.In(resources.SupportedTxTypes()...),
		),
	}

	chain := ChainsQ(r).Get(req.Data.Attributes.Network)
	if chain == nil {
		errs["data/attributes/network"] = fmt.Errorf("network %s is not supported", req.Data.Attributes.Network)
		return nil, nil, errs.Filter()
	}

	var txData interface{}

	switch chain.Type {
	case tokenmanager.NetworkType_EVM:
		var evm resources.EthTxData

		if err := json.Unmarshal(req.Data.Attributes.TxData, &evm); err != nil {
			return nil, nil, errors.Wrap(err, "failed to unmarshal tx_data")
		}

		for k, v := range validateBuildEvmDepositTx(req.Data.Attributes.TxType, evm) {
			errs[k] = v
		}

		txData = evm
	case tokenmanager.NetworkType_Solana:
		var solana resources.SolanaTxData
		if err := json.Unmarshal(req.Data.Attributes.TxData, &solana); err != nil {
			return nil, nil, errors.Wrap(err, "failed to unmarshal tx_data")
		}

		for k, v := range validateBuildSolanaDepositTx(req.Data.Attributes.TxType, solana) {
			errs[k] = v
		}

		txData = solana
	case tokenmanager.NetworkType_Near:
		var near resources.NearTxData
		if err := json.Unmarshal(req.Data.Attributes.TxData, &near); err != nil {
			return nil, nil, errors.Wrap(err, "failed to unmarshal tx_data")
		}

		for k, v := range validateBuildNearDepositTx(req.Data.Attributes.TxType, near) {
			errs[k] = v
		}

		txData = near
	default:
		errs["data/attributes/network"] = fmt.Errorf("unsupported network type, supported are %v", resources.SupportedNetworkTypesText())
	}

	return &req.Data, txData, errs.Filter()
}

func validateBuildEvmDepositTx(txType resources.TxType, evm resources.EthTxData) validation.Errors {
	if _, err := hexutil.Decode(evm.BundleData); err != nil {
		return validation.Errors{
			"data/attributes/tx_data/bundle_data": err,
		}
	}
	if _, err := hexutil.Decode(evm.BundleSalt); err != nil {
		return validation.Errors{
			"data/attributes/tx_data/bundle_salt": err,
		}
	}

	return validation.Errors{
		"data/attributes/tx_data/receiver": validation.Validate(evm.Receiver, validation.Required),
		"data/attributes/tx_data/amount": validation.Validate(evm.Amount,
			validation.When(txType != resources.TxTypeDepositNFT, validation.Required),
			validation.When(txType == resources.TxTypeDepositNFT, validation.Nil),
		),
		"data/attributes/tx_data/token_addr": validation.Validate(evm.TokenAddr,
			validation.When(txType != resources.TxTypeDepositNative, validation.Required),
			validation.When(txType == resources.TxTypeDepositNative, validation.Nil)),
		"data/attributes/tx_data/token_id": validation.Validate(evm.TokenId,
			validation.When(txType == resources.TxTypeDepositErc721 || txType == resources.TxTypeDepositErc1155,
				validation.Required),
			validation.When(txType == resources.TxTypeDepositNative || txType == resources.TxTypeDepositErc20,
				validation.Nil)),
		"data/attributes/tx_data/is_wrapped": validation.Validate(evm.IsWrapped,
			validation.When(txType != resources.TxTypeDepositNative, validation.In(true, false)),
			validation.When(txType == resources.TxTypeDepositNative, validation.Nil)),
	}
}

func validateBuildSolanaDepositTx(txType resources.TxType, solana resources.SolanaTxData) validation.Errors {
	if _, err := hexutil.Decode(solana.BundleData); err != nil {
		return validation.Errors{
			"data/attributes/tx_data/bundle_data": err,
		}
	}
	if _, err := hexutil.Decode(solana.BundleSeed); err != nil {
		return validation.Errors{
			"data/attributes/tx_data/bundle_salt": err,
		}
	}

	return validation.Errors{
		"data/attributes/tx_data/receiver": validation.Validate(solana.Receiver, validation.Required),
		"data/attributes/tx_data/amount": validation.Validate(solana.Amount,
			validation.When(txType == resources.TxTypeDepositNative, validation.Required),
			validation.When(txType != resources.TxTypeDepositNative, validation.Nil),
		),
		"data/attributes/tx_data/token_addr": validation.Validate(solana.TokenAddr,
			validation.When(txType == resources.TxTypeDepositNFT || txType == resources.TxTypeDepositFT,
				validation.Required),
			validation.When(txType == resources.TxTypeDepositNative,
				validation.Nil)),
	}
}

func validateBuildNearDepositTx(txType resources.TxType, near resources.NearTxData) validation.Errors {
	return validation.Errors{
		"data/attributes/tx_data/sender_public_key": validation.Validate(near.SenderPublicKey, validation.Required),
		"data/attributes/tx_data/receiver":          validation.Validate(near.Receiver, validation.Required),
		"data/attributes/tx_data/amount": validation.Validate(near.Amount,
			validation.When(txType != resources.TxTypeDepositNFT, validation.Required),
			validation.When(txType == resources.TxTypeDepositNFT, validation.Nil),
		),
		"data/attributes/tx_data/token_addr": validation.Validate(near.TokenAddr,
			validation.When(txType == resources.TxTypeDepositNFT || txType == resources.TxTypeDepositFT,
				validation.Required),
			validation.When(txType == resources.TxTypeDepositNative,
				validation.Nil)),
	}
}
