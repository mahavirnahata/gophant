package http

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type Context struct {
	Writer       http.ResponseWriter
	Request      *http.Request
	Params       map[string]string
	Status       int
	View         ViewRenderer
	Values       map[string]any
	Errors       []error
	Written      bool
	AutoViewName string
	Models       map[string]any
}

func NewContext(w http.ResponseWriter, r *http.Request, vr ViewRenderer) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		Params:  map[string]string{},
		Status:  http.StatusOK,
		View:    vr,
		Values:  map[string]any{},
		Errors:  []error{},
		Models:  map[string]any{},
	}
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) Header(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) StatusCode(code int) {
	c.Status = code
}

func (c *Context) JSON(code int, v any) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.StatusCode(code)
	c.Writer.WriteHeader(c.Status)
	c.Written = true
	_ = json.NewEncoder(c.Writer).Encode(v)
}

func (c *Context) Text(code int, text string) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.StatusCode(code)
	c.Writer.WriteHeader(c.Status)
	c.Written = true
	_, _ = c.Writer.Write([]byte(text))
}

func (c *Context) Redirect(code int, location string) {
	http.Redirect(c.Writer, c.Request, location, code)
	c.Written = true
}

func (c *Context) File(path string) {
	http.ServeFile(c.Writer, c.Request, path)
	c.Written = true
}

func (c *Context) Render(code int, template string, data map[string]any) {
	c.StatusCode(code)
	c.Writer.WriteHeader(c.Status)
	c.Written = true
	c.AutoViewName = ""
	if data == nil {
		data = map[string]any{}
	}
	for k, v := range c.Values {
		if _, ok := data[k]; !ok {
			data[k] = v
		}
	}
	_ = c.View.Render(c.Writer, template, data)
}

func (c *Context) Set(key string, val any) {
	c.Values[key] = val
}

func (c *Context) Get(key string) (any, bool) {
	v, ok := c.Values[key]
	return v, ok
}

func (c *Context) Error(err error) {
	if err == nil {
		return
	}
	c.Errors = append(c.Errors, err)
}

func (c *Context) SetModel(key string, val any) {
	c.Models[key] = val
}

func (c *Context) Model(key string) (any, bool) {
	v, ok := c.Models[key]
	return v, ok
}

func (c *Context) BindJSON(dest any) error {
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dest)
}

func (c *Context) BindForm(dest any) error {
	if err := c.Request.ParseForm(); err != nil {
		return err
	}
	return bindFormValues(c.Request.Form, dest)
}

func bindFormValues(values map[string][]string, dest any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return nil
	}
	rv = rv.Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		key := f.Tag.Get("form")
		if key == "" {
			key = strings.ToLower(f.Name)
		}
		vals, ok := values[key]
		if !ok || len(vals) == 0 {
			continue
		}
		assignField(rv.Field(i), vals[0])
	}
	return nil
}

func assignField(field reflect.Value, val string) {
	if !field.CanSet() {
		return
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
	case reflect.Bool:
		if b, err := strconv.ParseBool(val); err == nil {
			field.SetBool(b)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			field.SetInt(n)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, err := strconv.ParseUint(val, 10, 64); err == nil {
			field.SetUint(n)
		}
	case reflect.Float32, reflect.Float64:
		if n, err := strconv.ParseFloat(val, 64); err == nil {
			field.SetFloat(n)
		}
	}
}
