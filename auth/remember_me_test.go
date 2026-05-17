package auth

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gomvchttp "github.com/mahavirnahata/gophant/http"
)

func newRememberCtx() (*gomvchttp.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c := gomvchttp.NewContext(w, r, nil)
	return c, w
}

func TestRememberMe_Remember(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewRememberMeManager(db)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO remember_tokens (token, user_id, expires_at) VALUES (?, ?, ?)")).
		WithArgs(sqlmock.AnyArg(), "user-42", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	c, w := newRememberCtx()
	if err := mgr.Remember(c, "user-42"); err != nil {
		t.Fatalf("Remember: %v", err)
	}
	cookies := w.Result().Cookies()
	var found bool
	for _, ck := range cookies {
		if ck.Name == "remember_token" {
			found = true
			if len(ck.Value) != 64 {
				t.Fatalf("expected 64-char token in cookie, got %d", len(ck.Value))
			}
		}
	}
	if !found {
		t.Fatal("remember_token cookie not set")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

func TestRememberMe_ResolveUser_Valid(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewRememberMeManager(db)

	expiresAt := time.Now().Add(time.Hour)
	rows := sqlmock.NewRows([]string{"user_id", "expires_at"}).AddRow("user-42", expiresAt)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, expires_at FROM remember_tokens WHERE token = ? LIMIT 1")).
		WithArgs("mytoken").WillReturnRows(rows)

	c, _ := newRememberCtx()
	c.Request.AddCookie(&http.Cookie{Name: "remember_token", Value: "mytoken"})

	userID, err := mgr.ResolveUser(c)
	if err != nil {
		t.Fatalf("ResolveUser: %v", err)
	}
	if userID != "user-42" {
		t.Fatalf("expected user-42, got %s", userID)
	}
}

func TestRememberMe_ResolveUser_NoCookie(t *testing.T) {
	db, _ := newMockDB(t)
	mgr := NewRememberMeManager(db)
	c, _ := newRememberCtx()

	_, err := mgr.ResolveUser(c)
	if err != ErrRememberTokenInvalid {
		t.Fatalf("expected ErrRememberTokenInvalid, got %v", err)
	}
}

func TestRememberMe_ResolveUser_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewRememberMeManager(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, expires_at FROM remember_tokens")).
		WithArgs("badtoken").WillReturnError(sql.ErrNoRows)

	c, _ := newRememberCtx()
	c.Request.AddCookie(&http.Cookie{Name: "remember_token", Value: "badtoken"})

	_, err := mgr.ResolveUser(c)
	if err != ErrRememberTokenInvalid {
		t.Fatalf("expected ErrRememberTokenInvalid, got %v", err)
	}
}

func TestRememberMe_ResolveUser_Expired(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewRememberMeManager(db)

	expiredAt := time.Now().Add(-time.Hour)
	rows := sqlmock.NewRows([]string{"user_id", "expires_at"}).AddRow("user-42", expiredAt)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, expires_at FROM remember_tokens")).
		WithArgs("expiredtoken").WillReturnRows(rows)
	// Expect the cleanup DELETE
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM remember_tokens WHERE token = ?")).
		WithArgs("expiredtoken").WillReturnResult(sqlmock.NewResult(0, 1))

	c, _ := newRememberCtx()
	c.Request.AddCookie(&http.Cookie{Name: "remember_token", Value: "expiredtoken"})

	_, err := mgr.ResolveUser(c)
	if err != ErrRememberTokenInvalid {
		t.Fatalf("expected ErrRememberTokenInvalid for expired token, got %v", err)
	}
}

func TestRememberMe_Forget(t *testing.T) {
	db, mock := newMockDB(t)
	mgr := NewRememberMeManager(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM remember_tokens WHERE token = ?")).
		WithArgs("tok").WillReturnResult(sqlmock.NewResult(0, 1))

	c, w := newRememberCtx()
	c.Request.AddCookie(&http.Cookie{Name: "remember_token", Value: "tok"})
	if err := mgr.Forget(c); err != nil {
		t.Fatalf("Forget: %v", err)
	}

	var cleared bool
	for _, ck := range w.Result().Cookies() {
		if ck.Name == "remember_token" && ck.MaxAge == -1 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("expected cookie to be cleared (MaxAge=-1)")
	}
}
