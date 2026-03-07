# Queue

Use the in-memory queue for now:

```
q := queue.NewMemoryQueue(100)

_ = q.Push(func() error {
	// background work
	return nil
})

go q.Run()
```

Redis queue (payload-based):

```
client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
q := queue.NewRedisQueue(client, "gophant:queue")

_ = q.PushPayload([]byte("{\"job\":\"send_email\"}"))

go q.Run(func(payload []byte) error {
	// decode payload and execute job
	return nil
})
```

Typed jobs with registry:

```
// Define job
type SendWelcomeEmail struct {
	UserID int64 `json:"user_id"`
}

func (j *SendWelcomeEmail) Handle() error {
	return nil
}

reg := queue.NewRegistry()
reg.RegisterType(&SendWelcomeEmail{}, func() queue.JobHandler { return &SendWelcomeEmail{} })

job := &SendWelcomeEmail{UserID: 1}
payload, _ := reg.Serialize(job)
_ = q.PushPayload(payload)

go q.RunJobs(reg)
```

Retry + dead letter:

```
// Implement retryable job
type RetryableJob struct{}

func (j *RetryableJob) Handle() error { return nil }
func (j *RetryableJob) Retries() int { return 3 }
func (j *RetryableJob) Backoff() time.Duration { return time.Second }

reg := queue.NewRegistry()
reg.RegisterType(&RetryableJob{}, func() queue.JobHandler { return &RetryableJob{} })

dl := queue.NewDeadLetter(q, "gophant:dead")

go queue.RunWithRetry(q, reg, dl)
```
