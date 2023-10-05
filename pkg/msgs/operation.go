package msgs

import rarimocore "github.com/rarimo/rarimo-core/x/rarimocore/types"

const (
	MessageTypeTransferOp     MessageType = rarimocore.EventTypeNewOperation
	MessageTypeConfirmationOp MessageType = rarimocore.EventTypeNewConfirmation
)

type TransferOpMsg struct {
	TransferID      string `json:"transfer_id"`
	TransactionHash string `json:"transaction_hash"`
}

func (m TransferOpMsg) Message() Message {
	return marshalToMsg(m, MessageTypeTransferOp)
}

type ConfirmationOpMsg struct {
	ConfirmationID  string `json:"confirmation_id"`
	TransactionHash string `json:"transaction_hash"`
}

func (m ConfirmationOpMsg) Message() Message {
	return marshalToMsg(m, MessageTypeConfirmationOp)
}
