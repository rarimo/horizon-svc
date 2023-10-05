package rarimo

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/data/redis"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	"gitlab.com/distributed_lab/logan/v3"
)

type rarimoProducer struct {
	log       *logan.Entry
	chain     data.Chain
	kv        *redis.KeyValueProvider
	publisher services.QPublisher
}

func New(log *logan.Entry, chain data.Chain, kv *redis.KeyValueProvider, publisher services.QPublisher) types.Producer {
	return &rarimoProducer{
		log:       log.WithField("who", chain.Name+"_rarimo_bridge_events_producer"),
		chain:     chain,
		kv:        kv,
		publisher: publisher,
	}
}

func (p *rarimoProducer) Run(ctx context.Context) error {
	return nil
}
