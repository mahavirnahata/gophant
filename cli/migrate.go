package cli

import (
	"database/sql"

	"github.com/mahavirnahata/gophant"
	"github.com/mahavirnahata/gophant/db/migrate"
)

type MigrateOptions struct {
	Steps int
}

func MigrateUp(db *sql.DB) error {
	m := migrate.Migrator{DB: db}
	return m.Up(gophant.Migrations())
}

func MigrateDown(db *sql.DB, opts MigrateOptions) error {
	m := migrate.Migrator{DB: db}
	return m.Down(gophant.Migrations(), opts.Steps)
}
