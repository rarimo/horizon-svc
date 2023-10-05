package msgs

import tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"

const (
	MessageTypeSeedCreated MessageType = tokenmanager.EventTypeSeedCreated
	MessageTypeSeedRemoved MessageType = tokenmanager.EventTypeSeedRemoved
)

type SeedCreatedMessage struct {
	Seed      string `json:"seed"`
	ItemIndex string `json:"index"`
	TxHash    string `json:"tx_hash"`
}

func (m SeedCreatedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeSeedCreated)
}

type SeedRemovedMessage struct {
	Seed      string `json:"seed"`
	ItemIndex string `json:"index"`
	TxHash    string `json:"tx_hash"`
}

func (m SeedRemovedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeSeedRemoved)
}
