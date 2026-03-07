package scheduler

import (
	"sync"
	"time"
)

type Task func() error

type Job struct {
	Every   time.Duration
	Task    Task
	lastRun time.Time
}

type Scheduler struct {
	jobs []Job
	mu   sync.Mutex
}

func New() *Scheduler {
	return &Scheduler{jobs: []Job{}}
}

func (s *Scheduler) Every(d time.Duration, task Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, Job{Every: d, Task: task})
}

func (s *Scheduler) RunOnce() {
	s.mu.Lock()
	jobs := make([]Job, len(s.jobs))
	copy(jobs, s.jobs)
	s.mu.Unlock()

	now := time.Now()
	for i := range jobs {
		job := &jobs[i]
		if job.lastRun.IsZero() || now.Sub(job.lastRun) >= job.Every {
			_ = job.Task()
			job.lastRun = now
		}
	}

	s.mu.Lock()
	for i := range s.jobs {
		s.jobs[i].lastRun = jobs[i].lastRun
	}
	s.mu.Unlock()
}

func (s *Scheduler) Run(interval time.Duration, stop <-chan struct{}) {
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.RunOnce()
		case <-stop:
			return
		}
	}
}
