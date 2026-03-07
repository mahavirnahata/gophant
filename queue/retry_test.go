package queue

import (
	"errors"
	"testing"
	"time"
)

type retryJob struct {
	Failures int `json:"failures"`
	calls    int
}

func (j *retryJob) Handle() error {
	j.calls++
	if j.calls <= j.Failures {
		return errTest
	}
	return nil
}

func (j *retryJob) Retries() int           { return 3 }
func (j *retryJob) Backoff() time.Duration { return 0 }

var errTest = errors.New("fail")

func TestHandlePayloadWithRetry(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterType(&retryJob{}, func() JobHandler { return &retryJob{} })

	payload, _ := reg.Serialize(&retryJob{Failures: 1})
	if err := HandlePayloadWithRetry(reg, nil, payload); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}
