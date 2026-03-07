package http

import (
	"strings"
)

// AutoView sets the template name to be rendered after handler returns
// if no response was written.
func (c *Context) AutoView(view string) {
	c.AutoViewName = view
}

// InferView builds a view name based on controller + method.
// Example: UserController + Index -> "users/index.html"
func InferView(controllerName, methodName string) string {
	base := strings.TrimSuffix(controllerName, "Controller")
	if base == "" {
		return ""
	}
	return strings.ToLower(base) + "s/" + strings.ToLower(methodName) + ".html"
}
