package cache

import (
	"testing"
	"time"

	redismock "github.com/go-redis/redismock/v9"
)

func TestRedisCacheSetGet(t *testing.T) {
	client, mock := redismock.NewClientMock()
	store := NewRedisStore(client)

	mock.ExpectSet("gophant:cache:key", []byte(`"val"`), time.Minute).SetVal("OK")
	if err := store.Set("key", "val", time.Minute); err != nil {
		t.Fatalf("set error: %v", err)
	}

	mock.ExpectGet("gophant:cache:key").SetVal(`"val"`)
	v, ok := store.Get("key")
	if !ok || v.(string) != "val" {
		t.Fatalf("expected value")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRedisCacheFlushTag(t *testing.T) {
	client, mock := redismock.NewClientMock()
	store := NewRedisStore(client)

	mock.ExpectScan(0, "gophant:cache:tag:users:*", 0).SetVal([]string{"gophant:cache:tag:users:list"}, 0)
	mock.ExpectDel("gophant:cache:tag:users:list").SetVal(1)

	if err := store.FlushTag("users"); err != nil {
		t.Fatalf("flush tag error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
