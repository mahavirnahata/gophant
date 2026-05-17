package auth

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestEmailVerify_CreateToken(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewEmailVerificationManager(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM email_verifications WHERE email = ?")).
		WithArgs("user@example.com").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO email_verifications (email, token, created_at) VALUES (?, ?, ?)")).
		WithArgs("user@example.com", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	token, err := mgr.CreateToken("user@example.com")
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("expected 64-char token, got %d", len(token))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

func TestEmailVerify_Verify_Valid(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewEmailVerificationManager(db)

	rows := sqlmock.NewRows([]string{"created_at"}).AddRow(time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT created_at FROM email_verifications WHERE email = ? AND token = ? LIMIT 1")).
		WithArgs("user@example.com", "mytoken").WillReturnRows(rows)

	if err := mgr.Verify("user@example.com", "mytoken"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestEmailVerify_Verify_Invalid(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewEmailVerificationManager(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT created_at FROM email_verifications")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	err := mgr.Verify("user@example.com", "badtoken")
	if err != ErrVerifyTokenInvalid {
		t.Fatalf("expected ErrVerifyTokenInvalid, got %v", err)
	}
}

func TestEmailVerify_Verify_Expired(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewEmailVerificationManager(db)
	mgr.Expiry = time.Millisecond

	rows := sqlmock.NewRows([]string{"created_at"}).AddRow(time.Now().Add(-time.Hour))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT created_at FROM email_verifications")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(rows)

	err := mgr.Verify("user@example.com", "tok")
	if err != ErrVerifyTokenExpired {
		t.Fatalf("expected ErrVerifyTokenExpired, got %v", err)
	}
}

func TestEmailVerify_MarkVerified(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewEmailVerificationManager(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM email_verifications WHERE email = ?")).
		WithArgs("user@example.com").WillReturnResult(sqlmock.NewResult(0, 1))

	if err := mgr.MarkVerified("user@example.com"); err != nil {
		t.Fatalf("MarkVerified: %v", err)
	}
}

func TestEmailVerify_HasPending_True(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewEmailVerificationManager(db)

	rows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM email_verifications WHERE email = ?")).
		WithArgs("user@example.com").WillReturnRows(rows)

	has, err := mgr.HasPending("user@example.com")
	if err != nil {
		t.Fatalf("HasPending: %v", err)
	}
	if !has {
		t.Fatal("expected HasPending=true")
	}
}

func TestEmailVerify_HasPending_False(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewEmailVerificationManager(db)

	rows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM email_verifications WHERE email = ?")).
		WithArgs("user@example.com").WillReturnRows(rows)

	has, err := mgr.HasPending("user@example.com")
	if err != nil || has {
		t.Fatalf("expected HasPending=false, got has=%v err=%v", has, err)
	}
}

func TestEmailVerify_CustomTable(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewEmailVerificationManager(db)
	mgr.Table = "custom_verifications"

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM custom_verifications WHERE email = ?")).
		WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO custom_verifications")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := mgr.CreateToken("user@example.com"); err != nil {
		t.Fatalf("CreateToken with custom table: %v", err)
	}
}
