package auth

import (
	"net/http"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

type RequireOptions struct {
	RedirectTo string
}

func (a *Manager) Require(opts RequireOptions) gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			if a.Check(c) {
				next(c)
				return
			}
			if opts.RedirectTo != "" {
				c.Redirect(http.StatusFound, opts.RedirectTo)
				return
			}
			c.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		}
	}
}

func (a *Manager) RequireRole(role string) gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			if a.HasRole(c, role) {
				next(c)
				return
			}
			c.JSON(http.StatusForbidden, map[string]any{"error": "forbidden"})
		}
	}
}

func (a *Manager) RequireAnyRole(roles ...string) gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			for _, r := range roles {
				if a.HasRole(c, r) {
					next(c)
					return
				}
			}
			c.JSON(http.StatusForbidden, map[string]any{"error": "forbidden"})
		}
	}
}
