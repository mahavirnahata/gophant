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
	funcs    template.FuncMap
}

func New(basePath string) *Engine {
	return &Engine{
		basePath: basePath,
		funcs:    template.FuncMap{},
	}
}

// AddFunc registers a single template function. Must be called before Load().
func (e *Engine) AddFunc(name string, fn any) {
	e.funcs[name] = fn
}

// AddFuncs registers multiple template functions at once. Must be called before Load().
func (e *Engine) AddFuncs(funcs template.FuncMap) {
	for k, v := range funcs {
		e.funcs[k] = v
	}
}

func (e *Engine) Load(pattern string) error {
	files, err := e.findFiles(pattern)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("no view files found")
	}
	t, err := template.New("").Funcs(e.funcs).ParseFiles(files...)
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
