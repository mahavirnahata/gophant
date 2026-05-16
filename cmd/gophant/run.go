package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/mahavirnahata/gophant"
	"github.com/mahavirnahata/gophant/cli"
	"github.com/mahavirnahata/gophant/config"
	"github.com/mahavirnahata/gophant/jobs"
	"github.com/mahavirnahata/gophant/queue"
	"github.com/redis/go-redis/v9"
)

func runCommand(cmd string) error {
	switch cmd {
	case "serve":
		app := gophant.New()
		return cli.Serve(app, cli.ServeOptions{})
	case "migrate":
		db, err := openDB()
		if err != nil {
			return err
		}
		return cli.MigrateUp(db)
	case "migrate:rollback":
		db, err := openDB()
		if err != nil {
			return err
		}
		steps := 1
		if len(os.Args) >= 3 {
			if n, err := strconv.Atoi(os.Args[2]); err == nil {
				steps = n
			}
		}
		return cli.MigrateDown(db, cli.MigrateOptions{Steps: steps})
	case "migrate:fresh":
		db, err := openDB()
		if err != nil {
			return err
		}
		return cli.MigrateFresh(db)
	case "migrate:status":
		db, err := openDB()
		if err != nil {
			return err
		}
		return cli.MigrateStatus(db, cli.StatusOptions{JSON: hasFlag("--json")})
	case "queue:work":
		return runQueueWorkers()
	case "queue:retry":
		return retryDead()
	case "cache:clear":
		return cacheClear()
	case "route:list":
		app := gophant.New()
		cli.RouteList(app.Router)
		return nil
	case "db:seed":
		db, err := openDB()
		if err != nil {
			return err
		}
		return cli.SeedRun(db)
	default:
		return errors.New("unknown command")
	}
}

func openDB() (*sql.DB, error) {
	cfg := config.Load()
	if cfg.DBDriver == "" || cfg.DBDsn == "" {
		return nil, fmt.Errorf("db connection not configured: set DB_DRIVER and DB_DSN")
	}
	db, err := sql.Open(cfg.DBDriver, cfg.DBDsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func runQueueWorkers() error {
	cfg := config.Load()
	if cfg.RedisAddr == "" {
		return fmt.Errorf("redis not configured: set REDIS_ADDR")
	}
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	q := queue.NewRedisQueue(client, cfg.QueueKey)
	dl := queue.NewDeadLetter(q, cfg.QueueDeadPrefix)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	opts := queue.WorkOptions{
		Once:  hasFlag("--once"),
		Sleep: timeSecondsFlag("--sleep"),
	}
	return queue.RunWorkers(ctx, q, jobs.Registry, dl, cfg.QueueWorkers, opts)
}

func retryDead() error {
	cfg := config.Load()
	if cfg.RedisAddr == "" {
		return fmt.Errorf("redis not configured: set REDIS_ADDR")
	}
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	q := queue.NewRedisQueue(client, cfg.QueueKey)
	dl := queue.NewDeadLetter(q, cfg.QueueDeadPrefix)
	max := intFlag("--max", 0)
	count, err := queue.RetryDead(q, dl, max)
	if err != nil {
		return err
	}
	fmt.Printf("requeued %d jobs\n", count)
	return nil
}

func cacheClear() error {
	app := gophant.New()
	if err := app.Cache.Flush(); err != nil {
		return err
	}
	fmt.Println("cache cleared")
	return nil
}
