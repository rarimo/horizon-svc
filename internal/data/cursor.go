package data

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const separator = ":"

type Cursor struct {
	PageNumber uint64
	ItemIndex  uint64
}

func (c Cursor) String() string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d%s%d", c.PageNumber, separator, c.ItemIndex)))
}

func NewCursor(page, itemIndex uint64) *Cursor {
	return &Cursor{
		PageNumber: page,
		ItemIndex:  itemIndex,
	}
}

func DecodeCursor(cursor string) (*Cursor, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode cursor", logan.F{
			"cursor": cursor,
		})
	}

	result := Cursor{}

	res := strings.Split(string(decoded), ":")
	if len(res) != 2 {
		return nil, errors.From(errors.New("invalid cursor"), logan.F{
			"cursor": cursor,
		})
	}

	result.PageNumber, err = strconv.ParseUint(res[0], 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse page", logan.F{
			"cursor": cursor,
		})
	}

	result.ItemIndex, err = strconv.ParseUint(res[1], 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse itemIndex", logan.F{
			"cursor": cursor,
		})
	}

	return &result, nil
}
