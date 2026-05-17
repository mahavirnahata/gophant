package db

import (
	"reflect"
	"strings"
	"time"
)

// StructToMap converts a struct to map[string]any using json struct tags.
// Fields tagged with json:"-" are skipped. Unexported fields are skipped.
// The omitempty tag option is honoured — zero-value fields are excluded.
//
//	type User struct {
//	    ID    int    `json:"id,omitempty"`
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//	m := db.StructToMap(User{Name: "Alice", Email: "a@x.com"})
//	// → map[string]any{"name": "Alice", "email": "a@x.com"}
func StructToMap(v any) map[string]any {
	return structToMap(reflect.ValueOf(v))
}

func structToMap(rv reflect.Value) map[string]any {
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return map[string]any{}
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return map[string]any{}
	}
	rt := rv.Type()
	out := make(map[string]any, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, opts, _ := strings.Cut(tag, ",")
		if name == "" {
			name = field.Name
		}
		omitempty := strings.Contains(opts, "omitempty")
		fv := rv.Field(i)
		if omitempty && isZero(fv) {
			continue
		}
		out[name] = fv.Interface()
	}
	return out
}

// StructToMapExclude converts a struct to map[string]any, skipping the named fields.
// Useful for updates where you don't want to overwrite id/created_at.
//
//	m := db.StructToMapExclude(user, "id", "created_at")
func StructToMapExclude(v any, exclude ...string) map[string]any {
	skip := make(map[string]struct{}, len(exclude))
	for _, e := range exclude {
		skip[e] = struct{}{}
	}
	m := StructToMap(v)
	for k := range skip {
		delete(m, k)
	}
	return m
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			return t.IsZero()
		}
		return false
	}
	return false
}
