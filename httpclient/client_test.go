package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	})
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer secret" {
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(401)
	})
	return httptest.NewServer(mux)
}

func TestClientGet(t *testing.T) {
	srv := testServer()
	defer srv.Close()

	client := New().WithBaseURL(srv.URL)
	resp, err := client.Get(context.Background(), "/ping")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !resp.OK() {
		t.Fatalf("expected 2xx, got %d", resp.StatusCode)
	}
	var result map[string]string
	if err := resp.JSON(&result); err != nil {
		t.Fatalf("JSON decode: %v", err)
	}
	if result["message"] != "pong" {
		t.Fatalf("expected pong, got %q", result["message"])
	}
}

func TestClientPost(t *testing.T) {
	srv := testServer()
	defer srv.Close()

	client := New().WithBaseURL(srv.URL)
	resp, err := client.Post(context.Background(), "/echo", map[string]string{"name": "Alice"})
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	var result map[string]string
	resp.JSON(&result)
	if result["name"] != "Alice" {
		t.Fatalf("expected Alice, got %q", result["name"])
	}
}

func TestClientWithToken(t *testing.T) {
	srv := testServer()
	defer srv.Close()

	client := New().WithBaseURL(srv.URL).WithToken("secret")
	resp, err := client.Get(context.Background(), "/auth")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClientCloneIsolation(t *testing.T) {
	base := New().WithBaseURL("http://a.com")
	child := base.WithHeader("X-Custom", "yes")

	if _, ok := base.headers["X-Custom"]; ok {
		t.Fatal("WithHeader should not mutate the original client")
	}
	if child.headers["X-Custom"] != "yes" {
		t.Fatal("child should have the extra header")
	}
}
