# Repository Structure

## Framework packages

| Package | Role |
|---|---|
| `auth/` | Session auth (`Manager`), gate/abilities, API bearer tokens |
| `cache/` | `Cache` with memory + Redis stores, tag invalidation, response cache middleware |
| `cli/` | Internal command implementations (migrate, serve, schedule, cache:clear) |
| `cmd/gophant` | CLI binary + `make:*` code generators |
| `config/` | `.env` loader + `Config` struct |
| `container/` | Reflect-based IoC container for controller dependency injection |
| `db/` | `DB` wrapper, query builder, `TxQuery`, `Model` helpers, `Repository` with hooks |
| `db/migrate/` | `Migrator` — Up/Down/Fresh/Status, dialect-aware (MySQL + PostgreSQL) |
| `db/schema/` | `Blueprint`-style DDL builder |
| `http/` | `Router`, `Context`, `Middleware`, model binding, auto-routes |
| `middleware/` | Recover, Logger, CSRF, SecurityHeaders, ErrorHandler, RateLimit |
| `queue/` | Memory + Redis queues, job registry, retry + dead-letter |
| `scheduler/` | `Every(duration, task)` in-process scheduler + cron wrapper |
| `security/` | CSRF token gen/verify, bcrypt helpers, secure headers |
| `session/` | HMAC-signed cookie sessions, memory + Redis stores, flash data |
| `testkit/` | HTTP test helpers |
| `validation/` | Chainable validators with human-readable messages |
| `view/` | `html/template` wrapper with FuncMap support |

## Recommended app layout

```
myapp/
├── main.go                    # gophant.New() + app.Run()
├── routes.go                  # gophant.RegisterRoutes(...)
├── schedule.go                # gophant.RegisterSchedule(...)
├── .env
├── app/
│   ├── Controllers/           # HTTP controllers
│   ├── Models/                # db.Model wrappers
│   ├── Services/              # business logic
│   ├── Policies/              # gate policies
│   └── jobs/                  # queue jobs + registry.go
├── database/
│   └── migrations/            # timestamped migration files
└── views/
    ├── layouts/
    └── *.html
```

## Key design patterns

**init() registration** — Routes, migrations, and schedules are registered in `init()` functions (like database/sql drivers). The framework collects them and applies them during `New()`.

**Shared router state** — `Router.Group()` shares the `routes` slice and `namedRoutes`/`routeSet` maps between parent and child, so named routes defined inside a group are visible globally.

**Middleware order** — Applied innermost-first (last registered = outermost execution). Global middleware runs before route-level middleware.

**Session flash** — Flash values (`:flash:key`) are automatically injected into `c.Values` as `flash_key` at request start, so templates receive `{{ .flash_error }}` etc. without extra controller code.
