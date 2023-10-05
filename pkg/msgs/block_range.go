package msgs

import (
	"encoding/json"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

const MessageTypeBlockRange MessageType = "block_range"

type BlockRangeMessage struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

func (br BlockRangeMessage) Message() Message {
	return marshalToMsg(br, MessageTypeBlockRange)
}

type rawMessage struct {
	Raw  interface{} `json:"raw"`
	Type MessageType `json:"type"`
}

func marshalToMsg(v interface{}, msgType MessageType) Message {
	raw, err := json.Marshal(rawMessage{
		Raw:  v,
		Type: msgType,
	})
	if err != nil {
		panic(errors.Wrap(err, "failed to marshal message"))
	}

	return Message{raw: raw, typ: msgType}
}
