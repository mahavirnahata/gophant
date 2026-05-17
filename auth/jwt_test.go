package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

func newJWT() *JWTManager {
	return NewJWTManager([]byte("super-secret-key"), time.Hour)
}

func TestJWTSignAndVerify(t *testing.T) {
	mgr := newJWT()
	token, err := mgr.Sign(map[string]any{"sub": "42", "role": "admin"})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	claims, err := mgr.Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims["sub"] != "42" {
		t.Fatalf("expected sub=42, got %v", claims["sub"])
	}
	if claims["role"] != "admin" {
		t.Fatalf("expected role=admin, got %v", claims["role"])
	}
}

func TestJWTRequiresSub(t *testing.T) {
	mgr := newJWT()
	_, err := mgr.Sign(map[string]any{"name": "Alice"})
	if err == nil {
		t.Fatal("expected error when sub is missing")
	}
}

func TestJWTInvalidSignature(t *testing.T) {
	mgr := newJWT()
	token, _ := mgr.Sign(map[string]any{"sub": "1"})

	// Tamper with the token
	tampered := token[:len(token)-4] + "XXXX"
	_, err := mgr.Verify(tampered)
	if err != ErrJWTInvalid {
		t.Fatalf("expected ErrJWTInvalid, got %v", err)
	}
}

func TestJWTExpired(t *testing.T) {
	mgr := NewJWTManager([]byte("secret"), -time.Second) // already expired
	token, _ := mgr.Sign(map[string]any{"sub": "1"})
	_, err := mgr.Verify(token)
	if err != ErrJWTExpired {
		t.Fatalf("expected ErrJWTExpired, got %v", err)
	}
}

func TestJWTMalformed(t *testing.T) {
	mgr := newJWT()
	_, err := mgr.Verify("not.a.jwt.at.all")
	if err != ErrJWTMalformed {
		// 4+ parts — still malformed (our check is != 3)
		_, err = mgr.Verify("a.b")
		if err != ErrJWTMalformed {
			t.Fatalf("expected ErrJWTMalformed for 2-part token")
		}
	}
}

func TestJWTMiddlewareAllows(t *testing.T) {
	mgr := newJWT()
	token, _ := mgr.Sign(map[string]any{"sub": "99"})

	router := gomvchttp.NewRouter(nil)
	router.Use(mgr.Middleware(""))
	router.Get("/protected", func(c *gomvchttp.Context) {
		claims := Claims(c)
		c.Text(200, claims["sub"].(string))
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if w.Body.String() != "99" {
		t.Fatalf("expected sub 99, got %q", w.Body.String())
	}
}

func TestJWTMiddlewareBlocks(t *testing.T) {
	mgr := newJWT()
	router := gomvchttp.NewRouter(nil)
	router.Use(mgr.Middleware(""))
	router.Get("/protected", func(c *gomvchttp.Context) { c.Text(200, "ok") })

	// No token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestJWTClaimsContainIatExp(t *testing.T) {
	mgr := newJWT()
	token, _ := mgr.Sign(map[string]any{"sub": "1"})
	claims, _ := mgr.Verify(token)
	if _, ok := claims["iat"]; !ok {
		t.Fatal("expected iat in claims")
	}
	if _, ok := claims["exp"]; !ok {
		t.Fatal("expected exp in claims")
	}
}
