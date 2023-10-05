package handlers

import (
	"fmt"
	"net/http"

	"gitlab.com/distributed_lab/urlval"

	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"github.com/rarimo/horizon-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type nftListRequest struct {
	Chain          string `url:"chain"`
	TokenIndex     string `url:"token_index"`
	AccountAddress string `url:"account_address"`

	RawCursor *string `page:"cursor"`
	Cursor    *data.Cursor
	Limit     uint64 `page:"limit" default:"15"`
}

func newNftListRequest(r *http.Request) (*nftListRequest, error) {
	var request nftListRequest

	err := urlval.Decode(r.URL.Query(), &request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode url")
	}

	request.Chain = chi.URLParam(r, "chain")
	request.TokenIndex = chi.URLParam(r, "token_index")
	request.AccountAddress = chi.URLParam(r, "account_address")

	err = validation.Errors{
		"chain":           validation.Validate(request.Chain, validation.Required),
		"token_index":     validation.Validate(request.TokenIndex, validation.Required),
		"account_address": validation.Validate(request.AccountAddress, validation.Required),
		"page[limit]":     validatePageSize(request.Limit),
		"page[cursor]":    validation.Validate(request.Cursor, validation.NilOrNotEmpty),
	}.Filter()
	if err != nil {
		return nil, errors.Wrap(err, "request is invalid")
	}

	if request.RawCursor != nil && *request.RawCursor != "" {
		if request.Cursor, err = data.DecodeCursor(*request.RawCursor); err != nil {
			return nil, validation.Errors{
				"page[cursor]": err,
			}
		}
	}

	return &request, nil
}

func NftList(w http.ResponseWriter, r *http.Request) {
	req, err := newNftListRequest(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	item, err := Core(r).TokenManager().GetItem(r.Context(), req.TokenIndex)
	if err != nil {
		Log(r).WithError(err).Error("failed to get item from token manager")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if item == nil {
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"token_index": errors.New("no item with such index"),
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

	collectionData, err := Core(r).TokenManager().CollectionData(r.Context(), onChainItem.Chain, onChainItem.Address)
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

	if !isNftTokenType(collectionData.TokenType) {
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"token_index": errors.Wrap(err, "item is not of NFT type"),
		})...)
		return
	}

	opts := &types.NftsOpts{
		AccountAddress: req.AccountAddress,
		Chain:          req.Chain,
		TokenAddress:   onChainItem.Address,
		Limit:          req.Limit,
	}
	if req.Cursor != nil {
		opts.PageNumber = &req.Cursor.PageNumber
		opts.ItemIndex = &req.Cursor.ItemIndex
	}

	nfts, cursor, err := ProxyRepo(r).Get(req.Chain).NftList(r.Context(), opts)
	if err != nil {
		Log(r).WithError(err).Debug("failed to get nfts")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	resp := resources.NftListResponse{
		Data: make([]resources.Nft, len(nfts)),
		Links: &resources.Links{
			Self: fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(req)),
		},
	}

	if len(nfts) == 0 {
		ape.Render(w, resp)
		return
	}

	for i, nft := range nfts {
		resp.Data[i] = *newNftModel(*nft)
	}

	if cursor != nil {
		c := cursor.String()
		req.RawCursor = &c
		resp.Links.Next = fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(req))
	}

	ape.Render(w, resp)
}

func newNftModel(value data.Nft) *resources.Nft {
	result := &resources.Nft{
		Key: resources.NewStringKey(value.ID, resources.NFTS),
		Attributes: resources.NftAttributes{
			CollectionName: value.CollectionName,
			Description:    value.Description,
			ImageUrl:       value.ImageURL,
			Name:           value.Name,
			Attributes:     make([]resources.NftMetadataAttribute, len(value.Attributes)),
		},
	}

	for i, attr := range value.Attributes {
		result.Attributes.Attributes[i] = resources.NftMetadataAttribute{
			TraitType: attr.Trait,
			Value:     attr.Value,
		}
	}

	return result
}
