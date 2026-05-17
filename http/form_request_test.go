package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mahavirnahata/gophant/validation"
)

type createUserRequest struct {
	Name  string `form:"name"`
	Email string `form:"email"`
	admin bool   // unexported, should not be bound
}

func (r *createUserRequest) Rules() validation.Rules {
	return validation.Rules{
		"name":  {validation.Required(), validation.Min(2)},
		"email": {validation.Required(), validation.Email()},
	}
}

func (r *createUserRequest) Authorize(c *Context) bool { return true }

type restrictedRequest struct {
	createUserRequest
}

func (r *restrictedRequest) Authorize(c *Context) bool { return false }

func TestValidateFormRequest_Valid(t *testing.T) {
	body := strings.NewReader("name=Alice&email=alice@example.com")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := makeCtx(w, req)

	fr, ok := ValidateFormRequest[*createUserRequest](c)
	if !ok {
		t.Fatalf("expected success, got: %s", w.Body.String())
	}
	if fr.Name != "Alice" {
		t.Fatalf("expected Name=Alice, got %q", fr.Name)
	}
	if fr.Email != "alice@example.com" {
		t.Fatalf("expected Email populated, got %q", fr.Email)
	}
}

func TestValidateFormRequest_ValidationFails(t *testing.T) {
	body := strings.NewReader("name=A&email=notanemail")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := makeCtx(w, req)

	_, ok := ValidateFormRequest[*createUserRequest](c)
	if ok {
		t.Fatal("expected validation to fail")
	}
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestValidateFormRequest_Unauthorized(t *testing.T) {
	body := strings.NewReader("name=Alice&email=alice@example.com")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := makeCtx(w, req)

	_, ok := ValidateFormRequest[*restrictedRequest](c)
	if ok {
		t.Fatal("expected authorization to fail")
	}
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
