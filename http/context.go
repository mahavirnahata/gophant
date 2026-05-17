package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mahavirnahata/gophant/validation"
)

var zeroTime = time.Time{}

type Context struct {
	Writer       http.ResponseWriter
	Request      *http.Request
	Params       map[string]string
	Status       int
	View         ViewRenderer
	Values       map[string]any
	Errors       []error
	Written      bool
	Aborted      bool
	AutoViewName string
	Models       map[string]any

	// lazy JSON body cache
	bodyParsed bool
	bodyData   map[string]any
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

// ── Request helpers ──────────────────────────────────────────────────────────

// Context returns the request's context (carries deadlines, cancellation, tracing).
func (c *Context) Context() context.Context {
	return c.Request.Context()
}

// WithContext replaces the request's context (e.g., to inject a deadline).
func (c *Context) WithContext(ctx context.Context) {
	c.Request = c.Request.WithContext(ctx)
}

// Method returns the HTTP method of the current request.
func (c *Context) Method() string {
	return c.Request.Method
}

// Path returns the URL path of the current request.
func (c *Context) Path() string {
	return c.Request.URL.Path
}

// Param returns a route parameter (e.g., {id} → c.Param("id")).
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// Query returns a query-string value.
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// QueryDefault returns a query-string value or a fallback if not set.
func (c *Context) QueryDefault(key, fallback string) string {
	if v := c.Request.URL.Query().Get(key); v != "" {
		return v
	}
	return fallback
}

// Input returns the first non-empty value for key, checked in order:
// route params → query string → JSON body → form values.
func (c *Context) Input(key string) string {
	if v, ok := c.Params[key]; ok {
		return v
	}
	if v := c.Request.URL.Query().Get(key); v != "" {
		return v
	}
	c.parseBody()
	if c.bodyData != nil {
		if v, ok := c.bodyData[key]; ok {
			return fmt.Sprintf("%v", v)
		}
	}
	if err := c.Request.ParseForm(); err == nil {
		if v := c.Request.FormValue(key); v != "" {
			return v
		}
	}
	return ""
}

// GetHeader returns a request header value.
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// IsJSON reports whether the request body is JSON (Content-Type: application/json).
func (c *Context) IsJSON() bool {
	ct := c.Request.Header.Get("Content-Type")
	return strings.Contains(ct, "application/json")
}

// IsAJAX reports whether the request was sent by XHR (X-Requested-With: XMLHttpRequest).
func (c *Context) IsAJAX() bool {
	return c.Request.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// IP returns the client IP, honouring X-Real-IP and X-Forwarded-For proxy headers.
func (c *Context) IP() string {
	if ip := c.Request.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if forwarded := c.Request.Header.Get("X-Forwarded-For"); forwarded != "" {
		if comma := strings.Index(forwarded, ","); comma != -1 {
			return strings.TrimSpace(forwarded[:comma])
		}
		return strings.TrimSpace(forwarded)
	}
	// Strip port from RemoteAddr
	addr := c.Request.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// parseBody lazily decodes a JSON request body and caches the result.
func (c *Context) parseBody() {
	if c.bodyParsed {
		return
	}
	c.bodyParsed = true
	if !c.IsJSON() {
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		return
	}
	// Restore body so downstream code can read it again (e.g., BindJSON).
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	_ = json.Unmarshal(body, &c.bodyData)
}

// ── Response helpers ─────────────────────────────────────────────────────────

// Header sets a response header.
func (c *Context) Header(key, value string) {
	c.Writer.Header().Set(key, value)
}

// StatusCode sets the HTTP status code for the response (call before writing).
func (c *Context) StatusCode(code int) {
	c.Status = code
}

// JSON writes a JSON response. No-op if a response was already sent.
func (c *Context) JSON(code int, v any) {
	if c.Written {
		return
	}
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.StatusCode(code)
	c.Writer.WriteHeader(c.Status)
	c.Written = true
	_ = json.NewEncoder(c.Writer).Encode(v)
}

// Text writes a plain-text response. No-op if a response was already sent.
func (c *Context) Text(code int, text string) {
	if c.Written {
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.StatusCode(code)
	c.Writer.WriteHeader(c.Status)
	c.Written = true
	_, _ = c.Writer.Write([]byte(text))
}

// Redirect sends an HTTP redirect. No-op if a response was already sent.
func (c *Context) Redirect(code int, location string) {
	if c.Written {
		return
	}
	http.Redirect(c.Writer, c.Request, location, code)
	c.Written = true
}

// File serves a file from disk.
func (c *Context) File(path string) {
	http.ServeFile(c.Writer, c.Request, path)
	c.Written = true
}

// Render executes a named HTML template. No-op if a response was already sent.
func (c *Context) Render(code int, template string, data map[string]any) {
	if c.Written {
		return
	}
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

// URL generates a URL for a named route (requires the router to be in context).
func (c *Context) URL(name string, params ...string) string {
	if v, ok := c.Values["_router"]; ok {
		if r, ok := v.(*Router); ok {
			return r.URL(name, params...)
		}
	}
	return ""
}

// ── Context value store ──────────────────────────────────────────────────────

func (c *Context) Set(key string, val any) {
	c.Values[key] = val
}

func (c *Context) Get(key string) (any, bool) {
	v, ok := c.Values[key]
	return v, ok
}

// Abort marks this request as aborted. Subsequent middleware that checks c.Aborted will stop.
func (c *Context) Abort() {
	c.Aborted = true
}

// AbortWithStatus marks the request aborted and writes an HTTP status with no body.
func (c *Context) AbortWithStatus(code int) {
	c.Aborted = true
	if !c.Written {
		c.Written = true
		c.Writer.WriteHeader(code)
	}
}

// Cookie returns the named request cookie.
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// SetCookie adds a Set-Cookie response header.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// FormFile returns the first uploaded file for the given field name.
func (c *Context) FormFile(field string) (*multipart.FileHeader, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	_, fh, err := c.Request.FormFile(field)
	return fh, err
}

// SaveFile copies an uploaded file to dest on disk.
func (c *Context) SaveFile(fh *multipart.FileHeader, dest string) error {
	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

// Back redirects the client to the Referer URL, or to fallback if Referer is empty.
func (c *Context) Back(fallback string) {
	ref := c.Request.Referer()
	if ref == "" {
		ref = fallback
	}
	c.Redirect(http.StatusFound, ref)
}

// Error appends an error to the context error list (handled by ErrorHandler middleware).
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

// ── Body binding ─────────────────────────────────────────────────────────────

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

// ── Response envelope helpers ─────────────────────────────────────────────────

// Success sends a 200 JSON envelope: {"data": v}.
func (c *Context) Success(v any) {
	c.JSON(http.StatusOK, map[string]any{"data": v})
}

// Created sends a 201 JSON envelope: {"data": v}.
func (c *Context) Created(v any) {
	c.JSON(http.StatusCreated, map[string]any{"data": v})
}

// Fail sends an error JSON envelope: {"error": message}.
func (c *Context) Fail(code int, message string) {
	c.JSON(code, map[string]string{"error": message})
}

// NoContent sends a 204 response with no body.
func (c *Context) NoContent() {
	c.Writer.WriteHeader(http.StatusNoContent)
	c.Written = true
}

// Paginate sends a paginated JSON envelope:
//
//	{"data": items, "meta": {"total": t, "page": p, "per_page": pp, "pages": n}}
func (c *Context) Paginate(items any, page, perPage, total int) {
	pages := 0
	if perPage > 0 {
		pages = (total + perPage - 1) / perPage
	}
	c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{
			"total":    total,
			"page":     page,
			"per_page": perPage,
			"pages":    pages,
		},
	})
}

// ── Validation shortcut ───────────────────────────────────────────────────────

// Validate runs validation rules against the request. If validation fails it
// writes a 422 JSON response with the errors and returns false. On success it
// returns the validated field map and true.
//
//	fields, ok := c.Validate(validation.Rules{
//	    "email": {validation.Required(), validation.Email()},
//	    "name":  {validation.Required(), validation.MinLength(2)},
//	})
//	if !ok { return }
func (c *Context) Validate(rules validation.Rules) (map[string]string, bool) {
	v := validation.New(c.Request)
	for field, fieldRules := range rules {
		v.Field(field, fieldRules...)
	}
	if v.Fails() {
		c.JSON(http.StatusUnprocessableEntity, map[string]any{
			"message": "validation failed",
			"errors":  v.Errors(),
		})
		return nil, false
	}
	return v.Data(), true
}

// ── File responses ────────────────────────────────────────────────────────────

// Download sends a file as an attachment (triggers browser download).
// filename overrides the name shown in the Save-As dialog; pass "" to use the file's base name.
func (c *Context) Download(path, filename string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	name := filename
	if name == "" {
		name = filepath.Base(path)
	}
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(c.Writer, c.Request, name, zeroTime, f)
	c.Written = true
	return nil
}

// Inline serves a file inline (e.g. PDFs, images rendered in the browser).
// contentType defaults to application/octet-stream when empty.
func (c *Context) Inline(path, contentType string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	ct := contentType
	if ct == "" {
		ct = "application/octet-stream"
	}
	name := filepath.Base(path)
	c.Writer.Header().Set("Content-Type", ct)
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, name))
	http.ServeContent(c.Writer, c.Request, name, zeroTime, f)
	c.Written = true
	return nil
}
