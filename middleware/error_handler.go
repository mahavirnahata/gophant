package middleware

import (
	"net/http"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

type ErrorHandlerFunc func(*gomvchttp.Context, []error)

func ErrorHandler(handler ErrorHandlerFunc) gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			next(c)
			if len(c.Errors) == 0 || c.Written {
				return
			}
			if handler != nil {
				handler(c, c.Errors)
				return
			}
			c.JSON(http.StatusInternalServerError, map[string]any{
				"error": "internal error",
			})
		}
	}
}
