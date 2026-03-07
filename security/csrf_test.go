package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRFToken(t *testing.T) {
	secret := []byte("secret")
	tok, err := GenerateToken(secret)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if err := VerifyToken(secret, tok); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestExtractToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-CSRF-Token", "abc")
	if v := ExtractToken(req); v != "abc" {
		t.Fatalf("expected header token")
	}
}
