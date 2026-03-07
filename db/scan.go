package db

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func scanRowsIntoStructs(rows *sql.Rows, dest any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Slice {
		return errors.New("dest must be pointer to slice")
	}
	sliceVal := rv.Elem()
	elemType := sliceVal.Type().Elem()
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	fieldMap := buildFieldMap(elemType)

	for rows.Next() {
		values := make([]any, len(cols))
		pointers := make([]any, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return err
		}

		item := reflect.New(elemType).Elem()
		for i, col := range cols {
			if idx, ok := fieldMap[normalize(col)]; ok {
				field := item.Field(idx)
				assignValue(field, values[i])
			}
		}

		if sliceVal.Type().Elem().Kind() == reflect.Pointer {
			sliceVal.Set(reflect.Append(sliceVal, item.Addr()))
		} else {
			sliceVal.Set(reflect.Append(sliceVal, item))
		}
	}
	return rows.Err()
}

func scanRowIntoStruct(rows *sql.Rows, dest any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return errors.New("dest must be pointer to struct")
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	fieldMap := buildFieldMap(rv.Elem().Type())

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	values := make([]any, len(cols))
	pointers := make([]any, len(cols))
	for i := range values {
		pointers[i] = &values[i]
	}
	if err := rows.Scan(pointers...); err != nil {
		return err
	}

	for i, col := range cols {
		if idx, ok := fieldMap[normalize(col)]; ok {
			field := rv.Elem().Field(idx)
			assignValue(field, values[i])
		}
	}
	return nil
}

func buildFieldMap(t reflect.Type) map[string]int {
	m := map[string]int{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		key := f.Tag.Get("db")
		if key == "-" {
			continue
		}
		if key == "" {
			key = f.Name
		}
		m[normalize(key)] = i
	}
	return m
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func assignValue(field reflect.Value, val any) {
	if !field.CanSet() {
		return
	}
	if val == nil {
		return
	}

	if field.Kind() == reflect.Pointer {
		elem := reflect.New(field.Type().Elem())
		assignValue(elem.Elem(), val)
		field.Set(elem)
		return
	}

	switch v := val.(type) {
	case []byte:
		if field.Kind() == reflect.String {
			field.SetString(string(v))
			return
		}
	case int64:
		setInt(field, v)
		return
	case float64:
		setFloat(field, v)
		return
	case bool:
		if field.Kind() == reflect.Bool {
			field.SetBool(v)
			return
		}
	}

	rv := reflect.ValueOf(val)
	if rv.Type().AssignableTo(field.Type()) {
		field.Set(rv)
		return
	}
	if rv.Type().ConvertibleTo(field.Type()) {
		field.Set(rv.Convert(field.Type()))
		return
	}
}

func setInt(field reflect.Value, v int64) {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(uint64(v))
	case reflect.String:
		field.SetString(fmt.Sprintf("%d", v))
	}
}

func setFloat(field reflect.Value, v float64) {
	switch field.Kind() {
	case reflect.Float32, reflect.Float64:
		field.SetFloat(v)
	case reflect.String:
		field.SetString(fmt.Sprintf("%f", v))
	}
}
