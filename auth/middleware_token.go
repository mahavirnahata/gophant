package auth

import (
	"net/http"
	"strings"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

type BearerOptions struct {
	UnauthorizedStatus int
}

func Bearer(service *TokenService, opts BearerOptions) gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			token := extractBearer(c.Request)
			info, err := service.Authenticate(token)
			if err != nil {
				status := http.StatusUnauthorized
				if opts.UnauthorizedStatus != 0 {
					status = opts.UnauthorizedStatus
				}
				c.JSON(status, map[string]any{"error": "unauthorized"})
				return
			}
			c.Set("auth_user_id", info.UserID)
			c.Set("auth_abilities", info.Abilities)
			next(c)
		}
	}
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
