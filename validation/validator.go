package validation

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mahavirnahata/gophant/db"
)

type Rule func(string) string
type RuleWith func(string, *Validator) string

// Rules maps field names to their validation rules for use with c.Validate().
type Rules map[string][]Rule

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
// Special rules: Bail() stops after the first error; Sometimes() skips if field is absent.
func (v *Validator) Field(name string, rules ...Rule) *Validator {
	_, present := v.data[name]
	val := v.data[name]
	bail := false

	for _, rule := range rules {
		code := rule(val)
		switch code {
		case "_bail_":
			bail = true
		case "_skip_":
			if !present {
				return v
			}
		case "":
			// no error
		default:
			v.errors[name] = append(v.errors[name], v.formatError(name, code))
			if bail {
				return v
			}
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
// Data returns all validated field values (the raw input map).
func (v *Validator) Data() map[string]string {
	out := make(map[string]string, len(v.data))
	for k, val := range v.data {
		out[k] = val
	}
	return out
}

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
	"required":   "The :field field is required.",
	"email":      "The :field field must be a valid email address.",
	"min":        "The :field field must be at least :value characters.",
	"max":        "The :field field must not exceed :value characters.",
	"numeric":    "The :field field must be a number.",
	"alpha":      "The :field field must only contain letters.",
	"alpha_num":  "The :field field must only contain letters and numbers.",
	"in":         "The selected :field is invalid.",
	"not_in":     "The selected :field is invalid.",
	"regex":      "The :field field format is invalid.",
	"confirmed":  "The :field confirmation does not match.",
	"different":  "The :field field and :value must be different.",
	"unique":     "The :field has already been taken.",
	"url":        "The :field field must be a valid URL.",
	"uuid":       "The :field field must be a valid UUID.",
	"json":       "The :field field must be valid JSON.",
	"boolean":    "The :field field must be true or false.",
	"date":       "The :field field must be a valid date.",
	"before":     "The :field must be a date before :value.",
	"after":      "The :field must be a date after :value.",
	"between":    "The :field must be between :value.",
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

// Different checks that the field value is different from another field.
func Different(other string) RuleWith {
	return func(val string, v *Validator) string {
		if val == v.data[other] {
			return "different:" + other
		}
		return ""
	}
}

// Bail signals Field() to stop validating after the first error for this field.
// Place it first in the rule list.
func Bail() Rule {
	return func(_ string) string { return "_bail_" }
}

// Sometimes causes Field() to skip all rules when the field is not present in the input.
// Place it first in the rule list.
func Sometimes() Rule {
	return func(_ string) string { return "_skip_" }
}

// commonLayouts are tried in order when no layout is specified for Date/Before/After.
var commonLayouts = []string{
	"2006-01-02",
	"2006-01-02 15:04:05",
	time.RFC3339,
	"01/02/2006",
	"02-01-2006",
}

func parseDate(val string, layouts []string) (time.Time, bool) {
	for _, l := range layouts {
		if t, err := time.Parse(l, val); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// Date validates that the value is a parseable date. Accepts common formats by default;
// pass an explicit Go time layout to restrict to that format.
func Date(layout ...string) Rule {
	layouts := commonLayouts
	if len(layout) > 0 {
		layouts = layout
	}
	return func(val string) string {
		if val == "" {
			return ""
		}
		if _, ok := parseDate(val, layouts); !ok {
			return "date"
		}
		return ""
	}
}

// Before validates that the date value is before ref.
// ref must be parseable by the same layouts.
func Before(ref string, layout ...string) Rule {
	layouts := commonLayouts
	if len(layout) > 0 {
		layouts = layout
	}
	refTime, _ := parseDate(ref, layouts)
	return func(val string) string {
		if val == "" {
			return ""
		}
		t, ok := parseDate(val, layouts)
		if !ok || !t.Before(refTime) {
			return "before:" + ref
		}
		return ""
	}
}

// After validates that the date value is after ref.
func After(ref string, layout ...string) Rule {
	layouts := commonLayouts
	if len(layout) > 0 {
		layouts = layout
	}
	refTime, _ := parseDate(ref, layouts)
	return func(val string) string {
		if val == "" {
			return ""
		}
		t, ok := parseDate(val, layouts)
		if !ok || !t.After(refTime) {
			return "after:" + ref
		}
		return ""
	}
}

// Between validates that the numeric value is between min and max (inclusive).
func Between(min, max float64) Rule {
	return func(val string) string {
		if val == "" {
			return ""
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil || f < min || f > max {
			return fmt.Sprintf("between:%v,%v", min, max)
		}
		return ""
	}
}

// NotIn validates that the value is not one of the listed options.
// (Distinct from the existing NotIn — this is the named version matching defaultMessages.)
func NotInList(list ...string) Rule {
	set := map[string]bool{}
	for _, v := range list {
		set[v] = true
	}
	return func(val string) string {
		if val == "" {
			return ""
		}
		if set[val] {
			return "not_in"
		}
		return ""
	}
}
