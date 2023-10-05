/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

// transaction parameters for solana deposit tx
type SolanaTxData struct {
	// The amount of the token to send (string containing a decimal number with precision specified by the contract)
	Amount *string `json:"amount,omitempty"`
	// bundle data as for calling Deposit* methods [(more info)](https://rarimo.gitlab.io/docs/docs/overview/bundling)
	BundleData string `json:"bundle_data"`
	// bundle seed as for calling Deposit* methods [(more info)](https://rarimo.gitlab.io/docs/docs/overview/bundling)
	BundleSeed string `json:"bundle_seed"`
	// The address of the receiver
	Receiver string `json:"receiver"`
	// Network which the transfer is to be consumed on
	TargetNetwork string `json:"target_network"`
	// [ OPTIONAL ] address that identifies the token to be deposited. If not provided then the native token is used.
	TokenAddr *string `json:"token_addr,omitempty"`
}
