/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "github.com/ethereum/go-ethereum/common"

// transaction parameters for ethereum deposit tx
type EthTxData struct {
	// The amount of the token to send (string containing a decimal number with precision specified by the contract)
	Amount *string `json:"amount,omitempty"`
	// bundle data as for calling Deposit* methods [(more info)](https://rarimo.gitlab.io/docs/docs/overview/bundling)
	BundleData string `json:"bundle_data"`
	// bundle salt as for calling Deposit* methods [(more info)](https://rarimo.gitlab.io/docs/docs/overview/bundling)
	BundleSalt string `json:"bundle_salt"`
	// indicates that the deposited token is wrapped
	IsWrapped *bool `json:"is_wrapped,omitempty"`
	// The address of the receiver
	Receiver string `json:"receiver"`
	// Network which the transfer is to be consumed on
	TargetNetwork string `json:"target_network"`
	// [ OPTIONAL ] contract address that identifies the token to be deposited. If not provided then the native token is used.
	TokenAddr *common.Address `json:"token_addr,omitempty"`
	// [ OPTIONAL ] hex-encoded token identifier. Should be provided if contract is provided and is of type ERC721 or ERC1155.
	TokenId *string `json:"token_id,omitempty"`
}
