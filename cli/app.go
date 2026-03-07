package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mahavirnahata/gophant"
	"github.com/mahavirnahata/gophant/config"
	"github.com/mahavirnahata/gophant/db"
	"github.com/mahavirnahata/gophant/db/migrate"
)

type AppFactory func() *gophant.App

type AppOptions struct {
	Migrations []migrate.Migration
	Seed       func(*db.DB) error
}

func RunApp(args []string, factory AppFactory, opts AppOptions) error {
	if factory == nil {
		return errors.New("app factory required")
	}
	if len(args) == 0 {
		return errors.New("command required")
	}

	cmd := args[0]
	switch cmd {
	case "serve":
		app := factory()
		if app.Config != nil && app.Config.DBDriver != "" && app.Config.DBDsn != "" {
			dbConn, err := openDBFromConfig(app.Config)
			if err != nil {
				return err
			}
			app.DB = dbConn
			db.SetDefaultDB(dbConn)
		}
		return Serve(app, ServeOptions{})
	case "migrate":
		app := factory()
		dbConn, err := openDBFromConfig(app.Config)
		if err != nil {
			return err
		}
		app.DB = dbConn
		db.SetDefaultDB(dbConn)
		m := migrate.Migrator{DB: dbConn.Conn}
		return m.Up(opts.Migrations)
	case "migrate:rollback":
		app := factory()
		dbConn, err := openDBFromConfig(app.Config)
		if err != nil {
			return err
		}
		app.DB = dbConn
		db.SetDefaultDB(dbConn)
		m := migrate.Migrator{DB: dbConn.Conn}
		return m.Down(opts.Migrations, 1)
	case "seed":
		if opts.Seed == nil {
			return errors.New("no seed function provided")
		}
		app := factory()
		dbConn, err := openDBFromConfig(app.Config)
		if err != nil {
			return err
		}
		app.DB = dbConn
		db.SetDefaultDB(dbConn)
		return opts.Seed(dbConn)
	case "schedule:run":
		return ScheduleRunOnce()
	case "schedule:work":
		return ScheduleRunLoop()
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func openDBFromConfig(cfg *config.Config) (*db.DB, error) {
	if cfg == nil || cfg.DBDriver == "" || cfg.DBDsn == "" {
		return nil, errors.New("DB_DRIVER and DB_DSN required")
	}
	return db.Open(cfg.DBDriver, cfg.DBDsn, db.QuestionDialect{})
}

// Run is a convenience wrapper: it reads os.Args, defaults to serve,
// and runs with migrations and seeders.
func Run(factory AppFactory, migrations []migrate.Migration, seed func(*db.DB) error) error {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"serve"}
	}
	return RunApp(args, factory, AppOptions{
		Migrations: migrations,
		Seed:       seed,
	})
}

func flagValue(args []string, name string) (string, bool) {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1], true
		}
		if strings.HasPrefix(a, name+"=") {
			return strings.TrimPrefix(a, name+"="), true
		}
	}
	return "", false
}
