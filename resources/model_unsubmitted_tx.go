/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type UnsubmittedTx struct {
	Key
	Attributes UnsubmittedTxAttributes `json:"attributes"`
}
type UnsubmittedTxResponse struct {
	Data     UnsubmittedTx `json:"data"`
	Included Included      `json:"included"`
}

type UnsubmittedTxListResponse struct {
	Data     []UnsubmittedTx `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *UnsubmittedTxListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *UnsubmittedTxListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustUnsubmittedTx - returns UnsubmittedTx from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustUnsubmittedTx(key Key) *UnsubmittedTx {
	var unsubmittedTx UnsubmittedTx
	if c.tryFindEntry(key, &unsubmittedTx) {
		return &unsubmittedTx
	}
	return nil
}
