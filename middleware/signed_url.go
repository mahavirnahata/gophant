package middleware

import (
	"errors"
	"net/http"

	gomvchttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/security"
)

// SignedURL returns a middleware that rejects requests with invalid or expired
// signed URLs. Use it on routes that require signed access.
//
//	signer := security.NewURLSigner([]byte(cfg.AppKey))
//	router.Get("/invoice/{id}/download", handler, middleware.SignedURL(signer))
func SignedURL(signer *security.URLSigner) gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			if err := signer.Verify(c.Request.URL.String()); err != nil {
				if errors.Is(err, security.ErrSignatureExpired) {
					c.JSON(http.StatusForbidden, map[string]string{"error": "link has expired"})
				} else {
					c.JSON(http.StatusForbidden, map[string]string{"error": "invalid signature"})
				}
				return
			}
			next(c)
		}
	}
}
