// Package httpclient provides a fluent HTTP client for outbound requests.
//
// Usage:
//
//	client := httpclient.New().
//	    WithBaseURL("https://api.example.com").
//	    WithToken("my-token").
//	    WithTimeout(10 * time.Second)
//
//	resp, err := client.Get(ctx, "/users")
//	var users []User
//	resp.JSON(&users)
//
//	resp2, err := client.Post(ctx, "/users", map[string]string{"name": "Alice"})
//	fmt.Println(resp2.OK()) // true if 2xx
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// Client is a reusable HTTP client with fluent configuration.
type Client struct {
	http    *http.Client
	baseURL string
	headers map[string]string
}

// New returns a Client with a 30-second timeout.
func New() *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		headers: map[string]string{"Content-Type": "application/json"},
	}
}

// WithBaseURL sets the URL prefix prepended to every request path.
func (c *Client) WithBaseURL(url string) *Client {
	cp := c.clone()
	cp.baseURL = url
	return cp
}

// WithTimeout overrides the default request timeout.
func (c *Client) WithTimeout(d time.Duration) *Client {
	cp := c.clone()
	cp.http = &http.Client{Timeout: d}
	return cp
}

// WithHeader adds a default header sent with every request.
func (c *Client) WithHeader(key, val string) *Client {
	cp := c.clone()
	cp.headers[key] = val
	return cp
}

// WithToken sets an Authorization: Bearer <token> header.
func (c *Client) WithToken(token string) *Client {
	return c.WithHeader("Authorization", "Bearer "+token)
}

// WithBasicAuth sets an Authorization: Basic header.
func (c *Client) WithBasicAuth(user, pass string) *Client {
	req, _ := http.NewRequest("GET", "http://x", nil)
	req.SetBasicAuth(user, pass)
	return c.WithHeader("Authorization", req.Header.Get("Authorization"))
}

// Response wraps *http.Response with helpers.
type Response struct {
	*http.Response
	body []byte
}

// OK reports whether the status code is 2xx.
func (r *Response) OK() bool { return r.StatusCode >= 200 && r.StatusCode < 300 }

// JSON decodes the response body into v.
func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.body, v)
}

// String returns the response body as a string.
func (r *Response) String() string { return string(r.body) }

// Bytes returns the raw response body.
func (r *Response) Bytes() []byte { return r.body }

func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*Response, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Response{Response: resp, body: b}, nil
}

func jsonBody(v any) (io.Reader, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// Get sends a GET request to path.
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

// Post sends a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body any) (*Response, error) {
	r, err := jsonBody(body)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPost, path, r)
}

// Put sends a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body any) (*Response, error) {
	r, err := jsonBody(body)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPut, path, r)
}

// Patch sends a PATCH request with a JSON body.
func (c *Client) Patch(ctx context.Context, path string, body any) (*Response, error) {
	r, err := jsonBody(body)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPatch, path, r)
}

// Delete sends a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil)
}

func (c *Client) clone() *Client {
	cp := &Client{http: c.http, baseURL: c.baseURL, headers: make(map[string]string, len(c.headers))}
	for k, v := range c.headers {
		cp.headers[k] = v
	}
	return cp
}
