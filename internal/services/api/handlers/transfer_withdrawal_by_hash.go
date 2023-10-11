package handlers

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/services/api/handlers/sse"
	"github.com/rarimo/horizon-svc/resources"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
	"time"
)

const (
	RenderTimeout = 10 * time.Second
)

type transferWithdrawalByHashRequest struct {
	Hash string
}

func newTransferWithdrawalByHashRequest(r *http.Request) (*transferWithdrawalByHashRequest, error) {
	id := chi.URLParam(r, "hash")

	err := validation.Errors{
		"hash": validation.Validate(id, validation.Required),
	}.Filter()
	if err != nil {
		return nil, errors.Wrap(err, "request is invalid")
	}

	return &transferWithdrawalByHashRequest{
		Hash: id,
	}, nil
}

func TransferWithdrawalByHash(w http.ResponseWriter, r *http.Request) {
	req, err := newTransferWithdrawalByHashRequest(r)
	if err != nil {
		sse.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	renderTransferWithdrawal(w, r, req)

	for {
		renderTransferWithdrawal(w, r, req)
		w.(http.Flusher).Flush()

		// Check for client disconnection using the context
		select {
		case <-r.Context().Done():
			return
		default:
		}

		time.Sleep(RenderTimeout)
	}

}

func renderTransferWithdrawal(w http.ResponseWriter, r *http.Request, req *transferWithdrawalByHashRequest) {
	withdrawal, err := getTransferWithdrawalByHash(r, req)
	if err != nil {
		Log(r).WithError(err).Error("failed to get transfer withdrawal by hash")
		sse.RenderErr(w, problems.InternalError())
		return
	}
	if withdrawal == nil {
		Log(r).WithFields(logan.F{"hash": req.Hash}).Error("transfer withdrawal not found")
		sse.RenderErr(w, problems.NotFound())
	} else {
		Log(r).WithFields(logan.F{"hash": req.Hash}).Error("found transfer withdrawal")
		sse.Render(w, withdrawal)
		return
	}
}

func getTransferWithdrawalByHash(r *http.Request, req *transferWithdrawalByHashRequest) (*resources.TransferWithdrawalResponse, error) {
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

	return &resources.TransferWithdrawalResponse{
		Data: resources.TransferWithdrawal{
			Key: resources.Key{
				ID:   transfer.Origin,
				Type: resources.TRANSFER_WITHDRAWALS,
			},
			Attributes: resources.TransferWithdrawalAttributes{
				CreatedAt: withdrawal.CreatedAt,
				Hash:      withdrawal.Hash.String,
				Origin:    transfer.Origin,
				Success:   withdrawal.Success.Bool,
			},
			Relationships: resources.TransferWithdrawalRelationships{
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
