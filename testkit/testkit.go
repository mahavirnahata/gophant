// Package testkit provides helpers for testing Gophant HTTP handlers.
//
// Usage:
//
//	client := testkit.New(app.Router)
//	resp := client.Get("/users")
//	resp.AssertStatus(t, 200).AssertBodyContains(t, "Alice")
package testkit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

// TestClient sends requests to a Router without a real TCP connection.
type TestClient struct {
	router  *gomvchttp.Router
	headers map[string]string
	cookies []*http.Cookie
}

// New returns a TestClient wrapping the given router.
func New(router *gomvchttp.Router) *TestClient {
	return &TestClient{
		router:  router,
		headers: map[string]string{},
	}
}

// WithHeader returns a copy of the client with an extra request header set.
func (c *TestClient) WithHeader(key, val string) *TestClient {
	cp := &TestClient{router: c.router, headers: make(map[string]string, len(c.headers)+1), cookies: c.cookies}
	for k, v := range c.headers {
		cp.headers[k] = v
	}
	cp.headers[key] = val
	return cp
}

// WithCookie returns a copy of the client with an extra request cookie.
func (c *TestClient) WithCookie(cookie *http.Cookie) *TestClient {
	cp := &TestClient{router: c.router, headers: c.headers, cookies: append(append([]*http.Cookie{}, c.cookies...), cookie)}
	return cp
}

func (c *TestClient) do(method, path string, body io.Reader, contentType string) *Response {
	req := httptest.NewRequest(method, path, body)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	c.router.ServeHTTP(rec, req)
	return &Response{ResponseRecorder: rec}
}

// Get sends a GET request to path.
func (c *TestClient) Get(path string) *Response {
	return c.do(http.MethodGet, path, nil, "")
}

// Delete sends a DELETE request to path.
func (c *TestClient) Delete(path string) *Response {
	return c.do(http.MethodDelete, path, nil, "")
}

// Post sends a POST request with a JSON body.
func (c *TestClient) Post(path string, body any) *Response {
	b, _ := json.Marshal(body)
	return c.do(http.MethodPost, path, bytes.NewReader(b), "application/json")
}

// PostForm sends a POST request with a URL-encoded form body.
func (c *TestClient) PostForm(path string, data map[string]string) *Response {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	return c.do(http.MethodPost, path, strings.NewReader(form.Encode()), "application/x-www-form-urlencoded")
}

// Put sends a PUT request with a JSON body.
func (c *TestClient) Put(path string, body any) *Response {
	b, _ := json.Marshal(body)
	return c.do(http.MethodPut, path, bytes.NewReader(b), "application/json")
}

// Patch sends a PATCH request with a JSON body.
func (c *TestClient) Patch(path string, body any) *Response {
	b, _ := json.Marshal(body)
	return c.do(http.MethodPatch, path, bytes.NewReader(b), "application/json")
}

// Response wraps httptest.ResponseRecorder with assertion helpers.
type Response struct {
	*httptest.ResponseRecorder
}

// Code returns the response HTTP status code.
func (r *Response) Code() int {
	return r.ResponseRecorder.Code
}

// Body returns the response body as a string.
func (r *Response) Body() string {
	return r.ResponseRecorder.Body.String()
}

// AssertStatus fails the test if the status code does not match.
func (r *Response) AssertStatus(t *testing.T, code int) *Response {
	t.Helper()
	if r.Code() != code {
		t.Errorf("expected status %d, got %d (body: %s)", code, r.Code(), r.Body())
	}
	return r
}

// AssertBodyContains fails the test if sub is not found in the response body.
func (r *Response) AssertBodyContains(t *testing.T, sub string) *Response {
	t.Helper()
	if !strings.Contains(r.Body(), sub) {
		t.Errorf("expected body to contain %q, got: %s", sub, r.Body())
	}
	return r
}

// AssertJSON fails if the response body cannot be decoded into v, or if the
// assertion function (if provided) returns false.
func (r *Response) AssertJSON(t *testing.T, v any) *Response {
	t.Helper()
	if err := json.Unmarshal(r.ResponseRecorder.Body.Bytes(), v); err != nil {
		t.Errorf("AssertJSON: could not decode body: %v (body: %s)", err, r.Body())
	}
	return r
}

// AssertRedirect fails if the response is not a redirect to location.
func (r *Response) AssertRedirect(t *testing.T, location string) *Response {
	t.Helper()
	code := r.Code()
	if code < 300 || code >= 400 {
		t.Errorf("expected redirect (3xx), got %d", code)
		return r
	}
	got := r.Header().Get("Location")
	if got != location {
		t.Errorf("expected redirect to %q, got %q", location, got)
	}
	return r
}

// AssertHeader fails if the named response header does not equal val.
func (r *Response) AssertHeader(t *testing.T, key, val string) *Response {
	t.Helper()
	got := r.Header().Get(key)
	if got != val {
		t.Errorf("expected header %s=%q, got %q", key, val, got)
	}
	return r
}

// JSON decodes the response body into v.
func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.ResponseRecorder.Body.Bytes(), v)
}
