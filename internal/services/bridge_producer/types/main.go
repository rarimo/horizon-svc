package types

import (
	"context"
)

type Producer interface {
	Run(ctx context.Context) error
}

type Producerer interface {
	Get(chain string) Producer
}

type Cursorer interface {
	GetStartCursor(ctx context.Context) (int64, error)
	SetStartCursor(ctx context.Context, value int64) error
}
