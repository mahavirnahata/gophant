package db

import (
	"testing"
)

func makeTxQuery() *TxQuery {
	d := &DB{Dialect: QuestionDialect{}}
	return &TxQuery{db: d, table: "users", selectCols: []string{"*"}}
}

func TestTxQuery_WhereNull(t *testing.T) {
	q := makeTxQuery()
	q.WhereNull("deleted_at")
	sql, args := q.buildSelect()
	if !contains(sql, "IS NULL") {
		t.Fatalf("expected IS NULL, got: %s", sql)
	}
	if len(args) != 0 {
		t.Fatalf("IS NULL should have no args, got %d", len(args))
	}
}

func TestTxQuery_WhereNotNull(t *testing.T) {
	q := makeTxQuery()
	q.WhereNotNull("deleted_at")
	sql, _ := q.buildSelect()
	if !contains(sql, "IS NOT NULL") {
		t.Fatalf("expected IS NOT NULL, got: %s", sql)
	}
}

func TestTxQuery_WhereBetween(t *testing.T) {
	q := makeTxQuery()
	q.WhereBetween("age", 18, 65)
	sql, args := q.buildSelect()
	if !contains(sql, "BETWEEN") {
		t.Fatalf("expected BETWEEN, got: %s", sql)
	}
	if len(args) != 2 {
		t.Fatalf("BETWEEN should have 2 args, got %d", len(args))
	}
}

func TestTxQuery_WhereIn(t *testing.T) {
	q := makeTxQuery()
	q.WhereIn("id", []any{1, 2, 3})
	sql, args := q.buildSelect()
	if !contains(sql, "IN") {
		t.Fatalf("expected IN, got: %s", sql)
	}
	if len(args) != 3 {
		t.Fatalf("IN should have 3 args, got %d", len(args))
	}
}

func TestTxQuery_Latest(t *testing.T) {
	q := makeTxQuery()
	q.Latest()
	sql, _ := q.buildSelect()
	if !contains(sql, "created_at DESC") {
		t.Fatalf("expected created_at DESC, got: %s", sql)
	}
}

func TestTxQuery_LatestCustomCol(t *testing.T) {
	q := makeTxQuery()
	q.Latest("published_at")
	sql, _ := q.buildSelect()
	if !contains(sql, "published_at DESC") {
		t.Fatalf("expected published_at DESC, got: %s", sql)
	}
}

func TestTxQuery_Oldest(t *testing.T) {
	q := makeTxQuery()
	q.Oldest()
	sql, _ := q.buildSelect()
	if !contains(sql, "created_at ASC") {
		t.Fatalf("expected created_at ASC, got: %s", sql)
	}
}

func TestTxQuery_OrderBy(t *testing.T) {
	q := makeTxQuery()
	q.OrderBy("name ASC")
	sql, _ := q.buildSelect()
	if !contains(sql, "name ASC") {
		t.Fatalf("expected name ASC in: %s", sql)
	}
}

func TestTxQuery_LimitOffset(t *testing.T) {
	q := makeTxQuery()
	q.Limit(10).Offset(20)
	sql, _ := q.buildSelect()
	if !contains(sql, "LIMIT 10") {
		t.Fatalf("expected LIMIT 10 in: %s", sql)
	}
	if !contains(sql, "OFFSET 20") {
		t.Fatalf("expected OFFSET 20 in: %s", sql)
	}
}

func TestTxQuery_WhereNull_AndWhere_Combined(t *testing.T) {
	q := makeTxQuery()
	q.WhereNull("deleted_at").Where("active", "=", 1)
	sql, args := q.buildSelect()
	if !contains(sql, "IS NULL") {
		t.Fatalf("expected IS NULL in: %s", sql)
	}
	if !contains(sql, "AND") {
		t.Fatalf("expected AND in: %s", sql)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
}
