package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

const flashPrefix = "_flash:"

type Manager struct {
	Store      Store
	CookieName string
	Secret     []byte
	Secure     bool
	SameSite   http.SameSite
	MaxAge     time.Duration
}

type Session struct {
	ID        string
	Values    map[string]any
	manager   *Manager
	changed   bool
	destroyed bool
	newID     bool
}

var errInvalidCookie = errors.New("invalid session cookie")

func NewManager(secret []byte) *Manager {
	return &Manager{
		Store:      NewMemoryStore(),
		CookieName: "_gophant_session",
		Secret:     secret,
		Secure:     false,
		SameSite:   http.SameSiteLaxMode,
		MaxAge:     7 * 24 * time.Hour,
	}
}

func (m *Manager) Middleware() gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			sess, _ := m.load(c.Request)
			c.Set("session", sess)

			// Auto-inject flash values into context so templates get {{ .flash_error }} etc.
			for k, v := range sess.Values {
				if strings.HasPrefix(k, flashPrefix) {
					c.Set("flash_"+strings.TrimPrefix(k, flashPrefix), v)
				}
			}

			wrapper := &sessionWriter{ResponseWriter: c.Writer, mgr: m, sess: sess}
			c.Writer = wrapper

			next(c)

			if !wrapper.wroteHeader {
				wrapper.ensureCookie()
			}
		}
	}
}

// FromContext retrieves the current session from the request context.
func FromContext(c *gomvchttp.Context) *Session {
	if v, ok := c.Get("session"); ok {
		if s, ok := v.(*Session); ok {
			return s
		}
	}
	return &Session{Values: map[string]any{}}
}

// ── Value accessors ──────────────────────────────────────────────────────────

func (s *Session) Get(key string) (any, bool) {
	v, ok := s.Values[key]
	return v, ok
}

func (s *Session) Set(key string, val any) {
	s.Values[key] = val
	s.changed = true
}

func (s *Session) Delete(key string) {
	delete(s.Values, key)
	s.changed = true
}

// All returns a copy of all non-flash session values.
func (s *Session) All() map[string]any {
	out := make(map[string]any, len(s.Values))
	for k, v := range s.Values {
		if !strings.HasPrefix(k, flashPrefix) {
			out[k] = v
		}
	}
	return out
}

// ── Flash ────────────────────────────────────────────────────────────────────

// Flash stores a value that persists through exactly one redirect.
// Retrieve it on the next request with GetFlash.
func (s *Session) Flash(key string, val any) {
	s.Set(flashPrefix+key, val)
}

// GetFlash reads a flashed value and removes it from the session (delete-on-read).
func (s *Session) GetFlash(key string) (any, bool) {
	v, ok := s.Get(flashPrefix + key)
	if ok {
		s.Delete(flashPrefix + key)
	}
	return v, ok
}

// FlashString is a convenience wrapper that returns the flash value as a string.
func (s *Session) FlashString(key string) string {
	v, ok := s.GetFlash(key)
	if !ok {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}

// HasFlash reports whether a flash value exists for key.
func (s *Session) HasFlash(key string) bool {
	_, ok := s.Values[flashPrefix+key]
	return ok
}

// ── Lifecycle ────────────────────────────────────────────────────────────────

func (s *Session) Regenerate() {
	if s.manager == nil {
		return
	}
	oldID := s.ID
	s.ID = newID()
	s.newID = true
	if oldID != "" {
		_ = s.manager.Store.Delete(oldID)
	}
}

func (s *Session) Destroy() {
	s.destroyed = true
}

// ── Internal ─────────────────────────────────────────────────────────────────

func (m *Manager) load(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(m.CookieName)
	if err != nil || cookie.Value == "" {
		return m.newSession(), nil
	}
	id, err := m.verifyCookie(cookie.Value)
	if err != nil {
		return m.newSession(), err
	}
	values, ok := m.Store.Get(id)
	if !ok {
		values = map[string]any{}
	}
	return &Session{ID: id, Values: values, manager: m}, nil
}

func (m *Manager) newSession() *Session {
	id := newID()
	return &Session{ID: id, Values: map[string]any{}, manager: m, newID: true}
}

func newID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (m *Manager) setCookie(w http.ResponseWriter, id string) {
	value := m.signCookie(id)
	http.SetCookie(w, &http.Cookie{
		Name:     m.CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.Secure,
		SameSite: m.SameSite,
		Expires:  time.Now().Add(m.MaxAge),
	})
}

func (m *Manager) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.Secure,
		SameSite: m.SameSite,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

type sessionWriter struct {
	http.ResponseWriter
	mgr         *Manager
	sess        *Session
	wroteHeader bool
}

func (w *sessionWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.ensureCookie()
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *sessionWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (w *sessionWriter) ensureCookie() {
	if w.sess == nil || w.mgr == nil {
		return
	}
	if w.sess.destroyed {
		_ = w.mgr.Store.Delete(w.sess.ID)
		w.mgr.clearCookie(w.ResponseWriter)
		return
	}
	if w.sess.changed || w.sess.newID {
		_ = w.mgr.Store.Save(w.sess.ID, w.sess.Values)
		w.mgr.setCookie(w.ResponseWriter, w.sess.ID)
	}
}

func (m *Manager) signCookie(id string) string {
	mac := hmac.New(sha256.New, m.Secret)
	_, _ = mac.Write([]byte(id))
	sig := mac.Sum(nil)
	return id + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func (m *Manager) verifyCookie(value string) (string, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return "", errInvalidCookie
	}
	id := parts[0]
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errInvalidCookie
	}
	mac := hmac.New(sha256.New, m.Secret)
	_, _ = mac.Write([]byte(id))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return "", errInvalidCookie
	}
	return id, nil
}
