package resource_test

import (
	"encoding/json"
	"testing"

	"github.com/mahavirnahata/gophant/db"
	"github.com/mahavirnahata/gophant/resource"
)

func userTransformer(u map[string]any) map[string]any {
	return map[string]any{
		"id":   u["id"],
		"name": u["name"],
	}
}

func TestOne(t *testing.T) {
	row := map[string]any{"id": 1, "name": "Alice", "password": "secret"}
	out := resource.One(row, userTransformer)
	if out["id"] != 1 {
		t.Fatalf("expected id=1, got %v", out["id"])
	}
	if out["name"] != "Alice" {
		t.Fatalf("expected name=Alice, got %v", out["name"])
	}
	if _, ok := out["password"]; ok {
		t.Fatal("password should be omitted by transformer")
	}
}

func TestMany(t *testing.T) {
	rows := []map[string]any{
		{"id": 1, "name": "Alice", "password": "s1"},
		{"id": 2, "name": "Bob", "password": "s2"},
	}
	out := resource.Many(rows, userTransformer)
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}
	if out[1]["name"] != "Bob" {
		t.Fatalf("expected Bob, got %v", out[1]["name"])
	}
	if _, ok := out[0]["password"]; ok {
		t.Fatal("password should be omitted")
	}
}

func TestManyEmpty(t *testing.T) {
	out := resource.Many([]map[string]any{}, userTransformer)
	if len(out) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(out))
	}
}

func TestCollection(t *testing.T) {
	rows := []map[string]any{{"id": 1, "name": "Alice"}}
	out := resource.Collection(rows, userTransformer, nil)
	data, ok := out["data"]
	if !ok {
		t.Fatal("expected 'data' key in collection")
	}
	if _, ok := out["meta"]; ok {
		t.Fatal("meta should be absent when nil is passed")
	}
	items := data.([]map[string]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestCollectionWithMeta(t *testing.T) {
	rows := []map[string]any{{"id": 1, "name": "Alice"}}
	meta := map[string]any{"total": 100, "page": 1}
	out := resource.Collection(rows, userTransformer, meta)
	if out["meta"] == nil {
		t.Fatal("expected meta to be present")
	}
	m := out["meta"].(map[string]any)
	if m["total"] != 100 {
		t.Fatalf("expected total=100, got %v", m["total"])
	}
}

func TestFromPage(t *testing.T) {
	page := db.Page{
		Data:    []map[string]any{{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}},
		Total:   45,
		Page:    3,
		PerPage: 15,
	}
	out := resource.FromPage(page, userTransformer)

	data := out["data"].([]map[string]any)
	if len(data) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(data))
	}

	meta := out["meta"].(map[string]any)
	if meta["total"] != 45 {
		t.Fatalf("expected total=45, got %v", meta["total"])
	}
	if meta["page"] != 3 {
		t.Fatalf("expected page=3, got %v", meta["page"])
	}
	if meta["per_page"] != 15 {
		t.Fatalf("expected per_page=15, got %v", meta["per_page"])
	}
	if meta["pages"] != 3 { // ceil(45/15) = 3
		t.Fatalf("expected pages=3, got %v", meta["pages"])
	}
}

func TestFromPageZeroPerPage(t *testing.T) {
	page := db.Page{Data: nil, Total: 0, Page: 1, PerPage: 0}
	out := resource.FromPage(page, userTransformer)
	meta := out["meta"].(map[string]any)
	if meta["pages"] != 0 {
		t.Fatalf("expected pages=0 when per_page=0, got %v", meta["pages"])
	}
}

func TestCollectionJSONSerializable(t *testing.T) {
	rows := []map[string]any{{"id": 1, "name": "Alice"}}
	out := resource.Collection(rows, userTransformer, map[string]any{"total": 1})
	_, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("collection should be JSON-serializable: %v", err)
	}
}

func TestOneGenericTypes(t *testing.T) {
	type Item struct{ Value int }
	fn := func(it Item) map[string]any { return map[string]any{"v": it.Value} }
	out := resource.One(Item{Value: 42}, fn)
	if out["v"] != 42 {
		t.Fatalf("expected v=42, got %v", out["v"])
	}
}
