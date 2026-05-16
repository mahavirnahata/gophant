package middleware

import (
	"fmt"
	"net/http"
	"strings"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

// CORSConfig holds the configuration for the CORS middleware.
type CORSConfig struct {
	// AllowOrigins is a list of allowed origins. Use "*" to allow all.
	AllowOrigins []string
	// AllowMethods is a list of allowed HTTP methods.
	AllowMethods []string
	// AllowHeaders is a list of allowed request headers.
	AllowHeaders []string
	// ExposeHeaders is a list of headers the browser may access.
	ExposeHeaders []string
	// AllowCredentials enables Access-Control-Allow-Credentials.
	AllowCredentials bool
	// MaxAge is the preflight cache duration in seconds.
	MaxAge int
}

// DefaultCORSConfig returns a permissive CORS config suitable for public APIs.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		MaxAge:       86400,
	}
}

// CORS returns a middleware that adds Cross-Origin Resource Sharing headers.
// Pass DefaultCORSConfig() for a sensible default.
func CORS(cfg CORSConfig) gomvchttp.Middleware {
	if len(cfg.AllowMethods) == 0 {
		cfg = DefaultCORSConfig()
	}
	methods := strings.Join(cfg.AllowMethods, ", ")
	headers := strings.Join(cfg.AllowHeaders, ", ")
	expose := strings.Join(cfg.ExposeHeaders, ", ")

	allowAll := len(cfg.AllowOrigins) == 1 && cfg.AllowOrigins[0] == "*"

	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			origin := c.GetHeader("Origin")

			if origin != "" {
				originAllowed := false
				if allowAll {
					originAllowed = true
				} else {
					for _, o := range cfg.AllowOrigins {
						if o == origin {
							originAllowed = true
							break
						}
					}
				}

				if originAllowed {
					if allowAll && !cfg.AllowCredentials {
						c.Header("Access-Control-Allow-Origin", "*")
					} else {
						c.Header("Access-Control-Allow-Origin", origin)
						c.Header("Vary", "Origin")
					}
				}
			}

			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			if expose != "" {
				c.Header("Access-Control-Expose-Headers", expose)
			}

			// Preflight request — respond and stop the chain.
			if c.Request.Method == http.MethodOptions {
				c.Header("Access-Control-Allow-Methods", methods)
				c.Header("Access-Control-Allow-Headers", headers)
				if cfg.MaxAge > 0 {
					c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.MaxAge))
				}
				c.Written = true
				c.Writer.WriteHeader(http.StatusNoContent)
				return
			}

			next(c)
		}
	}
}
