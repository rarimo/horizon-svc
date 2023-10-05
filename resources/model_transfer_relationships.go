/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type TransferRelationships struct {
	Creator  Relation  `json:"creator"`
	Receiver *Relation `json:"receiver,omitempty"`
	Token    *Relation `json:"token,omitempty"`
	Tx       *Relation `json:"tx,omitempty"`
}
