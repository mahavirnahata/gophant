package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"
)

const csrfCookieName = "_gophant_csrf"

var errInvalidToken = errors.New("invalid csrf token")

func GenerateToken(secret []byte) (string, error) {
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sig := signCSRF(secret, nonce)
	payload := append(sig, nonce...)
	return base64.StdEncoding.EncodeToString(payload), nil
}

func VerifyToken(secret []byte, token string) error {
	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil || len(data) < sha256.Size+32 {
		return errInvalidToken
	}
	sig := data[:sha256.Size]
	nonce := data[sha256.Size:]
	expected := signCSRF(secret, nonce)
	if !hmac.Equal(sig, expected) {
		return errInvalidToken
	}
	return nil
}

func signCSRF(secret, nonce []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(nonce)
	return mac.Sum(nil)
}

func SetCSRFCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(12 * time.Hour),
	})
}

func ExtractToken(r *http.Request) string {
	if v := r.Header.Get("X-CSRF-Token"); v != "" {
		return v
	}
	if err := r.ParseForm(); err == nil {
		if v := r.FormValue("_token"); v != "" {
			return v
		}
	}
	return ""
}

func CookieToken(r *http.Request) string {
	c, err := r.Cookie(csrfCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

func IsUnsafeMethod(method string) bool {
	m := strings.ToUpper(method)
	return m == http.MethodPost || m == http.MethodPut || m == http.MethodPatch || m == http.MethodDelete
}
