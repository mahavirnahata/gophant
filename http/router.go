package http

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/mahavirnahata/gophant/container"
)

type Handler func(*Context)

type Middleware func(Handler) Handler

type route struct {
	method     string
	pattern    string
	segments   []string
	handler    Handler
	name       string
	middleware []Middleware
}

type Router struct {
	routes           []route
	middleware       []Middleware
	view             ViewRenderer
	basePath         string
	notFound         Handler
	methodNotAllowed Handler
	routeSet         map[string]bool
}

func NewRouter(view ViewRenderer) *Router {
	r := &Router{view: view, routeSet: map[string]bool{}}
	r.notFound = func(c *Context) {
		c.Text(http.StatusNotFound, "Not Found")
	}
	r.methodNotAllowed = func(c *Context) {
		c.Text(http.StatusMethodNotAllowed, "Method Not Allowed")
	}
	return r
}

func (r *Router) Use(m Middleware) {
	r.middleware = append(r.middleware, m)
}

func (r *Router) NotFound(h Handler) {
	r.notFound = h
}

func (r *Router) MethodNotAllowed(h Handler) {
	r.methodNotAllowed = h
}

func (r *Router) add(method, pattern string, h Handler, m []Middleware) {
	pattern = r.withBase(pattern)
	key := method + " " + pattern
	if r.routeSet == nil {
		r.routeSet = map[string]bool{}
	}
	if r.routeSet[key] {
		// duplicate route, ignore
		return
	}
	r.routeSet[key] = true
	segments := splitPattern(pattern)
	r.routes = append(r.routes, route{method: method, pattern: pattern, segments: segments, handler: h, middleware: m})
}

func (r *Router) withBase(pattern string) string {
	if r.basePath == "" {
		return pattern
	}
	if pattern == "/" {
		return r.basePath
	}
	return strings.TrimRight(r.basePath, "/") + pattern
}

func (r *Router) Get(pattern string, h Handler, m ...Middleware) {
	r.add(http.MethodGet, pattern, h, m)
}
func (r *Router) Post(pattern string, h Handler, m ...Middleware) {
	r.add(http.MethodPost, pattern, h, m)
}
func (r *Router) Put(pattern string, h Handler, m ...Middleware) {
	r.add(http.MethodPut, pattern, h, m)
}
func (r *Router) Patch(pattern string, h Handler, m ...Middleware) {
	r.add(http.MethodPatch, pattern, h, m)
}
func (r *Router) Delete(pattern string, h Handler, m ...Middleware) {
	r.add(http.MethodDelete, pattern, h, m)
}

// Resource registers conventional REST routes for a controller.
// Index   GET    /resource
// Show    GET    /resource/{id}
// Store   POST   /resource
// Update  PUT    /resource/{id}
// Destroy DELETE /resource/{id}
func (r *Router) Resource(resource string, controller any) {
	base := "/" + strings.Trim(resource, "/")
	v := reflect.ValueOf(controller)

	if h, ok := methodHandler(v, "Index"); ok {
		r.Get(base, h)
	}
	if h, ok := methodHandler(v, "Show"); ok {
		r.Get(base+"/{id}", h)
	}
	if h, ok := methodHandler(v, "Store"); ok {
		r.Post(base, h)
	}
	if h, ok := methodHandler(v, "Update"); ok {
		r.Put(base+"/{id}", h)
	}
	if h, ok := methodHandler(v, "Destroy"); ok {
		r.Delete(base+"/{id}", h)
	}
}

