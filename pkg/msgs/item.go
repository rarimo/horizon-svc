package msgs

import tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"

const (
	MessageTypeItemCreated            MessageType = tokenmanager.EventTypeItemCreated
	MessageTypeItemRemoved            MessageType = tokenmanager.EventTypeItemRemoved
	MessageTypeItemOnChainDataCreated MessageType = tokenmanager.EventTypeOnChainItemCreated
	MessageTypeItemOnChainDataRemoved MessageType = tokenmanager.EventTypeOnChainItemRemoved
)

type ItemCreatedMessage struct {
	Index string `json:"index"`
}

func (m ItemCreatedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeItemCreated)
}

type ItemRemovedMessage struct {
	Index string `json:"index"`
}

func (m ItemRemovedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeItemRemoved)
}

type ItemOnChainDataCreatedMessage struct {
	ItemIndex string `json:"index"`
	Chain     string `json:"chain"`
}

func (m ItemOnChainDataCreatedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeItemOnChainDataCreated)
}

type ItemOnChainDataRemovedMessage struct {
	ItemIndex string `json:"index"`
	Chain     string `json:"chain"`
}

func (m ItemOnChainDataRemovedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeItemOnChainDataRemoved)
}
