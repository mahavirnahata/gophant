package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	Client *redis.Client
	Prefix string
	TTL    time.Duration
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{Client: client, Prefix: "gophant:session:", TTL: 7 * 24 * time.Hour}
}

func (s *RedisStore) key(id string) string {
	return s.Prefix + id
}

func (s *RedisStore) Get(id string) (map[string]any, bool) {
	ctx := context.Background()
	val, err := s.Client.Get(ctx, s.key(id)).Result()
	if err != nil {
		return nil, false
	}
	out := map[string]any{}
	if err := json.Unmarshal([]byte(val), &out); err != nil {
		return nil, false
	}
	return out, true
}

func (s *RedisStore) Save(id string, values map[string]any) error {
	ctx := context.Background()
	b, err := json.Marshal(values)
	if err != nil {
		return err
	}
	return s.Client.Set(ctx, s.key(id), b, s.TTL).Err()
}

func (s *RedisStore) Delete(id string) error {
	ctx := context.Background()
	return s.Client.Del(ctx, s.key(id)).Err()
}
