package db

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// ── Soft Delete ───────────────────────────────────────────────────────────────

func TestSoftDeleteQueryExcludesDeleted(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	conn := &DB{Conn: sqlDB, Dialect: QuestionDialect{}}
	m := &Model{DB: conn, Table: "posts", SoftDelete: true}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM posts WHERE deleted_at IS NULL")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	rows, err := m.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}

func TestWithTrashedIncludesDeleted(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	conn := &DB{Conn: sqlDB, Dialect: QuestionDialect{}}
	m := &Model{DB: conn, Table: "posts", SoftDelete: true}

	// WithTrashed returns a plain query without soft-delete filter.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM posts")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

	rows, err := m.WithTrashed().Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestOnlyTrashedFilter(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	conn := &DB{Conn: sqlDB, Dialect: QuestionDialect{}}
	m := &Model{DB: conn, Table: "posts", SoftDelete: true}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM posts WHERE deleted_at IS NOT NULL")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	rows, err := m.OnlyTrashed().Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 trashed row, got %d", len(rows))
	}
}

func TestDestroyWithSoftDelete(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	conn := &DB{Conn: sqlDB, Dialect: QuestionDialect{}}
	m := &Model{DB: conn, Table: "posts", SoftDelete: true}

	// Destroy with soft delete should UPDATE deleted_at, not DELETE.
	mock.ExpectExec(regexp.QuoteMeta("UPDATE posts SET deleted_at")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := m.Destroy(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestDestroyHardDeleteWhenNoSoftDelete(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	conn := &DB{Conn: sqlDB, Dialect: QuestionDialect{}}
	m := &Model{DB: conn, Table: "posts"} // SoftDelete = false (default)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM posts WHERE id = ?")).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := m.Destroy(int64(5)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

// ── FirstOrCreate ─────────────────────────────────────────────────────────────

func TestFirstOrCreateReturnsExisting(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	conn := &DB{Conn: sqlDB, Dialect: QuestionDialect{}}
	m := &Model{DB: conn, Table: "users"}

	// Row already exists.
	mock.ExpectQuery("SELECT \\* FROM users WHERE email = \\?").
		WithArgs("alice@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "alice@example.com"))

	row, created, err := m.FirstOrCreate(
		map[string]any{"email": "alice@example.com"},
		map[string]any{"name": "Alice"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Fatal("should not have created a new row")
	}
	if row["email"] != "alice@example.com" {
		t.Fatalf("unexpected row: %v", row)
	}
}

// ── UpdateOrCreate ────────────────────────────────────────────────────────────

func TestUpdateOrCreateCreatesWhenMissing(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	conn := &DB{Conn: sqlDB, Dialect: QuestionDialect{}}
	m := &Model{DB: conn, Table: "settings"}

	// First query: row not found.
	mock.ExpectQuery("SELECT \\* FROM settings WHERE").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Insert merged data.
	mock.ExpectExec("INSERT INTO settings").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Re-fetch by id.
	mock.ExpectQuery("SELECT \\* FROM settings WHERE id = \\?").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "value"}).AddRow(1, "theme", "dark"))

	row, err := m.UpdateOrCreate(
		map[string]any{"key": "theme"},
		map[string]any{"value": "dark"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row["value"] != "dark" {
		t.Fatalf("expected value=dark, got %v", row["value"])
	}
}
