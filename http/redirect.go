package http

import "net/http"

// RedirectResponse is returned by c.To() / c.Back() and supports fluent
// flash-data chaining before the redirect is sent.
//
//	c.To("/dashboard").With("success", "Profile updated!").Send()
//	c.Back("/home").WithErrors(v.Errors()).Send()
type RedirectResponse struct {
	c        *Context
	location string
	code     int
}

// To begins a fluent redirect to the given location (default 302).
//
//	c.To("/login").With("error", "Unauthorized").Send()
func (c *Context) To(location string) *RedirectResponse {
	return &RedirectResponse{c: c, location: location, code: http.StatusFound}
}

// Route redirects to a named route.
//
//	c.Route("users.show", "42").With("success", "Done").Send()
func (c *Context) Route(name string, params ...string) *RedirectResponse {
	return c.To(c.URL(name, params...))
}

// WithStatus sets the HTTP status code on the redirect (default 302).
func (r *RedirectResponse) WithStatus(code int) *RedirectResponse {
	r.code = code
	return r
}

// With stores a flash value in the session that survives one request.
//
//	c.To("/home").With("success", "Saved!").Send()
func (r *RedirectResponse) With(key string, value any) *RedirectResponse {
	r.c.Flash(key, value)
	return r
}

// WithError is a shorthand for With("error", message).
func (r *RedirectResponse) WithError(message string) *RedirectResponse {
	return r.With("error", message)
}

// WithSuccess is a shorthand for With("success", message).
func (r *RedirectResponse) WithSuccess(message string) *RedirectResponse {
	return r.With("success", message)
}

// WithErrors stores a map of validation errors in the flash (keyed as "errors").
//
//	c.Back("/form").WithErrors(v.Errors()).Send()
func (r *RedirectResponse) WithErrors(errors map[string][]string) *RedirectResponse {
	return r.With("errors", errors)
}

// WithInput re-flashes all current request inputs (form values) so the next
// page can repopulate form fields.
func (r *RedirectResponse) WithInput() *RedirectResponse {
	_ = r.c.Request.ParseForm()
	input := map[string]string{}
	for k, vals := range r.c.Request.Form {
		if len(vals) > 0 {
			input[k] = vals[0]
		}
	}
	return r.With("_old_input", input)
}

// Send performs the redirect. Always call this at the end of the chain.
func (r *RedirectResponse) Send() {
	r.c.Redirect(r.code, r.location)
}
