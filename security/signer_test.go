package security

import (
	"strings"
	"testing"
	"time"
)

func newSigner() *URLSigner {
	return NewURLSigner([]byte("test-secret-key"))
}

func TestSign_And_Verify(t *testing.T) {
	s := newSigner()
	url, err := s.Sign("/invoice/42", time.Hour)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if err := s.Verify(url); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	s := newSigner()
	url, _ := s.Sign("/invoice/42", time.Hour)
	tampered := url[:len(url)-4] + "XXXX"
	err := s.Verify(tampered)
	if err != ErrSignatureInvalid {
		t.Fatalf("expected ErrSignatureInvalid, got %v", err)
	}
}

func TestVerify_Expired(t *testing.T) {
	s := newSigner()
	url, _ := s.Sign("/invoice/42", -time.Second)
	err := s.Verify(url)
	if err != ErrSignatureExpired {
		t.Fatalf("expected ErrSignatureExpired, got %v", err)
	}
}

func TestVerify_NoSignature(t *testing.T) {
	s := newSigner()
	err := s.Verify("https://example.com/path?foo=bar")
	if err != ErrSignatureInvalid {
		t.Fatalf("expected ErrSignatureInvalid, got %v", err)
	}
}

func TestSign_NoExpiry(t *testing.T) {
	s := newSigner()
	url, _ := s.Sign("/download", 0)
	if strings.Contains(url, "_expires") {
		t.Fatal("no-expiry URL should not contain _expires param")
	}
	if err := s.Verify(url); err != nil {
		t.Fatalf("no-expiry URL should verify: %v", err)
	}
}

func TestTemporaryURL(t *testing.T) {
	s := newSigner()
	url, err := s.TemporaryURL("/download", map[string]string{"file": "report.pdf"}, time.Hour)
	if err != nil {
		t.Fatalf("TemporaryURL: %v", err)
	}
	if !strings.Contains(url, "file=report.pdf") {
		t.Fatalf("expected file param in URL: %s", url)
	}
	if err := s.Verify(url); err != nil {
		t.Fatalf("TemporaryURL should verify: %v", err)
	}
}

func TestSign_PreservesExistingParams(t *testing.T) {
	s := newSigner()
	url, _ := s.Sign("/page?tab=2", time.Hour)
	if !strings.Contains(url, "tab=2") {
		t.Fatalf("existing query params should be preserved: %s", url)
	}
	if err := s.Verify(url); err != nil {
		t.Fatalf("URL with existing params should verify: %v", err)
	}
}

func TestSign_DifferentSecrets(t *testing.T) {
	s1 := NewURLSigner([]byte("secret-1"))
	s2 := NewURLSigner([]byte("secret-2"))
	url, _ := s1.Sign("/path", time.Hour)
	err := s2.Verify(url)
	if err != ErrSignatureInvalid {
		t.Fatalf("different secret should reject signature, got %v", err)
	}
}
