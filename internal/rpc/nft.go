package rpc

type FetchNftResultSolana struct {
}

type FetchNftResult struct {
	ResponsePaginatedResponse
	Owner string `json:"owner,omitempty"`
	Nfts  []Nft  `json:"assets,omitempty"`
}

type Nft struct {
	Name              string `json:"name,omitempty"`
	CollectionName    string `json:"collectionName,omitempty"`
	TokenAddress      string `json:"tokenAddress,omitempty"`
	CollectionTokenId string `json:"collectionTokenId,omitempty"`
	CollectionAddress string `json:"collectionAddress,omitempty"`
	ImageURL          string `json:"imageUrl,omitempty"`
	Traits            Traits `json:"traits,omitempty"`
	Chain             string `json:"chain,omitempty"`
	Network           string `json:"network,omitempty"`
	Description       string `json:"description,omitempty"`
}
