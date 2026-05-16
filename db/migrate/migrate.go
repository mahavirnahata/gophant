package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"
)

type Migration struct {
	ID   string
	Up   func(*sql.DB) error
	Down func(*sql.DB) error
}

// Dialect lets the Migrator generate correct placeholders for the target database.
// Use QuestionDialect for MySQL/SQLite and DollarDialect for PostgreSQL.
type Dialect interface {
	Placeholder(n int) string
}

type QuestionDialect struct{}

func (d QuestionDialect) Placeholder(_ int) string { return "?" }

type DollarDialect struct{}

func (d DollarDialect) Placeholder(n int) string { return "$" + strconv.Itoa(n) }

type Migrator struct {
	DB      *sql.DB
	Dialect Dialect
}

func (m *Migrator) ph(n int) string {
	if m.Dialect == nil {
		return "?"
	}
	return m.Dialect.Placeholder(n)
}


func (m *Migrator) EnsureTable() error {
	_, err := m.DB.Exec(`CREATE TABLE IF NOT EXISTS migrations (id VARCHAR(255) PRIMARY KEY, batch INT NOT NULL, ran_at TIMESTAMP NOT NULL)`)
	return err
}

func (m *Migrator) AppliedIDs() (map[string]bool, error) {
	rows, err := m.DB.Query(`SELECT id FROM migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func (m *Migrator) AppliedList() ([]string, error) {
	rows, err := m.DB.Query(`SELECT id FROM migrations ORDER BY batch DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (m *Migrator) NextBatch() (int, error) {
	row := m.DB.QueryRow(`SELECT COALESCE(MAX(batch), 0) + 1 FROM migrations`)
	var batch int
	return batch, row.Scan(&batch)
}

func (m *Migrator) Up(migrations []Migration) error {
	if err := m.EnsureTable(); err != nil {
		return err
	}
	applied, err := m.AppliedIDs()
	if err != nil {
		return err
	}
	batch, err := m.NextBatch()
	if err != nil {
		return err
	}

	toRun := pending(migrations, applied)
	for _, mig := range toRun {
		if mig.Up == nil {
			return errors.New("missing Up for migration " + mig.ID)
		}
		if err := mig.Up(m.DB); err != nil {
			return err
		}
		_, err = m.DB.Exec(
				fmt.Sprintf(`INSERT INTO migrations (id, batch, ran_at) VALUES (%s, %s, %s)`, m.ph(1), m.ph(2), m.ph(3)),
				mig.ID, batch, time.Now(),
			)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) Down(migrations []Migration, steps int) error {
	if steps < 1 {
		steps = 1
	}
	rows, err := m.DB.Query(`SELECT id, batch FROM migrations ORDER BY batch DESC, id DESC`)
	if err != nil {
		return err
	}
	defer rows.Close()

	applied := []string{}
	for rows.Next() {
		var id string
		var batch int
		if err := rows.Scan(&id, &batch); err != nil {
			return err
		}
		applied = append(applied, id)
	}

	migMap := map[string]Migration{}
	for _, mig := range migrations {
		migMap[mig.ID] = mig
	}

	count := 0
	for _, id := range applied {
		mig, ok := migMap[id]
		if !ok || mig.Down == nil {
			continue
		}
		if err := mig.Down(m.DB); err != nil {
			return err
		}
		_, err := m.DB.Exec(fmt.Sprintf(`DELETE FROM migrations WHERE id = %s`, m.ph(1)), id)
		if err != nil {
			return err
		}
		count++
		if count >= steps {
			break
		}
	}
	return nil
}

func (m *Migrator) Fresh(migrations []Migration) error {
	if err := m.EnsureTable(); err != nil {
		return err
	}
	applied, err := m.AppliedList()
	if err != nil {
		return err
	}
	if len(applied) > 0 {
		if err := m.Down(migrations, len(applied)); err != nil {
			return err
		}
	}
	return m.Up(migrations)
}

func (m *Migrator) Status(migrations []Migration) ([]string, []string, error) {
	if err := m.EnsureTable(); err != nil {
		return nil, nil, err
	}
	appliedMap, err := m.AppliedIDs()
	if err != nil {
		return nil, nil, err
	}
	applied := []string{}
	pending := []string{}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].ID < migrations[j].ID })
	for _, mig := range migrations {
		if appliedMap[mig.ID] {
			applied = append(applied, mig.ID)
		} else {
			pending = append(pending, mig.ID)
		}
	}
	return applied, pending, nil
}

func pending(all []Migration, applied map[string]bool) []Migration {
	out := []Migration{}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	for _, mig := range all {
		if !applied[mig.ID] {
			out = append(out, mig)
		}
	}
	return out
}
