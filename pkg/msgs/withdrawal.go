package msgs

import "github.com/rarimo/horizon-svc/internal/data"

const (
	MessageTypeWithdrawal MessageType = "bridge_withdrawal"
)

type WithdrawalMsg struct {
	Hash        data.WithdrawalID `json:"hash"`
	BlockHeight int64             `json:"block_height"`
	TxResult    []byte            `json:"tx_result"`
	Success     bool              `json:"success"`
}

func (m WithdrawalMsg) Message() Message {
	return marshalToMsg(m, MessageTypeWithdrawal)
}
