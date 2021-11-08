package gosql

import (
	"context"
	"database/sql"
	"fmt"
)

// WeakTx represents SQL interface like sql.DB or sql.Tx.
type WeakTx interface {
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

// WithTx represents wrapper for code that should use transaction.
func WithTx(b TxBeginner, fn func(tx *sql.Tx) error) error {
	tx, err := b.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			// Try to rollback transaction on panic.
			_ = tx.Rollback()
			panic(r)
		}
	}()
	if err := fn(tx); err != nil {
		// Try to rollback transaction on error.
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// WithEnsuredTx ensures that code uses sql.Tx.
func WithEnsuredTx(tx WeakTx, fn func(tx *sql.Tx) error) error {
	switch v := tx.(type) {
	case *sql.Tx:
		return fn(v)
	case TxBeginner:
		return WithTx(v, fn)
	default:
		panic(fmt.Errorf("unsupported type: %T", v))
	}
}

// DB represents wrapper for sql.DB with additional builder and
// read-only connection.
type DB struct {
	// Read-write connection.
	*sql.DB
	// Read-only connection.
	RO *sql.DB
	// Builder contains builder for specified database dialect.
	Builder
}

// Test *DB for interfaces.
var (
	_ WeakTx     = &DB{}
	_ TxBeginner = &DB{}
)
