/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type TransferWithdrawal struct {
	Key
	Attributes    TransferWithdrawalAttributes    `json:"attributes"`
	Relationships TransferWithdrawalRelationships `json:"relationships"`
}
type TransferWithdrawalResponse struct {
	Data     TransferWithdrawal `json:"data"`
	Included Included           `json:"included"`
}

type TransferWithdrawalListResponse struct {
	Data     []TransferWithdrawal `json:"data"`
	Included Included             `json:"included"`
	Links    *Links               `json:"links"`
	Meta     json.RawMessage      `json:"meta,omitempty"`
}

func (r *TransferWithdrawalListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *TransferWithdrawalListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustTransferWithdrawal - returns TransferWithdrawal from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustTransferWithdrawal(key Key) *TransferWithdrawal {
	var transferWithdrawal TransferWithdrawal
	if c.tryFindEntry(key, &transferWithdrawal) {
		return &transferWithdrawal
	}
	return nil
}