func methodHandler(v reflect.Value, name string) (Handler, bool) {
	m := v.MethodByName(name)
	if !m.IsValid() {
		return nil, false
	}
	mt := m.Type()
	if mt.NumIn() == 1 && mt.In(0) == reflect.TypeOf(&Context{}) && mt.NumOut() == 0 {
		return func(c *Context) {
			m.Call([]reflect.Value{reflect.ValueOf(c)})
		}, true
	}
	if mt.NumIn() == 2 && mt.In(0) == reflect.TypeOf(&Context{}) {
		modelType := mt.In(1)
		return func(c *Context) {
			if err := bindModel(c, modelType); err != nil {
				c.Error(err)
				return
			}
			modelVal := reflect.ValueOf(c.Models[modelType.Name()])
			m.Call([]reflect.Value{reflect.ValueOf(c), modelVal})
		}, true
	}
	if mt.NumIn() >= 2 && mt.In(0) == reflect.TypeOf(&Context{}) {
		return func(c *Context) {
			contVal, _ := c.Get("container")
			cont, _ := contVal.(*container.Container)
			args := []reflect.Value{reflect.ValueOf(c)}
			for i := 1; i < mt.NumIn(); i++ {
				t := mt.In(i)
				if t == reflect.TypeOf(&Context{}) {
					args = append(args, reflect.ValueOf(c))
					continue
				}
				if cont != nil {
					if val, err := cont.Resolve(t); err == nil {
						args = append(args, reflect.ValueOf(val))
						continue
					}
				}
				args = append(args, reflect.Zero(t))
			}
			m.Call(args)
		}, true
	}
	if mt.NumIn() == 1 && mt.In(0) == reflect.TypeOf(&Context{}) && mt.NumOut() == 1 {
		return func(c *Context) {
			out := m.Call([]reflect.Value{reflect.ValueOf(c)})
			applyAutoView(c, out[0].Interface())
		}, true
	}
	if mt.NumIn() == 1 && mt.In(0) == reflect.TypeOf(&Context{}) && mt.NumOut() == 2 {
		return func(c *Context) {
			out := m.Call([]reflect.Value{reflect.ValueOf(c)})
			if err, ok := out[1].Interface().(error); ok && err != nil {
				c.Error(err)
				return
			}
			applyAutoView(c, out[0].Interface())
		}, true
	}
	return nil, false
}

func applyAutoView(c *Context, v any) {
	switch val := v.(type) {
	case string:
		c.AutoViewName = val
	case map[string]any:
		for k, v := range val {
			c.Set(k, v)
		}
		if c.AutoViewName == "" {
			// leave empty unless explicitly set
		}
	default:
		if val != nil {
			c.Set("data", val)
		}
	}
}

func (r *Router) Group(prefix string, fn func(*Router)) {
	child := &Router{
		routes:           r.routes,
		middleware:       r.middleware,
		view:             r.view,
		basePath:         r.withBase(prefix),
		notFound:         r.notFound,
		methodNotAllowed: r.methodNotAllowed,
		routeSet:         r.routeSet,
	}
	fn(child)
	r.routes = child.routes
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := cleanPath(req.URL.Path)
	method := req.Method

	var matched []route
	for _, rt := range r.routes {
		if rt.method != method {
			if matchSegments(rt.segments, splitPattern(path)) {
				matched = append(matched, rt)
			}
			continue
		}
		params, ok := matchParams(rt.segments, splitPattern(path))
		if !ok {
			continue
		}

		ctx := NewContext(w, req, r.view)
		ctx.Params = params
		h := r.applyMiddleware(rt.handler, append(r.middleware, rt.middleware...)...)
		h(ctx)
		if !ctx.Written && ctx.AutoViewName != "" {
			ctx.Render(http.StatusOK, ctx.AutoViewName, nil)
		}
		return
	}

	ctx := NewContext(w, req, r.view)
	if len(matched) > 0 {
		r.methodNotAllowed(ctx)
		return
	}
	r.notFound(ctx)
}

func (r *Router) applyMiddleware(h Handler, m ...Middleware) Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func cleanPath(path string) string {
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimRight(path, "/")
	}
	return path
}

func splitPattern(pattern string) []string {
	pattern = cleanPath(pattern)
	if pattern == "/" {
		return []string{}
	}
	return strings.Split(strings.Trim(pattern, "/"), "/")
}

func matchSegments(routeSegments, pathSegments []string) bool {
	if len(routeSegments) != len(pathSegments) {
		return false
	}
	for i := range routeSegments {
		if strings.HasPrefix(routeSegments[i], "{") && strings.HasSuffix(routeSegments[i], "}") {
			continue
		}
		if routeSegments[i] != pathSegments[i] {
			return false
		}
	}
	return true
}

func matchParams(routeSegments, pathSegments []string) (map[string]string, bool) {
	if len(routeSegments) != len(pathSegments) {
		return nil, false
	}
	params := map[string]string{}
	for i := range routeSegments {
		seg := routeSegments[i]
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			key := strings.TrimSuffix(strings.TrimPrefix(seg, "{"), "}")
			params[key] = pathSegments[i]
			continue
		}
		if seg != pathSegments[i] {
			return nil, false
		}
	}
	return params, true
}
