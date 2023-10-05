package msgs

import (
	"context"
	"time"

	"github.com/adjust/rmq/v5"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
)

type MultiPublisher struct {
	log            *logan.Entry
	queues         []rmq.Queue
	queueNames     []string
	attempts       uint64
	minRetryPeriod time.Duration
	maxRetryPeriod time.Duration
}

func (m *MultiPublisher) PublishMsgs(ctx context.Context, msgs ...Message) error {
	publishable := make([]string, len(msgs))

	for i, msg := range msgs {
		publishable[i] = msg.String()
	}

	m.log.
		WithField("messages", len(msgs)).
		WithField("queues", m.queues).
		Debug("starting publishing messages to queues")

	cursor := 0
	running.WithThreshold(ctx, m.log, "publishing", func(ctx context.Context) (bool, error) {
		for ; cursor < len(m.queues); cursor++ {
			queue := m.queues[cursor]
			if err := queue.Publish(publishable...); err != nil {
				return false, errors.Wrap(err, "failed to publish messages", logan.F{
					"queue": m.queueNames[cursor],
				})
			}
		}

		return true, nil
	}, m.minRetryPeriod, m.maxRetryPeriod, m.attempts)

	if cursor < len(m.queues) {
		return errors.New("failed to publish messages to all queues")
	}

	return nil
}
