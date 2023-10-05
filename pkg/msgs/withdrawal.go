package msgs

const (
	MessageTypeWithdrawal MessageType = "bridge_withdrawal"
)

type WithdrawalMsg struct {
	Hash        string `json:"hash"`
	BlockHeight int64  `json:"block_height"`
	TxResult    []byte `json:"tx_result"`
	Success     bool   `json:"success"`
}

func (m WithdrawalMsg) Message() Message {
	return marshalToMsg(m, MessageTypeWithdrawal)
}
