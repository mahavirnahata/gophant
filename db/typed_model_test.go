package db

import (
	"encoding/json"
	"testing"
)

type testUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestMapToTyped(t *testing.T) {
	m := map[string]any{"id": float64(1), "name": "Alice", "email": "alice@example.com"}
	u, err := mapToTyped[testUser](m)
	if err != nil {
		t.Fatalf("mapToTyped: %v", err)
	}
	if u.ID != 1 || u.Name != "Alice" || u.Email != "alice@example.com" {
		t.Fatalf("unexpected: %+v", u)
	}
}

func TestMapToTypedNil(t *testing.T) {
	_, err := mapToTyped[testUser](nil)
	if err == nil {
		t.Fatal("expected error for nil map")
	}
}

func TestMapsToTyped(t *testing.T) {
	rows := []map[string]any{
		{"id": float64(1), "name": "Alice", "email": "a@x.com"},
		{"id": float64(2), "name": "Bob", "email": "b@x.com"},
	}
	users, err := mapsToTyped[testUser](rows)
	if err != nil {
		t.Fatalf("mapsToTyped: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2, got %d", len(users))
	}
	if users[1].Name != "Bob" {
		t.Fatalf("expected Bob, got %s", users[1].Name)
	}
}

func TestMapsToTypedEmpty(t *testing.T) {
	users, err := mapsToTyped[testUser](nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected empty, got %d", len(users))
	}
}

func TestMustToTyped(t *testing.T) {
	m := map[string]any{"id": float64(5), "name": "Carol", "email": "carol@example.com"}
	u := MustToTyped[testUser](m)
	if u.ID != 5 {
		t.Fatalf("expected id=5, got %d", u.ID)
	}
}

func TestMustToTypedPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil map")
		}
	}()
	MustToTyped[testUser](nil)
}

func TestToTyped(t *testing.T) {
	m := map[string]any{"id": float64(9), "name": "Dan", "email": "dan@example.com"}
	u, err := ToTyped[testUser](m)
	if err != nil {
		t.Fatalf("ToTyped: %v", err)
	}
	if u.Name != "Dan" {
		t.Fatalf("expected Dan, got %s", u.Name)
	}
}

func TestTypedPageJSONSerializable(t *testing.T) {
	page := TypedPage[testUser]{
		Data:    []testUser{{ID: 1, Name: "Alice"}},
		Total:   50,
		Page:    1,
		PerPage: 15,
	}
	_, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("TypedPage should be JSON serializable: %v", err)
	}
}
