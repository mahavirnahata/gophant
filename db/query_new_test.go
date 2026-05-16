package db

import (
	"fmt"
	"testing"
)

func TestWhereNull(t *testing.T) {
	d := &DB{Dialect: QuestionDialect{}}
	q := &Query{db: d, table: "users", selectCols: []string{"*"}}
	q.WhereNull("deleted_at")

	sql, args := q.buildSelect()
	if !contains(sql, "IS NULL") {
		t.Fatalf("expected IS NULL in query, got: %s", sql)
	}
	if len(args) != 0 {
		t.Fatalf("IS NULL should have no args, got %d", len(args))
	}
}

func TestWhereNotNull(t *testing.T) {
	d := &DB{Dialect: QuestionDialect{}}
	q := &Query{db: d, table: "users", selectCols: []string{"*"}}
	q.WhereNotNull("deleted_at")

	sql, _ := q.buildSelect()
	if !contains(sql, "IS NOT NULL") {
		t.Fatalf("expected IS NOT NULL in query, got: %s", sql)
	}
}

func TestWhereBetween(t *testing.T) {
	d := &DB{Dialect: QuestionDialect{}}
	q := &Query{db: d, table: "orders", selectCols: []string{"*"}}
	q.WhereBetween("price", 10, 100)

	sql, args := q.buildSelect()
	if !contains(sql, "BETWEEN") {
		t.Fatalf("expected BETWEEN in query, got: %s", sql)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != 10 || args[1] != 100 {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestWhereBetweenPostgres(t *testing.T) {
	d := &DB{Dialect: DollarDialect{}}
	q := &Query{db: d, table: "orders", selectCols: []string{"*"}}
	q.WhereBetween("price", 10, 100)

	sql, args := q.buildSelect()
	if !contains(sql, "$1") || !contains(sql, "$2") {
		t.Fatalf("expected $1 and $2 placeholders, got: %s", sql)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
}

func TestLatestOldest(t *testing.T) {
	d := &DB{Dialect: QuestionDialect{}}

	q := &Query{db: d, table: "posts", selectCols: []string{"*"}}
	q.Latest()
	sql, _ := q.buildSelect()
	if !contains(sql, "created_at DESC") {
		t.Fatalf("expected created_at DESC, got: %s", sql)
	}

	q2 := &Query{db: d, table: "posts", selectCols: []string{"*"}}
	q2.Oldest("updated_at")
	sql2, _ := q2.buildSelect()
	if !contains(sql2, "updated_at ASC") {
		t.Fatalf("expected updated_at ASC, got: %s", sql2)
	}
}

func TestWhereNullCombinedWithOtherConditions(t *testing.T) {
	d := &DB{Dialect: QuestionDialect{}}
	q := &Query{db: d, table: "users", selectCols: []string{"*"}}
	q.Where("active", "=", 1).WhereNull("deleted_at")

	sql, args := q.buildSelect()
	if !contains(sql, "IS NULL") {
		t.Fatalf("expected IS NULL, got: %s", sql)
	}
	if !contains(sql, "WHERE") {
		t.Fatalf("expected WHERE clause, got: %s", sql)
	}
	// Should have 1 arg for the active=1 condition; deleted_at IS NULL adds none
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d: %v", len(args), args)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func TestBuildSelectWithBetweenAndNull(t *testing.T) {
	d := &DB{Dialect: QuestionDialect{}}
	q := &Query{db: d, table: "products", selectCols: []string{"id", "name"}}
	q.WhereBetween("price", 5, 50).WhereNull("archived_at").OrderBy("price ASC").Limit(10)

	sql, args := q.buildSelect()
	expected := fmt.Sprintf("SELECT id,name FROM products WHERE price BETWEEN ? AND ? AND archived_at IS NULL ORDER BY price ASC LIMIT 10")
	if sql != expected {
		t.Fatalf("unexpected sql:\n got:  %s\n want: %s", sql, expected)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
}
