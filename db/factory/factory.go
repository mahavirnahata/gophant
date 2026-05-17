// Package factory provides a Laravel-style model factory for generating test data.
//
// Define a factory for each model, then use it in tests:
//
//	var UserFactory = factory.New(func(f *factory.Context) map[string]any {
//	    return map[string]any{
//	        "name":  f.Seq("User %d"),
//	        "email": f.Seq("user%d@example.com"),
//	        "role":  "member",
//	    }
//	})
//
//	// In tests:
//	row  := UserFactory.Make()                     // map, no DB
//	id   := UserFactory.Create(db, "users")        // inserts, returns ID
//	rows := UserFactory.MakeMany(5)                // []map[string]any
//	ids  := UserFactory.CreateMany(5, db, "users") // inserts 5, returns IDs
package factory

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
)

// Context is passed to the definition function to help generate unique values.
type Context struct {
	seq int64 // current sequence number
	rng *rand.Rand
}

// Seq formats a string with the current sequence number.
// Use it to generate unique fields: f.Seq("user%d@example.com")
func (c *Context) Seq(format string) string {
	return fmt.Sprintf(format, c.seq)
}

// Int returns a random int in [min, max].
func (c *Context) Int(min, max int) int {
	return min + c.rng.Intn(max-min+1)
}

// Pick returns a random element from the given values.
func (c *Context) Pick(values ...string) string {
	return values[c.rng.Intn(len(values))]
}

// Word returns a random word from a small built-in list (useful for names, titles).
func (c *Context) Word() string {
	words := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel"}
	return words[c.rng.Intn(len(words))]
}

// Email returns a unique fake email address.
func (c *Context) Email() string {
	return fmt.Sprintf("user%d@example.com", c.seq)
}

// Name returns a unique fake full name.
func (c *Context) Name() string {
	first := []string{"Alice", "Bob", "Carol", "Dan", "Eve", "Frank", "Grace", "Hank"}
	last := []string{"Smith", "Jones", "Lee", "Kim", "Chen", "Brown", "Davis", "Wilson"}
	return first[c.rng.Intn(len(first))] + " " + last[c.rng.Intn(len(last))]
}

// Sentence returns a fake sentence with n words.
func (c *Context) Sentence(n int) string {
	words := make([]string, n)
	for i := range words {
		words[i] = c.Word()
	}
	s := strings.Join(words, " ")
	return strings.ToUpper(s[:1]) + s[1:] + "."
}

// Definition is the function that produces a base attribute map for a factory.
type Definition func(c *Context) map[string]any

// Factory generates model attribute maps and optionally persists them.
type Factory struct {
	definition Definition
	overrides  map[string]any
	seq        atomic.Int64
}

// New creates a Factory from a definition function.
func New(def Definition) *Factory {
	return &Factory{definition: def}
}

// With returns a new Factory with the given attribute overrides applied on top
// of the definition. The original Factory is not mutated.
//
//	admin := UserFactory.With(map[string]any{"role": "admin"})
//	row   := admin.Make()
func (f *Factory) With(overrides map[string]any) *Factory {
	merged := make(map[string]any, len(f.overrides)+len(overrides))
	for k, v := range f.overrides {
		merged[k] = v
	}
	for k, v := range overrides {
		merged[k] = v
	}
	return &Factory{definition: f.definition, overrides: merged}
}

// Make generates a single attribute map without persisting it.
func (f *Factory) Make() map[string]any {
	n := f.seq.Add(1)
	ctx := &Context{seq: n, rng: rand.New(rand.NewSource(n))}
	attrs := f.definition(ctx)
	for k, v := range f.overrides {
		attrs[k] = v
	}
	return attrs
}

// MakeMany generates n attribute maps without persisting them.
func (f *Factory) MakeMany(n int) []map[string]any {
	out := make([]map[string]any, n)
	for i := range out {
		out[i] = f.Make()
	}
	return out
}

// Create inserts a single row into table using db and returns the last-insert ID.
func (f *Factory) Create(db *sql.DB, table string) (int64, error) {
	attrs := f.Make()
	return insertRow(db, table, attrs)
}

// CreateMany inserts n rows and returns their IDs.
func (f *Factory) CreateMany(n int, db *sql.DB, table string) ([]int64, error) {
	ids := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		id, err := f.Create(db, table)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// CreateAndReturn inserts a row and reads it back from the DB.
func (f *Factory) CreateAndReturn(db *sql.DB, table, pkCol string) (map[string]any, error) {
	id, err := f.Create(db, table)
	if err != nil {
		return nil, err
	}
	row := db.QueryRow(fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table, pkCol), id)
	cols, err := columnNames(db, table)
	if err != nil {
		return nil, err
	}
	return scanOne(row, cols)
}

// Reset resets the sequence counter (useful between test runs).
func (f *Factory) Reset() { f.seq.Store(0) }

// ── helpers ───────────────────────────────────────────────────────────────────

func insertRow(db *sql.DB, table string, attrs map[string]any) (int64, error) {
	cols := make([]string, 0, len(attrs))
	vals := make([]any, 0, len(attrs))
	placeholders := make([]string, 0, len(attrs))
	for k, v := range attrs {
		cols = append(cols, k)
		vals = append(vals, v)
		placeholders = append(placeholders, "?")
	}
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(cols, ","), strings.Join(placeholders, ","))
	res, err := db.Exec(q, vals...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func columnNames(db *sql.DB, table string) ([]string, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 0", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return rows.Columns()
}

func scanOne(row *sql.Row, cols []string) (map[string]any, error) {
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := row.Scan(ptrs...); err != nil {
		return nil, err
	}
	out := make(map[string]any, len(cols))
	for i, col := range cols {
		out[col] = vals[i]
	}
	return out, nil
}
