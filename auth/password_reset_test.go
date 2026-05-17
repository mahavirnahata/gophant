package auth

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestPasswordReset_CreateToken(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewPasswordResetManager(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM password_resets WHERE email = ?")).
		WithArgs("user@example.com").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO password_resets (email, token, created_at) VALUES (?, ?, ?)")).
		WithArgs("user@example.com", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	token, err := mgr.CreateToken("user@example.com")
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if len(token) != 64 { // 32 bytes → 64-char hex
		t.Fatalf("expected 64-char token, got %d", len(token))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPasswordReset_TokenLength(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewPasswordResetManager(db)
	mgr.TokenSize = 16

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM password_resets WHERE email = ?")).
		WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO password_resets")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	token, _ := mgr.CreateToken("user@example.com")
	if len(token) != 32 { // 16 bytes → 32-char hex
		t.Fatalf("expected 32-char token, got %d", len(token))
	}
}

func TestPasswordReset_Verify_Valid(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewPasswordResetManager(db)

	createdAt := time.Now()
	rows := sqlmock.NewRows([]string{"created_at"}).AddRow(createdAt)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT created_at FROM password_resets WHERE email = ? AND token = ? LIMIT 1")).
		WithArgs("user@example.com", "mytoken").
		WillReturnRows(rows)

	if err := mgr.Verify("user@example.com", "mytoken"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestPasswordReset_Verify_InvalidToken(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewPasswordResetManager(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT created_at FROM password_resets")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	err := mgr.Verify("user@example.com", "badtoken")
	if err != ErrTokenInvalid {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestPasswordReset_Verify_Expired(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewPasswordResetManager(db)
	mgr.Expiry = time.Millisecond

	createdAt := time.Now().Add(-time.Hour) // well in the past
	rows := sqlmock.NewRows([]string{"created_at"}).AddRow(createdAt)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT created_at FROM password_resets")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	err := mgr.Verify("user@example.com", "mytoken")
	if err != ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestPasswordReset_Delete(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewPasswordResetManager(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM password_resets WHERE email = ?")).
		WithArgs("user@example.com").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := mgr.Delete("user@example.com"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPasswordReset_CustomTable(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewPasswordResetManager(db)
	mgr.Table = "custom_resets"

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM custom_resets WHERE email = ?")).
		WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO custom_resets")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := mgr.CreateToken("user@example.com"); err != nil {
		t.Fatalf("CreateToken with custom table: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
