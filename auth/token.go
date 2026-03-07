package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/mahavirnahata/gophant/db"
)

type TokenService struct {
	DB        *db.DB
	Table     string
	HashKey   string
	UserKey   string
	NameKey   string
	Abilities string
	ExpiresAt string
}

type TokenInfo struct {
	UserID    string
	Abilities []string
}

func NewTokenService(conn *db.DB) *TokenService {
	return &TokenService{
		DB:        conn,
		Table:     "personal_access_tokens",
		HashKey:   "token_hash",
		UserKey:   "user_id",
		NameKey:   "name",
		Abilities: "abilities",
		ExpiresAt: "expires_at",
	}
}

func (s *TokenService) Create(userID string, name string, abilities []string, ttl time.Duration) (string, error) {
	if s == nil || s.DB == nil {
		return "", errors.New("token service not configured")
	}
	plain, err := randomToken()
	if err != nil {
		return "", err
	}
	hash := hashToken(plain)
	abilitiesJSON, _ := json.Marshal(abilities)

	data := map[string]any{
		s.HashKey:   hash,
		s.UserKey:   userID,
		s.NameKey:   name,
		s.Abilities: string(abilitiesJSON),
	}
	if ttl > 0 {
		data[s.ExpiresAt] = time.Now().Add(ttl)
	}

	_, err = s.DB.Table(s.Table).Insert(data)
	if err != nil {
		return "", err
	}
	return plain, nil
}

func (s *TokenService) Authenticate(bearer string) (*TokenInfo, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("token service not configured")
	}
	if bearer == "" {
		return nil, errors.New("token required")
	}
	hash := hashToken(bearer)

	row, err := s.DB.Table(s.Table).Where(s.HashKey, "=", hash).First()
	if err != nil {
		return nil, err
	}

	if exp, ok := row[s.ExpiresAt]; ok {
		if t, ok := toTime(exp); ok && time.Now().After(t) {
			return nil, errors.New("token expired")
		}
	}

	info := &TokenInfo{UserID: toString(row[s.UserKey])}
	if ab, ok := row[s.Abilities]; ok {
		_ = json.Unmarshal([]byte(toString(ab)), &info.Abilities)
	}
	return info, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

func toTime(v any) (time.Time, bool) {
	switch t := v.(type) {
	case time.Time:
		return t, true
	case []byte:
		parsed, err := time.Parse(time.RFC3339, string(t))
		return parsed, err == nil
	case string:
		parsed, err := time.Parse(time.RFC3339, t)
		return parsed, err == nil
	default:
		return time.Time{}, false
	}
}
