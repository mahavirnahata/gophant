package gophant

import "github.com/mahavirnahata/gophant/db/migrate"

var migrationRegistry []migrate.Migration

// RegisterMigrations lets applications register migrations during init().
func RegisterMigrations(ms ...migrate.Migration) {
	migrationRegistry = append(migrationRegistry, ms...)
}

// Migrations returns all registered migrations.
func Migrations() []migrate.Migration {
	return migrationRegistry
}

// SetMigrations overwrites the registry (useful for tests).
func SetMigrations(ms []migrate.Migration) {
	migrationRegistry = ms
}
