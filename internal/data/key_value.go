package data

import (
	"context"
	"time"
)

type KeyValue struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type KeyValueQ interface {
	Upsert(ctx context.Context, kv KeyValue) error
	Get(ctx context.Context, key string) (*KeyValue, error)
	Remove(ctx context.Context, key string) error
}
