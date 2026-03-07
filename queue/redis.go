package queue

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type PayloadHandler func([]byte) error

type RedisQueue struct {
	Client *redis.Client
	Key    string
	ctx    context.Context
}

func NewRedisQueue(client *redis.Client, key string) *RedisQueue {
	if key == "" {
		key = "gophant:queue"
	}
	return &RedisQueue{Client: client, Key: key, ctx: context.Background()}
}

func (q *RedisQueue) PushPayload(payload []byte) error {
	if len(payload) == 0 {
		return errors.New("payload required")
	}
	return q.Client.RPush(q.ctx, q.Key, payload).Err()
}

func (q *RedisQueue) Run(handler PayloadHandler) error {
	return q.RunWithContext(q.ctx, handler)
}

func (q *RedisQueue) RunWithContext(ctx context.Context, handler PayloadHandler) error {
	if handler == nil {
		return errors.New("handler required")
	}
	for {
		res, err := q.Client.BLPop(ctx, 0, q.Key).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		}
		if len(res) < 2 {
			continue
		}
		payload := []byte(res[1])
		_ = handler(payload)
	}
}

func (q *RedisQueue) Pop(ctx context.Context, timeout time.Duration) ([]byte, bool, error) {
	res, err := q.Client.BLPop(ctx, timeout, q.Key).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, false, nil
		}
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if len(res) < 2 {
		return nil, false, nil
	}
	return []byte(res[1]), true, nil
}

func (q *RedisQueue) Close() error {
	return nil
}

func (q *RedisQueue) RunJobs(reg *Registry) error {
	if reg == nil {
		return errors.New("registry required")
	}
	return q.RunWithContext(q.ctx, func(payload []byte) error {
		job, err := reg.Deserialize(payload)
		if err != nil {
			return err
		}
		return job.Handle()
	})
}
