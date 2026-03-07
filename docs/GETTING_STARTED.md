# Getting Started → Production Deploy

This guide takes you from an empty folder to a production‑ready Gophant app.

## 1) Create a new project

```bash
mkdir myapp && cd myapp

go mod init myapp

go get github.com/mahavirnahata/gophant
```

## 2) Create `main.go`

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

## 3) Run in 5 minutes

```bash
go run .
```

Open:
```
http://localhost:8080
```

## 3) Configure `.env`

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

CACHE_DRIVER=memory
CACHE_PREFIX=gophant:cache:
```

## 4) Build a CLI

```bash
go build -o gophant-cli ./cmd/gophant
./gophant-cli serve
```

## 5) Production deploy

```bash
go build -o app ./cmd/yourapp
./app
```

Recommendations:
- Set `APP_ENV=production`
- Set a strong `APP_KEY`
- Use a real DB and Redis
- Use a process manager (systemd, supervisor)

## Error Handling

Gophant includes a default error handler middleware. In controllers you can do:

```go
if err != nil {
    c.Error(err)
    return
}
```

If no response was written, Gophant returns a `500` JSON error by default.
