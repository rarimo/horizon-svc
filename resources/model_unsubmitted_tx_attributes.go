/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type UnsubmittedTxAttributes struct {
	// contract to send tx to
	ContractAddr string `json:"contract_addr"`
	// hex-encoded transaction envelope
	Envelope string `json:"envelope"`
	// time when the transaction was generated
	GeneratedAt string `json:"generated_at"`
}
