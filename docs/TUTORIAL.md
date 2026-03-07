# Gophant Tutorial (Developer Friendly)

This tutorial teaches Go **while** building a small MVC app.

## Part 1: Hello Gophant

### 1) Create `main.go`

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

**Go concepts learned:**
- packages, imports
- structs + methods
- receiver `(h *HomeController)`

---

## Part 2: Views + Data

```go
func (h *HomeController) Index(c *gophanthttp.Context) {
	c.Render(200, "home.html", map[string]any{
		"title": "Gophant",
	})
}
```

**Go concepts learned:**
- map literals (`map[string]any`)
- template data binding

---

## Part 3: Validation

```go
func (h *HomeController) Store(c *gophanthttp.Context) {
	v := validation.New(c.Request).
		Field("email", validation.Required(), validation.Email())

	if v.Fails() {
		c.JSON(422, map[string]any{"errors": v.Errors()})
		return
	}
	c.JSON(200, map[string]any{"ok": true})
}
```

**Go concepts learned:**
- method chaining
- early return pattern

---

## Part 4: Binding Form Input

```go
type SignupForm struct {
	Email string `form:"email"`
	Age   int    `form:"age"`
}

func (h *HomeController) Store(c *gophanthttp.Context) {
	var form SignupForm
	_ = c.BindForm(&form)
}
```

**Go concepts learned:**
- struct tags
- type safety

---

## Part 5: DB Queries

```go
conn, _ := db.Open("mysql", dsn, db.QuestionDialect{})

users, _ := conn.Table("users").
	Where("active", "=", true).
	OrderBySafe("id", "DESC", []string{"id"}).
	Get()
```

**Go concepts learned:**
- chaining
- SQL placeholders

---

## Part 6: Auth + Policies

```go
app.Auth.Login(c, "1")

app.Gate.Define("view-admin", func(c *gophanthttp.Context) bool {
	return app.Auth.HasRole(c, "admin")
})
```

**Go concepts learned:**
- closures
- function types

---

## Part 7: Queue + Jobs

```go
type SendWelcomeEmail struct{ UserID int64 }
func (j *SendWelcomeEmail) Handle() error { return nil }

jobs.Registry.RegisterType(&SendWelcomeEmail{}, func() queue.JobHandler { return &SendWelcomeEmail{} })
```

**Go concepts learned:**
- interfaces
- registration patterns

---

## Part 8: Cache

```go
_ = app.Cache.Set("key", "value", time.Minute)
```

**Go concepts learned:**
- interfaces + composition

---

## Part 9: Testing Basics

```go
func TestSomething(t *testing.T) {
	if 1+1 != 2 { t.Fail() }
}
```

**Go concepts learned:**
- table tests and mocks

---

## Next Steps
- Build a small CRUD app.
- Add one feature at a time.
- Improve the framework only when you feel real friction.
