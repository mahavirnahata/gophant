package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// TxQuery is like Query but executes against a *sql.Tx (transaction).
// Obtain one with db.TxTable(tx, "table").
type TxQuery struct {
	tx         *sql.Tx
	db         *DB
	table      string
	selectCols []string
	joins      []join
	wheres     []condition
	orderBy    string
	limit      int
	offset     int
}

func (q *TxQuery) Where(col, op string, val any) *TxQuery {
	q.wheres = append(q.wheres, condition{col: col, op: op, val: val})
	return q
}

func (q *TxQuery) OrWhere(col, op string, val any) *TxQuery {
	q.wheres = append(q.wheres, condition{col: col, op: op, val: val, or: true})
	return q
}

func (q *TxQuery) Select(cols ...string) *TxQuery {
	if len(cols) > 0 {
		q.selectCols = cols
	}
	return q
}

func (q *TxQuery) Limit(n int) *TxQuery  { q.limit = n; return q }
func (q *TxQuery) Offset(n int) *TxQuery { q.offset = n; return q }

func (q *TxQuery) WhereNull(col string) *TxQuery {
	q.wheres = append(q.wheres, condition{col: col, op: "IS NULL"})
	return q
}

func (q *TxQuery) WhereNotNull(col string) *TxQuery {
	q.wheres = append(q.wheres, condition{col: col, op: "IS NOT NULL"})
	return q
}

func (q *TxQuery) WhereBetween(col string, min, max any) *TxQuery {
	q.wheres = append(q.wheres, condition{col: col, op: "BETWEEN", val: [2]any{min, max}})
	return q
}

func (q *TxQuery) WhereIn(col string, vals []any) *TxQuery {
	q.wheres = append(q.wheres, condition{col: col, op: "IN", val: vals})
	return q
}

func (q *TxQuery) OrderBy(order string) *TxQuery {
	q.orderBy = order
	return q
}

func (q *TxQuery) Latest(col ...string) *TxQuery {
	c := "created_at"
	if len(col) > 0 && col[0] != "" {
		c = col[0]
	}
	q.orderBy = c + " DESC"
	return q
}

func (q *TxQuery) Oldest(col ...string) *TxQuery {
	c := "created_at"
	if len(col) > 0 && col[0] != "" {
		c = col[0]
	}
	q.orderBy = c + " ASC"
	return q
}

func (q *TxQuery) Pluck(col string) ([]string, error) {
	q.selectCols = []string{col}
	rows, err := q.Get()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if v, ok := row[col]; ok && v != nil {
			out = append(out, fmt.Sprintf("%v", v))
		}
	}
	return out, nil
}

func (q *TxQuery) OrderBySafe(column, direction string, allowed []string) *TxQuery {
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

func (q *TxQuery) Get() ([]map[string]any, error) {
	query, args := q.buildSelect()
	rows, err := q.tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (q *TxQuery) First() (map[string]any, error) {
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

func (q *TxQuery) Insert(data map[string]any) (sql.Result, error) {
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
	return q.tx.Exec(query, vals...)
}

func (q *TxQuery) Update(data map[string]any) (sql.Result, error) {
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
	return q.tx.Exec(strings.TrimSpace(query), vals...)
}

func (q *TxQuery) Delete() (sql.Result, error) {
	whereSQL, whereArgs := q.buildWhere(1)
	query := fmt.Sprintf("DELETE FROM %s %s", q.table, whereSQL)
	return q.tx.Exec(strings.TrimSpace(query), whereArgs...)
}

func (q *TxQuery) buildSelect() (string, []any) {
	cols := strings.Join(q.selectCols, ",")
	whereSQL, args := q.buildWhere(1)
	query := fmt.Sprintf("SELECT %s FROM %s %s", cols, q.table, whereSQL)
	if q.orderBy != "" {
		query += " ORDER BY " + q.orderBy
	}
	if q.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	if q.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", q.offset)
	}
	return strings.TrimSpace(query), args
}

func (q *TxQuery) buildWhere(startIdx int) (string, []any) {
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
		case "BETWEEN":
			pair := w.val.([2]any)
			p1 := q.db.Dialect.Placeholder(idx)
			p2 := q.db.Dialect.Placeholder(idx + 1)
			sb.WriteString(fmt.Sprintf("%s BETWEEN %s AND %s", w.col, p1, p2))
			args = append(args, pair[0], pair[1])
			idx += 2
		case "IN":
			list, ok := w.val.([]any)
			if !ok || len(list) == 0 {
				sb.WriteString("1=0")
			} else {
				placeholders := make([]string, len(list))
				for i := range list {
					placeholders[i] = q.db.Dialect.Placeholder(idx)
					idx++
				}
				sb.WriteString(fmt.Sprintf("%s IN (%s)", w.col, strings.Join(placeholders, ",")))
				args = append(args, list...)
			}
		default:
			sb.WriteString(fmt.Sprintf("%s %s %s", w.col, w.op, q.db.Dialect.Placeholder(idx)))
			args = append(args, w.val)
			idx++
		}
	}
	return sb.String(), args
}
