package db

import (
	"testing"
	"time"
)

func TestCastBool(t *testing.T) {
	cases := []struct {
		in  any
		out bool
	}{
		{true, true}, {false, false},
		{int64(1), true}, {int64(0), false},
		{float64(1), true}, {float64(0), false},
		{"true", true}, {"false", false}, {"1", true}, {"0", false},
		{[]byte{1}, true}, {[]byte("0"), false},
		{nil, false},
	}
	for _, c := range cases {
		got := CastBool(c.in).(bool)
		if got != c.out {
			t.Fatalf("CastBool(%v) = %v, want %v", c.in, got, c.out)
		}
	}
}

func TestCastInt(t *testing.T) {
	cases := []struct {
		in  any
		out int64
	}{
		{int64(42), 42}, {float64(3.9), 3},
		{"99", 99}, {[]byte("7"), 7}, {nil, 0},
	}
	for _, c := range cases {
		got := CastInt(c.in).(int64)
		if got != c.out {
			t.Fatalf("CastInt(%v) = %v, want %v", c.in, got, c.out)
		}
	}
}

func TestCastFloat(t *testing.T) {
	if CastFloat(float64(3.14)).(float64) != 3.14 {
		t.Fatal("CastFloat float64")
	}
	if CastFloat(int64(5)).(float64) != 5.0 {
		t.Fatal("CastFloat int64")
	}
	if CastFloat("2.5").(float64) != 2.5 {
		t.Fatal("CastFloat string")
	}
	if CastFloat(nil).(float64) != 0 {
		t.Fatal("CastFloat nil")
	}
}

func TestCastString(t *testing.T) {
	if CastString("hello").(string) != "hello" {
		t.Fatal("string passthrough")
	}
	if CastString([]byte("bytes")).(string) != "bytes" {
		t.Fatal("[]byte cast")
	}
	if CastString(nil).(string) != "" {
		t.Fatal("nil cast")
	}
}

func TestCastJSON_Object(t *testing.T) {
	result := CastJSON(`{"key":"val"}`)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["key"] != "val" {
		t.Fatalf("expected key=val, got %v", m["key"])
	}
}

func TestCastJSON_Array(t *testing.T) {
	result := CastJSONSlice(`[1,2,3]`)
	s, ok := result.([]any)
	if !ok || len(s) != 3 {
		t.Fatalf("expected []any with 3 elements, got %T %v", result, result)
	}
}

func TestCastJSON_Bytes(t *testing.T) {
	result := CastJSON([]byte(`{"a":1}`))
	m, ok := result.(map[string]any)
	if !ok || m["a"] == nil {
		t.Fatalf("expected map from []byte, got %T", result)
	}
}

func TestCastJSON_Nil(t *testing.T) {
	if CastJSON(nil) != nil {
		t.Fatal("nil should return nil")
	}
}

func TestCastTime(t *testing.T) {
	fn := CastTime("")
	result := fn("2024-01-15 10:30:00")
	ts, ok := result.(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", result)
	}
	if ts.Year() != 2024 || ts.Month() != 1 || ts.Day() != 15 {
		t.Fatalf("unexpected time: %v", ts)
	}
}

func TestCastTime_PassThrough(t *testing.T) {
	fn := CastTime("")
	now := time.Now()
	result := fn(now)
	if result.(time.Time) != now {
		t.Fatal("time.Time should pass through unchanged")
	}
}

func TestCastTime_Nil(t *testing.T) {
	fn := CastTime("")
	result := fn(nil)
	if _, ok := result.(time.Time); !ok {
		t.Fatal("nil should return zero time.Time")
	}
}

func TestCastRow(t *testing.T) {
	row := map[string]any{
		"active":   int64(1),
		"price":    "9.99",
		"settings": `{"theme":"dark"}`,
		"name":     "Widget",
	}
	casts := Casts{
		"active":   CastBool,
		"price":    CastFloat,
		"settings": CastJSON,
	}
	out := CastRow(row, casts)
	if out["active"].(bool) != true {
		t.Fatal("active should be bool true")
	}
	if out["price"].(float64) != 9.99 {
		t.Fatalf("price should be 9.99, got %v", out["price"])
	}
	if out["name"].(string) != "Widget" {
		t.Fatal("non-cast fields should pass through")
	}
}

func TestCastRows(t *testing.T) {
	rows := []map[string]any{
		{"active": int64(1)},
		{"active": int64(0)},
	}
	out := CastRows(rows, Casts{"active": CastBool})
	if !out[0]["active"].(bool) || out[1]["active"].(bool) {
		t.Fatal("CastRows should cast each row")
	}
}
