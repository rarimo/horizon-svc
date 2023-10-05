/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type Token struct {
	Key
	Attributes    TokenAttributes    `json:"attributes"`
	Relationships TokenRelationships `json:"relationships"`
}
type TokenResponse struct {
	Data     Token    `json:"data"`
	Included Included `json:"included"`
}

type TokenListResponse struct {
	Data     []Token         `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *TokenListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *TokenListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustToken - returns Token from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustToken(key Key) *Token {
	var token Token
	if c.tryFindEntry(key, &token) {
		return &token
	}
	return nil
}
