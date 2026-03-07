package view

import (
	"errors"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

type Engine struct {
	basePath string
	tmpl     *template.Template
}

func New(basePath string) *Engine {
	return &Engine{basePath: basePath}
}

func (e *Engine) Load(pattern string) error {
	files, err := e.findFiles(pattern)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("no view files found")
	}
	t, err := template.ParseFiles(files...)
	if err != nil {
		return err
	}
	e.tmpl = t
	return nil
}

func (e *Engine) findFiles(pattern string) ([]string, error) {
	if strings.Contains(pattern, "**") {
		return walkFiles(e.basePath)
	}
	return filepath.Glob(filepath.Join(e.basePath, pattern))
}

func walkFiles(base string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".html") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (e *Engine) Render(w io.Writer, name string, data map[string]any) error {
	if e.tmpl == nil {
		return errors.New("view templates not loaded")
	}
	return e.tmpl.ExecuteTemplate(w, name, data)
}
