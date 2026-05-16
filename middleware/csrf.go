package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"

	gomvchttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/security"
)

type CSRFConfig struct {
	Secret       []byte
	SecureCookie bool
	// Except lists path prefixes that are exempt from CSRF verification (e.g., "/api").
	Except []string
	// Skip is a custom function; return true to bypass CSRF for a specific request.
	Skip func(*gomvchttp.Context) bool
}

func CSRF(cfg CSRFConfig) gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			if len(cfg.Secret) == 0 {
				c.Text(http.StatusInternalServerError, "CSRF secret not configured")
				return
			}

			// Check if this request is CSRF-exempt.
			exempt := false
			for _, prefix := range cfg.Except {
				if strings.HasPrefix(c.Request.URL.Path, prefix) {
					exempt = true
					break
				}
			}
			if !exempt && cfg.Skip != nil {
				exempt = cfg.Skip(c)
			}

			if !exempt {
				cookieToken := security.CookieToken(c.Request)
				if cookieToken == "" {
					tok, err := security.GenerateToken(cfg.Secret)
					if err != nil {
						c.Text(http.StatusInternalServerError, "CSRF token error")
						return
					}
					security.SetCSRFCookie(c.Writer, tok, cfg.SecureCookie)
					cookieToken = tok
				}
				c.Set("csrf", cookieToken)

				if security.IsUnsafeMethod(c.Request.Method) {
					requestToken := security.ExtractToken(c.Request)
					if requestToken == "" {
						c.Text(http.StatusForbidden, "CSRF token missing")
						return
					}
					if err := security.VerifyToken(cfg.Secret, requestToken); err != nil {
						c.Text(http.StatusForbidden, "CSRF token invalid")
						return
					}
					if !tokensMatch(cookieToken, requestToken) {
						c.Text(http.StatusForbidden, "CSRF token mismatch")
						return
					}
				}
			}

			next(c)
		}
	}
}

func tokensMatch(a, b string) bool {
	ba, errA := base64.StdEncoding.DecodeString(a)
	bb, errB := base64.StdEncoding.DecodeString(b)
	if errA != nil || errB != nil || len(ba) != len(bb) {
		return false
	}
	return subtle.ConstantTimeCompare(ba, bb) == 1
}
