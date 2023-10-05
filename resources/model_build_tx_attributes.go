/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type BuildTxAttributes struct {
	// network to send tx to
	Network string          `json:"network"`
	TxData  json.RawMessage `json:"tx_data"`
	// Type of the transaction
	TxType TxType `json:"tx_type"`
}
