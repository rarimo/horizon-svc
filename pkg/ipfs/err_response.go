package ipfs

import (
	"strings"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type errResponse struct {
	Message string `json:"Message"`
	Code    int    `json:"Code"`
	Type    string `json:"Type"`
}

func (e errResponse) ToError() error {
	if strings.Contains(e.Message, "no link named") {
		return ErrNotFound
	}

	return errors.From(errors.New(e.Message), logan.F{
		"code": e.Code,
		"type": e.Type,
	})
}
