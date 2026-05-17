package db

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// TypedModel is a type-safe wrapper around Model that returns T instead of
// map[string]any. Use json struct tags to control field mapping.
//
//	type User struct {
//	    ID    int    `json:"id"`
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	var Users = db.NewTypedModel[User](nil, "users")
//
//	user, err := Users.Find(1)   // returns User, not map[string]any
//	list, err := Users.Get()     // returns []User
type TypedModel[T any] struct {
	*Model
}

// NewTypedModel creates a type-safe model for the given table.
// Pass nil for conn to use the package-level default DB.
func NewTypedModel[T any](conn *DB, table string) *TypedModel[T] {
	return &TypedModel[T]{Model: NewModel(conn, table)}
}

// Find retrieves a row by primary key and scans it into T.
func (m *TypedModel[T]) Find(id any) (T, error) {
	row, err := m.Model.Find(id)
	if err != nil {
		var zero T
		return zero, err
	}
	return mapToTyped[T](row)
}

// FindOrFail is like Find but returns an error when the row is missing.
func (m *TypedModel[T]) FindOrFail(id any) (T, error) {
	row, err := m.Model.FindOrFail(id)
	if err != nil {
		var zero T
		return zero, err
	}
	return mapToTyped[T](row)
}

// First returns the first matching row as T.
func (m *TypedModel[T]) First() (T, error) {
	row, err := m.Model.First()
	if err != nil {
		var zero T
		return zero, err
	}
	return mapToTyped[T](row)
}

// Get returns all rows as []T.
func (m *TypedModel[T]) Get() ([]T, error) {
	rows, err := m.Model.Get()
	if err != nil {
		return nil, err
	}
	return mapsToTyped[T](rows)
}

// Query returns a TypedQuery scoped to this model's table with soft-delete filtering applied.
func (m *TypedModel[T]) Query() *TypedQuery[T] {
	return &TypedQuery[T]{q: m.Model.Query()}
}

// Where starts a typed query with a WHERE clause.
func (m *TypedModel[T]) Where(col, op string, val any) *TypedQuery[T] {
	return &TypedQuery[T]{q: m.Model.Query().Where(col, op, val)}
}

// FirstOrCreate finds the first row matching match or creates it with match+defaults.
// Returns the typed row and whether it was just created.
func (m *TypedModel[T]) FirstOrCreate(match, defaults map[string]any) (T, bool, error) {
	row, created, err := m.Model.FirstOrCreate(match, defaults)
	if err != nil {
		var zero T
		return zero, false, err
	}
	typed, err := mapToTyped[T](row)
	return typed, created, err
}

// TypedCreate inserts a struct as a new row, returning the last-insert ID.
// Uses json struct tags for column mapping. Fields tagged omitempty are skipped when zero.
//
//	id, err := Users.TypedCreate(User{Name: "Alice", Email: "a@x.com"})
func (m *TypedModel[T]) TypedCreate(v T) (int64, error) {
	return m.Model.Create(StructToMap(v))
}

// TypedSave updates the row by primary key using a struct value.
// Fields tagged omitempty are skipped when zero (useful for partial updates).
//
//	err := Users.TypedSave(42, User{Name: "Bob"})
func (m *TypedModel[T]) TypedSave(id any, v T) error {
	return m.Model.Save(id, StructToMap(v))
}

// TypedInsert inserts a struct without returning the ID.
func (m *TypedModel[T]) TypedInsert(v T) error {
	return m.Model.Insert(StructToMap(v))
}

// TypedObservedCreate runs observer hooks then inserts the struct.
func (m *TypedModel[T]) TypedObservedCreate(v T) (int64, error) {
	return m.Model.ObservedCreate(StructToMap(v))
}

// TypedObservedSave runs observer hooks then updates the row by primary key.
func (m *TypedModel[T]) TypedObservedSave(id any, v T) error {
	return m.Model.ObservedSave(id, StructToMap(v))
}

// Scope applies a query scope function to this TypedModel's base query.
//
//	active := func(q *db.Query) *db.Query { return q.Where("active", "=", 1) }
//	users, _ := Users.Scope(active).Get()
func (m *TypedModel[T]) Scope(fn func(*Query) *Query) *TypedQuery[T] {
	return &TypedQuery[T]{q: m.Model.Scope(fn)}
}

// Observe registers an observer on the underlying Model.
func (m *TypedModel[T]) Observe(o Observer) {
	m.Model.Observe(o)
}

// UpdateOrCreate finds the first row matching match and updates it, or creates it.
func (m *TypedModel[T]) UpdateOrCreate(match, data map[string]any) (T, error) {
	row, err := m.Model.UpdateOrCreate(match, data)
	if err != nil {
		var zero T
		return zero, err
	}
	return mapToTyped[T](row)
}

