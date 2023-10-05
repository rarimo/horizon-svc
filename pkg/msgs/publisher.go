package msgs

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/adjust/rmq/v5"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Publisher struct {
	log   *logan.Entry
	queue rmq.Queue
}

func NewPublisher(log *logan.Entry, redisClient *redis.Client, tag, queueName string) (*Publisher, error) {
	conn, err := rmq.OpenConnectionWithRedisClient(tag, redisClient, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open connection to redis")
	}

	queue, err := conn.OpenQueue(queueName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open queue", logan.F{
			"queue_name": queueName,
		})
	}

	return &Publisher{log: log, queue: queue}, nil
}

func (p *Publisher) PublishMsgs(ctx context.Context, msgs ...Message) error {
	if len(msgs) == 0 {
		p.log.Debug("No messages to publish")
		return nil
	}

	publishable := make([]string, len(msgs))

	for i, msg := range msgs {
		publishable[i] = msg.String()
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	p.log.WithField("messages", publishable).Debug("Publishing messages")

	err := p.queue.Publish(publishable...)
	if err != nil {
		return errors.Wrap(err, "failed to publish messages", logan.F{
			"messages": publishable,
		})
	}

	return nil
}
