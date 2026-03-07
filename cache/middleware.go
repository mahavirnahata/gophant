package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

type ResponseCacheOptions struct {
	TTL time.Duration
	Tag string
}

func ResponseCache(c *Cache, opts ResponseCacheOptions) gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(ctx *gomvchttp.Context) {
			if c == nil || opts.TTL <= 0 || ctx.Request.Method != http.MethodGet {
				next(ctx)
				return
			}

			key := cacheKey(ctx.Request)
			if opts.Tag != "" {
				key = tagKey(opts.Tag, key)
			}
			if val, ok := c.Get(key); ok {
				if b, ok := val.([]byte); ok {
					ctx.Header("Content-Type", "text/html; charset=utf-8")
					ctx.StatusCode(200)
					ctx.Writer.WriteHeader(ctx.Status)
					_, _ = ctx.Writer.Write(b)
					return
				}
			}

			rec := newResponseRecorder(ctx.Writer)
			ctx.Writer = rec
			next(ctx)

			if rec.status >= 200 && rec.status < 300 {
				_ = c.Set(key, rec.body, opts.TTL)
			}
		}
	}
}

func cacheKey(r *http.Request) string {
	b := sha1.Sum([]byte(r.Method + ":" + r.URL.RequestURI()))
	return "resp:" + hex.EncodeToString(b[:])
}

type responseRecorder struct {
	writer http.ResponseWriter
	status int
	body   []byte
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{writer: w, status: http.StatusOK}
}

func (r *responseRecorder) Header() http.Header {
	return r.writer.Header()
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.writer.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.writer.Write(b)
}
