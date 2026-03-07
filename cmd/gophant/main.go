package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mahavirnahata/gophant/cli"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	if cmd == "serve" || cmd == "migrate" || cmd == "migrate:rollback" || cmd == "queue:work" || cmd == "queue:retry" || cmd == "migrate:fresh" || cmd == "migrate:status" || cmd == "cache:clear" {
		if err := runCommand(cmd); err != nil {
			fatal(err)
		}
		return
	}

	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	name := os.Args[2]
	resource := hasFlag("--resource")

	switch cmd {
	case "make:controller":
		err := makeController(name, resource)
		if err != nil {
			fatal(err)
		}
	case "make:service":
		err := makeService(name)
		if err != nil {
			fatal(err)
		}
	case "make:model":
		err := makeModel(name)
		if err != nil {
			fatal(err)
		}
	case "make:migration":
		err := makeMigration(name)
		if err != nil {
			fatal(err)
		}
	case "make:policy":
		err := makePolicy(name)
		if err != nil {
			fatal(err)
		}
	case "make:job":
		err := makeJob(name)
		if err != nil {
			fatal(err)
		}
	case "make:auth":
		err := makeAuth()
		if err != nil {
			fatal(err)
		}
	case "make:routes":
		err := makeRoutes()
		if err != nil {
			fatal(err)
		}
	case "make:bootstrap":
		err := makeBootstrap()
		if err != nil {
			fatal(err)
		}
	case "make:schedule":
		err := makeSchedule()
		if err != nil {
			fatal(err)
		}
	case "schedule:run":
		err := cliScheduleRun()
		if err != nil {
			fatal(err)
		}
	case "schedule:work":
		err := cliScheduleWork()
		if err != nil {
			fatal(err)
		}
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("gophant make:controller Name")
	fmt.Println("gophant make:controller Name --resource")
	fmt.Println("gophant make:service Name")
	fmt.Println("gophant make:model Name")
	fmt.Println("gophant make:migration create_users_table")
	fmt.Println("gophant make:policy UserPolicy")
	fmt.Println("gophant make:job SendWelcomeEmail")
	fmt.Println("gophant make:auth")
	fmt.Println("gophant make:routes")
	fmt.Println("gophant make:bootstrap")
	fmt.Println("gophant make:schedule")
	fmt.Println("gophant schedule:run")
	fmt.Println("gophant schedule:work [--interval N]")
	fmt.Println("gophant serve")
	fmt.Println("gophant migrate")
	fmt.Println("gophant migrate:rollback [steps]")
	fmt.Println("gophant queue:work")
	fmt.Println("gophant queue:retry [--max N]")
	fmt.Println("gophant migrate:fresh")
	fmt.Println("gophant migrate:status")
	fmt.Println("gophant cache:clear")
}

func makeController(name string, resource bool) error {
	if name == "" {
		return errors.New("name required")
	}
	if !strings.HasSuffix(name, "Controller") {
		name += "Controller"
	}
	pkg := "controllers"
	filename := filepath.Join("app", "Controllers", toSnake(name)+".go")

	methods := ""
	if resource {
		methods = `
func (c *%s) Index(ctx *gophanthttp.Context) { ctx.Text(200, "index") }
func (c *%s) Show(ctx *gophanthttp.Context) { ctx.Text(200, "show") }
func (c *%s) Store(ctx *gophanthttp.Context) { ctx.Text(200, "store") }
func (c *%s) Update(ctx *gophanthttp.Context) { ctx.Text(200, "update") }
func (c *%s) Destroy(ctx *gophanthttp.Context) { ctx.Text(200, "destroy") }
`
	}

	return writeFile(filename, fmt.Sprintf(`package %s

import (
	gophanthttp "github.com/mahavirnahata/gophant/http"
)

type %s struct{}

func init() {
	gophanthttp.RegisterController(&%s{})
}

func (c *%s) Index(ctx *gophanthttp.Context) {
	ctx.Text(200, "ok")
}
%s
`, pkg, name, name, name, fmt.Sprintf(methods, name, name, name, name, name)))
}

func makeService(name string) error {
	if name == "" {
		return errors.New("name required")
	}
	pkg := "services"
	filename := filepath.Join("app", "Services", toSnake(name)+".go")

	return writeFile(filename, fmt.Sprintf(`package %s

type %s struct{}

func (s *%s) Example() string {
	return "ok"
}
`, pkg, name, name))
}

func makeModel(name string) error {
	if name == "" {
		return errors.New("name required")
	}
	pkg := "models"
	filename := filepath.Join("app", "Models", toSnake(name)+".go")

	return writeFile(filename, fmt.Sprintf(`package %s

import "github.com/mahavirnahata/gophant/db"

type %s struct {
	ID int64 `+"`db:\"id\"`"+`
}

func %sModel() *db.Model {
	return db.NewModel(nil, "%s")
}
`, pkg, name, name, toSnake(name)+"s"))
}

