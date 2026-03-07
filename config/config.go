package config

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName            string
	AppEnv             string
	AppKey             string
	BaseURL            string
	Addr               string
	LogLevel           string
	SessionCookie      string
	SessionSecure      bool
	SessionMaxAgeHours int
	DBDriver           string
	DBDsn              string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	QueueKey           string
	QueueWorkers       int
	QueueDeadPrefix    string
	CacheDriver        string
	CachePrefix        string
}

func Load() *Config {
	_ = LoadEnvFile(".env")

	cfg := &Config{
		AppName:            getenv("APP_NAME", "gophant"),
		AppEnv:             getenv("APP_ENV", "local"),
		AppKey:             getenv("APP_KEY", ""),
		BaseURL:            getenv("APP_URL", "http://localhost:8080"),
		Addr:               getenv("APP_ADDR", ":8080"),
		LogLevel:           getenv("LOG_LEVEL", "info"),
		SessionCookie:      getenv("SESSION_COOKIE", "_gophant_session"),
		SessionSecure:      getenvBool("SESSION_SECURE", false),
		SessionMaxAgeHours: getenvInt("SESSION_MAX_AGE_HOURS", 168),
		DBDriver:           getenv("DB_DRIVER", ""),
		DBDsn:              getenv("DB_DSN", ""),
		RedisAddr:          getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getenv("REDIS_PASSWORD", ""),
		RedisDB:            getenvInt("REDIS_DB", 0),
		QueueKey:           getenv("QUEUE_KEY", "gophant:queue"),
		QueueWorkers:       getenvInt("QUEUE_WORKERS", 1),
		QueueDeadPrefix:    getenv("QUEUE_DEAD_PREFIX", "gophant:dead"),
		CacheDriver:        getenv("CACHE_DRIVER", "memory"),
		CachePrefix:        getenv("CACHE_PREFIX", "gophant:cache:"),
	}

	if cfg.AppKey == "" {
		cfg.AppKey = randomKey()
	}
	return cfg
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		v = strings.ToLower(v)
		return v == "1" || v == "true" || v == "yes"
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func randomKey() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
