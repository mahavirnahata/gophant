// Package seed provides database seeding for development and testing.
//
// Register seeders in init() functions — they run in registration order:
//
//	func init() {
//	    seed.Register(&UserSeeder{})
//	}
//
//	type UserSeeder struct{}
//	func (s *UserSeeder) Run(db *sql.DB) error {
//	    _, err := db.Exec(`INSERT INTO users (name, email) VALUES ('Admin', 'admin@example.com')`)
//	    return err
//	}
package seed

import (
	"database/sql"
	"fmt"
	"log"
)

// Seeder populates a database with test or default data.
type Seeder interface {
	Run(db *sql.DB) error
}

var registered []namedSeeder

type namedSeeder struct {
	name    string
	seeder  Seeder
}

// Register adds a seeder to the global list.
func Register(s Seeder) {
	registered = append(registered, namedSeeder{
		name:   fmt.Sprintf("%T", s),
		seeder: s,
	})
}

// RunAll executes every registered seeder in registration order.
func RunAll(db *sql.DB) error {
	for _, ns := range registered {
		log.Printf("seeding: %s", ns.name)
		if err := ns.seeder.Run(db); err != nil {
			return fmt.Errorf("seeder %s: %w", ns.name, err)
		}
	}
	return nil
}

// Run executes a single seeder without registering it.
func Run(db *sql.DB, s Seeder) error {
	return s.Run(db)
}

// Seeders returns the names of all registered seeders.
func Seeders() []string {
	names := make([]string, len(registered))
	for i, ns := range registered {
		names[i] = ns.name
	}
	return names
}

// Reset clears the registered seeder list (useful in tests).
func Reset() { registered = nil }
