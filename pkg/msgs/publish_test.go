package msgs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/adjust/rmq/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func TestPublishBlockRangeMsg(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	conn, err := rmq.OpenConnectionWithRedisClient("tagtest", client, nil)
	if !assert.NoError(t, err) {
		return
	}

	q, err := conn.OpenQueue("test")
	if !assert.NoError(t, err) {
		return
	}

	start := int64(0)

	for i := int64(0); i < 5; i++ {
		msg := BlockRangeMessage{
			Start: start + i,
			End:   start + i + 1,
		}

		err = q.PublishBytes(msg.Message().raw)
		if !assert.NoError(t, err, "expected no error when publishing message") {
			return
		}
	}
}

func TestConsumeBlockRangeMsg(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	conn, err := rmq.OpenConnectionWithRedisClient("tagtest", client, nil)
	if !assert.NoError(t, err) {
		return
	}

	q, err := conn.OpenQueue("test")
	if !assert.NoError(t, err) {
		return
	}

	err = q.StartConsuming(10, 1*time.Second)
	if !assert.NoError(t, err) {
		return
	}

	_, err = q.AddBatchConsumer("test", 10, 1*time.Second, &consumer{
		handler: &logHandler{
			log: logan.New().WithField("who", "test"),
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	time.Sleep(10 * time.Second)
}

func TestPurge(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	conn, err := rmq.OpenConnectionWithRedisClient("tagtest", client, nil)
	if !assert.NoError(t, err) {
		return
	}

	q, err := conn.OpenQueue("test")
	if !assert.NoError(t, err) {
		return
	}

	_, err = q.PurgeRejected()
	if !assert.NoError(t, err) {
		return
	}
}

type logHandler struct {
	log *logan.Entry
}

func (l *logHandler) Handle(ctx context.Context, msgs []Message) error {
	l.log.WithFields(logan.F{
		"count": len(msgs),
	}).Info("received messages")

	for _, msg := range msgs {
		brmsg := msg.MustBlockRangeMessage()

		l.log.WithFields(logan.F{
			"start": brmsg.Start,
			"end":   brmsg.End,
		}).Info("received block range message")
	}

	return nil
}

type consumer struct {
	handler Handler
}

func (c *consumer) Consume(batch rmq.Deliveries) {
	msgs := make([]Message, len(batch))

	for i, del := range batch {
		payload := del.Payload()
		msg := Message{raw: []byte(payload)}
		msgs[i] = msg
	}

	err := c.handler.Handle(context.TODO(), msgs)
	if err != nil {
		panic(errors.Wrap(err, "failed to handle messages"))
	}

	errs := batch.Ack()
	if len(errs) > 0 {
		for i, err := range errs {
			fmt.Printf("failed to ack message %d: %s", i, err.Error())
		}
	}
}
