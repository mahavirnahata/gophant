package gophant

import (
	"log"
	"net/http"
	"time"

	"github.com/mahavirnahata/gophant/auth"
	"github.com/mahavirnahata/gophant/cache"
	"github.com/mahavirnahata/gophant/config"
	"github.com/mahavirnahata/gophant/container"
	"github.com/mahavirnahata/gophant/db"
	gomvchttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/middleware"
	"github.com/mahavirnahata/gophant/security"
	"github.com/mahavirnahata/gophant/session"
	"github.com/mahavirnahata/gophant/view"
	"github.com/redis/go-redis/v9"
)

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
}

func New() *App {
	cfg := config.Load()
	v := view.New("views")
	_ = v.Load("**/*.html")

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
	}

	app.Container.Set(app)
	app.Container.Set(app.Auth)
	app.Container.Set(app.Gate)
	app.Container.Set(app.Cache)
	app.Container.Set(app.Session)

	r.Use(func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			c.Set("app", app)
			c.Set("container", app.Container)
			next(c)
		}
	})

	applyRegisteredRoutes(app)

	return app
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

func (a *App) Run() {
	log.Printf("starting %s on %s", a.Config.AppName, a.Config.Addr)
	if err := http.ListenAndServe(a.Config.Addr, a.Router); err != nil {
		log.Fatal(err)
	}
}
