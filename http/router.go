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

// Route is returned by Get/Post/etc. and allows the caller to assign a name.
type Route struct {
	router      *Router
	idx         int
	namedRoutes map[string]string
	pattern     string
}

// Name registers a name for this route, enabling URL generation via Router.URL().
func (rt *Route) Name(name string) *Route {
	rt.namedRoutes[name] = rt.pattern
	if rt.idx >= 0 && rt.idx < len(rt.router.routes) {
		rt.router.routes[rt.idx].name = name
	}
	return rt
}

// RouteInfo is a read-only snapshot of a registered route (used by route:list).
type RouteInfo struct {
	Method  string
	Pattern string
	Name    string
}

type Router struct {
	routes            []route
	middleware        []Middleware
	groupMiddleware   []Middleware // extra middleware applied to every route in this group
	middlewareGroups  map[string][]Middleware
	view              ViewRenderer
	basePath          string
	notFound          Handler
	methodNotAllowed  Handler
	routeSet          map[string]bool
	namedRoutes       map[string]string
}

func NewRouter(view ViewRenderer) *Router {
	r := &Router{
		view:        view,
		routeSet:    map[string]bool{},
		namedRoutes: map[string]string{},
	}
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

// DefineMiddlewareGroup registers a named set of middleware.
// Apply a group by name with UseGroup inside a Group() call or via WithGroup().
//
//	r.DefineMiddlewareGroup("api", middleware.CORS(cfg), middleware.RateLimit(60, time.Minute))
//	r.Group("/api", routes, r.WithGroup("api")...)
func (r *Router) DefineMiddlewareGroup(name string, middleware ...Middleware) {
	if r.middlewareGroups == nil {
		r.middlewareGroups = map[string][]Middleware{}
	}
	r.middlewareGroups[name] = middleware
}

// WithGroup returns the middleware registered under name (panics if undefined).
func (r *Router) WithGroup(name string) []Middleware {
	if r.middlewareGroups == nil {
		panic("router: middleware group " + name + " is not defined")
	}
	mw, ok := r.middlewareGroups[name]
	if !ok {
		panic("router: middleware group " + name + " is not defined")
	}
	return mw
}

func (r *Router) NotFound(h Handler) {
	r.notFound = h
}

func (r *Router) MethodNotAllowed(h Handler) {
	r.methodNotAllowed = h
}

func (r *Router) add(method, pattern string, h Handler, m []Middleware) int {
	pattern = r.withBase(pattern)
	key := method + " " + pattern
	if r.routeSet == nil {
		r.routeSet = map[string]bool{}
	}
	if r.routeSet[key] {
		for i, rt := range r.routes {
			if rt.method == method && rt.pattern == pattern {
				return i
			}
		}
		return -1
	}
	r.routeSet[key] = true
	segments := splitPattern(pattern)
	// Merge group-level middleware (prepended) with route-level middleware.
	routeMiddleware := append(append([]Middleware{}, r.groupMiddleware...), m...)
	r.routes = append(r.routes, route{method: method, pattern: pattern, segments: segments, handler: h, middleware: routeMiddleware})
	return len(r.routes) - 1
}

func (r *Router) newRoute(pattern string, idx int) *Route {
	return &Route{router: r, idx: idx, namedRoutes: r.namedRoutes, pattern: r.withBase(pattern)}
}

func (r *Router) Get(pattern string, h Handler, m ...Middleware) *Route {
	idx := r.add(http.MethodGet, pattern, h, m)
	return r.newRoute(pattern, idx)
}

func (r *Router) Post(pattern string, h Handler, m ...Middleware) *Route {
	idx := r.add(http.MethodPost, pattern, h, m)
	return r.newRoute(pattern, idx)
}

func (r *Router) Put(pattern string, h Handler, m ...Middleware) *Route {
	idx := r.add(http.MethodPut, pattern, h, m)
	return r.newRoute(pattern, idx)
}

func (r *Router) Patch(pattern string, h Handler, m ...Middleware) *Route {
	idx := r.add(http.MethodPatch, pattern, h, m)
	return r.newRoute(pattern, idx)
}

func (r *Router) Delete(pattern string, h Handler, m ...Middleware) *Route {
	idx := r.add(http.MethodDelete, pattern, h, m)
	return r.newRoute(pattern, idx)
}

// Routes returns a snapshot of all registered routes (for route:list and introspection).
func (r *Router) Routes() []RouteInfo {
	out := make([]RouteInfo, len(r.routes))
	for i, rt := range r.routes {
		out[i] = RouteInfo{Method: rt.method, Pattern: rt.pattern, Name: rt.name}
	}
	return out
}

// URL generates a URL for a named route, substituting {param} segments in order.
//
//	r.URL("users.show", "42")  →  "/users/42"
func (r *Router) URL(name string, params ...string) string {
	pattern, ok := r.namedRoutes[name]
	if !ok {
		return ""
	}
	if pattern == "/" {
		return "/"
	}
	segments := strings.Split(strings.Trim(pattern, "/"), "/")
	pi := 0
	for i, seg := range segments {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			if pi < len(params) {
				segments[i] = params[pi]
				pi++
			}
		}
	}
	return "/" + strings.Join(segments, "/")
}

