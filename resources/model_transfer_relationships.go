/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type TransferRelationships struct {
	Creator  Relation  `json:"creator"`
	Item     *Relation `json:"item,omitempty"`
	Receiver *Relation `json:"receiver,omitempty"`
	Tx       *Relation `json:"tx,omitempty"`
}
