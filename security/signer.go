package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"strconv"
	"time"
)

// ErrSignatureInvalid is returned when a signed URL has a bad signature.
var ErrSignatureInvalid = errors.New("signed url: signature is invalid")

// ErrSignatureExpired is returned when a signed URL has passed its expiry time.
var ErrSignatureExpired = errors.New("signed url: url has expired")

// URLSigner generates and verifies HMAC-SHA256 signed URLs.
//
//	signer := security.NewURLSigner([]byte(cfg.AppKey))
//	link := signer.Sign("/invoice/42/download", 24*time.Hour)
//	// send link to user; verify on receipt:
//	if err := signer.Verify(link); err != nil { ... }
type URLSigner struct {
	secret []byte
}

// NewURLSigner returns a signer using the given secret key.
func NewURLSigner(secret []byte) *URLSigner {
	return &URLSigner{secret: secret}
}

// Sign appends _expires and _signature query parameters to rawURL.
// Pass 0 for expiry to create a URL that never expires.
func (s *URLSigner) Sign(rawURL string, expiry time.Duration) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	if expiry != 0 {
		exp := time.Now().Add(expiry).Unix()
		q.Set("_expires", strconv.FormatInt(exp, 10))
	}
	u.RawQuery = q.Encode()
	sig := s.sign(u.String())
	q.Set("_signature", sig)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// Verify checks the signature and expiry of a previously signed URL.
// Returns nil on success, ErrSignatureInvalid or ErrSignatureExpired on failure.
func (s *URLSigner) Verify(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ErrSignatureInvalid
	}
	q := u.Query()

	sig := q.Get("_signature")
	if sig == "" {
		return ErrSignatureInvalid
	}
	q.Del("_signature")
	u.RawQuery = q.Encode()

	expected := s.sign(u.String())
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return ErrSignatureInvalid
	}

	if expiresStr := q.Get("_expires"); expiresStr != "" {
		exp, err := strconv.ParseInt(expiresStr, 10, 64)
		if err != nil {
			return ErrSignatureInvalid
		}
		if time.Now().Unix() > exp {
			return ErrSignatureExpired
		}
	}
	return nil
}

// TemporaryURL is a convenience that builds and signs a URL with an expiry.
//
//	url := signer.TemporaryURL("/download", map[string]string{"file": "report.pdf"}, time.Hour)
func (s *URLSigner) TemporaryURL(path string, params map[string]string, expiry time.Duration) (string, error) {
	u := &url.URL{Path: path}
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}
	return s.Sign(u.String(), expiry)
}

func (s *URLSigner) sign(msg string) string {
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte(msg))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

