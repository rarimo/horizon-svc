package msgs

import (
	"encoding/json"

	"gitlab.com/distributed_lab/logan/v3"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

// TODO this whole file is a mess, need to refactor it

type MessageType string

type Message struct {
	raw json.RawMessage
	typ MessageType
}

func (m Message) String() string {
	return string(m.raw)
}

func (m *Message) Type() MessageType {
	if m.typ != "" {
		return m.typ
	}

	var raw rawMessage
	if err := json.Unmarshal(m.raw, &raw); err != nil {
		panic(errors.Wrap(err, "failed to unmarshal message", logan.F{
			"raw": string(m.raw),
		}))
	}

	m.typ = raw.Type
	return m.typ
}

func (m *Message) MustBlockRangeMessage() (msg BlockRangeMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeBlockRange)
	m.typ = rm.Type
	return
}

func (m *Message) MustTransferOpMessage() (msg TransferOpMsg) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeTransferOp)
	m.typ = rm.Type
	return
}

func (m *Message) MustConfirmationOpMessage() (msg ConfirmationOpMsg) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeConfirmationOp)
	m.typ = rm.Type
	return
}

func (m *Message) MustVoteOpMessage() (msg VoteOpMsg) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeVoteOp)
	m.typ = rm.Type
	return
}

func (m *Message) MustApprovalOpMessage() (msg ApprovalOpMsg) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeApprovalOp)
	m.typ = rm.Type
	return
}

func (m *Message) MustRejectionOpMessage() (msg RejectionOpMsg) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeRejectionOp)
	m.typ = rm.Type
	return
}

func (m *Message) MustCollectionCreatedMessage() (msg CollectionCreatedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeCollectionCreated)
	m.typ = rm.Type
	return
}

func (m *Message) MustCollectionRemovedMessage() (msg CollectionRemovedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeCollectionRemoved)
	m.typ = rm.Type
	return
}

func (m *Message) MustCollectionDataCreatedMessage() (msg CollectionDataCreatedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeCollectionDataCreated)
	m.typ = rm.Type
	return
}

func (m *Message) MustCollectionDataRemovedMessage() (msg CollectionDataRemovedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeCollectionDataRemoved)
	m.typ = rm.Type
	return
}

func (m *Message) MustCollectionDataUpdatedMessage() (msg CollectionDataUpdatedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeCollectionDataUpdated)
	m.typ = rm.Type
	return
}

func (m *Message) MustItemCreatedMessage() (msg ItemCreatedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeItemCreated)
	m.typ = rm.Type
	return
}

func (m *Message) MustItemRemovedMessage() (msg ItemRemovedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeItemRemoved)
	m.typ = rm.Type
	return
}

func (m *Message) MustItemOnChainDataCreatedMessage() (msg ItemOnChainDataCreatedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeItemOnChainDataCreated)
	m.typ = rm.Type
	return
}

func (m *Message) MustItemOnChainDataRemovedMessage() (msg ItemOnChainDataRemovedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeItemOnChainDataRemoved)
	m.typ = rm.Type
	return
}

func (m *Message) MustSeedCreatedMessage() (msg SeedCreatedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeSeedCreated)
	m.typ = rm.Type
	return
}

func (m *Message) MustSeedRemovedMessage() (msg SeedRemovedMessage) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeSeedRemoved)
	m.typ = rm.Type
	return
}

func (m *Message) MustWithdrawalMessage() (msg WithdrawalMsg) {
	rm := mustUnmarshal(m.raw, &msg)
	assertType(rm.Type, MessageTypeWithdrawal)
	m.typ = rm.Type
	return
}

func mustUnmarshal(raw json.RawMessage, v interface{}) rawMessage {
	rm := rawMessage{
		Raw: v,
	}
	if err := json.Unmarshal(raw, &rm); err != nil {
		panic(errors.Wrap(err, "failed to unmarshal", logan.F{"raw": string(raw)}))
	}

	return rm
}

func assertType(actual MessageType, expected MessageType) {
	if actual != expected {
		panic(errors.From(errors.New("wrong message type"), logan.F{"expected": expected, "actual": actual}))
	}
}
