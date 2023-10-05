/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type Item struct {
	Key
	Attributes    ItemAttributes    `json:"attributes"`
	Relationships ItemRelationships `json:"relationships"`
}
type ItemResponse struct {
	Data     Item     `json:"data"`
	Included Included `json:"included"`
}

type ItemListResponse struct {
	Data     []Item          `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *ItemListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *ItemListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustItem - returns Item from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustItem(key Key) *Item {
	var item Item
	if c.tryFindEntry(key, &item) {
		return &item
	}
	return nil
}
