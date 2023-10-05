/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "time"

type TransferAttributes struct {
	// Amount of tokens to be transferred
	Amount *string `json:"amount,omitempty"`
	// Additional data for the transfer [(more info)](https://rarimo.gitlab.io/docs/docs/overview/bundling)
	BundleData *string `json:"bundle_data,omitempty"`
	// Salt for the transfer's bundle [(more info)](https://rarimo.gitlab.io/docs/docs/overview/bundling)
	BundleSalt *string `json:"bundle_salt,omitempty"`
	// Time (UTC) of the transfer creation, RFC3339 format
	CreatedAt time.Time `json:"created_at"`
	// Number of the event in source chain's transaction
	EventId *string `json:"event_id,omitempty"`
	// Name of the source chain
	FromChain *string `json:"from_chain,omitempty"`
	// Identifier of the transfer origin
	Origin *string `json:"origin,omitempty"`
	// Shows state of the transfer
	Status TransferState `json:"status"`
	// Name of the destination chain
	ToChain *string `json:"to_chain,omitempty"`
}
