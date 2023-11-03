package handlers

import (
	"math/rand"
	"net/http"

	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/horizon-svc/internal/amount"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"github.com/rarimo/horizon-svc/resources"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type balanceRequest struct {
	Chain          string `url:"chain"`
	Index          string `url:"index"`
	AccountAddress string `url:"account_address"`
	TokenID        string `url:"token_id,omitempty"`
}

func newBalanceRequest(r *http.Request) (*balanceRequest, error) {
	var request balanceRequest

	request.Chain = chi.URLParam(r, "chain")
	request.Index = chi.URLParam(r, "index")
	request.AccountAddress = chi.URLParam(r, "account_address")
	request.TokenID = r.URL.Query().Get("token_id")

	return &request, validation.Errors{
		"chain":           validation.Validate(request.Chain, validation.Required),
		"index":           validation.Validate(request.Index, validation.Required),
		"account_address": validation.Validate(request.AccountAddress, validation.Required),
	}.Filter()
}

func Balance(w http.ResponseWriter, r *http.Request) {
	req, err := newBalanceRequest(r)

	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	item, err := Core(r).Tokenmanager().GetItem(r.Context(), req.Index)
	if err != nil {
		Log(r).WithError(err).Error("failed to get item")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if item == nil {
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"index": errors.New("no item with such index"),
		})...)
		return
	}

	onChainItem := findOnChainItem(*item, req.Chain)
	if onChainItem == nil {
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"chain": errors.New("no info about chain for item"),
		})...)
		return
	}

	collectionData, err := Core(r).Tokenmanager().GetCollectionData(r.Context(), onChainItem.Chain, onChainItem.Address)
	if err != nil {
		Log(r).WithError(err).Error("failed to get collection")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if collectionData == nil {
		Log(r).WithFields(logan.F{
			"collection": item.Collection,
			"chain":      onChainItem.Chain,
			"address":    onChainItem.Address,
		}).Error("collection data missing for chain+address")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	balance, err := ProxyRepo(r).Get(req.Chain).BalanceOf(r.Context(), &types.BalanceOfOpts{
		AccountAddress: req.AccountAddress,
		Chain:          onChainItem.Chain,
		Decimals:       collectionData.Decimals,
		TokenType:      collectionData.TokenType,
		TokenAddress:   onChainItem.Address,
		TokenID:        req.TokenID,
	})
	if err != nil {
		Log(r).WithError(err).Error("failed to get balance")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if balance == nil {
		Log(r).WithFields(logan.F{
			"chain":           onChainItem.Chain,
			"account_address": req.AccountAddress,
			"token_address":   onChainItem.Address,
			"token_id":        req.TokenID,
		}).Error("no balance found")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, newBalanceResponse(balance))
}

func newBalanceResponse(amount *amount.Amount) resources.BalanceResponse {
	return resources.BalanceResponse{
		Data: resources.Balance{
			Key: resources.NewKeyInt64(rand.Int63(), resources.BALANCES), // TODO - remove random key after adding aggregation
			Attributes: resources.BalanceAttributes{
				Amount: amount.String(),
			},
		},
	}
}

func findOnChainItem(item tokenmanager.Item, chain string) *tokenmanager.OnChainItemIndex {
	for _, onChainItem := range item.OnChain {
		if onChainItem.Chain == chain {
			return onChainItem
		}
	}
	return nil
}
