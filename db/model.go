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
	DB           *DB
	Table        string
	PrimaryKey   string // defaults to "id"
	Timestamps   bool   // auto-set created_at / updated_at
	SoftDelete   bool   // when true, Destroy sets deleted_at instead of deleting
	DeletedAtCol string // soft-delete column (default: "deleted_at")
	observers    []Observer
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

func (m *Model) deletedAtCol() string {
	if m.DeletedAtCol != "" {
		return m.DeletedAtCol
	}
	return "deleted_at"
}

// Query returns a fresh query builder for this table.
// If SoftDelete is true, rows where deleted_at IS NOT NULL are excluded automatically.
func (m *Model) Query() *Query {
	q := m.conn().Table(m.Table)
	if m.SoftDelete {
		q = q.WhereNull(m.deletedAtCol())
	}
	return q
}

// WithTrashed returns a query that includes soft-deleted rows.
func (m *Model) WithTrashed() *Query {
	return m.conn().Table(m.Table)
}

// OnlyTrashed returns a query scoped to soft-deleted rows only.
func (m *Model) OnlyTrashed() *Query {
	return m.conn().Table(m.Table).WhereNotNull(m.deletedAtCol())
}

// Restore un-deletes a soft-deleted row by primary key.
func (m *Model) Restore(id any) error {
	if !m.SoftDelete {
		return nil
	}
	pk := m.PrimaryKey
	if pk == "" {
		pk = "id"
	}
	_, err := m.conn().Table(m.Table).Where(pk, "=", id).Update(map[string]any{
		m.deletedAtCol(): nil,
	})
	return err
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
// When SoftDelete is true, it sets deleted_at = now() instead of hard-deleting.
func (m *Model) Destroy(id any) error {
	pk := m.PrimaryKey
	if pk == "" {
		pk = "id"
	}
	if m.SoftDelete {
		return m.Save(id, map[string]any{m.deletedAtCol(): time.Now()})
	}
	return m.Delete(pk, id)
}

// FirstOrCreate finds the first row matching match, or creates it with the
// union of match and defaults. Returns the row, whether it was created, and any error.
func (m *Model) FirstOrCreate(match map[string]any, defaults map[string]any) (map[string]any, bool, error) {
	q := m.Query()
	for col, val := range match {
		q = q.Where(col, "=", val)
	}
	row, err := q.First()
	if err == nil {
		return row, false, nil
	}

	data := make(map[string]any, len(match)+len(defaults))
	for k, v := range match {
		data[k] = v
	}
	for k, v := range defaults {
		data[k] = v
	}
	id, err := m.Create(data)
	if err != nil {
		return nil, false, err
	}
	pk := m.PrimaryKey
	if pk == "" {
		pk = "id"
	}
	created, err := m.Find(id)
	return created, true, err
}

// UpdateOrCreate finds the first row matching match and updates it with data,
// or creates a new row with the union of match and data. Returns the final row.
func (m *Model) UpdateOrCreate(match map[string]any, data map[string]any) (map[string]any, error) {
	q := m.Query()
	for col, val := range match {
		q = q.Where(col, "=", val)
	}
	row, err := q.First()
	if err == nil {
		pk := m.PrimaryKey
		if pk == "" {
			pk = "id"
		}
		if err := m.Save(row[pk], data); err != nil {
			return nil, err
		}
		return m.Find(row[pk])
	}

	merged := make(map[string]any, len(match)+len(data))
	for k, v := range match {
		merged[k] = v
	}
	for k, v := range data {
		merged[k] = v
	}
	id, err := m.Create(merged)
	if err != nil {
		return nil, err
	}
	return m.Find(id)
}

// Scope applies a named query scope (a function that modifies a Query).
// Use it to encapsulate common query conditions:
//
//	activeUsers := func(q *db.Query) *db.Query { return q.Where("active", "=", 1) }
//	rows, _ := model.Scope(activeUsers).Get()
func (m *Model) Scope(scope func(*Query) *Query) *Query {
	return scope(m.Query())
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
