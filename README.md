# Gophant — A Clean, Convention‑Over‑Config Go MVC Framework

Gophant is a **Go MVC framework** designed for clarity, safety, and speed. It favors **convention over configuration**, keeps the API explicit, and provides the productivity features you expect in a modern web framework: routing, controllers, views, validation, ORM helpers, migrations, queue, scheduler, cache, sessions, auth, and API tokens.

If you’re looking for a **developer‑friendly Go web framework** that feels familiar, structured, and production‑ready, Gophant is built for that.

## Why Gophant

- **Convention over configuration**: sensible defaults, fewer wires.
- **MVC done right**: controllers, views, models, middleware, policies.
- **Security first**: CSRF, secure headers, bcrypt, param binding.
- **Fast feedback**: clean CLI, generators, and strong testability.
- **Go‑native**: explicit types, minimal magic, easy to understand.

## Features (Short Summary)

- **Routing**: HTTP verbs, groups, middleware, and auto‑routing.
- **Controllers**: struct methods with dependency injection.
- **Views**: HTML templates + auto view rendering.
- **Request binding**: bind request data to structs.
- **Validation**: chainable validators with error maps.
- **ORM helpers**: model helpers, hooks, relations, eager loading.
- **Query builder**: safe, parameterized SQL with dialects.
- **Migrations**: schema builder + migrate / rollback / status.
- **Cache**: memory or Redis with tagging and response cache.
- **Queue**: in‑memory or Redis queues with retries + dead‑letter.
- **Scheduler**: periodic tasks with `schedule:run` / `schedule:work`.
- **Auth**: session auth, roles, gate/policies.
- **API tokens**: bearer tokens with abilities and expiry.
- **Security defaults**: CSRF, secure headers, session rotation.
- **Testing**: helpers + clean package‑level tests.

## 5‑Minute Quickstart

```bash
mkdir myapp && cd myapp
go mod init myapp
go get github.com/mahavirnahata/gophant
```

Create `main.go`:

```go
package main

import (
	"github.com/mahavirnahata/gophant"
	gophanthttp "github.com/mahavirnahata/gophant/http"
)

type HomeController struct{}

func (h *HomeController) Index(c *gophanthttp.Context) {
	c.Text(200, "Hello Gophant")
}

func main() {
	app := gophant.New()
	app.Router.Get("/", (&HomeController{}).Index)
	app.Run()
}
```

Run:

```bash
go run .
```

Open `http://localhost:8080`.

## Quick Example (Routing + Validation)

```go
import (
	"github.com/mahavirnahata/gophant"
	gophanthttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/validation"
)

type HomeController struct{}

func (h *HomeController) Store(c *gophanthttp.Context) {
	v := validation.New(c.Request).
		Field("email", validation.Required(), validation.Email())

	if v.Fails() {
		c.JSON(422, map[string]any{"errors": v.Errors()})
		return
	}
	c.JSON(200, map[string]any{"ok": true})
}

func main() {
	app := gophant.New()
	app.Router.Post("/submit", (&HomeController{}).Store)
	app.Run()
}
```

## CLI

Build once:

```bash
go build -o gophant ./cmd/gophant
./gophant serve
```

Common commands:

- `gophant serve`
- `gophant migrate`
- `gophant migrate:rollback [steps]`
- `gophant migrate:fresh`
- `gophant migrate:status`
- `gophant queue:work`
- `gophant queue:retry --max 100`
- `gophant cache:clear`
- `gophant make:controller User`
- `gophant make:model User`
- `gophant make:migration create_users_table`
- `gophant make:policy UserPolicy`
- `gophant make:job SendWelcomeEmail`
- `gophant make:auth`
- `gophant make:routes`
- `gophant make:schedule`

## Configuration (.env)

```
APP_NAME=gophant
APP_ENV=local
APP_KEY=change-me
APP_ADDR=:8080

DB_DRIVER=mysql
DB_DSN=user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
QUEUE_KEY=gophant:queue
QUEUE_WORKERS=1
QUEUE_DEAD_PREFIX=gophant:dead

CACHE_DRIVER=memory
CACHE_PREFIX=gophant:cache:
```

## Short Feature Examples

**Routing + Middleware**

```go
app.Router.Group("/admin", func(r *gophanthttp.Router) {
	r.Get("/dashboard", admin.Dashboard, app.Auth.RequireRole("admin"))
})
```

**Query Builder**

```go
conn, _ := db.Open("mysql", dsn, db.QuestionDialect{})
db.SetDefaultDB(conn)
users, _ := conn.Table("users").
	Where("active", "=", true).
	OrderBySafe("id", "DESC", []string{"id"}).
	Get()
```

**Model Helpers**

```go
var users []models.User
err := models.UserModel().
	Where("active", "=", true).
	OrderBySafe("id", "DESC", []string{"id"}).
	GetStructs(&users)
```

**Queue**

```go
q := queue.NewRedisQueue(client, "gophant:queue")
reg := queue.NewRegistry()
reg.RegisterType(&SendWelcomeEmail{}, func() queue.JobHandler { return &SendWelcomeEmail{} })

go queue.RunWithRetry(q, reg, queue.NewDeadLetter(q, "gophant:dead"))
```

**Cache**

```go
_ = app.Cache.Set("key", "value", time.Minute)
val, ok := app.Cache.Get("key")
```

## Security Defaults

- CSRF protection with `SameSite=Lax` cookies
- Secure headers (`nosniff`, `frame‑deny`, `referrer‑policy`)
- Password hashing via bcrypt
- Parameterized SQL only
- Session rotation on login

## Docs

- `docs/GETTING_STARTED.md`
- `docs/TUTORIAL.md`
- `docs/MAGIC.md`
- `docs/MIGRATIONS.md`
- `docs/QUEUE_USAGE.md`
- `docs/SCHEDULER.md`
- `docs/TOKENS.md`
- `docs/CLI.md`
- `docs/STRUCTURE.md`
- `docs/RELEASE_CHECKLIST.md`

## Public API Stability

We aim to keep these packages stable:

- `gophant`
- `gophant/http`
- `gophant/middleware`
- `gophant/db`
- `gophant/auth`
- `gophant/cache`
- `gophant/queue`
- `gophant/cli`

Internal helpers may change without notice.

## License

MIT
