package handlers

import (
	"net/http"
	"strconv"

	"github.com/rarimo/horizon-svc/internal/data"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/horizon-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type transferByIDRequest struct {
	ID string
}

func newByIDRequest(r *http.Request) (*transferByIDRequest, error) {
	id, err := hexutil.Decode(chi.URLParam(r, "id"))
	if err != nil {
		return nil, validation.Errors{
			"id": err,
		}
	}

	return &transferByIDRequest{
		ID: hexutil.Encode(id),
	}, nil
}

func TransferByID(w http.ResponseWriter, r *http.Request) {
	request, err := newByIDRequest(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	// TODO check ID format
	// TODO figure out how to save data models to cache and use CachedStorage
	transfer, err := Storage(r).TransferQ().TransferByIndexCtx(r.Context(), []byte(request.ID), false)
	if err != nil {
		panic(errors.Wrap(err, "failed to select transfers"))
	}

	if transfer == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, resources.TransferResponse{
		Data:     mustToTransferResource(*transfer),
		Included: resources.Included{},
	})
}

func mustToTransferResource(transfer data.Transfer) resources.Transfer {
	amount := transfer.Amount.String()
	bundleData := string(transfer.BundleData)
	bundleSalt := string(transfer.BundleSalt)
	eventID := strconv.FormatInt(transfer.EventID, 10)
	state, ok := resources.TransferStateFromInt(transfer.Status)
	if !ok {
		panic(errors.New("invalid transfer state"))
	}

	if !transfer.Creator.Valid {
		panic(errors.New("transfer creator invalid in db"))
	}

	return resources.Transfer{
		Key: resources.Key{
			ID:   strconv.FormatInt(transfer.ID, 10),
			Type: resources.TRANSFERS,
		},
		Attributes: resources.TransferAttributes{
			Amount:     &amount,
			BundleData: &bundleData,
			BundleSalt: &bundleSalt,
			CreatedAt:  transfer.CreatedAt,
			EventId:    &eventID,
			FromChain:  &transfer.FromChain,
			ToChain:    &transfer.ToChain,
			Origin:     &transfer.Origin,
			Status:     state,
		},
		Relationships: resources.TransferRelationships{
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
	}
}
