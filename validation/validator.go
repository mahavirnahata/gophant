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
	errors map[string][]string
	data   map[string]string
}

func New(r *http.Request) *Validator {
	_ = r.ParseForm()
	data := map[string]string{}
	for k, v := range r.Form {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}
	return &Validator{errors: map[string][]string{}, data: data}
}

func (v *Validator) Field(name string, rules ...Rule) *Validator {
	val := v.data[name]
	for _, rule := range rules {
		if msg := rule(val); msg != "" {
			v.errors[name] = append(v.errors[name], msg)
		}
	}
	return v
}

func (v *Validator) FieldWith(name string, rules ...RuleWith) *Validator {
	val := v.data[name]
	for _, rule := range rules {
		if msg := rule(val, v); msg != "" {
			v.errors[name] = append(v.errors[name], msg)
		}
	}
	return v
}

func (v *Validator) Fails() bool {
	return len(v.errors) > 0
}

func (v *Validator) Errors() map[string][]string {
	return v.errors
}

func (v *Validator) First(field string) string {
	if errs, ok := v.errors[field]; ok && len(errs) > 0 {
		return errs[0]
	}
	return ""
}

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
		if len(val) < n {
			return "min:" + strconv.Itoa(n)
		}
		return ""
	}
}

func Max(n int) Rule {
	return func(val string) string {
		if len(val) > n {
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