func makeMigration(name string) error {
	if name == "" {
		return errors.New("name required")
	}
	ts := time.Now().Format("20060102150405")
	filename := filepath.Join("database", "migrations", ts+"_"+toSnake(name)+".go")
	migID := ts + "_" + toSnake(name)
	return writeFile(filename, fmt.Sprintf(`package migrations

import (
	"database/sql"

	"github.com/mahavirnahata/gophant/db/migrate"
	"github.com/mahavirnahata/gophant/db/schema"
)

func init() {
	Register(migrate.Migration{
		ID: "%s",
		Up: func(db *sql.DB) error {
			b := schema.New("mysql")
			bp, sql := b.Build("example", func(t *schema.Blueprint) {
				t.Increments("id")
				t.String("name", 255)
				t.Timestamps()
				t.Unique("name")
				t.Index("name")
			})
			if _, err := db.Exec(sql); err != nil {
				return err
			}
			for _, idx := range b.Indexes(bp) {
				if _, err := db.Exec(idx); err != nil {
					return err
				}
			}
			return nil
		},
		Down: func(db *sql.DB) error {
			b := schema.New("mysql")
			_, err := db.Exec(b.Drop("example"))
			return err
		},
	})
}
`, migID))
}

func makePolicy(name string) error {
	if name == "" {
		return errors.New("name required")
	}
	if !strings.HasSuffix(name, "Policy") {
		name += "Policy"
	}
	filename := filepath.Join("app", "Policies", toSnake(name)+".go")
	return writeFile(filename, fmt.Sprintf(`package policies

import (
	gophanthttp "github.com/mahavirnahata/gophant/http"
)

type %s struct{}

func (p *%s) ViewAny(c *gophanthttp.Context) bool { return false }
func (p *%s) View(c *gophanthttp.Context) bool { return false }
func (p *%s) Create(c *gophanthttp.Context) bool { return false }
func (p *%s) Update(c *gophanthttp.Context) bool { return false }
func (p *%s) Delete(c *gophanthttp.Context) bool { return false }
`, name, name, name, name, name, name))
}

func makeJob(name string) error {
	if name == "" {
		return errors.New("name required")
	}
	if err := ensureJobsRegistry(); err != nil {
		return err
	}
	filename := filepath.Join("app", "jobs", toSnake(name)+".go")
	return writeFile(filename, fmt.Sprintf(`package jobs

import (
	"time"

	"github.com/mahavirnahata/gophant/queue"
)

type %s struct {
	// TODO: add fields
}

func (j *%s) Handle() error {
	return nil
}

func (j *%s) Retries() int {
	return 0
}

func (j *%s) Backoff() time.Duration {
	return 0
}

func init() {
	Registry.RegisterType(&%s{}, func() queue.JobHandler { return &%s{} })
}
	`, name, name, name, name, name, name))
}

