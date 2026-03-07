package queue

import (
	"errors"
	"sync"
)

type MemoryQueue struct {
	jobs   chan Job
	wg     sync.WaitGroup
	closed bool
	mu     sync.Mutex
}

func NewMemoryQueue(size int) *MemoryQueue {
	if size <= 0 {
		size = 100
	}
	return &MemoryQueue{jobs: make(chan Job, size)}
}

func (q *MemoryQueue) Push(job Job) error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return errors.New("queue closed")
	}
	q.wg.Add(1)
	q.mu.Unlock()
	q.jobs <- job
	return nil
}

func (q *MemoryQueue) Run() error {
	for job := range q.jobs {
		_ = job()
		q.wg.Done()
	}
	return nil
}

func (q *MemoryQueue) Close() error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return nil
	}
	q.closed = true
	close(q.jobs)
	q.mu.Unlock()
	q.wg.Wait()
	return nil
}
