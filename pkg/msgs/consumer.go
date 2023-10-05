package msgs

import (
	"context"
	"fmt"
	"time"

	"github.com/adjust/rmq/v5"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
)

type Handler interface {
	Handle(ctx context.Context, msgs []Message) error
}

type Consumer struct {
	log     *logan.Entry
	handler Handler
	cfg     ConsumerConfig

	queue rmq.Queue
	conn  rmq.Connection // for cleanup purposes
}

func NewConsumer(log *logan.Entry, cfg ConsumerConfig, handler Handler) *Consumer {
	return &Consumer{
		log:     log.WithField("who", cfg.Name),
		handler: handler,
		cfg:     cfg,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	c.init()

	defer func() {
		if rvr := recover(); rvr != nil {
			c.log.WithRecover(rvr).Error("consumer panicked")
		}
	}()

	for {
		if err := ctx.Err(); err != nil {
			c.log.Info("stopped by context")
			c.cleanup()
			return
		}

		running.UntilSuccess(ctx, c.log, "consumer",
			func(ctx context.Context) (bool, error) {
				err := c.runConsumingOnce(ctx)
				switch err {
				case nil:
					return true, nil
				case context.DeadlineExceeded, context.Canceled:
					return false, nil
				default:
					c.cleanup()
					c.init()
					return false, err
				}
			},
			c.cfg.MinRetryPeriod, c.cfg.MaxRetryPeriod)
	}
}

func (c *Consumer) cleanup() {
	c.log.Info("cleaning up")

	unacked, err := rmq.NewCleaner(c.conn).Clean()
	if err != nil {
		panic(errors.Wrap(err, "unexpected error on cleaning up connection and queue"))
	}

	c.log.WithField("cnt", unacked).Info("returned unacked back to queue")

	rejected, err := c.queue.PurgeRejected()
	if err != nil {
		panic(errors.Wrap(err, "unexpected error on cleaning up rejected messages"))
	}
	c.log.WithField("cnt", rejected).Info("cleaned up rejected messages")

	c.log.Info("shutting down connection")
	<-c.conn.StopAllConsuming()
	c.log.Info("connection shut down")
}

func (c *Consumer) init() {
	var err error

	c.conn, err = rmq.OpenConnectionWithRedisClient(
		fmt.Sprintf("%s-consumer", c.cfg.Name),
		c.cfg.RedisClient,
		nil,
	)
	if err != nil {
		panic(errors.Wrap(err, "failed to open connection"))
	}

	c.queue, err = c.conn.OpenQueue(c.cfg.Queue)
	if err != nil {
		panic(errors.Wrap(err, "failed to open queue", logan.F{
			"queue_name": fmt.Sprintf("%s-consumer-queue", c.cfg.Name),
		}))
	}
}

func (c *Consumer) runConsumingOnce(ctx context.Context) error {
	err := c.queue.StartConsuming(c.cfg.PrefetchLimit, c.cfg.PollDuration)
	if err != nil {
		return errors.Wrap(err, "failed to start consuming")
	}

	consumerName, err := c.queue.AddBatchConsumer(
		fmt.Sprintf("%s-consumer", c.cfg.Name),
		c.cfg.PrefetchLimit,
		c.cfg.PollDuration,
		&handlerConsumer{
			ctx:            ctx,
			handler:        c.handler,
			log:            c.log.WithField("consumer", c.cfg.Name),
			minRetryPeriod: c.cfg.MinRetryPeriod,
			maxRetryPeriod: c.cfg.MaxRetryPeriod,
			attempts:       c.cfg.RetryConsumeAttempts,
		})

	if err != nil {
		return errors.Wrap(err, "failed to add batch consumer")
	}

	c.log.WithField("consumer", consumerName).Info("added consumer")

	// only possible way to return from here is context cancellation
	// from outside and here we'll catch this stop
	// (same context is also provided to the consumer to control it)
	<-ctx.Done()

	return nil
}

type handlerConsumer struct {
	ctx     context.Context // fixme(hp) make a parameter after rmq update (if it will be)
	handler Handler
	log     *logan.Entry

	minRetryPeriod time.Duration
	maxRetryPeriod time.Duration
	attempts       uint64
}

func (h *handlerConsumer) Consume(batch rmq.Deliveries) {
	msgs := make([]Message, 0, len(batch))

	for _, delivery := range batch {
		msgs = append(msgs, Message{raw: []byte(delivery.Payload())})
	}

	h.log.Info("handling messages")

	failedHandling := false
	running.WithThreshold(h.ctx, h.log, "handle", func(ctx context.Context) (bool, error) {
		if err := h.handler.Handle(h.ctx, msgs); err != nil {
			attempt, ok := running.Attempt(h.ctx)
			h.log.WithFields(logan.F{
				"attempt": attempt,
				"ok":      ok,
			}).Debug("got attempt from ctx")
			if ok && attempt == h.attempts {
				failedHandling = true
			}
			return false, errors.Wrap(err, "failed to handle messages", logan.F{
				"attempt": attempt,
			})
		}
		return true, nil
	}, h.minRetryPeriod, h.maxRetryPeriod, h.attempts)

	if h.ctx.Err() != nil {
		h.log.WithError(h.ctx.Err()).Error("ctx canceled")
		return
	}

	if failedHandling {
		h.log.WithField("msgs", msgs).Debug("failed handling messages, should be retried manually")
	}

	h.log.WithField("success", !failedHandling).Info("handled, finalizing messages")

	running.UntilSuccess(h.ctx, h.log, "ack", func(ctx context.Context) (bool, error) {
		finalize := batch.Ack
		if failedHandling {
			finalize = batch.Reject
		}

		if errs := finalize(); len(errs) > 0 {
			for i, err := range errs {
				h.log.WithError(err).Errorf("failed to finalize message %d", i)
			}
			return false, nil
		}
		return true, nil
	}, h.minRetryPeriod, h.maxRetryPeriod)
}
