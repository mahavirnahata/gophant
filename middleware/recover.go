package middleware

import (
	"log"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

func Recover() gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("panic: %v", r)
					c.Text(500, "Internal Server Error")
				}
			}()
			next(c)
		}
	}
}
