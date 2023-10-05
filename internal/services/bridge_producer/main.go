package bridge_producer

import (
	"context"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/services"
	"github.com/rarimo/horizon-svc/internal/services/bridge_producer/types"
	"github.com/rarimo/horizon-svc/pkg/msgs"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"sync"
	"time"
)

func RunBridgeEventsProducer(ctx context.Context, cfg config.Config) {
	log := cfg.Log().WithField("who", cfg.BridgeProducer().RunnerName+"_bridge_events_producer")

	withdrawalsPublisher, err := msgs.NewPublisher(cfg.Log(), cfg.RedisClient(),
		cfg.BridgeProducer().RunnerName+"_withdrawals_publisher",
		cfg.BridgeProducer().WithdrawalsQueueName)
	if err != nil {
		panic(errors.Wrap(err, "failed to create withdrawals publisher"))
	}

	producer := &bridgeEventsProducer{
		log:       log,
		producers: newProducerer(cfg, withdrawalsPublisher),
	}

	producer.run(ctx)
}

type bridgeEventsProducer struct {
	log                  *logan.Entry
	chains               data.ChainsQ
	withdrawalsPublisher services.QPublisher
	producers            types.Producerer
}

func (c *bridgeEventsProducer) run(ctx context.Context) {
	wg := &sync.WaitGroup{}

	for _, chain := range c.chains.List() {
		wg.Add(1)

		c.log.WithField("chain", chain.Name).Info("starting producer")

		go func(chain data.Chain) {
			defer wg.Done()
			running.WithBackOff(ctx, c.log, chain.Name+"_producer", c.producers.Get(chain.Name).Run, 5*time.Second, 10*time.Second, time.Minute)
		}(chain)
	}
	wg.Wait()
}
