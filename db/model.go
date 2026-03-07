package db

import "errors"

type Model struct {
	DB    *DB
	Table string
}

func NewModel(db *DB, table string) *Model {
	return &Model{DB: db, Table: table}
}

func (m *Model) Query() *Query {
	db := m.DB
	if db == nil {
		db = DefaultDB
	}
	if db == nil {
		panic(errors.New("db not configured: call db.SetDefaultDB or use NewModel(db, table)"))
	}
	return db.Table(m.Table)
}

func (m *Model) Where(col, op string, val any) *Query {
	return m.Query().Where(col, op, val)
}

func (m *Model) OrderBySafe(column string, direction string, allowed []string) *Query {
	return m.Query().OrderBySafe(column, direction, allowed)
}

func (m *Model) Get() ([]map[string]any, error) {
	return m.Query().Get()
}

func (m *Model) GetStructs(dest any) error {
	return m.Query().GetStructs(dest)
}

func (m *Model) First() (map[string]any, error) {
	return m.Query().First()
}

func (m *Model) FirstStruct(dest any) error {
	return m.Query().FirstStruct(dest)
}

func (m *Model) Insert(data map[string]any) error {
	_, err := m.Query().Insert(data)
	return err
}

func (m *Model) Update(whereCol string, whereVal any, data map[string]any) error {
	_, err := m.Query().Where(whereCol, "=", whereVal).Update(data)
	return err
}

func (m *Model) Delete(whereCol string, whereVal any) error {
	_, err := m.Query().Where(whereCol, "=", whereVal).Delete()
	return err
}
