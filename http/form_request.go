package http

import (
	"net/http"
	"reflect"

	"github.com/mahavirnahata/gophant/validation"
)

// FormRequest is implemented by request objects that carry their own
// validation rules and authorization logic.
//
//	type CreateUserRequest struct {
//	    Name  string
//	    Email string
//	}
//
//	func (r *CreateUserRequest) Rules() validation.Rules {
//	    return validation.Rules{
//	        "name":  {validation.Required(), validation.MinLength(2)},
//	        "email": {validation.Required(), validation.Email()},
//	    }
//	}
//
//	func (r *CreateUserRequest) Authorize(c *Context) bool { return true }
//
//	// In controller:
//	req, ok := http.ValidateFormRequest[CreateUserRequest](c)
//	if !ok { return }
type FormRequest interface {
	Rules() validation.Rules
	Authorize(c *Context) bool
}

// ValidateFormRequest validates and populates T from the request.
// On failure it writes the appropriate HTTP response (403 or 422) and returns false.
//
//	req, ok := http.ValidateFormRequest[CreateUserRequest](c)
//	if !ok { return }
//	// req.Name, req.Email are populated
func ValidateFormRequest[T FormRequest](c *Context) (T, bool) {
	var req T
	// When T is a pointer type (common: *MyRequest), allocate the underlying struct.
	if rv := reflect.ValueOf(&req).Elem(); rv.Kind() == reflect.Pointer {
		rv.Set(reflect.New(rv.Type().Elem()))
		req = rv.Interface().(T)
	}

	// Populate struct from form / JSON body before calling Rules()
	// bindTarget: when req is already a pointer (common case), pass it directly;
	// otherwise take its address so BindJSON/BindForm receive a pointer to struct.
	var bindTarget any
	if reflect.TypeOf(req).Kind() == reflect.Pointer {
		bindTarget = req
	} else {
		bindTarget = &req
	}

	if c.IsJSON() {
		if err := c.BindJSON(bindTarget); err != nil {
			c.Fail(http.StatusUnprocessableEntity, "invalid JSON body")
			var zero T
			return zero, false
		}
	} else {
		if err := c.BindForm(bindTarget); err != nil {
			c.Fail(http.StatusUnprocessableEntity, "invalid form data")
			var zero T
			return zero, false
		}
	}

	if !req.Authorize(c) {
		c.Fail(http.StatusForbidden, "unauthorized")
		var zero T
		return zero, false
	}

	rules := req.Rules()
	if len(rules) > 0 {
		v := validation.New(c.Request)
		for field, fieldRules := range rules {
			v.Field(field, fieldRules...)
		}
		if v.Fails() {
			c.JSON(http.StatusUnprocessableEntity, map[string]any{
				"message": "validation failed",
				"errors":  v.Errors(),
			})
			var zero T
			return zero, false
		}
	}

	return req, true
}
