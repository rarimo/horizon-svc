/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type ItemChainMapping struct {
	Key
	Attributes    ItemChainMappingAttributes    `json:"attributes"`
	Relationships ItemChainMappingRelationships `json:"relationships"`
}
type ItemChainMappingResponse struct {
	Data     ItemChainMapping `json:"data"`
	Included Included         `json:"included"`
}

type ItemChainMappingListResponse struct {
	Data     []ItemChainMapping `json:"data"`
	Included Included           `json:"included"`
	Links    *Links             `json:"links"`
	Meta     json.RawMessage    `json:"meta,omitempty"`
}

func (r *ItemChainMappingListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *ItemChainMappingListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustItemChainMapping - returns ItemChainMapping from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustItemChainMapping(key Key) *ItemChainMapping {
	var itemChainMapping ItemChainMapping
	if c.tryFindEntry(key, &itemChainMapping) {
		return &itemChainMapping
	}
	return nil
}
