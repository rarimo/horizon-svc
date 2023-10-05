package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/proxy/types"
	"github.com/rarimo/horizon-svc/resources"
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3"
)

type nftMetadataRequest struct {
	Chain      string `url:"chain"`
	TokenIndex string `url:"token_index"`
	TokenID    string `url:"token_id"`
}

func newNftMetadataRequest(r *http.Request) (*nftMetadataRequest, error) {
	var request nftMetadataRequest

	request.Chain = chi.URLParam(r, "chain")
	request.TokenIndex = chi.URLParam(r, "token_index")
	request.TokenID = chi.URLParam(r, "token_id")

	return &request, validation.Errors{
		"chain":       validation.Validate(request.Chain, validation.Required),
		"token_index": validation.Validate(request.TokenIndex, validation.Required),
		"token_id":    validation.Validate(request.TokenID, validation.Required),
	}.Filter()
}

func NftMetadata(w http.ResponseWriter, r *http.Request) {
	req, err := newNftMetadataRequest(r)
	if err != nil {
		Log(r).WithError(err).Debug("failed to decode request")
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

	metadata, err := ProxyRepo(r).Get(req.Chain).NftMetadata(r.Context(), &types.NftMetadataOpts{
		Chain:        onChainItem.Chain,
		TokenType:    collectionData.TokenType,
		TokenAddress: onChainItem.Address,
		TokenID:      onChainItem.TokenID,
	})

	if err != nil {
		if errors.Cause(err) == types.ErrorNotFound {
			ape.RenderErr(w, problems.NotFound())
			return
		}

		Log(r).WithError(err).Error("failed to get metadata")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if metadata == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, newNftMetadataResponse(req.TokenID, metadata))
}

func newNftMetadataModel(tokenId string, value *data.NftMetadata) *resources.NftMetadata {
	result := resources.NftMetadata{
		Key: resources.NewStringKey(tokenId, resources.NFTS_METADATA),
		Attributes: resources.NftMetadataAttributes{
			Name:       value.Name,
			ImageUrl:   value.ImageURL,
			Attributes: make([]resources.NftMetadataAttribute, 0, len(value.Attributes)),
		},
	}

	result.Attributes.Description = data.GetOrDefaultStrPtr(value.Description, nil)
	result.Attributes.AnimationUrl = data.GetOrDefaultStrPtr(value.AnimationUrl, nil)
	result.Attributes.ExternalUrl = data.GetOrDefaultStrPtr(value.ExternalUrl, nil)

	if value.MetadataUrl != nil && *value.MetadataUrl != "" {
		result.Attributes.MetadataUrl = *value.MetadataUrl
	}

	for _, attr := range value.Attributes {
		result.Attributes.Attributes = append(result.Attributes.Attributes, resources.NftMetadataAttribute{
			TraitType: attr.Trait,
			Value:     attr.Value,
		})
	}

	return &result
}

func newNftMetadataResponse(tokenId string, model *data.NftMetadata) resources.NftMetadataResponse {
	return resources.NftMetadataResponse{
		Data: *newNftMetadataModel(tokenId, model),
	}
}

var nftTypes = map[tokenmanager.Type]interface{}{
	tokenmanager.Type_ERC721:       nil,
	tokenmanager.Type_ERC1155:      nil,
	tokenmanager.Type_NEAR_NFT:     nil,
	tokenmanager.Type_METAPLEX_NFT: nil,
}

func isNftTokenType(tokenType tokenmanager.Type) bool {
	_, ok := nftTypes[tokenType]
	return ok
}
