package db

import "database/sql"

type DB struct {
	Conn    *sql.DB
	Dialect Dialect
}

var DefaultDB *DB

func Open(driver, dsn string, dialect Dialect) (*DB, error) {
	conn, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if dialect == nil {
		dialect = QuestionDialect{}
	}
	return &DB{Conn: conn, Dialect: dialect}, nil
}

func (db *DB) Table(name string) *Query {
	return &Query{db: db, table: name, selectCols: []string{"*"}}
}

func SetDefaultDB(db *DB) {
	DefaultDB = db
}
