package queue

import (
	"context"
	"errors"
	"log"
	"time"
)

type WorkOptions struct {
	Once  bool
	Sleep time.Duration
}

func RunWorkers(ctx context.Context, q *RedisQueue, reg *Registry, dl *DeadLetter, workers int, opts WorkOptions) error {
	if q == nil || reg == nil {
		return errors.New("queue and registry required")
	}
	if workers < 1 {
		workers = 1
	}
	if opts.Once {
		return runOnce(ctx, q, reg, dl, opts)
	}
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			errCh <- runLoop(ctx, q, reg, dl, opts)
		}()
	}
	return <-errCh
}

func runLoop(ctx context.Context, q *RedisQueue, reg *Registry, dl *DeadLetter, opts WorkOptions) error {
	timeout := time.Duration(0)
	if opts.Sleep > 0 {
		timeout = opts.Sleep
	}
	for {
		payload, ok, err := q.Pop(ctx, timeout)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if err := HandlePayloadWithRetry(reg, dl, payload); err != nil {
			log.Printf("worker: job error: %v", err)
		}
	}
}

func runOnce(ctx context.Context, q *RedisQueue, reg *Registry, dl *DeadLetter, opts WorkOptions) error {
	timeout := time.Duration(0)
	if opts.Sleep > 0 {
		timeout = opts.Sleep
	}
	payload, ok, err := q.Pop(ctx, timeout)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return HandlePayloadWithRetry(reg, dl, payload)
}
