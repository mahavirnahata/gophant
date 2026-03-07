package http

var controllerRegistry []any

// RegisterController registers a controller instance for AutoDiscover.
// Use in init() of your controllers package.
func RegisterController(c any) {
	controllerRegistry = append(controllerRegistry, c)
}

// AutoDiscover registers routes for all registered controllers.
func (r *Router) AutoDiscover() {
	for _, c := range controllerRegistry {
		r.registerController(c)
	}
}
