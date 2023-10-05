/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type Transfer struct {
	Key
	Attributes    TransferAttributes    `json:"attributes"`
	Relationships TransferRelationships `json:"relationships"`
}
type TransferResponse struct {
	Data     Transfer `json:"data"`
	Included Included `json:"included"`
}

type TransferListResponse struct {
	Data     []Transfer      `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *TransferListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *TransferListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustTransfer - returns Transfer from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustTransfer(key Key) *Transfer {
	var transfer Transfer
	if c.tryFindEntry(key, &transfer) {
		return &transfer
	}
	return nil
}
