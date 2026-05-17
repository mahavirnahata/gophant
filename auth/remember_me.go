package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

const rememberCookieName = "remember_token"

// ErrRememberTokenInvalid is returned when the remember-me cookie token is missing or invalid.
var ErrRememberTokenInvalid = errors.New("remember: token is invalid or expired")

// RememberMeManager handles persistent "remember me" login via a long-lived cookie
// and a token stored in the database.
//
// Required table:
//
//	CREATE TABLE remember_tokens (
//	    token      VARCHAR(64)  NOT NULL PRIMARY KEY,
//	    user_id    VARCHAR(255) NOT NULL,
//	    expires_at DATETIME     NOT NULL
//	);
type RememberMeManager struct {
	DB         *sql.DB
	Table      string
	CookieName string
	Expiry     time.Duration // default: 30 days
	Secure     bool
}

// NewRememberMeManager returns a manager using the given DB connection.
func NewRememberMeManager(db *sql.DB) *RememberMeManager {
	return &RememberMeManager{
		DB:         db,
		Table:      "remember_tokens",
		CookieName: rememberCookieName,
		Expiry:     30 * 24 * time.Hour,
	}
}

// Remember sets a persistent remember-me cookie and stores the token in the DB.
// Call this after a successful login when the user ticked "remember me".
func (m *RememberMeManager) Remember(c *gomvchttp.Context, userID string) error {
	token, err := m.generateToken()
	if err != nil {
		return err
	}
	expiry := m.expiry()
	expiresAt := time.Now().Add(expiry)

	if _, err := m.DB.Exec(
		"INSERT INTO "+m.table()+" (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expiresAt,
	); err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     m.cookieName(),
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   m.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// ResolveUser reads the remember-me cookie and returns the associated user ID.
// Returns ErrRememberTokenInvalid if the cookie is missing, expired, or not found.
func (m *RememberMeManager) ResolveUser(c *gomvchttp.Context) (string, error) {
	cookie, err := c.Cookie(m.cookieName())
	if err != nil {
		return "", ErrRememberTokenInvalid
	}
	token := cookie.Value
	if token == "" {
		return "", ErrRememberTokenInvalid
	}

	var userID string
	var expiresAt time.Time
	err = m.DB.QueryRow(
		"SELECT user_id, expires_at FROM "+m.table()+" WHERE token = ? LIMIT 1",
		token,
	).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrRememberTokenInvalid
	}
	if err != nil {
		return "", err
	}
	if time.Now().After(expiresAt) {
		_ = m.Forget(c) // clean up expired token
		return "", ErrRememberTokenInvalid
	}
	return userID, nil
}

// Forget deletes the remember-me token from the DB and clears the cookie.
func (m *RememberMeManager) Forget(c *gomvchttp.Context) error {
	cookie, err := c.Cookie(m.cookieName())
	if err == nil && cookie.Value != "" {
		_, _ = m.DB.Exec("DELETE FROM "+m.table()+" WHERE token = ?", cookie.Value)
	}
	c.SetCookie(&http.Cookie{
		Name:     m.cookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.Secure,
	})
	return nil
}

// PurgeExpired deletes all expired tokens from the database.
// Call this periodically (e.g., via the scheduler).
func (m *RememberMeManager) PurgeExpired() error {
	_, err := m.DB.Exec(
		"DELETE FROM "+m.table()+" WHERE expires_at < ?", time.Now(),
	)
	return err
}

func (m *RememberMeManager) generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (m *RememberMeManager) table() string {
	if m.Table != "" {
		return m.Table
	}
	return "remember_tokens"
}

func (m *RememberMeManager) cookieName() string {
	if m.CookieName != "" {
		return m.CookieName
	}
	return rememberCookieName
}

func (m *RememberMeManager) expiry() time.Duration {
	if m.Expiry > 0 {
		return m.Expiry
	}
	return 30 * 24 * time.Hour
}
