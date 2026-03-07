package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mahavirnahata/gophant"
	"github.com/mahavirnahata/gophant/db/migrate"
)

type StatusOptions struct {
	JSON bool
}

func MigrateStatus(db *sql.DB, opts StatusOptions) error {
	m := migrate.Migrator{DB: db}
	applied, pending, err := m.Status(gophant.Migrations())
	if err != nil {
		return err
	}

	if opts.JSON {
		out := map[string]any{
			"applied": applied,
			"pending": pending,
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	fmt.Println("Applied:")
	for _, id := range applied {
		fmt.Println(" -", id)
	}

	fmt.Println("Pending:")
	for _, id := range pending {
		fmt.Println(" -", id)
	}

	return nil
}
