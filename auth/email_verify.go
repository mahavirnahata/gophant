package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

// ErrVerifyTokenInvalid is returned when the verification token does not match.
var ErrVerifyTokenInvalid = errors.New("email verification token is invalid")

// ErrVerifyTokenExpired is returned when the verification token has expired.
var ErrVerifyTokenExpired = errors.New("email verification token has expired")

// ErrAlreadyVerified is returned when the email is already verified.
var ErrAlreadyVerified = errors.New("email is already verified")

// EmailVerificationManager creates, verifies, and invalidates email-verification tokens.
//
// Required table:
//
//	CREATE TABLE email_verifications (
//	    email      VARCHAR(255) NOT NULL PRIMARY KEY,
//	    token      VARCHAR(64)  NOT NULL,
//	    created_at DATETIME     NOT NULL
//	);
//
// Your users table should have an email_verified_at DATETIME NULL column.
type EmailVerificationManager struct {
	DB        *sql.DB
	Table     string        // default: "email_verifications"
	Expiry    time.Duration // default: 24 hours
	TokenSize int           // bytes of randomness (hex-encoded). Default: 32 → 64-char token.
}

// NewEmailVerificationManager returns a manager using the given connection.
func NewEmailVerificationManager(db *sql.DB) *EmailVerificationManager {
	return &EmailVerificationManager{
		DB:        db,
		Table:     "email_verifications",
		Expiry:    24 * time.Hour,
		TokenSize: 32,
	}
}

// CreateToken generates and stores a verification token for the given email.
// Any previous token for that email is replaced.
func (m *EmailVerificationManager) CreateToken(email string) (string, error) {
	size := m.TokenSize
	if size <= 0 {
		size = 32
	}
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	now := time.Now()

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

// Verify checks that token matches the record for email and has not expired.
// Returns nil on success.
func (m *EmailVerificationManager) Verify(email, token string) error {
	var createdAt time.Time
	err := m.DB.QueryRow(
		"SELECT created_at FROM "+m.table()+" WHERE email = ? AND token = ? LIMIT 1",
		email, token,
	).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrVerifyTokenInvalid
	}
	if err != nil {
		return err
	}
	expiry := m.Expiry
	if expiry <= 0 {
		expiry = 24 * time.Hour
	}
	if time.Since(createdAt) > expiry {
		return ErrVerifyTokenExpired
	}
	return nil
}

// MarkVerified removes the pending token (call after Verify succeeds).
// Your application is responsible for setting email_verified_at on the user record.
func (m *EmailVerificationManager) MarkVerified(email string) error {
	_, err := m.DB.Exec("DELETE FROM "+m.table()+" WHERE email = ?", email)
	return err
}

// HasPending reports whether a pending (not yet verified) token exists for email.
func (m *EmailVerificationManager) HasPending(email string) (bool, error) {
	var count int
	err := m.DB.QueryRow(
		"SELECT COUNT(*) FROM "+m.table()+" WHERE email = ?", email,
	).Scan(&count)
	return count > 0, err
}

func (m *EmailVerificationManager) table() string {
	if m.Table != "" {
		return m.Table
	}
	return "email_verifications"
}
