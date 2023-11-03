package handlers

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.com/distributed_lab/kit/pgdb"

	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/urlval"
)

type transferListRequest struct {
	Origin    *string                  `filter:"origin"`
	RarimoTx  *string                  `filter:"rarimo_tx"`
	ChainTx   *string                  `filter:"chain_tx"`
	FromChain *string                  `filter:"from_chain"`
	ToChain   *string                  `filter:"to_chain"`
	Receiver  *string                  `filter:"receiver"`
	Status    *resources.TransferState `filter:"status"`
	Creator   *string                  `filter:"creator"`
	Before    *int64                   `filter:"bridged_before"`
	After     *int64                   `filter:"bridged_after"`
	ItemIndex *string                  `filter:"rarimo_item_index"`

	PageCursor uint64     `page:"cursor"`
	PageLimit  uint64     `page:"limit" default:"15"`
	Sorts      pgdb.Sorts `url:"sort" default:"-time"`
}

func newTransferListRequest(r *http.Request) (*transferListRequest, error) {
	var result transferListRequest

	err := urlval.Decode(r.URL.Query(), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func TransferList(w http.ResponseWriter, r *http.Request) {
	request, err := newTransferListRequest(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	// TODO should it be authorized ?
	//  how to authorize if so?
	// TODO figure out how to save data models to cache and use CachedStorage
	transfers, err := Storage(r).TransferQ().SelectCtx(r.Context(), createSelector(*request))

	if err != nil {
		panic(errors.Wrap(err, "failed to select transfers"))
	}

	response := resources.TransferListResponse{
		Data:     make([]resources.Transfer, 0, len(transfers)),
		Included: resources.Included{},
		Links: &resources.Links{
			Self: fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(request)),
		},
	}

	if len(transfers) == 0 {
		ape.Render(w, response)
		return
	}

	request.PageCursor = uint64(transfers[len(transfers)-1].ID)
	response.Links.Next = fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(request))

	_ = response.PutMeta(map[string]interface{}{
		"next_cursor": transfers[len(transfers)-1].ID,
	})

	for _, transfer := range transfers {
		response.Data = append(response.Data, mustToTransferResource(transfer))
	}

	ape.Render(w, response)
}

func createSelector(request transferListRequest) data.TransferSelector {
	sel := data.TransferSelector{
		Origin:           request.Origin,
		RarimoTx:         request.RarimoTx,
		ChainTx:          request.ChainTx,
		SourceChain:      request.FromChain,
		DestinationChain: request.ToChain,
		Receiver:         request.Receiver,
		Creator:          request.Creator,
		ItemIndex:        request.ItemIndex,
		PageCursor:       request.PageCursor,
		PageSize:         request.PageLimit,
		Sort:             request.Sorts,
	}

	if request.Status != nil {
		sel.Status = request.Status.Intp()
	}

	if request.Before != nil {
		before := time.Unix(*request.Before, 0)
		sel.Before = &before
	}

	if request.After != nil {
		after := time.Unix(*request.After, 0)
		sel.After = &after
	}

	return sel
}
