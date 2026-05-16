package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type join struct {
	kind   string
	table  string
	first  string
	op     string
	second string
}

type condition struct {
	col string
	op  string
	val any
	or  bool
}

type Query struct {
	db         *DB
	table      string
	selectCols []string
	joins      []join
	wheres     []condition
	orderBy    string
	limit      int
	offset     int
}

type Page struct {
	Data    []map[string]any
	Page    int
	PerPage int
	Total   int
}

func (q *Query) Select(cols ...string) *Query {
	if len(cols) > 0 {
		q.selectCols = cols
	}
	return q
}

func (q *Query) SelectRaw(expr string) *Query {
	if expr != "" {
		q.selectCols = []string{expr}
	}
	return q
}

func (q *Query) Where(col, op string, val any) *Query {
	q.wheres = append(q.wheres, condition{col: col, op: op, val: val})
	return q
}

func (q *Query) OrWhere(col, op string, val any) *Query {
	q.wheres = append(q.wheres, condition{col: col, op: op, val: val, or: true})
	return q
}

func (q *Query) WhereLike(col string, val any) *Query {
	return q.Where(col, "LIKE", val)
}

func (q *Query) OrWhereLike(col string, val any) *Query {
	return q.OrWhere(col, "LIKE", val)
}

func (q *Query) WhereIn(col string, vals []any) *Query {
	if len(vals) == 0 {
		return q
	}
	q.wheres = append(q.wheres, condition{col: col, op: "IN", val: vals})
	return q
}

func (q *Query) OrWhereIn(col string, vals []any) *Query {
	if len(vals) == 0 {
		return q
	}
	q.wheres = append(q.wheres, condition{col: col, op: "IN", val: vals, or: true})
	return q
}

// WhereNull adds a WHERE col IS NULL clause.
func (q *Query) WhereNull(col string) *Query {
	q.wheres = append(q.wheres, condition{col: col, op: "IS NULL"})
	return q
}

// WhereNotNull adds a WHERE col IS NOT NULL clause.
func (q *Query) WhereNotNull(col string) *Query {
	q.wheres = append(q.wheres, condition{col: col, op: "IS NOT NULL"})
	return q
}

// WhereBetween adds a WHERE col BETWEEN min AND max clause.
func (q *Query) WhereBetween(col string, min, max any) *Query {
	q.wheres = append(q.wheres, condition{col: col, op: "BETWEEN", val: [2]any{min, max}})
	return q
}

func (q *Query) Join(table, first, op, second string) *Query {
	q.joins = append(q.joins, join{kind: "INNER", table: table, first: first, op: op, second: second})
	return q
}

func (q *Query) LeftJoin(table, first, op, second string) *Query {
	q.joins = append(q.joins, join{kind: "LEFT", table: table, first: first, op: op, second: second})
	return q
}

func (q *Query) OrderBy(order string) *Query {
	q.orderBy = order
	return q
}

// Latest orders by col DESC (default: created_at). Equivalent to OrderBy("col DESC").
func (q *Query) Latest(col ...string) *Query {
	c := "created_at"
	if len(col) > 0 && col[0] != "" {
		c = col[0]
	}
	q.orderBy = c + " DESC"
	return q
}

// Oldest orders by col ASC (default: created_at).
func (q *Query) Oldest(col ...string) *Query {
	c := "created_at"
	if len(col) > 0 && col[0] != "" {
		c = col[0]
	}
	q.orderBy = c + " ASC"
	return q
}

func (q *Query) OrderBySafe(column string, direction string, allowed []string) *Query {
	if !isAllowed(column, allowed) {
		return q
	}
	dir := strings.ToUpper(direction)
	if dir != "ASC" && dir != "DESC" {
		dir = "ASC"
	}
	q.orderBy = column + " " + dir
	return q
}

func (q *Query) Limit(n int) *Query {
	q.limit = n
	return q
}

func (q *Query) Offset(n int) *Query {
	q.offset = n
	return q
}

func (q *Query) Get() ([]map[string]any, error) {
	query, args := q.buildSelect()
	rows, err := q.db.Conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (q *Query) GetStructs(dest any) error {
	query, args := q.buildSelect()
	rows, err := q.db.Conn.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return scanRowsIntoStructs(rows, dest)
}

func (q *Query) First() (map[string]any, error) {
	q.limit = 1
	rows, err := q.Get()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	return rows[0], nil
}

func (q *Query) FirstStruct(dest any) error {
	q.limit = 1
	query, args := q.buildSelect()
	rows, err := q.db.Conn.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return scanRowIntoStruct(rows, dest)
}

func (q *Query) Count() (int, error) {
	query, args := q.buildCount()
	row := q.db.Conn.QueryRow(query, args...)
	var total int
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (q *Query) Paginate(page, perPage int) (Page, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}
	count, err := q.Count()
	if err != nil {
		return Page{}, err
	}
	q.limit = perPage
	q.offset = (page - 1) * perPage
	data, err := q.Get()
	if err != nil {
		return Page{}, err
	}
	return Page{Data: data, Page: page, PerPage: perPage, Total: count}, nil
}

// Pluck returns all values of a single column as a string slice.
func (q *Query) Pluck(col string) ([]string, error) {
	q.selectCols = []string{col}
	rows, err := q.Get()
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, fmt.Sprintf("%v", row[col]))
	}
	return result, nil
}

