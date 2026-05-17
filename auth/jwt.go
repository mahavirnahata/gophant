package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

// Sentinel errors returned by JWTManager.Verify.
var (
	ErrJWTMalformed = errors.New("jwt: token is malformed")
	ErrJWTInvalid   = errors.New("jwt: signature is invalid")
	ErrJWTExpired   = errors.New("jwt: token has expired")
)

// TokenPair holds an access token and a refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expires
}

// JWTManager signs and verifies HS256 JWT tokens using only the standard library.
//
// Usage:
//
//	mgr := auth.NewJWTManager([]byte(cfg.AppKey), 24*time.Hour)
//	token, _ := mgr.Sign(map[string]any{"sub": "42", "role": "admin"})
//	claims, err := mgr.Verify(token)
type JWTManager struct {
	secret         []byte
	expiry         time.Duration
	refreshExpiry  time.Duration // default: 30 days
}

// NewJWTManager returns a manager that signs tokens with secret and expires
// them after expiry. The "sub" claim is required when calling Sign.
func NewJWTManager(secret []byte, expiry time.Duration) *JWTManager {
	return &JWTManager{secret: secret, expiry: expiry, refreshExpiry: 30 * 24 * time.Hour}
}

// WithRefreshExpiry sets the lifetime of refresh tokens (default: 30 days).
func (j *JWTManager) WithRefreshExpiry(d time.Duration) *JWTManager {
	j.refreshExpiry = d
	return j
}

// Sign creates a signed JWT for the given claims.
// "iat" (issued-at) and "exp" (expiry) are added automatically.
// Returns ErrJWTInvalid if "sub" is missing.
func (j *JWTManager) Sign(claims map[string]any) (string, error) {
	if _, ok := claims["sub"]; !ok {
		return "", errors.New("jwt: 'sub' claim is required")
	}
	now := time.Now()
	payload := make(map[string]any, len(claims)+2)
	for k, v := range claims {
		payload[k] = v
	}
	payload["iat"] = now.Unix()
	payload["exp"] = now.Add(j.expiry).Unix()

	header := jwtB64(`{"alg":"HS256","typ":"JWT"}`)
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	unsigned := header + "." + base64.RawURLEncoding.EncodeToString(body)
	sig := j.hmacSign(unsigned)
	return unsigned + "." + sig, nil
}

// SignPair creates an access token + opaque refresh token pair.
// The refresh token is a random 32-byte hex string stored alongside the
// access token's "sub" claim — your application must persist it (e.g., in
// a refresh_tokens table or Redis) and pass it to Refresh to rotate.
//
//	pair, _ := mgr.SignPair(map[string]any{"sub": "42"})
//	// store pair.RefreshToken in DB associated with user "42"
func (j *JWTManager) SignPair(claims map[string]any) (TokenPair, error) {
	access, err := j.Sign(claims)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := j.generateRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(j.expiry.Seconds()),
	}, nil
}

// Refresh issues a new access token using the given claims (retrieved from your
// storage layer after validating the refresh token). Rotate the refresh token
// in your storage at the same time to prevent reuse.
//
//	// In your refresh endpoint:
//	storedClaims := lookupRefreshToken(refreshToken) // your code
//	newPair, err := mgr.SignPair(storedClaims)
func (j *JWTManager) Refresh(claims map[string]any) (string, error) {
	return j.Sign(claims)
}

func (j *JWTManager) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Verify parses and validates a JWT. Returns the claims on success.
// Returns ErrJWTMalformed, ErrJWTInvalid, or ErrJWTExpired on failure.
func (j *JWTManager) Verify(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrJWTMalformed
	}
	unsigned := parts[0] + "." + parts[1]
	expected := j.hmacSign(unsigned)
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return nil, ErrJWTInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrJWTMalformed
	}
	var claims map[string]any
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, ErrJWTMalformed
	}
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, ErrJWTExpired
		}
	}
	return claims, nil
}

// Middleware returns a handler middleware that validates the Bearer token in
// the Authorization header and stores the parsed claims under claimsKey.
// If claimsKey is empty it defaults to "jwt_claims".
func (j *JWTManager) Middleware(claimsKey string) gomvchttp.Middleware {
	key := claimsKey
	if key == "" {
		key = "jwt_claims"
	}
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			raw := c.GetHeader("Authorization")
			if !strings.HasPrefix(raw, "Bearer ") {
				c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
				return
			}
			claims, err := j.Verify(strings.TrimPrefix(raw, "Bearer "))
			if err != nil {
				msg := "invalid token"
				if errors.Is(err, ErrJWTExpired) {
					msg = "token expired"
				}
				c.JSON(http.StatusUnauthorized, map[string]string{"error": msg})
				return
			}
			c.Set(key, claims)
			next(c)
		}
	}
}

// Claims is a convenience helper that retrieves JWT claims from the context.
// Returns nil if the middleware has not run or the token was invalid.
func Claims(c *gomvchttp.Context, key ...string) map[string]any {
	k := "jwt_claims"
	if len(key) > 0 && key[0] != "" {
		k = key[0]
	}
	v, _ := c.Get(k)
	claims, _ := v.(map[string]any)
	return claims
}

func (j *JWTManager) hmacSign(msg string) string {
	h := hmac.New(sha256.New, j.secret)
	h.Write([]byte(msg))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func jwtB64(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}
