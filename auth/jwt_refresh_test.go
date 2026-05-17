package auth

import (
	"testing"
	"time"
)

func TestSignPair_ReturnsTokens(t *testing.T) {
	mgr := newJWT()
	pair, err := mgr.SignPair(map[string]any{"sub": "42"})
	if err != nil {
		t.Fatalf("SignPair: %v", err)
	}
	if pair.AccessToken == "" {
		t.Fatal("expected non-empty AccessToken")
	}
	if pair.RefreshToken == "" {
		t.Fatal("expected non-empty RefreshToken")
	}
	if pair.ExpiresIn <= 0 {
		t.Fatalf("expected positive ExpiresIn, got %d", pair.ExpiresIn)
	}
}

func TestSignPair_AccessTokenIsValid(t *testing.T) {
	mgr := newJWT()
	pair, _ := mgr.SignPair(map[string]any{"sub": "99"})
	claims, err := mgr.Verify(pair.AccessToken)
	if err != nil {
		t.Fatalf("access token should verify: %v", err)
	}
	if claims["sub"] != "99" {
		t.Fatalf("expected sub=99, got %v", claims["sub"])
	}
}

func TestSignPair_RequiresSub(t *testing.T) {
	mgr := newJWT()
	_, err := mgr.SignPair(map[string]any{"name": "Alice"})
	if err == nil {
		t.Fatal("expected error when sub is missing")
	}
}

func TestSignPair_RefreshTokenLength(t *testing.T) {
	mgr := newJWT()
	pair, _ := mgr.SignPair(map[string]any{"sub": "1"})
	if len(pair.RefreshToken) != 64 { // 32 bytes → 64-char hex
		t.Fatalf("expected 64-char refresh token, got %d", len(pair.RefreshToken))
	}
}

func TestSignPair_UniqueRefreshTokens(t *testing.T) {
	mgr := newJWT()
	p1, _ := mgr.SignPair(map[string]any{"sub": "1"})
	p2, _ := mgr.SignPair(map[string]any{"sub": "1"})
	if p1.RefreshToken == p2.RefreshToken {
		t.Fatal("refresh tokens should be unique")
	}
}

func TestRefresh_IssuesNewAccessToken(t *testing.T) {
	mgr := newJWT()
	claims := map[string]any{"sub": "7", "role": "admin"}
	newToken, err := mgr.Refresh(claims)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	got, err := mgr.Verify(newToken)
	if err != nil {
		t.Fatalf("refreshed token should verify: %v", err)
	}
	if got["sub"] != "7" {
		t.Fatalf("expected sub=7, got %v", got["sub"])
	}
}

func TestWithRefreshExpiry(t *testing.T) {
	mgr := NewJWTManager([]byte("secret"), time.Hour).WithRefreshExpiry(7 * 24 * time.Hour)
	if mgr.refreshExpiry != 7*24*time.Hour {
		t.Fatalf("expected 7-day refresh expiry")
	}
}

func TestExpiresIn_MatchesExpiry(t *testing.T) {
	expiry := 2 * time.Hour
	mgr := NewJWTManager([]byte("secret"), expiry)
	pair, _ := mgr.SignPair(map[string]any{"sub": "1"})
	if pair.ExpiresIn != int64(expiry.Seconds()) {
		t.Fatalf("expected ExpiresIn=%d, got %d", int64(expiry.Seconds()), pair.ExpiresIn)
	}
}
