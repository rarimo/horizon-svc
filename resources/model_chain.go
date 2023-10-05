/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type Chain struct {
	Key
	Attributes    ChainAttributes    `json:"attributes"`
	Relationships ChainRelationships `json:"relationships"`
}
type ChainResponse struct {
	Data     Chain    `json:"data"`
	Included Included `json:"included"`
}

type ChainListResponse struct {
	Data     []Chain         `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *ChainListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *ChainListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustChain - returns Chain from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustChain(key Key) *Chain {
	var chain Chain
	if c.tryFindEntry(key, &chain) {
		return &chain
	}
	return nil
}
