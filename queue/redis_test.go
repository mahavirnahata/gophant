package queue

import (
	"context"
	"testing"
	"time"

	redismock "github.com/go-redis/redismock/v9"
)

func TestRedisQueuePop(t *testing.T) {
	client, mock := redismock.NewClientMock()
	q := NewRedisQueue(client, "q")

	mock.ExpectBLPop(time.Second, "q").SetVal([]string{"q", "payload"})
	payload, ok, err := q.Pop(context.Background(), time.Second)
	if err != nil || !ok {
		t.Fatalf("expected payload")
	}
	if string(payload) != "payload" {
		t.Fatalf("unexpected payload")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
