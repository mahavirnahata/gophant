package db

import "testing"

func TestOrderBySafe(t *testing.T) {
	db := &DB{Dialect: QuestionDialect{}}
	q := &Query{db: db, table: "users", selectCols: []string{"*"}}
	q.OrderBySafe("name", "DESC", []string{"name"})
	if q.orderBy != "name DESC" {
		t.Fatalf("expected safe orderby")
	}
	q.OrderBySafe("evil", "DESC", []string{"name"})
	if q.orderBy != "name DESC" {
		t.Fatalf("expected orderby unchanged")
	}
}

func TestBuildSelect(t *testing.T) {
	db := &DB{Dialect: QuestionDialect{}}
	q := &Query{db: db, table: "users", selectCols: []string{"id", "email"}}
	q.Where("id", ">", 1).OrderBy("id DESC").Limit(10).Offset(5)

	sql, args := q.buildSelect()
	if sql == "" || len(args) != 1 {
		t.Fatalf("unexpected sql or args")
	}
}
