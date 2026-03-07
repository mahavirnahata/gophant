package validation

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mahavirnahata/gophant/db"
)

func TestRulesBasic(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	v := New(req).
		Field("email", Required(), Email()).
		Field("age", Numeric())

	if !v.Fails() {
		t.Fatalf("expected validation to fail")
	}
}

func TestConfirmed(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	req.Form = map[string][]string{
		"password":              {"abc"},
		"password_confirmation": {"abc"},
	}
	v := New(req).
		FieldWith("password", Confirmed("password_confirmation"))

	if v.Fails() {
		t.Fatalf("expected confirmed to pass")
	}
}

func TestUnique(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT 1 FROM users WHERE email = \\? LIMIT 1").WithArgs("a@b.com").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	conn := &db.DB{Conn: sqlDB, Dialect: db.QuestionDialect{}}
	req := httptest.NewRequest("POST", "/", strings.NewReader("email=a@b.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	v := New(req).
		FieldWith("email", Unique(conn, "users", "email"))

	if !v.Fails() {
		t.Fatalf("expected unique to fail")
	}
}
