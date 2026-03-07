package schema

import (
	"fmt"
	"strings"
)

type Column struct {
	Name     string
	Type     string
	NotNull  bool
	Nullable bool
	Default  string
}

type Blueprint struct {
	table       string
	columns     []Column
	primary     string
	driver      string
	indexes     []string
	uniques     []string
	foreignKeys []string
}

type Builder struct {
	Driver string
}

func New(driver string) *Builder {
	return &Builder{Driver: strings.ToLower(driver)}
}

func (b *Builder) Create(table string, fn func(*Blueprint)) string {
	_, sql := b.Build(table, fn)
	return sql
}

func (b *Builder) Drop(table string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
}

func (b *Builder) Build(table string, fn func(*Blueprint)) (*Blueprint, string) {
	bp := &Blueprint{table: table, driver: b.Driver}
	fn(bp)
	cols := []string{}
	for _, c := range bp.columns {
		s := c.Name + " " + c.Type
		if c.NotNull {
			s += " NOT NULL"
		}
		if c.Nullable {
			s += " NULL"
		}
		if c.Default != "" {
			s += " DEFAULT " + c.Default
		}
		cols = append(cols, s)
	}
	if bp.primary != "" {
		cols = append(cols, "PRIMARY KEY ("+bp.primary+")")
	}
	for _, u := range bp.uniques {
		cols = append(cols, u)
	}
	for _, fk := range bp.foreignKeys {
		cols = append(cols, fk)
	}
	return bp, fmt.Sprintf("CREATE TABLE %s (%s)", table, strings.Join(cols, ", "))
}

func (bp *Blueprint) Increments(name string) {
	if name == "" {
		name = "id"
	}
	switch bp.driver {
	case "postgres", "postgresql":
		bp.columns = append(bp.columns, Column{Name: name, Type: "BIGSERIAL", NotNull: true})
	default:
		bp.columns = append(bp.columns, Column{Name: name, Type: "BIGINT AUTO_INCREMENT", NotNull: true})
	}
	bp.primary = name
}

func (bp *Blueprint) String(name string, size int) {
	if size <= 0 {
		size = 255
	}
	bp.columns = append(bp.columns, Column{Name: name, Type: fmt.Sprintf("VARCHAR(%d)", size)})
}

func (bp *Blueprint) Integer(name string) {
	bp.columns = append(bp.columns, Column{Name: name, Type: "INT"})
}

func (bp *Blueprint) Boolean(name string) {
	bp.columns = append(bp.columns, Column{Name: name, Type: "BOOLEAN"})
}

func (bp *Blueprint) Timestamp(name string) {
	bp.columns = append(bp.columns, Column{Name: name, Type: "TIMESTAMP"})
}

func (bp *Blueprint) Timestamps() {
	bp.Timestamp("created_at")
	bp.Timestamp("updated_at")
}

func (bp *Blueprint) Nullable(name string) {
	for i := range bp.columns {
		if bp.columns[i].Name == name {
			bp.columns[i].Nullable = true
			bp.columns[i].NotNull = false
			return
		}
	}
}

func (bp *Blueprint) Default(name, value string) {
	for i := range bp.columns {
		if bp.columns[i].Name == name {
			bp.columns[i].Default = value
			return
		}
	}
}

func (bp *Blueprint) Unique(column string, name ...string) {
	n := ""
	if len(name) > 0 && name[0] != "" {
		n = name[0]
	} else {
		n = "uniq_" + column
	}
	bp.uniques = append(bp.uniques, fmt.Sprintf("CONSTRAINT %s UNIQUE (%s)", n, column))
}

func (bp *Blueprint) Index(column string, name ...string) {
	n := ""
	if len(name) > 0 && name[0] != "" {
		n = name[0]
	} else {
		n = "idx_" + column
	}
	bp.indexes = append(bp.indexes, fmt.Sprintf("CREATE INDEX %s ON %s (%s)", n, bp.table, column))
}

func (bp *Blueprint) Foreign(column, refTable, refColumn string, onDelete string) {
	constraint := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", column, refTable, refColumn)
	if onDelete != "" {
		constraint += " ON DELETE " + onDelete
	}
	bp.foreignKeys = append(bp.foreignKeys, constraint)
}

func (b *Builder) Indexes(bp *Blueprint) []string {
	return bp.indexes
}
