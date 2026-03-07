package http

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type bindStruct struct {
	Email string `form:"email"`
	Age   int    `form:"age"`
}

type stubView struct{}

func (s stubView) Render(w io.Writer, name string, data map[string]any) error { return nil }

func TestBindJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"a@b.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, stubView{})

	var dst struct {
		Email string `json:"email"`
	}
	if err := ctx.BindJSON(&dst); err != nil {
		t.Fatalf("bind json error: %v", err)
	}
	if dst.Email != "a@b.com" {
		t.Fatalf("expected email, got %q", dst.Email)
	}
}

func TestBindForm(t *testing.T) {
	form := url.Values{}
	form.Set("email", "a@b.com")
	form.Set("age", "30")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req, stubView{})

	var dst bindStruct
	if err := ctx.BindForm(&dst); err != nil {
		t.Fatalf("bind form error: %v", err)
	}
	if dst.Email != "a@b.com" || dst.Age != 30 {
		t.Fatalf("unexpected values: %+v", dst)
	}
}
