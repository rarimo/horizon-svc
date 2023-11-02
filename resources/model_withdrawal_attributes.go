/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "time"

type WithdrawalAttributes struct {
	// Time (UTC) of the withdrawal creation, RFC3339 format
	CreatedAt time.Time `json:"created_at"`
	// Identifier of the withdrawal transaction
	Hash string `json:"hash"`
	// Identifier of the withdrawal origin (equals to the transfer origin)
	Origin string `json:"origin"`
	// Whether the withdrawal was successful
	Success bool `json:"success"`
}
