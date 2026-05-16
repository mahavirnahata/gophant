package cli

import (
	"database/sql"
	"fmt"

	"github.com/mahavirnahata/gophant/db/seed"
)

// SeedRun executes all registered seeders against the provided database.
func SeedRun(db *sql.DB) error {
	names := seed.Seeders()
	if len(names) == 0 {
		fmt.Println("No seeders registered.")
		return nil
	}
	fmt.Printf("Running %d seeder(s)...\n", len(names))
	return seed.RunAll(db)
}
