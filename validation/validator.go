package validation

import (
	"database/sql"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/mahavirnahata/gophant/db"
)

type Rule func(string) string
type RuleWith func(string, *Validator) string

type Validator struct {
	errors   map[string][]string
	data     map[string]string
	messages map[string]string // custom overrides keyed by "field.rule" or "rule"
}

// New creates a Validator from an HTTP request (form + query values).
func New(r *http.Request) *Validator {
	_ = r.ParseForm()
	data := map[string]string{}
	for k, v := range r.Form {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}
	return &Validator{errors: map[string][]string{}, data: data, messages: map[string]string{}}
}

// NewFromMap creates a Validator from an arbitrary string map (useful for validating
// decoded JSON or other non-request data).
func NewFromMap(data map[string]string) *Validator {
	return &Validator{errors: map[string][]string{}, data: data, messages: map[string]string{}}
}

// WithMessages sets custom error messages. Keys can be "field.rule" (e.g., "email.required")
// for field-specific overrides or just "rule" (e.g., "required") for global overrides.
// Supports :field, :value placeholders.
func (v *Validator) WithMessages(msgs map[string]string) *Validator {
	for k, msg := range msgs {
		v.messages[k] = msg
	}
	return v
}

// Field validates a single field against one or more rules.
func (v *Validator) Field(name string, rules ...Rule) *Validator {
	val := v.data[name]
	for _, rule := range rules {
		if code := rule(val); code != "" {
			v.errors[name] = append(v.errors[name], v.formatError(name, code))
		}
	}
	return v
}

// FieldWith validates a field against rules that have access to the whole Validator
// (e.g., Confirmed, Unique).
func (v *Validator) FieldWith(name string, rules ...RuleWith) *Validator {
	val := v.data[name]
	for _, rule := range rules {
		if code := rule(val, v); code != "" {
			v.errors[name] = append(v.errors[name], v.formatError(name, code))
		}
	}
	return v
}

// Fails reports whether any validation errors were collected.
func (v *Validator) Fails() bool {
	return len(v.errors) > 0
}

// Passes is the inverse of Fails.
func (v *Validator) Passes() bool {
	return !v.Fails()
}

// Errors returns the full map of field → []message errors.
func (v *Validator) Errors() map[string][]string {
	return v.errors
}

// First returns the first error message for a field, or empty string.
func (v *Validator) First(field string) string {
	if errs, ok := v.errors[field]; ok && len(errs) > 0 {
		return errs[0]
	}
	return ""
}

// Value returns the validated value for a field.
func (v *Validator) Value(field string) string {
	return v.data[field]
}

// ── Message formatting ───────────────────────────────────────────────────────

func (v *Validator) formatError(field, code string) string {
	parts := strings.SplitN(code, ":", 2)
	rule := parts[0]
	param := ""
	if len(parts) == 2 {
		param = parts[1]
	}

	// 1. "field.rule" custom override (most specific)
	if msg, ok := v.messages[field+"."+rule]; ok {
		return strings.ReplaceAll(msg, ":field", field)
	}
	// 2. "rule" custom override
	if msg, ok := v.messages[rule]; ok {
		msg = strings.ReplaceAll(msg, ":field", field)
		msg = strings.ReplaceAll(msg, ":value", param)
		return msg
	}
	// 3. Built-in default
	return buildDefault(field, rule, param)
}

var defaultMessages = map[string]string{
	"required":  "The :field field is required.",
	"email":     "The :field field must be a valid email address.",
	"min":       "The :field field must be at least :value characters.",
	"max":       "The :field field must not exceed :value characters.",
	"numeric":   "The :field field must be a number.",
	"alpha":     "The :field field must only contain letters.",
	"in":        "The selected :field is invalid.",
	"regex":     "The :field field format is invalid.",
	"confirmed": "The :field confirmation does not match.",
	"unique":    "The :field has already been taken.",
	"url":       "The :field field must be a valid URL.",
	"uuid":      "The :field field must be a valid UUID.",
	"json":      "The :field field must be valid JSON.",
	"boolean":   "The :field field must be true or false.",
	"date":      "The :field field must be a valid date.",
}

func buildDefault(field, rule, param string) string {
	msg, ok := defaultMessages[rule]
	if !ok {
		return "The " + field + " field is invalid."
	}
	msg = strings.ReplaceAll(msg, ":field", field)
	msg = strings.ReplaceAll(msg, ":value", param)
	return msg
}

