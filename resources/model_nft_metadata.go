/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type NftMetadata struct {
	Key
	Attributes NftMetadataAttributes `json:"attributes"`
}
type NftMetadataResponse struct {
	Data     NftMetadata `json:"data"`
	Included Included    `json:"included"`
}

type NftMetadataListResponse struct {
	Data     []NftMetadata   `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *NftMetadataListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *NftMetadataListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustNftMetadata - returns NftMetadata from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustNftMetadata(key Key) *NftMetadata {
	var nftMetadata NftMetadata
	if c.tryFindEntry(key, &nftMetadata) {
		return &nftMetadata
	}
	return nil
}
