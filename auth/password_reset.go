package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

// ErrTokenInvalid is returned when the reset token does not match any record.
var ErrTokenInvalid = errors.New("password reset token is invalid")

// ErrTokenExpired is returned when the reset token exists but has expired.
var ErrTokenExpired = errors.New("password reset token has expired")

// PasswordResetManager creates, verifies, and deletes password-reset tokens.
// Tokens are stored in a "password_resets" table:
//
//	CREATE TABLE password_resets (
//	  email      VARCHAR(255) NOT NULL,
//	  token      VARCHAR(64)  NOT NULL,
//	  created_at DATETIME     NOT NULL,
//	  PRIMARY KEY (email)
//	);
type PasswordResetManager struct {
	DB        *sql.DB
	Table     string        // default: "password_resets"
	Expiry    time.Duration // default: 1 hour
	TokenSize int           // bytes of randomness; hex-encoded → 2× length. Default: 32 → 64-char token.
}

// NewPasswordResetManager returns a manager using the given database connection.
func NewPasswordResetManager(db *sql.DB) *PasswordResetManager {
	return &PasswordResetManager{
		DB:        db,
		Table:     "password_resets",
		Expiry:    time.Hour,
		TokenSize: 32,
	}
}

// CreateToken generates and persists a reset token for the given email.
// Any previous token for that email is replaced (upsert).
func (m *PasswordResetManager) CreateToken(email string) (string, error) {
	token, err := m.generateToken()
	if err != nil {
		return "", err
	}
	now := time.Now()
	// Upsert: delete then insert (portable across MySQL/SQLite/PostgreSQL).
	if _, err := m.DB.Exec("DELETE FROM "+m.table()+" WHERE email = ?", email); err != nil {
		return "", err
	}
	if _, err := m.DB.Exec(
		"INSERT INTO "+m.table()+" (email, token, created_at) VALUES (?, ?, ?)",
		email, token, now,
	); err != nil {
		return "", err
	}
	return token, nil
}

// Verify checks that token belongs to email and has not expired.
// Returns nil on success, ErrTokenInvalid or ErrTokenExpired otherwise.
func (m *PasswordResetManager) Verify(email, token string) error {
	var createdAt time.Time
	err := m.DB.QueryRow(
		"SELECT created_at FROM "+m.table()+" WHERE email = ? AND token = ? LIMIT 1",
		email, token,
	).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTokenInvalid
	}
	if err != nil {
		return err
	}
	if time.Since(createdAt) > m.expiry() {
		return ErrTokenExpired
	}
	return nil
}

// Delete removes the reset token for email (call after a successful password change).
func (m *PasswordResetManager) Delete(email string) error {
	_, err := m.DB.Exec("DELETE FROM "+m.table()+" WHERE email = ?", email)
	return err
}

func (m *PasswordResetManager) table() string {
	if m.Table != "" {
		return m.Table
	}
	return "password_resets"
}

func (m *PasswordResetManager) expiry() time.Duration {
	if m.Expiry > 0 {
		return m.Expiry
	}
	return time.Hour
}

func (m *PasswordResetManager) generateToken() (string, error) {
	size := m.TokenSize
	if size <= 0 {
		size = 32
	}
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
