package storage

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestLocalDriverPutGetDelete(t *testing.T) {
	dir := t.TempDir()
	d := NewLocalDriver(dir, "/files")
	s := New(d)

	// Put
	if err := s.Put("sub/hello.txt", strings.NewReader("hello world")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Exists
	ok, err := s.Exists("sub/hello.txt")
	if err != nil || !ok {
		t.Fatalf("Exists: %v %v", ok, err)
	}

	// Get
	rc, err := s.Get("sub/hello.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer rc.Close()
	body, _ := io.ReadAll(rc)
	if string(body) != "hello world" {
		t.Fatalf("unexpected content: %q", body)
	}

	// URL
	if got := s.URL("sub/hello.txt"); got != "/files/sub/hello.txt" {
		t.Fatalf("URL: %q", got)
	}

	// Delete
	if err := s.Delete("sub/hello.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	ok, _ = s.Exists("sub/hello.txt")
	if ok {
		t.Fatal("file should not exist after Delete")
	}
}

func TestLocalDriverGetMissing(t *testing.T) {
	d := NewLocalDriver(t.TempDir(), "/f")
	_, err := d.Get("missing.txt")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalDriverDeleteMissing(t *testing.T) {
	d := NewLocalDriver(t.TempDir(), "/f")
	// Should not error on missing file.
	if err := d.Delete("ghost.txt"); err != nil {
		t.Fatalf("Delete missing: %v", err)
	}
}

func TestLocalDriverCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	d := NewLocalDriver(dir, "/f")
	if err := d.Put("a/b/c/file.txt", strings.NewReader("data")); err != nil {
		t.Fatalf("Put nested: %v", err)
	}
	if _, err := os.Stat(dir + "/a/b/c/file.txt"); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}
