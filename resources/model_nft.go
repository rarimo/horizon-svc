/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type Nft struct {
	Key
	Attributes NftAttributes `json:"attributes"`
}
type NftResponse struct {
	Data     Nft      `json:"data"`
	Included Included `json:"included"`
}

type NftListResponse struct {
	Data     []Nft           `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *NftListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *NftListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustNft - returns Nft from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustNft(key Key) *Nft {
	var nft Nft
	if c.tryFindEntry(key, &nft) {
		return &nft
	}
	return nil
}
