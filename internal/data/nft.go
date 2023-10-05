package data

// DEPRECATED
// TODO use aggregated Items instead of Nft and make new endpoints
type Nft struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    *string        `json:"description"`
	ImageURL       string         `json:"image_url"`
	CollectionName string         `json:"collection_name,omitempty"`
	Attributes     []NftAttribute `json:"attributes,omitempty"`
}
