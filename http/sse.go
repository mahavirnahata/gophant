package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSEWriter streams Server-Sent Events to the client.
// Obtain one from c.SSE() and call Send in a loop until the client disconnects.
//
//	router.Get("/events", func(c *http.Context) {
//	    sse := c.SSE()
//	    for {
//	        select {
//	        case <-c.Context().Done():
//	            return
//	        case msg := <-myChannel:
//	            if err := sse.Send(msg); err != nil {
//	                return
//	            }
//	        }
//	    }
//	})
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// SSE sets the appropriate SSE response headers and returns an SSEWriter.
// Returns nil if the underlying ResponseWriter does not support flushing.
func (c *Context) SSE() *SSEWriter {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return nil
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering
	c.Written = true
	return &SSEWriter{w: c.Writer, flusher: flusher}
}

// Send writes a data event with an optional event type and ID.
//
//	sse.Send(SSEEvent{Data: "hello"})
//	sse.Send(SSEEvent{Event: "update", Data: map[string]any{"id": 1}})
func (s *SSEWriter) Send(event SSEEvent) error {
	if event.ID != "" {
		fmt.Fprintf(s.w, "id: %s\n", event.ID)
	}
	if event.Event != "" {
		fmt.Fprintf(s.w, "event: %s\n", event.Event)
	}

	var data string
	switch v := event.Data.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		b, err := json.Marshal(event.Data)
		if err != nil {
			return err
		}
		data = string(b)
	}
	fmt.Fprintf(s.w, "data: %s\n\n", data)
	s.flusher.Flush()
	return nil
}

// SendString is a convenience for sending a plain string data event.
func (s *SSEWriter) SendString(data string) error {
	return s.Send(SSEEvent{Data: data})
}

// SendJSON is a convenience for sending a JSON-encoded data event.
func (s *SSEWriter) SendJSON(v any) error {
	return s.Send(SSEEvent{Data: v})
}

// Comment writes an SSE comment (begins with ':'). Useful as a keep-alive ping.
func (s *SSEWriter) Comment(msg string) {
	fmt.Fprintf(s.w, ": %s\n\n", msg)
	s.flusher.Flush()
}

// Retry instructs the client to reconnect after ms milliseconds on disconnect.
func (s *SSEWriter) Retry(ms int) {
	fmt.Fprintf(s.w, "retry: %d\n\n", ms)
	s.flusher.Flush()
}

// SSEEvent represents a single Server-Sent Event.
type SSEEvent struct {
	// ID is the optional event ID (used by clients to resume after reconnect).
	ID string
	// Event is the optional event type name (default: "message").
	Event string
	// Data is the event payload. Strings are sent as-is; everything else is JSON-encoded.
	Data any
}
