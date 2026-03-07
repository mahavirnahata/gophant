package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
	if err := LoadEnvFile(path); err != nil {
		t.Fatalf("load env: %v", err)
	}
	if os.Getenv("FOO") != "bar" {
		t.Fatalf("expected env")
	}
}
