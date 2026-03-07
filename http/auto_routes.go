package http

import (
	"reflect"
	"strings"
)

// AutoRoutes registers routes for controller methods based on naming convention.
// Examples:
// - UserController.Index  -> GET /users
// - UserController.Show   -> GET /users/{id}
// - UserController.Store  -> POST /users
// - UserController.Update -> PUT /users/{id}
// - UserController.Destroy-> DELETE /users/{id}
// - AuthController.Login  -> POST /login
// - AuthController.LoginForm -> GET /login
func (r *Router) AutoRoutes(controllers ...any) {
	for _, c := range controllers {
		r.registerController(c)
	}
}

func (r *Router) registerController(controller any) {
	v := reflect.ValueOf(controller)
	if !v.IsValid() {
		return
	}
	name := reflect.Indirect(v).Type().Name()
	base := controllerBase(name)

	// RESTful defaults
	if h, ok := methodHandler(v, "Index"); ok {
		r.Get("/"+base, h)
	}
	if h, ok := methodHandler(v, "Show"); ok {
		r.Get("/"+base+"/{id}", h)
	}
	if h, ok := methodHandler(v, "Store"); ok {
		r.Post("/"+base, h)
	}
	if h, ok := methodHandler(v, "Update"); ok {
		r.Put("/"+base+"/{id}", h)
	}
	if h, ok := methodHandler(v, "Destroy"); ok {
		r.Delete("/"+base+"/{id}", h)
	}

	// Custom actions
	methods := []string{"Login", "Register", "Logout", "Forgot", "Reset"}
	for _, m := range methods {
		if h, ok := methodHandler(v, m); ok {
			r.Post("/"+strings.ToLower(m), h)
		}
		if h, ok := methodHandler(v, m+"Form"); ok {
			r.Get("/"+strings.ToLower(m), h)
		}
	}
}

func controllerBase(name string) string {
	name = strings.TrimSuffix(name, "Controller")
	if name == "" {
		return ""
	}
	return strings.ToLower(name) + "s"
}
