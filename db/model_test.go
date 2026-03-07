package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type userModel struct {
	ID int64 `db:"id"`
}

func TestModelUsesDefaultDB(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	SetDefaultDB(&DB{Conn: sqlDB, Dialect: QuestionDialect{}})
	m := NewModel(nil, "users")

	mock.ExpectQuery("SELECT .* FROM users").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	var out []userModel
	if err := m.GetStructs(&out); err != nil {
		t.Fatalf("get structs: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 row")
	}
}
