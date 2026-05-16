package testkit_test

import (
	"net/http"
	"testing"

	gomvchttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/testkit"
)

func newRouter() *gomvchttp.Router {
	r := gomvchttp.NewRouter(nil)
	r.Get("/ping", func(c *gomvchttp.Context) {
		c.Text(200, "pong")
	})
	r.Post("/echo", func(c *gomvchttp.Context) {
		var body map[string]any
		_ = c.BindJSON(&body)
		c.JSON(200, body)
	})
	r.Get("/redirect", func(c *gomvchttp.Context) {
		c.Redirect(http.StatusFound, "/ping")
	})
	r.Get("/header-check", func(c *gomvchttp.Context) {
		c.Header("X-Custom", "hello")
		c.Text(200, "ok")
	})
	return r
}

func TestTestkitGet(t *testing.T) {
	client := testkit.New(newRouter())
	client.Get("/ping").AssertStatus(t, 200).AssertBodyContains(t, "pong")
}

func TestTestkitPost(t *testing.T) {
	client := testkit.New(newRouter())
	var result map[string]any
	client.Post("/echo", map[string]string{"name": "Alice"}).
		AssertStatus(t, 200).
		AssertJSON(t, &result)
	if result["name"] != "Alice" {
		t.Fatalf("expected name=Alice, got %v", result["name"])
	}
}

func TestTestkitRedirect(t *testing.T) {
	client := testkit.New(newRouter())
	client.Get("/redirect").AssertRedirect(t, "/ping")
}

func TestTestkitHeader(t *testing.T) {
	client := testkit.New(newRouter())
	client.Get("/header-check").AssertHeader(t, "X-Custom", "hello")
}

func TestTestkitWithHeader(t *testing.T) {
	r := gomvchttp.NewRouter(nil)
	r.Get("/auth", func(c *gomvchttp.Context) {
		if c.GetHeader("Authorization") == "Bearer token123" {
			c.Text(200, "ok")
		} else {
			c.Text(401, "unauthorized")
		}
	})

	client := testkit.New(r).WithHeader("Authorization", "Bearer token123")
	client.Get("/auth").AssertStatus(t, 200)
}

func TestTestkitPostForm(t *testing.T) {
	r := gomvchttp.NewRouter(nil)
	r.Post("/form", func(c *gomvchttp.Context) {
		name := c.Input("name")
		c.Text(200, "hello "+name)
	})

	client := testkit.New(r)
	client.PostForm("/form", map[string]string{"name": "Bob"}).
		AssertStatus(t, 200).
		AssertBodyContains(t, "hello Bob")
}

func TestTestkit404(t *testing.T) {
	client := testkit.New(newRouter())
	client.Get("/does-not-exist").AssertStatus(t, 404)
}
