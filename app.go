package gophant

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mahavirnahata/gophant/auth"
	"github.com/mahavirnahata/gophant/cache"
	"github.com/mahavirnahata/gophant/config"
	"github.com/mahavirnahata/gophant/container"
	"github.com/mahavirnahata/gophant/db"
	"github.com/mahavirnahata/gophant/events"
	gomvchttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/mail"
	"github.com/mahavirnahata/gophant/middleware"
	"github.com/mahavirnahata/gophant/security"
	"github.com/mahavirnahata/gophant/session"
	"github.com/mahavirnahata/gophant/storage"
	"github.com/mahavirnahata/gophant/view"
	"github.com/redis/go-redis/v9"
)

// App is the root application object. Create one with New() and call Run().
type App struct {
	Config    *config.Config
	View      *view.Engine
	Router    *gomvchttp.Router
	Session   *session.Manager
	Auth      *auth.Manager
	Gate      *auth.Gate
	Cache     *cache.Cache
	DB        *db.DB
	Container *container.Container
	Mailer    *mail.Mailer
	Events    *events.Bus
	Storage   *storage.Storage
}

func New() *App {
	cfg := config.Load()

	v := view.New("views")

	r := gomvchttp.NewRouter(v)
	r.Use(middleware.Recover())
	r.Use(middleware.Logger())
	r.Use(middleware.SecurityHeaders(security.DefaultHeaders()))
	r.Use(middleware.ErrorHandler(nil))

	sess := session.NewManager([]byte(cfg.AppKey))
	sess.CookieName = cfg.SessionCookie
	sess.Secure = cfg.SessionSecure || cfg.AppEnv == "production"
	sess.MaxAge = time.Duration(cfg.SessionMaxAgeHours) * time.Hour
	r.Use(sess.Middleware())

	r.Use(middleware.CSRF(middleware.CSRFConfig{
		Secret:       []byte(cfg.AppKey),
		SecureCookie: cfg.AppEnv == "production",
	}))

	app := &App{
		Config:    cfg,
		View:      v,
		Router:    r,
		Session:   sess,
		Auth:      auth.NewManager(),
		Gate:      auth.NewGate(),
		Cache:     newCache(cfg),
		Container: container.New(),
		Mailer:    newMailer(cfg),
		Events:    events.NewBus(),
		Storage:   storage.New(storage.NewLocalDriver("storage/app", "/storage")),
	}

	// Auto-connect database when DB_DRIVER and DB_DSN are configured.
	if cfg.DBDriver != "" && cfg.DBDsn != "" {
		conn, err := db.Open(cfg.DBDriver, cfg.DBDsn, nil)
		if err != nil {
			log.Printf("db: failed to connect (%s): %v", cfg.DBDriver, err)
		} else {
			conn.Conn.SetMaxOpenConns(cfg.DBMaxOpenConns)
			conn.Conn.SetMaxIdleConns(cfg.DBMaxIdleConns)
			if cfg.DBConnMaxLifetime > 0 {
				conn.Conn.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetime) * time.Second)
			}
			app.DB = conn
			db.SetDefaultDB(conn)
		}
	}

	// Register built-in template functions after the router exists so url() can close over it.
	v.AddFunc("url", r.URL)
	v.AddFunc("asset", func(path string) string { return "/" + path })
	_ = v.Load("**/*.html")

	app.Container.Set(app)
	app.Container.Set(app.Auth)
	app.Container.Set(app.Gate)
	app.Container.Set(app.Cache)
	app.Container.Set(app.Session)
	app.Container.Set(app.Mailer)
	app.Container.Set(app.Events)
	app.Container.Set(app.Storage)

	r.Use(func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			c.Set("app", app)
			c.Set("container", app.Container)
			c.Set("_router", r)
			next(c)
		}
	})

	applyRegisteredRoutes(app)

	return app
}

// HealthCheck registers a liveness endpoint (default: GET /health) that returns
// {"status":"ok"} with a 200. Pass a custom path to override.
func (a *App) HealthCheck(path ...string) {
	p := "/health"
	if len(path) > 0 && path[0] != "" {
		p = path[0]
	}
	a.Router.Get(p, func(c *gomvchttp.Context) {
		c.JSON(200, map[string]string{"status": "ok"})
	})
}

// Run starts the HTTP server and blocks until SIGINT or SIGTERM is received,
// then drains in-flight requests with a 30-second timeout.
func (a *App) Run() {
	srv := &http.Server{
		Addr:         a.Config.Addr,
		Handler:      a.Router,
		ReadTimeout:  time.Duration(a.Config.ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.Config.ServerWriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(a.Config.ServerIdleTimeout) * time.Second,
	}

	go func() {
		log.Printf("starting %s on %s (env=%s)", a.Config.AppName, a.Config.Addr, a.Config.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("received %s, shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}

func newMailer(cfg *config.Config) *mail.Mailer {
	var driver mail.Driver
	switch cfg.MailDriver {
	case "smtp":
		driver = mail.NewSMTPDriver(mail.SMTPConfig{
			Host:     cfg.MailHost,
			Port:     cfg.MailPort,
			Username: cfg.MailUsername,
			Password: cfg.MailPassword,
			From:     fmt.Sprintf("%s <%s>", cfg.MailFromName, cfg.MailFromAddress),
		})
	case "null":
		driver = mail.NewNullDriver()
	default:
		driver = mail.NewLogDriver()
	}
	m := mail.New(driver)
	m.From = fmt.Sprintf("%s <%s>", cfg.MailFromName, cfg.MailFromAddress)
	return m
}

func newCache(cfg *config.Config) *cache.Cache {
	switch cfg.CacheDriver {
	case "redis":
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		store := cache.NewRedisStore(client)
		store.Prefix = cfg.CachePrefix
		return cache.New(store)
	default:
		return cache.New(cache.NewMemoryStore())
	}
}
