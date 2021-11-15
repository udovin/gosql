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

// BeginTxOption represents option for BeginTx.
type BeginTxOption func(ctx *context.Context, opts **sql.TxOptions)

// WithContext represents context option for BeginTx.
func WithContext(ctx context.Context) BeginTxOption {
	return func(txCtx *context.Context, _ **sql.TxOptions) {
		*txCtx = ctx
	}
}

// WithTxOptions represents TxOptions option for BeginTx.
func WithTxOptions(opts *sql.TxOptions) BeginTxOption {
	return func(_ *context.Context, txOpts **sql.TxOptions) {
		*txOpts = opts
	}
}

// WithTx represents wrapper for code that should use transaction.
func WithTx(
	b TxBeginner, fn func(tx *sql.Tx) error, options ...BeginTxOption,
) error {
	var ctx context.Context
	var opts *sql.TxOptions
	for _, option := range options {
		option(&ctx, &opts)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	tx, err := b.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	rollback := true
	defer func() {
		if rollback {
			// Try to rollback transaction on error or panic.
			_ = tx.Rollback()
		}
	}()
	if err := fn(tx); err != nil {
		return err
	}
	rollback = false
	return tx.Commit()
}

// WithEnsuredTx ensures that code uses sql.Tx.
func WithEnsuredTx(
	tx WeakTx, fn func(tx *sql.Tx) error, options ...BeginTxOption,
) error {
	switch v := tx.(type) {
	case *sql.Tx:
		return fn(v)
	case TxBeginner:
		return WithTx(v, fn, options...)
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

// BeginTx starts new transaction.
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if opts != nil && opts.ReadOnly {
		return d.RO.BeginTx(ctx, opts)
	}
	return d.DB.BeginTx(ctx, opts)
}

// Test *DB for interfaces.
var (
	_ WeakTx     = &DB{}
	_ TxBeginner = &DB{}
)
