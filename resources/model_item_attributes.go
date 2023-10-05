/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type ItemAttributes struct {
	// unique index of the item saved on core
	Index string `json:"index"`
	// free form JSON object representing item's metadata saved on core
	Metadata json.RawMessage `json:"metadata"`
}
