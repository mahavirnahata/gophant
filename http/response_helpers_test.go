package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mahavirnahata/gophant/validation"
)

func makeCtx(w http.ResponseWriter, r *http.Request) *Context {
	return NewContext(w, r, nil)
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	c.Success(map[string]any{"id": 1})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["data"] == nil {
		t.Fatal("expected 'data' key")
	}
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodPost, "/", nil))
	c.Created(map[string]any{"id": 2})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestFail(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	c.Fail(http.StatusNotFound, "not found")

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "not found" {
		t.Fatalf("expected 'not found', got %q", body["error"])
	}
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodDelete, "/", nil))
	c.NoContent()

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatal("expected empty body")
	}
}

func TestPaginate(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	c.Paginate([]string{"a", "b"}, 2, 10, 45)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["data"] == nil {
		t.Fatal("expected 'data' key")
	}
	meta := body["meta"].(map[string]any)
	if meta["total"].(float64) != 45 {
		t.Fatalf("expected total=45, got %v", meta["total"])
	}
	if meta["pages"].(float64) != 5 { // ceil(45/10)
		t.Fatalf("expected pages=5, got %v", meta["pages"])
	}
}

func TestPaginate_ZeroPerPage(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	c.Paginate([]string{}, 1, 0, 0)
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	meta := body["meta"].(map[string]any)
	if meta["pages"].(float64) != 0 {
		t.Fatalf("expected pages=0, got %v", meta["pages"])
	}
}

func TestValidate_Passes(t *testing.T) {
	body := strings.NewReader("name=Alice&email=alice@example.com")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := makeCtx(w, req)

	rules := validation.Rules{
		"name":  {validation.Required()},
		"email": {validation.Required(), validation.Email()},
	}
	data, ok := c.Validate(rules)
	if !ok {
		t.Fatalf("expected validation to pass, body=%s", w.Body.String())
	}
	if data["name"] != "Alice" {
		t.Fatalf("expected name=Alice, got %q", data["name"])
	}
}

func TestValidate_Fails(t *testing.T) {
	body := strings.NewReader("name=")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := makeCtx(w, req)

	_, ok := c.Validate(validation.Rules{"name": {validation.Required()}})
	if ok {
		t.Fatal("expected validation to fail")
	}
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}
