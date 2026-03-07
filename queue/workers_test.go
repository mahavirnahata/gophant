package queue

import (
	"context"
	"testing"
	"time"

	redismock "github.com/go-redis/redismock/v9"
)

type workerJob struct{}

var workerCalled bool

func (j *workerJob) Handle() error {
	workerCalled = true
	return nil
}

func TestRunWorkersOnce(t *testing.T) {
	client, mock := redismock.NewClientMock()
	q := NewRedisQueue(client, "main")

	reg := NewRegistry()
	reg.RegisterType(&workerJob{}, func() JobHandler { return &workerJob{} })

	payload, _ := reg.Serialize(&workerJob{})
	mock.ExpectBLPop(time.Second, "main").SetVal([]string{"main", string(payload)})

	opts := WorkOptions{Once: true, Sleep: time.Second}
	workerCalled = false

	err := RunWorkers(context.Background(), q, reg, nil, 1, opts)
	if err != nil {
		t.Fatalf("run workers: %v", err)
	}
	if !workerCalled {
		t.Fatalf("expected job to run")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
