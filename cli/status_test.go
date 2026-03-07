package cli

import (
	"io"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mahavirnahata/gophant"
	"github.com/mahavirnahata/gophant/db/migrate"
)

func TestMigrateStatusJSON(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	gophant.SetMigrations([]migrate.Migration{{ID: "001"}})

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT id FROM migrations").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("001"))

	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := MigrateStatus(sqlDB, StatusOptions{JSON: true}); err != nil {
		t.Fatalf("status: %v", err)
	}
	_ = w.Close()
	out, _ := io.ReadAll(r)
	if len(out) == 0 {
		t.Fatalf("expected output")
	}
}
