package db

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// CastRow applies type casts to a raw map[string]any returned from the database.
// Define casts as a map of column name → CastType.
//
//	casts := db.Casts{
//	    "settings":   db.CastJSON,
//	    "created_at": db.CastTime("2006-01-02 15:04:05"),
//	    "price":      db.CastFloat,
//	    "active":     db.CastBool,
//	}
//	row, _ := ProductModel.Find(1)
//	row = db.CastRow(row, casts)
func CastRow(row map[string]any, casts Casts) map[string]any {
	if row == nil || len(casts) == 0 {
		return row
	}
	out := make(map[string]any, len(row))
	for k, v := range row {
		if cast, ok := casts[k]; ok {
			out[k] = cast(v)
		} else {
			out[k] = v
		}
	}
	return out
}

// CastRows applies casts to a slice of rows.
func CastRows(rows []map[string]any, casts Casts) []map[string]any {
	out := make([]map[string]any, len(rows))
	for i, row := range rows {
		out[i] = CastRow(row, casts)
	}
	return out
}

// Casts maps column names to cast functions.
type Casts map[string]CastFunc

// CastFunc converts a raw database value to a typed Go value.
type CastFunc func(v any) any

// CastJSON unmarshals a JSON string column into map[string]any.
var CastJSON CastFunc = func(v any) any {
	if v == nil {
		return nil
	}
	var s string
	switch t := v.(type) {
	case string:
		s = t
	case []byte:
		s = string(t)
	default:
		return v
	}
	var out any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return v
	}
	return out
}

// CastJSONSlice unmarshals a JSON array column into []any.
var CastJSONSlice CastFunc = func(v any) any {
	result := CastJSON(v)
	if slice, ok := result.([]any); ok {
		return slice
	}
	return []any{}
}

// CastBool casts a value to bool (handles 0/1, "0"/"1", "true"/"false", []byte).
var CastBool CastFunc = func(v any) any {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	case float64:
		return t != 0
	case string:
		b, err := strconv.ParseBool(t)
		if err != nil {
			return t != "" && t != "0"
		}
		return b
	case []byte:
		return len(t) > 0 && t[0] != 0 && string(t) != "0"
	}
	return false
}

// CastInt casts a value to int64.
var CastInt CastFunc = func(v any) any {
	if v == nil {
		return int64(0)
	}
	switch t := v.(type) {
	case int64:
		return t
	case float64:
		return int64(t)
	case string:
		n, _ := strconv.ParseInt(t, 10, 64)
		return n
	case []byte:
		n, _ := strconv.ParseInt(string(t), 10, 64)
		return n
	}
	return int64(0)
}

// CastFloat casts a value to float64.
var CastFloat CastFunc = func(v any) any {
	if v == nil {
		return float64(0)
	}
	switch t := v.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	case []byte:
		f, _ := strconv.ParseFloat(string(t), 64)
		return f
	}
	return float64(0)
}

// CastString casts a value to string.
var CastString CastFunc = func(v any) any {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// CastTime returns a CastFunc that parses a string/[]byte column into time.Time
// using the given layout (defaults to "2006-01-02 15:04:05").
//
//	db.CastTime("")                          // default layout
//	db.CastTime(time.RFC3339)                // ISO 8601
//	db.CastTime("2006-01-02")               // date only
func CastTime(layout string) CastFunc {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return func(v any) any {
		if v == nil {
			return time.Time{}
		}
		var s string
		switch t := v.(type) {
		case time.Time:
			return t
		case string:
			s = t
		case []byte:
			s = string(t)
		default:
			return v
		}
		parsed, err := time.Parse(layout, s)
		if err != nil {
			return v
		}
		return parsed
	}
}