// Chunk processes rows in batches of size, calling fn for each batch.
// Stops early if fn returns an error.
func (q *Query) Chunk(size int, fn func([]map[string]any) error) error {
	offset := 0
	for {
		batch := *q
		batch.limit = size
		batch.offset = offset
		rows, err := batch.Get()
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		if err := fn(rows); err != nil {
			return err
		}
		if len(rows) < size {
			break
		}
		offset += size
	}
	return nil
}

func (q *Query) Insert(data map[string]any) (sql.Result, error) {
	cols := make([]string, 0, len(data))
	vals := make([]any, 0, len(data))
	placeholders := make([]string, 0, len(data))
	idx := 1
	for k, v := range data {
		cols = append(cols, k)
		vals = append(vals, v)
		placeholders = append(placeholders, q.db.Dialect.Placeholder(idx))
		idx++
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", q.table, strings.Join(cols, ","), strings.Join(placeholders, ","))
	return q.db.Conn.Exec(query, vals...)
}

func (q *Query) Update(data map[string]any) (sql.Result, error) {
	setParts := make([]string, 0, len(data))
	vals := make([]any, 0, len(data)+len(q.wheres))
	idx := 1
	for k, v := range data {
		setParts = append(setParts, fmt.Sprintf("%s = %s", k, q.db.Dialect.Placeholder(idx)))
		vals = append(vals, v)
		idx++
	}

	whereSQL, whereArgs := q.buildWhere(idx)
	vals = append(vals, whereArgs...)
	query := fmt.Sprintf("UPDATE %s SET %s %s", q.table, strings.Join(setParts, ","), whereSQL)
	return q.db.Conn.Exec(strings.TrimSpace(query), vals...)
}

func (q *Query) Delete() (sql.Result, error) {
	whereSQL, whereArgs := q.buildWhere(1)
	query := fmt.Sprintf("DELETE FROM %s %s", q.table, whereSQL)
	return q.db.Conn.Exec(strings.TrimSpace(query), whereArgs...)
}

func (q *Query) buildSelect() (string, []any) {
	cols := strings.Join(q.selectCols, ",")
	joinSQL := q.buildJoins()
	whereSQL, args := q.buildWhere(1)
	query := fmt.Sprintf("SELECT %s FROM %s %s %s", cols, q.table, joinSQL, whereSQL)
	if q.orderBy != "" {
		query += " ORDER BY " + q.orderBy
	}
	if q.limit > 0 {
		query += " LIMIT " + fmt.Sprintf("%d", q.limit)
	}
	if q.offset > 0 {
		query += " OFFSET " + fmt.Sprintf("%d", q.offset)
	}
	return strings.TrimSpace(query), args
}

func (q *Query) buildCount() (string, []any) {
	joinSQL := q.buildJoins()
	whereSQL, args := q.buildWhere(1)
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s %s %s", q.table, joinSQL, whereSQL)
	return strings.TrimSpace(query), args
}

func (q *Query) buildJoins() string {
	if len(q.joins) == 0 {
		return ""
	}
	parts := make([]string, 0, len(q.joins))
	for _, j := range q.joins {
		parts = append(parts, fmt.Sprintf("%s JOIN %s ON %s %s %s", j.kind, j.table, j.first, j.op, j.second))
	}
	return strings.Join(parts, " ")
}

func (q *Query) buildWhere(startIdx int) (string, []any) {
	if len(q.wheres) == 0 {
		return "", nil
	}
	args := make([]any, 0, len(q.wheres))
	idx := startIdx
	var sb strings.Builder
	sb.WriteString("WHERE ")
	for i, w := range q.wheres {
		if i > 0 {
			if w.or {
				sb.WriteString(" OR ")
			} else {
				sb.WriteString(" AND ")
			}
		}

		switch w.op {
		case "IS NULL", "IS NOT NULL":
			sb.WriteString(fmt.Sprintf("%s %s", w.col, w.op))
		case "IN":
			list, ok := w.val.([]any)
			if !ok || len(list) == 0 {
				sb.WriteString("1=0")
				continue
			}
			placeholders := make([]string, len(list))
			for i := range list {
				placeholders[i] = q.db.Dialect.Placeholder(idx)
				idx++
			}
			sb.WriteString(fmt.Sprintf("%s IN (%s)", w.col, strings.Join(placeholders, ",")))
			args = append(args, list...)
		case "BETWEEN":
			pair, ok := w.val.([2]any)
			if !ok {
				continue
			}
			sb.WriteString(fmt.Sprintf("%s BETWEEN %s AND %s", w.col, q.db.Dialect.Placeholder(idx), q.db.Dialect.Placeholder(idx+1)))
			args = append(args, pair[0], pair[1])
			idx += 2
		default:
			sb.WriteString(fmt.Sprintf("%s %s %s", w.col, w.op, q.db.Dialect.Placeholder(idx)))
			args = append(args, w.val)
			idx++
		}
	}
	return sb.String(), args
}

func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := []map[string]any{}
	for rows.Next() {
		values := make([]any, len(cols))
		pointers := make([]any, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		row := map[string]any{}
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func isAllowed(val string, allowed []string) bool {
	for _, a := range allowed {
		if val == a {
			return true
		}
	}
	return false
}