// Resource registers conventional REST routes for a controller struct.
// Recognized methods: Index, Show, Store, Update, Destroy.
func (r *Router) Resource(resource string, controller any) {
	base := "/" + strings.Trim(resource, "/")
	v := reflect.ValueOf(controller)

	if h, ok := methodHandler(v, "Index"); ok {
		r.Get(base, h).Name(resource + ".index")
	}
	if h, ok := methodHandler(v, "Create"); ok {
		r.Get(base+"/create", h).Name(resource + ".create")
	}
	if h, ok := methodHandler(v, "Show"); ok {
		r.Get(base+"/{id}", h).Name(resource + ".show")
	}
	if h, ok := methodHandler(v, "Store"); ok {
		r.Post(base, h).Name(resource + ".store")
	}
	if h, ok := methodHandler(v, "Edit"); ok {
		r.Get(base+"/{id}/edit", h).Name(resource + ".edit")
	}
	if h, ok := methodHandler(v, "Update"); ok {
		r.Put(base+"/{id}", h).Name(resource + ".update")
	}
	if h, ok := methodHandler(v, "Destroy"); ok {
		r.Delete(base+"/{id}", h).Name(resource + ".destroy")
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
	default:
		if val != nil {
			c.Set("data", val)
		}
	}
}

// Group creates a route group with a shared prefix and optional middleware.
// Routes defined inside fn inherit the global middleware plus any group-level middleware.
func (r *Router) Group(prefix string, fn func(*Router), middleware ...Middleware) {
	child := &Router{
		routes:           r.routes,
		middleware:       r.middleware,
		groupMiddleware:  append(append([]Middleware{}, r.groupMiddleware...), middleware...),
		view:             r.view,
		basePath:         r.withBase(prefix),
		notFound:         r.notFound,
		methodNotAllowed: r.methodNotAllowed,
		routeSet:         r.routeSet,
		namedRoutes:      r.namedRoutes,
	}
	fn(child)
	r.routes = child.routes
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := cleanPath(req.URL.Path)
	method := req.Method

	// Form method spoofing: allow _method=PUT|PATCH|DELETE from HTML POST forms.
	if method == http.MethodPost {
		if m := req.FormValue("_method"); m != "" {
			switch strings.ToUpper(m) {
			case http.MethodPut, http.MethodPatch, http.MethodDelete:
				req.Method = strings.ToUpper(m)
				method = req.Method
			}
		}
	}

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
		// Apply global middleware so CORS/auth middleware still run (e.g. OPTIONS preflight).
		r.applyMiddleware(r.methodNotAllowed, r.middleware...)(ctx)
		return
	}
	r.applyMiddleware(r.notFound, r.middleware...)(ctx)
}

func (r *Router) applyMiddleware(h Handler, m ...Middleware) Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
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
