package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/resources"
	tokentypes "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/urlval"
)

type chainListRequest struct {
	pgdb.OffsetPageParams
	IncludeItems bool `include:"items"`
}

func newChainListRequest(r *http.Request) (*chainListRequest, error) {
	var request chainListRequest

	err := urlval.Decode(r.URL.Query(), &request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode url")
	}

	if request.Order != "desc" {
		return nil, validation.Errors{
			"page[order]": errors.New("parameter is not supported for this endpoint"),
		}
	}

	return &request, nil
}

func ChainList(w http.ResponseWriter, r *http.Request) {
	req, err := newChainListRequest(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	chains := ChainsQ(r).Page(int(req.PageNumber), int(req.Limit))

	response := resources.ChainListResponse{
		Data:     make([]resources.Chain, len(chains)),
		Included: resources.Included{},
	}

	if len(chains) == 0 {
		ape.Render(w, response)
		return
	}

	params, err := Core(r).TokenManager().GetParams(r.Context())
	if err != nil {
		Log(r).WithError(err).Error("failed to get params from token manager")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if len(params.Networks) != len(chains) {
		Log(r).WithFields(logan.F{
			"params.networks": fmt.Sprintf("%+v", params.Networks),
			"chains":          fmt.Sprintf("%+v", chains),
		}).Error("mismatched number of chains in horizon and in core")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	networksMap := make(map[string]tokentypes.Network)

	for _, net := range params.Networks {
		chain := ChainsQ(r).Get(net.Name)
		if chain == nil {
			Log(r).WithFields(logan.F{"network": net.Name}).Info("chain not found")
			continue
		}
		networksMap[net.Name] = *net
	}

	for i, c := range chains {
		net := networksMap[c.Name]

		// TODO figure out how to save data models to cache and use CachedStorage
		icms, err := Storage(r).
			ItemChainMappingQ().
			ItemChainMappingsByNetworkCtx(r.Context(), c.ID, false)
		if err != nil {
			Log(r).WithError(err).Error("failed to select item chain mappings by network")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		itemKeys := make([]resources.Key, len(icms))
		for i, icm := range icms {
			itemKeys[i] = resources.Key{
				ID:   strconv.FormatInt(icm.Item, 10),
				Type: resources.ITEMS,
			}

			if req.IncludeItems {
				// TODO figure out how to save data models to cache and use CachedStorage
				item, err := Storage(r).ItemQ().ItemByIDCtx(r.Context(), icm.Item, false)
				if err != nil {
					Log(r).WithError(err).WithField("id", icm.Item).Error("failed to get item")
					ape.RenderErr(w, problems.InternalError())
					return
				}

				if item == nil {
					Log(r).WithFields(logan.F{
						"item": icm.Item,
					}).Warn("item not found")
					continue
				}

				response.Included.Add(itemResourceP(*item))
			}
		}

		bridgeParams := net.GetBridgeParams()

		if bridgeParams == nil {
			Log(r).WithFields(logan.F{
				"network": net.Name,
			}).Debug("bridge params not found")
			continue
		}

		response.Data[i] = resources.Chain{
			Key: resources.Key{
				ID:   strconv.FormatInt(int64(c.ID), 10),
				Type: resources.CHAINS,
			},
			Attributes: resources.ChainAttributes{
				BridgeContract: bridgeParams.Contract,
				ChainParams:    c.ChainParams,
				ChainType:      c.Type,
				Icon:           c.Icon,
				Name:           c.Name,
			},
			Relationships: resources.ChainRelationships{
				Items: resources.RelationCollection{
					Data: itemKeys,
				},
			},
		}
	}

	ape.Render(w, response)
}

func itemResourceP(item data.Item) *resources.Item {
	return &resources.Item{
		Key: resources.Key{
			ID:   strconv.FormatInt(item.ID, 10),
			Type: resources.ITEMS,
		},
		Attributes: resources.ItemAttributes{
			Index:    string(item.Index),
			Metadata: json.RawMessage(item.Metadata),
		},
		Relationships: resources.ItemRelationships{
			Collection: resources.Relation{
				Data: &resources.Key{
					ID:   strconv.FormatInt(item.Collection.Int64, 10),
					Type: resources.COLLECTIONS,
				},
			},
		},
	}
}
