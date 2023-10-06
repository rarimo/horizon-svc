package types

import (
	"context"
	"github.com/olegfomenko/solana-go"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"strconv"
)

type Producer interface {
	Run(ctx context.Context) error
}

type Producerer interface {
	Get(chain string) Producer
}

type Cursorer interface {
	GetStartCursor(ctx context.Context) (*Cursor, error)
	SetStartCursor(ctx context.Context, value *Cursor) error
}

type Cursor struct {
	Value string
}

func NewCursor(value string) *Cursor {
	return &Cursor{
		Value: value,
	}
}

func (c *Cursor) SetInt64(value int64) *Cursor {
	return NewCursor(strconv.FormatInt(value, 10))
}

func (c *Cursor) SetUint64(value uint64) *Cursor {
	return NewCursor(strconv.FormatUint(value, 10))
}

func (c *Cursor) SetSignature(value solana.Signature) *Cursor {
	return NewCursor(value.String())
}

func (c *Cursor) String() string {
	return c.Value
}

func (c *Cursor) Int() int64 {
	res, err := strconv.ParseInt(c.Value, 10, 64)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse int64", logan.F{
			"value": c.Value,
		}))
	}

	return res
}

func (c *Cursor) Uint() uint64 {
	return uint64(c.Int())
}

func (c *Cursor) Signature() solana.Signature {
	if c.Value == "" {
		return solana.Signature{}
	}

	res, err := solana.SignatureFromBase58(c.Value)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse signature", logan.F{
			"value": c.Value,
		}))
	}

	return res
}
