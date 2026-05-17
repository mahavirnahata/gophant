package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/security"
)

func makeSignedRouter(signer *security.URLSigner) *gomvchttp.Router {
	router := gomvchttp.NewRouter(nil)
	router.Use(SignedURL(signer))
	router.Get("/download", func(c *gomvchttp.Context) { c.Text(200, "file content") })
	return router
}

func TestSignedURL_AllowsValidSignature(t *testing.T) {
	signer := security.NewURLSigner([]byte("secret"))
	router := makeSignedRouter(signer)

	signedPath, _ := signer.Sign("/download", time.Hour)
	req := httptest.NewRequest(http.MethodGet, signedPath, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestSignedURL_RejectsUnsigned(t *testing.T) {
	signer := security.NewURLSigner([]byte("secret"))
	router := makeSignedRouter(signer)

	req := httptest.NewRequest(http.MethodGet, "/download", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestSignedURL_RejectsExpired(t *testing.T) {
	signer := security.NewURLSigner([]byte("secret"))
	router := makeSignedRouter(signer)

	signedPath, _ := signer.Sign("/download", -time.Second)
	req := httptest.NewRequest(http.MethodGet, signedPath, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestSignedURL_RejectsTamperedSignature(t *testing.T) {
	signer := security.NewURLSigner([]byte("secret"))
	router := makeSignedRouter(signer)

	signedPath, _ := signer.Sign("/download", time.Hour)
	tampered := signedPath[:len(signedPath)-4] + "XXXX"
	req := httptest.NewRequest(http.MethodGet, tampered, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
