package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryInsertHooks(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))

	called := false
	repo := &Repository{
		DB:    &DB{Conn: sqlDB, Dialect: QuestionDialect{}},
		Table: "users",
		Hooks: Hooks{
			BeforeCreate: func(data map[string]any) error {
				called = true
				return nil
			},
		},
	}

	if err := repo.Insert(map[string]any{"name": "a"}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if !called {
		t.Fatalf("expected hook")
	}
}
