package cli

import (
	"database/sql"

	"github.com/mahavirnahata/gophant"
	"github.com/mahavirnahata/gophant/db/migrate"
)

func MigrateFresh(db *sql.DB) error {
	m := migrate.Migrator{DB: db}
	return m.Fresh(gophant.Migrations())
}
