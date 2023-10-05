package msgs

import rarimocore "github.com/rarimo/rarimo-core/x/rarimocore/types"

const (
	MessageTypeVoteOp      MessageType = rarimocore.EventTypeVoted
	MessageTypeApprovalOp  MessageType = rarimocore.EventTypeOperationApproved
	MessageTypeRejectionOp MessageType = rarimocore.EventTypeOperationRejected
)

type VoteOpMsg struct {
	OperationID     string `json:"operation_id"`
	OperationType   string `json:"operation_type"`
	VotingChoice    string `json:"voting_choice"`
	TransactionHash string `json:"transaction_hash"`
}

func (m VoteOpMsg) Message() Message {
	return marshalToMsg(m, MessageTypeVoteOp)
}

type ApprovalOpMsg struct {
	OperationID     string `json:"operation_id"`
	OperationType   string `json:"operation_type"`
	TransactionHash string `json:"transaction_hash"`
}

func (m ApprovalOpMsg) Message() Message {
	return marshalToMsg(m, MessageTypeApprovalOp)
}

type RejectionOpMsg struct {
	OperationID     string `json:"operation_id"`
	OperationType   string `json:"operation_type"`
	TransactionHash string `json:"transaction_hash"`
}

func (m RejectionOpMsg) Message() Message {
	return marshalToMsg(m, MessageTypeRejectionOp)
}
