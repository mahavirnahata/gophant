package middleware

import (
	"log/slog"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

// Logger returns a middleware that logs each request using slog.
// Output: method, path, status, and duration at INFO level.
func Logger() gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			start := time.Now()
			next(c)
			slog.Info("request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Status,
				"duration", time.Since(start).String(),
			)
		}
	}
}
