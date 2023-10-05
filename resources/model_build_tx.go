/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type BuildTx struct {
	Key
	Attributes    BuildTxAttributes    `json:"attributes"`
	Relationships BuildTxRelationships `json:"relationships"`
}
type BuildTxRequest struct {
	Data     BuildTx  `json:"data"`
	Included Included `json:"included"`
}

type BuildTxListRequest struct {
	Data     []BuildTx       `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *BuildTxListRequest) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *BuildTxListRequest) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustBuildTx - returns BuildTx from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustBuildTx(key Key) *BuildTx {
	var buildTx BuildTx
	if c.tryFindEntry(key, &buildTx) {
		return &buildTx
	}
	return nil
}
