package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewRender(t *testing.T) {
	dir := t.TempDir()
	tpl := `{{ define "a" }}Hello {{ .name }}{{ end }}`
	if err := os.WriteFile(filepath.Join(dir, "a.html"), []byte(tpl), 0o644); err != nil {
		t.Fatalf("write tpl: %v", err)
	}

	v := New(dir)
	if err := v.Load("*.html"); err != nil {
		t.Fatalf("load: %v", err)
	}

	var out strings.Builder
	if err := v.Render(&out, "a", map[string]any{"name": "Go"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	if out.String() != "Hello Go" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}
