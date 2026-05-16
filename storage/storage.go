// Package storage provides a unified file-storage abstraction.
// The default driver writes to the local filesystem. Swap to an S3-compatible
// driver by implementing the Driver interface.
//
// Usage:
//
//	disk := storage.New(storage.NewLocalDriver("storage/app", "/files"))
//	disk.Put("avatars/alice.png", file)
//	url  := disk.URL("avatars/alice.png")   // → "/files/avatars/alice.png"
//	exists, _ := disk.Exists("avatars/alice.png")
package storage

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotFound is returned by Get when the path does not exist.
var ErrNotFound = errors.New("storage: file not found")

// Driver is the interface that every storage backend must implement.
type Driver interface {
	// Put stores content at path, creating intermediate directories as needed.
	Put(path string, content io.Reader) error
	// Get returns a ReadCloser for the file at path.
	Get(path string) (io.ReadCloser, error)
	// Delete removes the file at path.
	Delete(path string) error
	// Exists reports whether path exists.
	Exists(path string) (bool, error)
	// URL returns a public URL for the file at path.
	URL(path string) string
}

// Storage wraps a Driver with convenience methods.
type Storage struct {
	driver Driver
}

// New returns a Storage backed by driver.
func New(driver Driver) *Storage { return &Storage{driver: driver} }

func (s *Storage) Put(path string, content io.Reader) error    { return s.driver.Put(path, content) }
func (s *Storage) Get(path string) (io.ReadCloser, error)      { return s.driver.Get(path) }
func (s *Storage) Delete(path string) error                    { return s.driver.Delete(path) }
func (s *Storage) Exists(path string) (bool, error)            { return s.driver.Exists(path) }
func (s *Storage) URL(path string) string                      { return s.driver.URL(path) }

// PutFile reads src entirely into path. Convenience wrapper around Put.
func (s *Storage) PutFile(path string, src io.Reader) error {
	return s.driver.Put(path, src)
}

// ── Local Driver ──────────────────────────────────────────────────────────────

// LocalDriver stores files on the local filesystem.
type LocalDriver struct {
	root    string // absolute or relative filesystem root
	baseURL string // URL prefix for URL()
}

// NewLocalDriver creates a driver rooted at root.
// baseURL is prepended to paths returned by URL() (e.g. "/storage").
func NewLocalDriver(root, baseURL string) *LocalDriver {
	return &LocalDriver{
		root:    root,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (d *LocalDriver) fullPath(path string) string {
	return filepath.Join(d.root, filepath.FromSlash(path))
}

func (d *LocalDriver) Put(path string, content io.Reader) error {
	full := d.fullPath(path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, content)
	return err
}

func (d *LocalDriver) Get(path string) (io.ReadCloser, error) {
	f, err := os.Open(d.fullPath(path))
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return f, err
}

func (d *LocalDriver) Delete(path string) error {
	err := os.Remove(d.fullPath(path))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (d *LocalDriver) Exists(path string) (bool, error) {
	_, err := os.Stat(d.fullPath(path))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (d *LocalDriver) URL(path string) string {
	return d.baseURL + "/" + strings.TrimLeft(path, "/")
}
