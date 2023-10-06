package cursorer

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"time"
)

func NewCursorer(log *logan.Entry, kv *redis.KeyValueProvider, cursorKey, initialCursor string) types.Cursorer {
	return &cursorer{
		log:           log,
		kv:            kv,
		initialCursor: initialCursor,
		cursorKey:     cursorKey,
	}
}

type cursorer struct {
	kv            *redis.KeyValueProvider
	log           *logan.Entry
	initialCursor string
	cursorKey     string
	createdAt     time.Time
}

func (c *cursorer) GetStartCursor(ctx context.Context) (*types.Cursor, error) {
	startCursorKV, err := c.getStartCursorKv(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cursor", logan.F{
			"cursor_name": c.cursorKey,
		})
	}

	return types.NewCursor(startCursorKV.Value), nil
}

func (c *cursorer) SetStartCursor(ctx context.Context, value *types.Cursor) error {
	if c.createdAt.IsZero() {
		_, err := c.getStartCursorKv(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get cursor", logan.F{
				"cursor_name": c.cursorKey,
			})
		}
	}

	err := c.kv.Upsert(ctx, data.KeyValue{
		Key:       c.cursorKey,
		Value:     value.String(),
		CreatedAt: c.createdAt,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return errors.Wrap(err, "failed to set cursor", logan.F{
			"cursor_name": c.cursorKey,
			"raw":         value,
		})
	}

	return nil
}

func (c *cursorer) getStartCursorKv(ctx context.Context) (*data.KeyValue, error) {
	startCursorKV, err := c.kv.Get(ctx, c.cursorKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cursor", logan.F{
			"cursor_name": c.cursorKey,
		})
	}

	if startCursorKV == nil {
		now := time.Now()

		startCursorKV = &data.KeyValue{
			Key:       c.cursorKey,
			Value:     "",
			CreatedAt: now,
			UpdatedAt: now,
		}

		if c.initialCursor != "" {
			startCursorKV.Value = c.initialCursor
		}
	}

	if c.createdAt.IsZero() {
		c.createdAt = startCursorKV.CreatedAt
	}

	return startCursorKV, nil
}
