package db

import (
	"context"
	"database/sql"
)

// DB wraps *sql.DB with a dialect for placeholder generation.
type DB struct {
	Conn    *sql.DB
	Dialect Dialect
}

// DefaultDB is used by Model when no explicit DB is provided.
var DefaultDB *DB

// Open opens a database connection. Blank-import the driver before calling this.
//
//	import _ "github.com/go-sql-driver/mysql"
//	conn, err := db.Open("mysql", dsn, db.QuestionDialect{})
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

// Table returns a new Query builder for the given table.
func (db *DB) Table(name string) *Query {
	return &Query{db: db, table: name, selectCols: []string{"*"}}
}

// TxTable returns a Query builder that runs against an existing transaction.
// Use inside Transaction() to keep queries within the same tx.
func (db *DB) TxTable(tx *sql.Tx, name string) *TxQuery {
	return &TxQuery{tx: tx, db: db, table: name, selectCols: []string{"*"}}
}

// SetDefaultDB sets the package-level default DB used by Model.
func SetDefaultDB(db *DB) {
	DefaultDB = db
}

// Transaction runs fn inside a database transaction. If fn returns an error
// the transaction is rolled back automatically; otherwise it is committed.
//
//	err := app.DB.Transaction(func(tx *sql.Tx) error {
//	    _, err := app.DB.TxTable(tx, "orders").Insert(order)
//	    return err
//	})
func (db *DB) Transaction(fn func(tx *sql.Tx) error) error {
	return db.TransactionContext(context.Background(), nil, fn)
}

// TransactionContext is like Transaction but accepts a context and tx options.
func (db *DB) TransactionContext(ctx context.Context, opts *sql.TxOptions, fn func(tx *sql.Tx) error) error {
	tx, err := db.Conn.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
