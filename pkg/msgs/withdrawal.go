package msgs

const (
	MessageTypeWithdrawal MessageType = "bridge_withdrawal"
)

type WithdrawalMsg struct {
	Origin  string `json:"origin"`
	Hash    string `json:"hash"`
	Success bool   `json:"success"`
}

func (m WithdrawalMsg) Message() Message {
	return marshalToMsg(m, MessageTypeWithdrawal)
}
