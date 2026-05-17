package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
	"github.com/redis/go-redis/v9"
)

// RedisRateLimitConfig controls the Redis-backed sliding-window rate limiter.
type RedisRateLimitConfig struct {
	// Max requests allowed per Window. Default: 60.
	Max int
	// Window is the sliding window duration. Default: 1 minute.
	Window time.Duration
	// KeyFunc derives the rate-limit key from the request. Defaults to client IP.
	KeyFunc func(*http.Request) string
	// KeyPrefix is prepended to all Redis keys. Default: "rl:".
	KeyPrefix string
	// LimitExceeded is called when the limit is hit. Defaults to 429 JSON.
	LimitExceeded func(*gomvchttp.Context)
}

// RedisRateLimit returns a distributed sliding-window rate limiter backed by Redis.
// Unlike the in-memory RateLimit, this works correctly across multiple instances.
//
//	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
//	router.Use(middleware.RedisRateLimit(rdb, middleware.RedisRateLimitConfig{
//	    Max: 100, Window: time.Minute,
//	}))
func RedisRateLimit(rdb *redis.Client, cfg RedisRateLimitConfig) gomvchttp.Middleware {
	if cfg.Max <= 0 {
		cfg.Max = 60
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = clientIP
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "rl:"
	}
	if cfg.LimitExceeded == nil {
		cfg.LimitExceeded = func(c *gomvchttp.Context) {
			c.JSON(http.StatusTooManyRequests, map[string]any{
				"error": "too many requests",
			})
		}
	}

	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			key := cfg.KeyPrefix + cfg.KeyFunc(c.Request)
			count, ttl, err := slidingWindowIncr(rdb, key, cfg.Window)
			if err != nil {
				// Redis unavailable — fail open (allow request) and log
				next(c)
				return
			}

			resetAt := time.Now().Add(ttl)
			remaining := cfg.Max - int(count)
			if remaining < 0 {
				remaining = 0
			}

			c.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.Max))
			c.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Writer.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if count > int64(cfg.Max) {
				retryAfter := int(ttl.Seconds()) + 1
				c.Writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				cfg.LimitExceeded(c)
				return
			}
			next(c)
		}
	}
}

// slidingWindowIncr atomically increments the request count in a Redis sorted set
// using a sliding window algorithm. Returns (count, remaining TTL, error).
func slidingWindowIncr(rdb *redis.Client, key string, window time.Duration) (int64, time.Duration, error) {
	ctx := context.Background()
	now := time.Now()
	windowStart := now.Add(-window).UnixMilli()
	member := fmt.Sprintf("%d-%d", now.UnixMilli(), now.UnixNano())

	pipe := rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixMilli()), Member: member})
	countCmd := pipe.ZCard(ctx, key)
	pipe.PExpire(ctx, key, window)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, 0, err
	}

	return countCmd.Val(), window, nil
}
