package validation

import (
	"net/http/httptest"
	"testing"
)

// ── Bail and Sometimes ───────────────────────────────────────────────────────

func TestBailStopsAfterFirstError(t *testing.T) {
	v := NewFromMap(map[string]string{"name": ""}).
		Field("name", Bail(), Required(), Min(3))

	errs := v.Errors()["name"]
	if len(errs) != 1 {
		t.Fatalf("bail: expected 1 error, got %d", len(errs))
	}
}

func TestSometimesSkipsAbsentField(t *testing.T) {
	v := NewFromMap(map[string]string{}).
		Field("phone", Sometimes(), Required(), Min(10))

	if v.Fails() {
		t.Fatalf("sometimes: expected no errors when field absent, got: %v", v.Errors())
	}
}

func TestSometimesValidatesWhenPresent(t *testing.T) {
	v := NewFromMap(map[string]string{"phone": "123"}).
		Field("phone", Sometimes(), Min(10))

	if !v.Fails() {
		t.Fatal("sometimes: expected error when field present but too short")
	}
}

// ── Date / Before / After ────────────────────────────────────────────────────

func TestDateValid(t *testing.T) {
	v := NewFromMap(map[string]string{"dob": "1990-01-15"}).
		Field("dob", Date())
	if v.Fails() {
		t.Fatalf("expected valid date to pass: %v", v.Errors())
	}
}

func TestDateInvalid(t *testing.T) {
	v := NewFromMap(map[string]string{"dob": "not-a-date"}).
		Field("dob", Date())
	if !v.Fails() {
		t.Fatal("expected invalid date to fail")
	}
}

func TestDateEmptySkipped(t *testing.T) {
	v := NewFromMap(map[string]string{"dob": ""}).
		Field("dob", Date())
	if v.Fails() {
		t.Fatal("empty value should skip date rule")
	}
}

func TestBefore(t *testing.T) {
	v := NewFromMap(map[string]string{"start": "2020-01-01"}).
		Field("start", Before("2025-01-01"))
	if v.Fails() {
		t.Fatalf("expected 2020 to be before 2025: %v", v.Errors())
	}

	v2 := NewFromMap(map[string]string{"start": "2030-01-01"}).
		Field("start", Before("2025-01-01"))
	if !v2.Fails() {
		t.Fatal("expected 2030 to fail before(2025)")
	}
}

func TestAfter(t *testing.T) {
	v := NewFromMap(map[string]string{"end": "2030-06-01"}).
		Field("end", After("2025-01-01"))
	if v.Fails() {
		t.Fatalf("expected 2030 to be after 2025: %v", v.Errors())
	}

	v2 := NewFromMap(map[string]string{"end": "2020-01-01"}).
		Field("end", After("2025-01-01"))
	if !v2.Fails() {
		t.Fatal("expected 2020 to fail after(2025)")
	}
}

// ── Between ──────────────────────────────────────────────────────────────────

func TestBetween(t *testing.T) {
	v := NewFromMap(map[string]string{"age": "25"}).
		Field("age", Between(18, 65))
	if v.Fails() {
		t.Fatalf("expected 25 to be between 18 and 65: %v", v.Errors())
	}

	v2 := NewFromMap(map[string]string{"age": "10"}).
		Field("age", Between(18, 65))
	if !v2.Fails() {
		t.Fatal("expected 10 to fail between(18, 65)")
	}

	v3 := NewFromMap(map[string]string{"age": "70"}).
		Field("age", Between(18, 65))
	if !v3.Fails() {
		t.Fatal("expected 70 to fail between(18, 65)")
	}
}

// ── Different ────────────────────────────────────────────────────────────────

func TestDifferent(t *testing.T) {
	v := NewFromMap(map[string]string{
		"old_password": "abc",
		"new_password": "xyz",
	}).FieldWith("new_password", Different("old_password"))

	if v.Fails() {
		t.Fatalf("different values should pass: %v", v.Errors())
	}

	v2 := NewFromMap(map[string]string{
		"old_password": "same",
		"new_password": "same",
	}).FieldWith("new_password", Different("old_password"))

	if !v2.Fails() {
		t.Fatal("same values should fail Different check")
	}
}

