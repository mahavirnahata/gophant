package migrate

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestStatus(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT id FROM migrations").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	m := Migrator{DB: sqlDB}
	applied, pending, err := m.Status([]Migration{{ID: "001"}})
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if len(applied) != 0 || len(pending) != 1 {
		t.Fatalf("unexpected status")
	}
}
