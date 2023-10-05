package resources

import (
	"encoding/json"
	"strconv"

	rarimocore "github.com/rarimo/rarimo-core/x/rarimocore/types"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type TransferState rarimocore.OpStatus

const (
	TransferStateInitialized = TransferState(rarimocore.OpStatus_INITIALIZED)
	TransferStateApproved    = TransferState(rarimocore.OpStatus_APPROVED)
	TransferStateNotApproved = TransferState(rarimocore.OpStatus_NOT_APPROVED)
	TransferStateSigned      = TransferState(rarimocore.OpStatus_SIGNED)
)

var transferStateIntStr = map[TransferState]string{
	TransferStateInitialized: "initialized",
	TransferStateNotApproved: "not_approved",
	TransferStateApproved:    "approved",
	TransferStateSigned:      "signed",
}

var transferStateStrInt = map[string]TransferState{
	"initialized":  TransferStateInitialized,
	"not_approved": TransferStateNotApproved,
	"approved":     TransferStateApproved,
	"signed":       TransferStateSigned,
}

func (t TransferState) String() string {
	return transferStateIntStr[t]
}

func (t TransferState) MarshalJSON() ([]byte, error) {
	return json.Marshal(Flag{
		Name:  transferStateIntStr[t],
		Value: int32(t),
	})
}

func (t *TransferState) UnmarshalJSON(b []byte) error {
	var res Flag
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}

	*t = TransferState(res.Value)
	return nil
}

func (t *TransferState) UnmarshalText(b []byte) error {
	typ, err := strconv.ParseInt(string(b), 0, 0)
	if err != nil {
		return err
	}

	if _, ok := transferStateIntStr[TransferState(typ)]; !ok {
		return errors.From(errors.New("unsupported value"), logan.F{
			"supported": []TransferState{
				TransferStateInitialized,
				TransferStateNotApproved,
				TransferStateApproved,
				TransferStateSigned,
			},
		})
	}

	*t = TransferState(typ)
	return nil
}

func TransferStateFromInt(raw int) (TransferState, bool) {
	if _, ok := transferStateIntStr[TransferState(raw)]; !ok {
		return 0, false
	}

	return TransferState(raw), true
}

func (t TransferState) Int() int {
	return int(t)
}

func (t TransferState) Intp() *int {
	i := t.Int()
	return &i
}
