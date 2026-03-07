# Magic Features (Convention‑Based)

Gophant supports opt‑in "magic" that reduces boilerplate while keeping Go explicit when needed.

## Resource Routes

```go
app.Router.Resource("users", &UserController{})
```

Generates:
- `GET /users` → `Index`
- `GET /users/{id}` → `Show`
- `POST /users` → `Store`
- `PUT /users/{id}` → `Update`
- `DELETE /users/{id}` → `Destroy`

You can still add explicit custom routes with `Get`, `Post`, etc.

## Resource Controller Generator

```bash
gophant make:controller UserController --resource
```

## AutoRoutes

```go
app.Router.AutoRoutes(&UserController{}, &AuthController{})
```

Registers routes by convention using controller method names.

## Auto View Rendering

You can return a `map[string]any` and Gophant will render a view automatically
if you set `c.AutoView("path.html")` or if you return a string view name.

```go
func (c *HomeController) Index(ctx *gophanthttp.Context) map[string]any {
	ctx.AutoView("home/index.html")
	return map[string]any{"title": "Home"}
}
```

## Implicit Model Binding

Register a model binder once:

```go
http.RegisterModelBinder(models.User{}, http.DefaultModelBinder("users", "id"))
```

Then your controller can receive the model directly:

```go
func (c *UserController) Show(ctx *gophanthttp.Context, user models.User) {
	ctx.JSON(200, user)
}
```

## Auto Dependency Injection

Bind services in the container:

```go
app := gophant.New()
app.Container.Bind(&MyService{}, func() any { return &MyService{} })
```

Then receive them in controller methods:

```go
func (c *UserController) Show(ctx *gophanthttp.Context, svc *MyService) {
	// svc injected automatically
}
```

## Controller Auto‑Discovery

Register controllers once (usually in `init()`):

```go
func init() {
	http.RegisterController(&UserController{})
	http.RegisterController(&AuthController{})
}
```

Then enable auto‑discovery:

```go
app.Router.AutoDiscover()
```
