package queue

import (
	"testing"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

func TestRetryDead(t *testing.T) {
	client, mock := redismock.NewClientMock()
	q := NewRedisQueue(client, "main")
	dl := NewDeadLetter(q, "dead")

	mock.ExpectRPop("dead:main").SetVal("job1")
	mock.ExpectRPush("main", []byte("job1")).SetVal(1)
	mock.ExpectRPop("dead:main").SetErr(redis.Nil)

	count, err := RetryDead(q, dl, 0)
	if err != nil {
		t.Fatalf("retry dead: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1, got %d", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
