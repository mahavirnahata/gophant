package db

import (
	"testing"
	"time"
)

type product struct {
	ID        int       `json:"id,omitempty"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Active    bool      `json:"active"`
	Secret    string    `json:"-"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	internal  string    // unexported — must be skipped
}

func TestStructToMap_Basic(t *testing.T) {
	p := product{ID: 1, Name: "Widget", Price: 9.99, Active: true}
	m := StructToMap(p)
	if m["id"].(int) != 1 {
		t.Fatalf("expected id=1, got %v", m["id"])
	}
	if m["name"].(string) != "Widget" {
		t.Fatalf("expected name=Widget, got %v", m["name"])
	}
}

func TestStructToMap_OmitsJSONDash(t *testing.T) {
	p := product{Name: "x", Secret: "topsecret"}
	m := StructToMap(p)
	if _, ok := m["secret"]; ok {
		t.Fatal("json:\"-\" field should be omitted")
	}
	if _, ok := m["-"]; ok {
		t.Fatal("json:\"-\" field should be omitted")
	}
}

func TestStructToMap_OmitemptySkipsZero(t *testing.T) {
	p := product{Name: "x"} // ID=0, omitempty → skip
	m := StructToMap(p)
	if _, ok := m["id"]; ok {
		t.Fatal("id=0 with omitempty should be skipped")
	}
}

func TestStructToMap_OmitemptyIncludesNonZero(t *testing.T) {
	p := product{ID: 5, Name: "x"}
	m := StructToMap(p)
	if m["id"].(int) != 5 {
		t.Fatalf("non-zero omitempty field should be included, got %v", m["id"])
	}
}

func TestStructToMap_SkipsUnexported(t *testing.T) {
	p := product{Name: "x", internal: "hidden"}
	m := StructToMap(p)
	if _, ok := m["internal"]; ok {
		t.Fatal("unexported field should be skipped")
	}
}

func TestStructToMap_OmitemptyTime(t *testing.T) {
	p := product{Name: "x"} // CreatedAt is zero time
	m := StructToMap(p)
	if _, ok := m["created_at"]; ok {
		t.Fatal("zero time with omitempty should be skipped")
	}
}

func TestStructToMap_NonZeroTime(t *testing.T) {
	p := product{Name: "x", CreatedAt: time.Now()}
	m := StructToMap(p)
	if _, ok := m["created_at"]; !ok {
		t.Fatal("non-zero time with omitempty should be included")
	}
}

func TestStructToMap_Pointer(t *testing.T) {
	p := &product{ID: 3, Name: "ptr"}
	m := StructToMap(p)
	if m["name"].(string) != "ptr" {
		t.Fatalf("pointer struct should be dereferenced, got %v", m["name"])
	}
}

func TestStructToMap_NilPointer(t *testing.T) {
	var p *product
	m := StructToMap(p)
	if len(m) != 0 {
		t.Fatal("nil pointer should produce empty map")
	}
}

func TestStructToMapExclude(t *testing.T) {
	p := product{ID: 1, Name: "x", Price: 5.0}
	m := StructToMapExclude(p, "id", "price")
	if _, ok := m["id"]; ok {
		t.Fatal("id should be excluded")
	}
	if _, ok := m["price"]; ok {
		t.Fatal("price should be excluded")
	}
	if m["name"].(string) != "x" {
		t.Fatal("name should remain")
	}
}
