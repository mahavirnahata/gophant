# Queue Usage

This guide shows how to enqueue, run workers, and handle retries.

## 1) Define a job

```go
type SendWelcomeEmail struct {
	UserID int64 `json:"user_id"`
}

func (j *SendWelcomeEmail) Handle() error {
	// do work
	return nil
}
```

## 2) Register job type

```go
jobs.Registry.RegisterType(&SendWelcomeEmail{}, func() queue.JobHandler { return &SendWelcomeEmail{} })
```

## 3) Enqueue job

```go
payload, _ := jobs.Registry.Serialize(&SendWelcomeEmail{UserID: 1})
_ = q.PushPayload(payload)
```

## 4) Run workers

```bash
gophant queue:work
```

With flags:

```bash
gophant queue:work --once
gophant queue:work --sleep 5
gophant queue:work --interval 5
```

## 5) Retries and dead‑letter

Implement retry behavior:

```go
func (j *SendWelcomeEmail) Retries() int { return 3 }
func (j *SendWelcomeEmail) Backoff() time.Duration { return 2 * time.Second }
```

Retry failed jobs:

```bash
gophant queue:retry --max 100
```
