package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type userRow struct {
	ID    int64  `db:"id"`
	Email string `db:"email"`
}

func TestScanRowsIntoStructs(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "a@b.com")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	db := &DB{Conn: sqlDB, Dialect: QuestionDialect{}}
	q := db.Table("users")

	var out []userRow
	if err := q.GetStructs(&out); err != nil {
		t.Fatalf("get structs: %v", err)
	}
	if len(out) != 1 || out[0].Email != "a@b.com" {
		t.Fatalf("unexpected output")
	}
}
