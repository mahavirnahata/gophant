package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

type rateBucket struct {
	count   int
	resetAt time.Time
}

// RateLimitByKey returns a rate limiter that uses keyFn to derive the bucket key,
// allowing per-user or per-route limiting instead of per-IP.
//
//	r.Use(middleware.RateLimitByKey(100, time.Minute, func(r *http.Request) string {
//	    return r.Header.Get("X-User-ID")  // limit per authenticated user
//	}))
func RateLimitByKey(requests int, window time.Duration, keyFn func(*http.Request) string) gomvchttp.Middleware {
	var mu sync.Mutex
	buckets := map[string]*rateBucket{}

	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			mu.Lock()
			for k, b := range buckets {
				if now.After(b.resetAt) {
					delete(buckets, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			key := keyFn(c.Request)
			now := time.Now()

			mu.Lock()
			b, ok := buckets[key]
			if !ok || now.After(b.resetAt) {
				b = &rateBucket{count: 0, resetAt: now.Add(window)}
				buckets[key] = b
			}
			b.count++
			remaining := requests - b.count
			resetAt := b.resetAt
			mu.Unlock()

			retryAfter := int(time.Until(resetAt).Seconds()) + 1
			c.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(requests))
			if remaining < 0 {
				remaining = 0
			}
			c.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Writer.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if b.count > requests {
				c.Writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				c.JSON(http.StatusTooManyRequests, map[string]any{
					"error":       "too many requests",
					"retry_after": retryAfter,
				})
				return
			}
			next(c)
		}
	}
}

// RateLimit returns a middleware that allows at most `requests` requests per `window`
// from a single IP address. Excess requests receive a 429 with a Retry-After header.
//
// The key is the client IP (X-Real-IP → X-Forwarded-For → RemoteAddr).
// A background goroutine prunes stale buckets every window interval.
//
//	r.Use(middleware.RateLimit(60, time.Minute))          // 60 req/min globally
//	r.Post("/login", h, middleware.RateLimit(5, time.Minute)) // 5 req/min on a route
func RateLimit(requests int, window time.Duration) gomvchttp.Middleware {
	var mu sync.Mutex
	buckets := map[string]*rateBucket{}

	// Background pruner — removes expired buckets to prevent memory growth.
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			mu.Lock()
			for k, b := range buckets {
				if now.After(b.resetAt) {
					delete(buckets, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			key := clientIP(c.Request)
			now := time.Now()

			mu.Lock()
			b, ok := buckets[key]
			if !ok || now.After(b.resetAt) {
				b = &rateBucket{count: 0, resetAt: now.Add(window)}
				buckets[key] = b
			}
			b.count++
			remaining := requests - b.count
			resetAt := b.resetAt
			mu.Unlock()

			retryAfter := int(time.Until(resetAt).Seconds()) + 1

			c.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(requests))
			if remaining < 0 {
				remaining = 0
			}
			c.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Writer.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if b.count > requests {
				c.Writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				c.JSON(http.StatusTooManyRequests, map[string]any{
					"error":       "too many requests",
					"retry_after": retryAfter,
				})
				return
			}

			next(c)
		}
	}
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if i := strings.Index(fwd, ","); i != -1 {
			return strings.TrimSpace(fwd[:i])
		}
		return strings.TrimSpace(fwd)
	}
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}
	return addr
}
