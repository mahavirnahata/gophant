package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	Client *redis.Client
	Prefix string
	ctx    context.Context
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{Client: client, Prefix: "gophant:cache:", ctx: context.Background()}
}

func (r *RedisStore) key(key string) string {
	return r.Prefix + key
}

func (r *RedisStore) Get(key string) (any, bool) {
	val, err := r.Client.Get(r.ctx, r.key(key)).Result()
	if err != nil {
		return nil, false
	}
	var out any
	if err := json.Unmarshal([]byte(val), &out); err != nil {
		return nil, false
	}
	return out, true
}

func (r *RedisStore) Set(key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client.Set(r.ctx, r.key(key), b, ttl).Err()
}

func (r *RedisStore) Delete(key string) error {
	return r.Client.Del(r.ctx, r.key(key)).Err()
}

func (r *RedisStore) Flush() error {
	iter := r.Client.Scan(r.ctx, 0, r.Prefix+"*", 0).Iterator()
	for iter.Next(r.ctx) {
		if err := r.Client.Del(r.ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (r *RedisStore) FlushTag(tag string) error {
	if tag == "" {
		return nil
	}
	iter := r.Client.Scan(r.ctx, 0, r.Prefix+tagKey(tag, "*"), 0).Iterator()
	for iter.Next(r.ctx) {
		if err := r.Client.Del(r.ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
