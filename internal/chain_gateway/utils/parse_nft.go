package utils

import (
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/rpc"
)

func ParseNftList(input []rpc.Nft) []*data.Nft {
	nfts := make([]*data.Nft, len(input))

	for i, nft := range input {
		id := nft.CollectionTokenId
		if id == "" {
			id = nft.TokenAddress
		}

		nfts[i] = &data.Nft{
			ID:             id,
			Name:           nft.Name,
			Description:    &nft.Description,
			ImageURL:       nft.ImageURL,
			CollectionName: nft.CollectionName,
			Attributes:     make([]data.NftAttribute, len(nft.Traits.Value)),
		}

		for j, trait := range nft.Traits.Value {
			nfts[i].Attributes[j] = data.NftAttribute{
				Trait: trait.TraitType,
				Value: string(*trait.Value),
			}
		}
	}

	return nfts
}
