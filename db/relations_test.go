package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHasMany(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"user_id", "title"}).AddRow(1, "a").AddRow(1, "b")
	mock.ExpectQuery(`SELECT .* FROM posts WHERE user_id IN \(\?,\?\)`).WillReturnRows(rows)

	conn := &DB{Conn: sqlDB, Dialect: QuestionDialect{}}
	out, err := HasMany(conn, "posts", "user_id", []any{1, 2})
	if err != nil {
		t.Fatalf("hasmany: %v", err)
	}
	if len(out[1]) != 2 {
		t.Fatalf("expected 2 children")
	}
}
