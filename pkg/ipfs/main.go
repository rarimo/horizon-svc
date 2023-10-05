package ipfs

import (
	"context"
	"errors"
	"io"
)

type Gateway interface {
	Get(ctx context.Context, resourceID string) ([]byte, error)
	GetReader(ctx context.Context, resourceID string) (io.ReadCloser, error)
}

var ErrNotFound = errors.New("not found")
