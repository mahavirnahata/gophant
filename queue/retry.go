package queue

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RetryHandler interface {
	Handle() error
	Retries() int
	Backoff() time.Duration
}

type DeadLetter struct {
	Queue  *RedisQueue
	Prefix string
}

func NewDeadLetter(q *RedisQueue, prefix string) *DeadLetter {
	if prefix == "" {
		prefix = "gophant:dead"
	}
	return &DeadLetter{Queue: q, Prefix: prefix}
}

func (d *DeadLetter) Push(data []byte) error {
	if d.Queue == nil {
		return errors.New("queue required")
	}
	key := d.Prefix + ":" + d.Queue.Key
	return d.Queue.Client.RPush(d.Queue.ctx, key, data).Err()
}

func (d *DeadLetter) Key() string {
	if d == nil || d.Queue == nil {
		return ""
	}
	return d.Prefix + ":" + d.Queue.Key
}

func RunWithRetry(q *RedisQueue, reg *Registry, dl *DeadLetter) error {
	return RunWithRetryContext(context.Background(), q, reg, dl)
}

func RunWithRetryContext(ctx context.Context, q *RedisQueue, reg *Registry, dl *DeadLetter) error {
	if q == nil || reg == nil {
		return errors.New("queue and registry required")
	}
	return q.RunWithContext(ctx, func(payload []byte) error {
		return HandlePayloadWithRetry(reg, dl, payload)
	})
}

func HandlePayloadWithRetry(reg *Registry, dl *DeadLetter, payload []byte) error {
	job, err := reg.Deserialize(payload)
	if err != nil {
		if dl != nil {
			_ = dl.Push(payload)
		}
		return err
	}

	if rj, ok := job.(RetryHandler); ok {
		attempts := 0
		err = rj.Handle()
		for err != nil && attempts < rj.Retries() {
			attempts++
			time.Sleep(rj.Backoff())
			err = rj.Handle()
		}
		if err != nil && dl != nil {
			_ = dl.Push(payload)
		}
		return err
	}

	if err := job.Handle(); err != nil {
		if dl != nil {
			_ = dl.Push(payload)
		}
		return err
	}
	return nil
}

func RetryDead(q *RedisQueue, dl *DeadLetter, max int) (int, error) {
	if q == nil || dl == nil {
		return 0, errors.New("queue and deadletter required")
	}
	key := dl.Key()
	if key == "" {
		return 0, errors.New("deadletter key missing")
	}
	if max < 0 {
		max = 0
	}
	count := 0
	for {
		if max > 0 && count >= max {
			break
		}
		val, err := q.Client.RPop(q.ctx, key).Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				break
			}
			return count, err
		}
		if err := q.Client.RPush(q.ctx, q.Key, val).Err(); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
