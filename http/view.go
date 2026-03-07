package http

import "io"

type ViewRenderer interface {
	Render(w io.Writer, name string, data map[string]any) error
}
