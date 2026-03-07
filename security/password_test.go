package security

import "testing"

func TestPasswordHash(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if !CheckPassword(hash, "secret") {
		t.Fatalf("expected password to match")
	}
}
