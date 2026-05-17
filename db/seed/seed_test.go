package seed

import (
	"database/sql"
	"errors"
	"testing"
)

type countSeeder struct{ ran int }

func (s *countSeeder) Run(_ *sql.DB) error {
	s.ran++
	return nil
}

type failSeeder struct{}

func (s *failSeeder) Run(_ *sql.DB) error {
	return errors.New("seeder failed")
}

func setup() {
	Reset()
}

func TestRegisterAndRunAll(t *testing.T) {
	setup()
	s := &countSeeder{}
	Register(s)
	if err := RunAll(nil); err != nil {
		t.Fatalf("RunAll: %v", err)
	}
	if s.ran != 1 {
		t.Fatalf("expected seeder to run once, ran %d times", s.ran)
	}
}

func TestRunAllOrder(t *testing.T) {
	setup()
	order := []string{}
	makeSeeder := func(name string) Seeder {
		return &namedRunSeeder{name: name, order: &order}
	}
	Register(makeSeeder("first"))
	Register(makeSeeder("second"))
	Register(makeSeeder("third"))

	if err := RunAll(nil); err != nil {
		t.Fatalf("RunAll: %v", err)
	}
	if order[0] != "first" || order[1] != "second" || order[2] != "third" {
		t.Fatalf("wrong order: %v", order)
	}
}

type namedRunSeeder struct {
	name  string
	order *[]string
}

func (s *namedRunSeeder) Run(_ *sql.DB) error {
	*s.order = append(*s.order, s.name)
	return nil
}

func TestRunAllStopsOnError(t *testing.T) {
	setup()
	s1 := &countSeeder{}
	Register(s1)
	Register(&failSeeder{})
	s3 := &countSeeder{}
	Register(s3)

	err := RunAll(nil)
	if err == nil {
		t.Fatal("expected error from failSeeder")
	}
	if s1.ran != 1 {
		t.Fatalf("first seeder should have run once, ran %d", s1.ran)
	}
	if s3.ran != 0 {
		t.Fatalf("third seeder should not run after failure")
	}
}

func TestRunSingle(t *testing.T) {
	s := &countSeeder{}
	if err := Run(nil, s); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if s.ran != 1 {
		t.Fatalf("expected ran=1, got %d", s.ran)
	}
}

func TestRunSingleError(t *testing.T) {
	err := Run(nil, &failSeeder{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSeeders(t *testing.T) {
	setup()
	Register(&countSeeder{})
	Register(&failSeeder{})
	names := Seeders()
	if len(names) != 2 {
		t.Fatalf("expected 2 seeders, got %d", len(names))
	}
}

func TestReset(t *testing.T) {
	setup()
	Register(&countSeeder{})
	Reset()
	if len(Seeders()) != 0 {
		t.Fatal("expected empty seeder list after Reset")
	}
}

func TestRunAllEmpty(t *testing.T) {
	setup()
	if err := RunAll(nil); err != nil {
		t.Fatalf("RunAll on empty list: %v", err)
	}
}
