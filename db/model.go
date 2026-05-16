package db

import (
	"errors"
	"time"
)

// Model is a thin table-gateway helper. Embed it in your own struct or use
// db.NewModel(conn, "table") directly.
//
//	func UserModel() *db.Model { return db.NewModel(nil, "users") }
type Model struct {
	DB         *DB
	Table      string
	PrimaryKey string // defaults to "id"
	Timestamps bool   // auto-set created_at / updated_at
}

func NewModel(db *DB, table string) *Model {
	return &Model{DB: db, Table: table, PrimaryKey: "id"}
}

func (m *Model) conn() *DB {
	if m.DB != nil {
		return m.DB
	}
	if DefaultDB != nil {
		return DefaultDB
	}
	panic(errors.New("db not configured: call db.SetDefaultDB or use NewModel(db, table)"))
}

// Query returns a fresh query builder for this table.
func (m *Model) Query() *Query {
	return m.conn().Table(m.Table)
}

// ── Finders ──────────────────────────────────────────────────────────────────

// Find retrieves a row by primary key. Returns nil, sql.ErrNoRows when not found.
func (m *Model) Find(id any) (map[string]any, error) {
	pk := m.PrimaryKey
	if pk == "" {
		pk = "id"
	}
	return m.Query().Where(pk, "=", id).First()
}

// FindOrFail is like Find but returns an error with a descriptive message on miss.
func (m *Model) FindOrFail(id any) (map[string]any, error) {
	row, err := m.Find(id)
	if err != nil {
		return nil, errors.New(m.Table + " not found")
	}
	return row, nil
}

// Where starts a query with a WHERE clause.
func (m *Model) Where(col, op string, val any) *Query {
	return m.Query().Where(col, op, val)
}

// OrderBySafe starts a query with a safe ORDER BY clause.
func (m *Model) OrderBySafe(column, direction string, allowed []string) *Query {
	return m.Query().OrderBySafe(column, direction, allowed)
}

// Get returns all rows in the table.
func (m *Model) Get() ([]map[string]any, error) {
	return m.Query().Get()
}

// GetStructs scans all rows into a slice of structs.
func (m *Model) GetStructs(dest any) error {
	return m.Query().GetStructs(dest)
}

// First returns the first row (LIMIT 1).
func (m *Model) First() (map[string]any, error) {
	return m.Query().First()
}

// FirstStruct scans the first row into a struct.
func (m *Model) FirstStruct(dest any) error {
	return m.Query().FirstStruct(dest)
}

// ── Mutations ────────────────────────────────────────────────────────────────

// Create inserts a row and returns the last-insert ID.
// If Timestamps is true, created_at and updated_at are set automatically.
func (m *Model) Create(data map[string]any) (int64, error) {
	if m.Timestamps {
		now := time.Now()
		if _, ok := data["created_at"]; !ok {
			data["created_at"] = now
		}
		if _, ok := data["updated_at"]; !ok {
			data["updated_at"] = now
		}
	}
	res, err := m.Query().Insert(data)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Insert inserts a row (no return value). Prefer Create when you need the ID.
func (m *Model) Insert(data map[string]any) error {
	if m.Timestamps {
		now := time.Now()
		if _, ok := data["created_at"]; !ok {
			data["created_at"] = now
		}
		if _, ok := data["updated_at"]; !ok {
			data["updated_at"] = now
		}
	}
	_, err := m.Query().Insert(data)
	return err
}

// Update updates rows matching whereCol = whereVal.
// If Timestamps is true, updated_at is set automatically.
func (m *Model) Update(whereCol string, whereVal any, data map[string]any) error {
	if m.Timestamps {
		if _, ok := data["updated_at"]; !ok {
			data["updated_at"] = time.Now()
		}
	}
	_, err := m.Query().Where(whereCol, "=", whereVal).Update(data)
	return err
}

// Save is a convenience for updating by primary key.
func (m *Model) Save(id any, data map[string]any) error {
	pk := m.PrimaryKey
	if pk == "" {
		pk = "id"
	}
	return m.Update(pk, id, data)
}

// Delete removes rows matching whereCol = whereVal.
func (m *Model) Delete(whereCol string, whereVal any) error {
	_, err := m.Query().Where(whereCol, "=", whereVal).Delete()
	return err
}

// Destroy removes a row by primary key.
func (m *Model) Destroy(id any) error {
	pk := m.PrimaryKey
	if pk == "" {
		pk = "id"
	}
	return m.Delete(pk, id)
}

// Count returns the number of rows in the table.
func (m *Model) Count() (int, error) {
	return m.Query().Count()
}

// Exists reports whether any row matches the given column/value.
func (m *Model) Exists(col string, val any) (bool, error) {
	count, err := m.Query().Where(col, "=", val).Count()
	return count > 0, err
}
