package handlers

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/services/api/handlers/sse"
	"github.com/rarimo/horizon-svc/resources"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
)

type withdrawalByHashRequest struct {
	Hash string
}

func newWithdrawalByHashRequest(r *http.Request) (*withdrawalByHashRequest, error) {
	hash := chi.URLParam(r, "hash")

	err := validation.Errors{
		"hash": validation.Validate(hash, validation.Required),
	}.Filter()
	if err != nil {
		return nil, errors.Wrap(err, "request is invalid")
	}

	return &withdrawalByHashRequest{
		hash,
	}, nil
}

func WithdrawalByHash(w http.ResponseWriter, r *http.Request) {
	makeResponse := func() interface{} {
		req, err := newWithdrawalByHashRequest(r)
		if err != nil {
			return sse.ToErrorResponse(problems.BadRequest(err)...)
		}

		withdrawal, err := getWithdrawalByHashResponse(r, req)
		if err != nil {
			Log(r).WithError(err).Error("failed to get withdrawal by hash")
			return sse.ToErrorResponse(problems.InternalError())
		}
		if withdrawal == nil {
			return sse.ToErrorResponse(problems.NotFound())
		}

		return withdrawal
	}

	sse.ServeEvents(w, r, makeResponse)
}

func getWithdrawalByHashResponse(r *http.Request, req *withdrawalByHashRequest) (*resources.WithdrawalResponse, error) {
	transfers, err := Storage(r).TransferQ().SelectCtx(r.Context(), data.TransferSelector{
		ChainTx: &req.Hash,
	})
	if err != nil {
		panic(errors.Wrap(err, "failed to select transfers"))
	}
	if transfers == nil || len(transfers) == 0 {
		return nil, nil
	}

	transfer := transfers[0]

	withdrawal, err := Storage(r).WithdrawalQ().WithdrawalByOriginCtx(r.Context(), hexutil.MustDecode(transfer.Origin), false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select withdrawal")
	}
	if withdrawal == nil {
		return nil, nil
	}

	return &resources.WithdrawalResponse{
		Data: resources.Withdrawal{
			Key: resources.Key{
				ID:   transfer.Origin,
				Type: resources.WITHDRAWALS,
			},
			Attributes: resources.WithdrawalAttributes{
				CreatedAt: withdrawal.CreatedAt,
				Hash:      withdrawal.Hash.String,
				Origin:    transfer.Origin,
				Success:   withdrawal.Success.Bool,
			},
			Relationships: resources.WithdrawalRelationships{
				Creator: resources.Relation{
					Data: &resources.Key{
						ID:   transfer.Creator.String,
						Type: resources.ACCOUNTS,
					},
				},
				Receiver: &resources.Relation{
					Data: &resources.Key{
						ID:   transfer.Receiver,
						Type: resources.ACCOUNT_EXTERNAL_IDS,
					},
				},
				Token: &resources.Relation{
					Data: &resources.Key{
						ID:   transfer.TokenIndex,
						Type: resources.TOKENS,
					},
				},
				Tx: &resources.Relation{
					Data: &resources.Key{
						ID:   transfer.RarimoTxHash(),
						Type: resources.TRANSACTIONS,
					},
				},
			},
		},
		Included: resources.Included{},
	}, nil
}
