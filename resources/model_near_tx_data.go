/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

// transaction parameters for ethereum deposit tx
type NearTxData struct {
	// The amount to be transferred (string containing a decimal number with precision specified by the contract)
	Amount *string `json:"amount,omitempty"`
	// indicates that the deposited token is wrapped
	IsWrapped *bool `json:"is_wrapped,omitempty"`
	// The address of the receiver
	Receiver string `json:"receiver"`
	// The base64-encoded public key of the sender: base64(publicKeyBase58)
	SenderPublicKey string `json:"sender_public_key"`
	// Network which the transfer is to be consumed on
	TargetNetwork string `json:"target_network"`
	// [ OPTIONAL ] should be provided in case of FT and NFT deposit.
	TokenAddr *string `json:"token_addr,omitempty"`
	// [ OPTIONAL ] token identifier. Shpuld be provided in case of NFT deposit.
	TokenId *string `json:"token_id,omitempty"`
}
