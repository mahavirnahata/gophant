package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

func makeRateLimitRouter(max int, window time.Duration) *gomvchttp.Router {
	router := gomvchttp.NewRouter(nil)
	router.Use(RateLimit(max, window))
	router.Get("/test", func(c *gomvchttp.Context) { c.Text(200, "ok") })
	return router
}

func TestRateLimit_Allows(t *testing.T) {
	router := makeRateLimitRouter(5, time.Minute)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "1.2.3.4:9999"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRateLimit_Blocks(t *testing.T) {
	router := makeRateLimitRouter(2, time.Minute)

	send := func() int {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	if send() != 200 {
		t.Fatal("first request should pass")
	}
	if send() != 200 {
		t.Fatal("second request should pass")
	}
	if send() != http.StatusTooManyRequests {
		t.Fatal("third request should be blocked")
	}
}

func TestRateLimit_SetsHeaders(t *testing.T) {
	router := makeRateLimitRouter(10, time.Minute)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "5.5.5.5:1000"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("X-RateLimit-Limit") != "10" {
		t.Fatalf("expected X-RateLimit-Limit=10, got %q", w.Header().Get("X-RateLimit-Limit"))
	}
	if w.Header().Get("X-RateLimit-Remaining") == "" {
		t.Fatal("expected X-RateLimit-Remaining header")
	}
}

func TestRateLimit_DifferentIPsAreIndependent(t *testing.T) {
	router := makeRateLimitRouter(1, time.Minute)

	send := func(ip string) int {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip + ":9999"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	send("192.168.1.1") // exhausts 192.168.1.1's bucket
	if send("192.168.1.1") == 200 {
		t.Fatal("same IP second request should be blocked")
	}
	if send("192.168.1.2") != 200 {
		t.Fatal("different IP should not be blocked")
	}
}

func TestRateLimitByKey_UsesCustomKey(t *testing.T) {
	router := gomvchttp.NewRouter(nil)
	router.Use(RateLimitByKey(1, time.Minute, func(r *http.Request) string {
		return r.Header.Get("X-User-ID")
	}))
	router.Get("/test", func(c *gomvchttp.Context) { c.Text(200, "ok") })

	send := func(userID string) int {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-ID", userID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	send("user-1")
	if send("user-1") != http.StatusTooManyRequests {
		t.Fatal("same user second request should be blocked")
	}
	if send("user-2") != 200 {
		t.Fatal("different user should not be blocked")
	}
}
