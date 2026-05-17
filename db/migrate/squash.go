package migrate

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Squash collapses all currently applied migrations into a single baseline migration
// with the given ID (e.g. "0000_squash_baseline"). The baseline's Up function
// is the provided schema dump (raw SQL that recreates the full schema from scratch).
//
// After squashing:
//   - All existing migration records are deleted from the migrations table.
//   - A single record for squashID is inserted with the current batch number.
//   - Future migrations added after the squash point still run normally.
//
// Typical workflow:
//  1. Dump your current schema: mysqldump --no-data mydb > schema.sql
//  2. Call Squash with that SQL as the baseline.
//  3. Commit the new baseline migration file and delete the old ones.
//
//	err := migrator.Squash("0000_squash_baseline", schemaSQL, migrations)
func (m *Migrator) Squash(squashID string, schemaSQL string, all []Migration) error {
	if err := m.EnsureTable(); err != nil {
		return err
	}

	// Determine which migrations are already applied.
	applied, err := m.AppliedIDs()
	if err != nil {
		return err
	}

	// Find the highest batch number.
	var maxBatch int
	row := m.DB.QueryRow(`SELECT COALESCE(MAX(batch), 0) FROM migrations`)
	_ = row.Scan(&maxBatch)

	// Remove all existing migration records.
	if _, err := m.DB.Exec(`DELETE FROM migrations`); err != nil {
		return err
	}

	// Insert the squash baseline record.
	now := time.Now()
	if _, err := m.DB.Exec(
		fmt.Sprintf(`INSERT INTO migrations (id, batch, ran_at) VALUES (%s, %s, %s)`,
			m.ph(1), m.ph(2), m.ph(3)),
		squashID, maxBatch+1, now,
	); err != nil {
		return err
	}

	// Re-insert records for migrations that were applied AFTER the squash point
	// (i.e., migrations not yet squashed). Identify by checking which IDs from
	// all[] were applied but are NOT in the pre-squash set — these need to remain tracked.
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	batch := maxBatch + 1
	for _, mg := range all {
		if mg.ID == squashID {
			continue
		}
		if !applied[mg.ID] {
			continue // not yet applied, will run normally in future
		}
		// Re-record it so rollback still works.
		if _, err := m.DB.Exec(
			fmt.Sprintf(`INSERT INTO migrations (id, batch, ran_at) VALUES (%s, %s, %s)`,
				m.ph(1), m.ph(2), m.ph(3)),
			mg.ID, batch, now,
		); err != nil {
			return err
		}
	}
	return nil
}

// SquashedMigration returns a Migration whose Up is the provided SQL schema dump.
// Use this to register the baseline after squashing.
//
//	gophant.RegisterMigrations(migrate.SquashedMigration("0000_squash_baseline", schemaSQL))
func SquashedMigration(id, schemaSQL string) Migration {
	return Migration{
		ID: id,
		Up: func(db *sql.DB) error {
			_, err := db.Exec(schemaSQL)
			return err
		},
		Down: func(db *sql.DB) error {
			return fmt.Errorf("squashed baseline %q cannot be rolled back — restore from a DB backup", id)
		},
	}
}
