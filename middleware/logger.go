package middleware

import (
	"log"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

func Logger() gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			start := time.Now()
			next(c)
			dur := time.Since(start)
			log.Printf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, c.Status, dur)
		}
	}
}
