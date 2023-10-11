/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "time"

type TransferWithdrawalAttributes struct {
	// Time (UTC) of the transfer withdrawal creation, RFC3339 format
	CreatedAt time.Time `json:"created_at"`
	// Identifier of the transfer withdrawal transaction
	Hash string `json:"hash"`
	// Identifier of the transfer origin
	Origin string `json:"origin"`
	// Whether the transfer withdrawal was successful
	Success bool `json:"success"`
}