func makeAuth() error {
	if err := os.MkdirAll(filepath.Join("app", "Controllers"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join("views", "auth"), 0o755); err != nil {
		return err
	}

	ctrlPath := filepath.Join("app", "Controllers", "auth_controller.go")
	if _, err := os.Stat(ctrlPath); err == nil {
		return fmt.Errorf("file exists: %s", ctrlPath)
	}
	ctrl := `package controllers

import (
	"net/http"

	"github.com/mahavirnahata/gophant"
	gophanthttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/security"
	"github.com/mahavirnahata/gophant/validation"
)

type AuthController struct{}

func (a *AuthController) LoginForm(c *gophanthttp.Context) {
	c.Render(200, "auth/login.html", map[string]any{})
}

func (a *AuthController) RegisterForm(c *gophanthttp.Context) {
	c.Render(200, "auth/register.html", map[string]any{})
}

func (a *AuthController) ForgotForm(c *gophanthttp.Context) {
	c.Render(200, "auth/forgot.html", map[string]any{})
}

func (a *AuthController) Login(c *gophanthttp.Context) {
	v := validation.New(c.Request).
		Field("email", validation.Required(), validation.Email()).
		Field("password", validation.Required(), validation.Min(6))

	if v.Fails() {
		c.JSON(422, map[string]any{"errors": v.Errors()})
		return
	}

	appVal, _ := c.Get("app")
	app := appVal.(*gophant.App)
	app.Auth.Login(c, "1")
	c.Redirect(http.StatusFound, "/")
}

func (a *AuthController) Register(c *gophanthttp.Context) {
	v := validation.New(c.Request).
		Field("email", validation.Required(), validation.Email()).
		Field("password", validation.Required(), validation.Min(6)).
		FieldWith("password", validation.Confirmed("password_confirmation"))

	if v.Fails() {
		c.JSON(422, map[string]any{"errors": v.Errors()})
		return
	}

	_ = security.HashPassword(c.Request.FormValue("password"))
	c.Redirect(http.StatusFound, "/login")
}

func (a *AuthController) Logout(c *gophanthttp.Context) {
	appVal, _ := c.Get("app")
	app := appVal.(*gophant.App)
	app.Auth.Logout(c)
	c.Redirect(http.StatusFound, "/")
}
`
	if err := os.WriteFile(ctrlPath, []byte(ctrl), 0o644); err != nil {
		return err
	}

	tmpls := map[string]string{
		"views/auth/login.html":    loginTemplate,
		"views/auth/register.html": registerTemplate,
		"views/auth/forgot.html":   forgotTemplate,
	}
	for path, content := range tmpls {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("file exists: %s", path)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}

	return nil
}

const loginTemplate = `{{ define "auth/login.html" }}
<!doctype html>
<html>
  <head><meta charset="utf-8" /><title>Login</title></head>
  <body>
    <h1>Login</h1>
    <form method="POST" action="/login">
      <input type="hidden" name="_token" value="{{ .csrf }}" />
      <input name="email" placeholder="Email" />
      <input name="password" type="password" placeholder="Password" />
      <button type="submit">Login</button>
    </form>
  </body>
</html>
{{ end }}`

const registerTemplate = `{{ define "auth/register.html" }}
<!doctype html>
<html>
  <head><meta charset="utf-8" /><title>Register</title></head>
  <body>
    <h1>Register</h1>
    <form method="POST" action="/register">
      <input type="hidden" name="_token" value="{{ .csrf }}" />
      <input name="email" placeholder="Email" />
      <input name="password" type="password" placeholder="Password" />
      <input name="password_confirmation" type="password" placeholder="Confirm Password" />
      <button type="submit">Register</button>
    </form>
  </body>
</html>
{{ end }}`

const forgotTemplate = `{{ define "auth/forgot.html" }}
<!doctype html>
<html>
  <head><meta charset="utf-8" /><title>Forgot Password</title></head>
  <body>
    <h1>Forgot Password</h1>
    <form method="POST" action="/forgot">
      <input type="hidden" name="_token" value="{{ .csrf }}" />
      <input name="email" placeholder="Email" />
      <button type="submit">Send Reset Link</button>
    </form>
  </body>
</html>
{{ end }}`

func makeRoutes() error {
	path := filepath.Join("routes.go")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file exists: %s", path)
	}
	content := `package main

import (
	"github.com/mahavirnahata/gophant"
)

func init() {
	gophant.RegisterRoutes(func(app *gophant.App) {
		// app.Router.AutoDiscover()
		// app.Router.Get("/", (&controllers.HomeController{}).Index)
	})
}
`
	return os.WriteFile(path, []byte(content), 0o644)
}

func makeBootstrap() error {
	path := filepath.Join("main.go")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file exists: %s", path)
	}
	content := `package main

import "github.com/mahavirnahata/gophant"

func main() {
	app := gophant.New()
	app.Run()
}
`
	return os.WriteFile(path, []byte(content), 0o644)
}

func makeSchedule() error {
	path := filepath.Join("schedule.go")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file exists: %s", path)
	}
	content := `package main

import (
	"time"

	"github.com/mahavirnahata/gophant"
	"github.com/mahavirnahata/gophant/scheduler"
)

func init() {
	gophant.RegisterSchedule(func(s *scheduler.Scheduler) {
		s.Every(time.Minute, func() error {
			// do something
			return nil
		})
	})
}
`
	return os.WriteFile(path, []byte(content), 0o644)
}

func cliScheduleRun() error {
	return cli.ScheduleRunOnce()
}

func cliScheduleWork() error {
	return cli.ScheduleRunLoop()
}

func ensureJobsRegistry() error {
	path := filepath.Join("app", "jobs", "registry.go")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(`package jobs

import "github.com/mahavirnahata/gophant/queue"

var Registry = queue.NewRegistry()
`), 0o644)
}

func writeFile(path string, contents string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file exists: %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(contents), 0o644)
}

func toSnake(name string) string {
	name = strings.TrimSpace(name)
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	s := re.ReplaceAllString(name, "${1}_${2}")
	s = strings.ToLower(s)
	return s
}

func hasFlag(name string) bool {
	for _, a := range os.Args {
		if a == name {
			return true
		}
	}
	return false
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