// ── Built-in rules ───────────────────────────────────────────────────────────

func Required() Rule {
	return func(val string) string {
		if strings.TrimSpace(val) == "" {
			return "required"
		}
		return ""
	}
}

func Email() Rule {
	re := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	return func(val string) string {
		if val == "" {
			return ""
		}
		if !re.MatchString(val) {
			return "email"
		}
		return ""
	}
}

func Min(n int) Rule {
	return func(val string) string {
		if len([]rune(val)) < n {
			return "min:" + strconv.Itoa(n)
		}
		return ""
	}
}

func Max(n int) Rule {
	return func(val string) string {
		if len([]rune(val)) > n {
			return "max:" + strconv.Itoa(n)
		}
		return ""
	}
}

func Numeric() Rule {
	return func(val string) string {
		if val == "" {
			return ""
		}
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			return "numeric"
		}
		return ""
	}
}

func Alpha() Rule {
	re := regexp.MustCompile(`^[A-Za-z]+$`)
	return func(val string) string {
		if val == "" {
			return ""
		}
		if !re.MatchString(val) {
			return "alpha"
		}
		return ""
	}
}

func AlphaNum() Rule {
	re := regexp.MustCompile(`^[A-Za-z0-9]+$`)
	return func(val string) string {
		if val == "" {
			return ""
		}
		if !re.MatchString(val) {
			return "alpha_num"
		}
		return ""
	}
}

func In(list ...string) Rule {
	set := map[string]bool{}
	for _, v := range list {
		set[v] = true
	}
	return func(val string) string {
		if val == "" {
			return ""
		}
		if !set[val] {
			return "in"
		}
		return ""
	}
}

func NotIn(list ...string) Rule {
	set := map[string]bool{}
	for _, v := range list {
		set[v] = true
	}
	return func(val string) string {
		if val == "" {
			return ""
		}
		if set[val] {
			return "in"
		}
		return ""
	}
}

func Regex(re *regexp.Regexp) Rule {
	return func(val string) string {
		if val == "" {
			return ""
		}
		if !re.MatchString(val) {
			return "regex"
		}
		return ""
	}
}

func URL() Rule {
	re := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return func(val string) string {
		if val == "" {
			return ""
		}
		if !re.MatchString(val) {
			return "url"
		}
		return ""
	}
}

func UUID() Rule {
	re := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return func(val string) string {
		if val == "" {
			return ""
		}
		if !re.MatchString(val) {
			return "uuid"
		}
		return ""
	}
}

func Boolean() Rule {
	return func(val string) string {
		if val == "" {
			return ""
		}
		lower := strings.ToLower(val)
		switch lower {
		case "true", "false", "1", "0", "yes", "no":
			return ""
		}
		return "boolean"
	}
}

func MinValue(n float64) Rule {
	return func(val string) string {
		if val == "" {
			return ""
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil || f < n {
			return "min:" + strconv.FormatFloat(n, 'f', -1, 64)
		}
		return ""
	}
}

func MaxValue(n float64) Rule {
	return func(val string) string {
		if val == "" {
			return ""
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil || f > n {
			return "max:" + strconv.FormatFloat(n, 'f', -1, 64)
		}
		return ""
	}
}

// Confirmed checks that the field matches another field named field+"_confirmation".
func Confirmed(field string) RuleWith {
	return func(val string, v *Validator) string {
		if val == "" {
			return ""
		}
		other := v.data[field]
		if val != other {
			return "confirmed"
		}
		return ""
	}
}

// Same checks that the field value equals the value of another field.
func Same(other string) RuleWith {
	return func(val string, v *Validator) string {
		if val != v.data[other] {
			return "confirmed"
		}
		return ""
	}
}

// Unique checks that the value does not already exist in the given column.
func Unique(conn *db.DB, table, column string) RuleWith {
	return func(val string, v *Validator) string {
		if val == "" || conn == nil {
			return ""
		}
		query := "SELECT 1 FROM " + table + " WHERE " + column + " = " + conn.Dialect.Placeholder(1) + " LIMIT 1"
		row := conn.Conn.QueryRow(query, val)
		var tmp int
		if err := row.Scan(&tmp); err != nil {
			if err == sql.ErrNoRows {
				return ""
			}
			return "unique"
		}
		return "unique"
	}
}
