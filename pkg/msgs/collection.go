package msgs

import tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"

const (
	MessageTypeCollectionCreated     MessageType = tokenmanager.EventTypeCollectionCreated
	MessageTypeCollectionRemoved     MessageType = tokenmanager.EventTypeCollectionRemoved
	MessageTypeCollectionDataCreated MessageType = tokenmanager.EventTypeCollectionDataCreated
	MessageTypeCollectionDataRemoved MessageType = tokenmanager.EventTypeCollectionDataRemoved
	MessageTypeCollectionDataUpdated MessageType = tokenmanager.EventTypeCollectionDataUpdated
)

type CollectionCreatedMessage struct {
	Index string `json:"index"`
}

func (m CollectionCreatedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeCollectionCreated)
}

type CollectionRemovedMessage struct {
	Index string `json:"index"`
}

func (m CollectionRemovedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeCollectionRemoved)
}

type CollectionDataCreatedMessage struct {
	CollectionIndex string `json:"index"`
	Chain           string `json:"chain"`
}

func (m CollectionDataCreatedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeCollectionDataCreated)
}

type CollectionDataRemovedMessage struct {
	CollectionIndex string `json:"index"`
	Chain           string `json:"chain"`
}

func (m CollectionDataRemovedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeCollectionDataRemoved)
}

type CollectionDataUpdatedMessage struct {
	CollectionIndex string `json:"index"`
	Chain           string `json:"chain"`
}

func (m CollectionDataUpdatedMessage) Message() Message {
	return marshalToMsg(m, MessageTypeCollectionDataUpdated)
}
