package db

import "fmt"

func normalizeKey(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case int:
		return fmt.Sprintf("%d", t)
	case int8:
		return fmt.Sprintf("%d", t)
	case int16:
		return fmt.Sprintf("%d", t)
	case int32:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case uint:
		return fmt.Sprintf("%d", t)
	case uint8:
		return fmt.Sprintf("%d", t)
	case uint16:
		return fmt.Sprintf("%d", t)
	case uint32:
		return fmt.Sprintf("%d", t)
	case uint64:
		return fmt.Sprintf("%d", t)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// HasMany loads related rows and groups them by local key.
func HasMany(conn *DB, table, foreignKey string, localIDs []any) (map[any][]map[string]any, error) {
	if conn == nil || len(localIDs) == 0 {
		return map[any][]map[string]any{}, nil
	}
	idMap := map[string]any{}
	for _, id := range localIDs {
		idMap[normalizeKey(id)] = id
	}
	q := conn.Table(table).WhereIn(foreignKey, localIDs)
	rows, err := q.Get()
	if err != nil {
		return nil, err
	}
	out := map[any][]map[string]any{}
	for _, row := range rows {
		key := row[foreignKey]
		nk := normalizeKey(key)
		if orig, ok := idMap[nk]; ok {
			out[orig] = append(out[orig], row)
			continue
		}
		out[key] = append(out[key], row)
	}
	return out, nil
}

// BelongsTo loads parent rows by id.
func BelongsTo(conn *DB, table, ownerKey string, ids []any) (map[any]map[string]any, error) {
	if conn == nil || len(ids) == 0 {
		return map[any]map[string]any{}, nil
	}
	idMap := map[string]any{}
	for _, id := range ids {
		idMap[normalizeKey(id)] = id
	}
	q := conn.Table(table).WhereIn(ownerKey, ids)
	rows, err := q.Get()
	if err != nil {
		return nil, err
	}
	out := map[any]map[string]any{}
	for _, row := range rows {
		key := row[ownerKey]
		nk := normalizeKey(key)
		if orig, ok := idMap[nk]; ok {
			out[orig] = row
			continue
		}
		out[key] = row
	}
	return out, nil
}

// EagerLoadHasMany attaches child rows to each parent row under key.
func EagerLoadHasMany(parents []map[string]any, key string, foreignKey string, conn *DB, table string, localKey string) error {
	ids := make([]any, 0, len(parents))
	for _, p := range parents {
		if v, ok := p[localKey]; ok {
			ids = append(ids, v)
		}
	}
	children, err := HasMany(conn, table, foreignKey, ids)
	if err != nil {
		return err
	}
	for _, p := range parents {
		id := p[localKey]
		p[key] = children[id]
	}
	return nil
}

func (q *Query) DebugSQL() string {
	sql, args := q.buildSelect()
	return fmt.Sprintf("%s | args=%v", sql, args)
}
