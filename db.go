package gosql

import (
	"context"
	"database/sql"
)

// Runner represents SQL runner interface like sql.DB or sql.Tx.
type Runner interface {
	// Exec executes a query that doesn't return rows.
	Exec(query string, args ...interface{}) (sql.Result, error)
	// Query executes a query that returns rows, typically a SELECT.
	Query(query string, args ...interface{}) (*sql.Rows, error)
	// QueryRow executes a query that is expected to return at most one row.
	QueryRow(query string, args ...interface{}) *sql.Row
}

// TxBeginner represents object that can start new transaction like sql.DB.
type TxBeginner interface {
	// Begin starts a transaction.
	Begin() (*sql.Tx, error)
	// BeginTx starts a transaction.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