// ── Human-readable messages ──────────────────────────────────────────────────

func TestHumanReadableMessages(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	v := New(req).
		Field("email", Required()).
		Field("password", Required(), Min(8))

	errs := v.Errors()
	emailErrs := errs["email"]
	if len(emailErrs) == 0 || emailErrs[0] != "The email field is required." {
		t.Fatalf("unexpected email error: %v", emailErrs)
	}
}

func TestCustomMessages(t *testing.T) {
	v := NewFromMap(map[string]string{}).
		WithMessages(map[string]string{
			"email.required": "We need your email!",
		}).
		Field("email", Required())

	if v.First("email") != "We need your email!" {
		t.Fatalf("expected custom message, got: %s", v.First("email"))
	}
}

// ── AlphaNum ─────────────────────────────────────────────────────────────────

func TestAlphaNum(t *testing.T) {
	valid := NewFromMap(map[string]string{"user": "abc123"}).Field("user", AlphaNum())
	if valid.Fails() {
		t.Fatal("alphanumeric should pass")
	}

	invalid := NewFromMap(map[string]string{"user": "abc 123!"}).Field("user", AlphaNum())
	if !invalid.Fails() {
		t.Fatal("non-alphanumeric should fail")
	}
}

// ── URL and UUID ─────────────────────────────────────────────────────────────

func TestURL(t *testing.T) {
	valid := NewFromMap(map[string]string{"site": "https://example.com"}).Field("site", URL())
	if valid.Fails() {
		t.Fatal("valid URL should pass")
	}

	invalid := NewFromMap(map[string]string{"site": "not-a-url"}).Field("site", URL())
	if !invalid.Fails() {
		t.Fatal("invalid URL should fail")
	}
}

func TestUUID(t *testing.T) {
	valid := NewFromMap(map[string]string{"id": "550e8400-e29b-41d4-a716-446655440000"}).Field("id", UUID())
	if valid.Fails() {
		t.Fatal("valid UUID should pass")
	}

	invalid := NewFromMap(map[string]string{"id": "not-a-uuid"}).Field("id", UUID())
	if !invalid.Fails() {
		t.Fatal("invalid UUID should fail")
	}
}

// ── MinValue / MaxValue ───────────────────────────────────────────────────────

func TestMinMaxValue(t *testing.T) {
	v := NewFromMap(map[string]string{"price": "5.0"}).Field("price", MinValue(10))
	if !v.Fails() {
		t.Fatal("5.0 < 10 should fail MinValue(10)")
	}

	v2 := NewFromMap(map[string]string{"price": "15.0"}).Field("price", MaxValue(10))
	if !v2.Fails() {
		t.Fatal("15.0 > 10 should fail MaxValue(10)")
	}

	v3 := NewFromMap(map[string]string{"price": "10.0"}).
		Field("price", MinValue(5), MaxValue(20))
	if v3.Fails() {
		t.Fatal("10.0 is within [5,20] and should pass")
	}
}

// ── NotIn ─────────────────────────────────────────────────────────────────────

func TestNotIn(t *testing.T) {
	v := NewFromMap(map[string]string{"role": "admin"}).Field("role", NotIn("admin", "root"))
	if !v.Fails() {
		t.Fatal("'admin' in list should fail NotIn")
	}

	v2 := NewFromMap(map[string]string{"role": "user"}).Field("role", NotIn("admin", "root"))
	if v2.Fails() {
		t.Fatal("'user' not in list should pass NotIn")
	}
}

// ── Passes and Value ─────────────────────────────────────────────────────────

func TestPassesAndValue(t *testing.T) {
	v := NewFromMap(map[string]string{"name": "Alice"}).Field("name", Required())
	if !v.Passes() {
		t.Fatal("expected Passes() to be true")
	}
	if v.Value("name") != "Alice" {
		t.Fatalf("expected Value(name) = Alice, got %q", v.Value("name"))
	}
}
