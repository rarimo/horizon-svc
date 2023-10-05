package rd

import (
	"context"
	"time"
)

type MutexOpts struct {
	Expiry time.Duration
	Tries  int
}

type Mutex interface {
	Name() string
	Value() string
	Until() time.Time
	LockContext(ctx context.Context) error
	UnlockContext(ctx context.Context) (bool, error)
	ExtendContext(ctx context.Context) (bool, error)
}
