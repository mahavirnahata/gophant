package middleware

import (
	gomvchttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/security"
)

func SecurityHeaders(cfg security.HeaderConfig) gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			security.ApplyHeaders(c.Writer, cfg)
			next(c)
		}
	}
}