// TypedQuery wraps Query and returns T / []T from terminal methods.
type TypedQuery[T any] struct {
	q *Query
}

func (q *TypedQuery[T]) Where(col, op string, val any) *TypedQuery[T] {
	q.q = q.q.Where(col, op, val)
	return q
}

func (q *TypedQuery[T]) WhereIn(col string, vals []any) *TypedQuery[T] {
	q.q = q.q.WhereIn(col, vals)
	return q
}

func (q *TypedQuery[T]) WhereNull(col string) *TypedQuery[T] {
	q.q = q.q.WhereNull(col)
	return q
}

func (q *TypedQuery[T]) WhereNotNull(col string) *TypedQuery[T] {
	q.q = q.q.WhereNotNull(col)
	return q
}

func (q *TypedQuery[T]) WhereBetween(col string, min, max any) *TypedQuery[T] {
	q.q = q.q.WhereBetween(col, min, max)
	return q
}

func (q *TypedQuery[T]) OrderBy(order string) *TypedQuery[T] {
	q.q = q.q.OrderBy(order)
	return q
}

func (q *TypedQuery[T]) Latest(col ...string) *TypedQuery[T] {
	q.q = q.q.Latest(col...)
	return q
}

func (q *TypedQuery[T]) Limit(n int) *TypedQuery[T] {
	q.q = q.q.Limit(n)
	return q
}

func (q *TypedQuery[T]) Offset(n int) *TypedQuery[T] {
	q.q = q.q.Offset(n)
	return q
}

func (q *TypedQuery[T]) Select(cols ...string) *TypedQuery[T] {
	q.q = q.q.Select(cols...)
	return q
}

// First returns the first matching row as T.
func (q *TypedQuery[T]) First() (T, error) {
	row, err := q.q.First()
	if err != nil {
		var zero T
		return zero, err
	}
	return mapToTyped[T](row)
}

// Get returns all matching rows as []T.
func (q *TypedQuery[T]) Get() ([]T, error) {
	rows, err := q.q.Get()
	if err != nil {
		return nil, err
	}
	return mapsToTyped[T](rows)
}

// Count delegates to the underlying Query.
func (q *TypedQuery[T]) Count() (int, error) {
	return q.q.Count()
}

// Paginate returns a Page of typed results.
func (q *TypedQuery[T]) Paginate(page, perPage int) (TypedPage[T], error) {
	total, err := q.q.Count()
	if err != nil {
		return TypedPage[T]{}, err
	}
	rows, err := q.Limit(perPage).Offset((page - 1) * perPage).Get()
	if err != nil {
		return TypedPage[T]{}, err
	}
	return TypedPage[T]{Data: rows, Total: total, Page: page, PerPage: perPage}, nil
}

// TypedPage mirrors db.Page but carries []T instead of []map[string]any.
type TypedPage[T any] struct {
	Data    []T
	Total   int
	Page    int
	PerPage int
}

// ── internal helpers ──────────────────────────────────────────────────────────

// mapToTyped converts a map[string]any to T via JSON round-trip.
// This respects all standard json: struct tags developers already use.
func mapToTyped[T any](m map[string]any) (T, error) {
	var zero T
	if m == nil {
		return zero, errors.New("db: nil row")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return zero, err
	}
	if err := json.Unmarshal(b, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}

func mapsToTyped[T any](rows []map[string]any) ([]T, error) {
	out := make([]T, 0, len(rows))
	for _, row := range rows {
		t, err := mapToTyped[T](row)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// TypedFind is a package-level helper for one-off typed lookups without a TypedModel.
//
//	user, err := db.TypedFind[User](conn, "users", "id", 42)
func TypedFind[T any](conn *DB, table, col string, val any) (T, error) {
	row, err := conn.Table(table).Where(col, "=", val).First()
	if err != nil {
		var zero T
		return zero, err
	}
	return mapToTyped[T](row)
}

// TypedGet is a package-level helper for one-off typed queries without a TypedModel.
//
//	users, err := db.TypedGet[User](conn.Table("users").Where("active","=",1))
func TypedGet[T any](q *Query) ([]T, error) {
	rows, err := q.Get()
	if err != nil {
		return nil, err
	}
	return mapsToTyped[T](rows)
}

// ToTyped converts a single map[string]any to T (useful in resource transformers).
func ToTyped[T any](m map[string]any) (T, error) {
	return mapToTyped[T](m)
}

// MustToTyped is like ToTyped but panics on error.
func MustToTyped[T any](m map[string]any) T {
	t, err := mapToTyped[T](m)
	if err != nil {
		panic(err)
	}
	return t
}

// TypedTransaction runs fn inside a transaction, providing a typed helper.
func TypedTransaction(conn *DB, fn func(tx *sql.Tx) error) error {
	return conn.Transaction(fn)
}
