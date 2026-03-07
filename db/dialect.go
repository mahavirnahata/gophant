package db

import "strconv"

type Dialect interface {
	Placeholder(n int) string
}

type QuestionDialect struct{}

func (d QuestionDialect) Placeholder(n int) string { return "?" }

type DollarDialect struct{}

func (d DollarDialect) Placeholder(n int) string { return "$" + strconv.Itoa(n) }
