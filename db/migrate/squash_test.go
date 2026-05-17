package migrate

import (
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func newMigratorMock(t *testing.T) (*Migrator, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &Migrator{DB: db}, mock
}

func TestSquashedMigration_Up(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectExec("CREATE TABLE users").WillReturnResult(sqlmock.NewResult(0, 0))

	m := SquashedMigration("0000_baseline", "CREATE TABLE users (id INT)")
	if m.ID != "0000_baseline" {
		t.Fatalf("unexpected ID: %s", m.ID)
	}
	if err := m.Up(db); err != nil {
		t.Fatalf("Up: %v", err)
	}
}

func TestSquashedMigration_DownErrors(t *testing.T) {
	m := SquashedMigration("0000_baseline", "CREATE TABLE users (id INT)")
	err := m.Down(&sql.DB{})
	if err == nil {
		t.Fatal("expected error from squashed Down")
	}
}

func TestSquash_ClearsAndReinserts(t *testing.T) {
	mg, mock := newMigratorMock(t)

	// EnsureTable
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnResult(sqlmock.NewResult(0, 0))
	// AppliedIDs
	mock.ExpectQuery("SELECT id FROM migrations").WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow("0001_create_users").AddRow("0002_add_email"),
	)
	// MAX(batch)
	mock.ExpectQuery("SELECT COALESCE").WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(2))
	// DELETE all
	mock.ExpectExec("DELETE FROM migrations").WillReturnResult(sqlmock.NewResult(0, 2))
	// INSERT baseline
	mock.ExpectExec("INSERT INTO migrations").WithArgs("0000_baseline", 3, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Re-insert applied migrations (0001, 0002 both applied)
	mock.ExpectExec("INSERT INTO migrations").WithArgs("0001_create_users", 3, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO migrations").WithArgs("0002_add_email", 3, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	migrations := []Migration{
		{ID: "0001_create_users", Up: nil, Down: nil},
		{ID: "0002_add_email", Up: nil, Down: nil},
	}
	if err := mg.Squash("0000_baseline", "CREATE TABLE users (id INT)", migrations); err != nil {
		t.Fatalf("Squash: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}
