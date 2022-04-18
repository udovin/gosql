package gosql

import (
	"context"
	"database/sql"
)

// Runner represents SQL interface like sql.DB or sql.Tx.
type Runner interface {
	// ExecContext executes a query that doesn't return rows.
	ExecContext(
		ctx context.Context, query string, args ...any,
	) (sql.Result, error)
	// QueryContext executes a query that returns rows, typically a SELECT.
	QueryContext(
		ctx context.Context, query string, args ...any,
	) (*sql.Rows, error)
	// QueryRowContext executes a query that is expected to return at
	// most one row.
	QueryRowContext(
		ctx context.Context, query string, args ...any,
	) *sql.Row
}

// TxBeginner represents object that can start new transaction like sql.DB.
type TxBeginner interface {
	// BeginTx starts a transaction.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// BeginTxOption represents option for BeginTx.
type BeginTxOption func(opts **sql.TxOptions)

// WithTxOptions represents TxOptions option for BeginTx.
func WithTxOptions(opts *sql.TxOptions) BeginTxOption {
	return func(txOpts **sql.TxOptions) {
		*txOpts = opts
	}
}

// WithReadOnly represents readonly mode option for BeginTx.
func WithReadOnly(readOnly bool) BeginTxOption {
	return func(txOpts **sql.TxOptions) {
		if *txOpts == nil {
			*txOpts = &sql.TxOptions{}
		}
		(*txOpts).ReadOnly = readOnly
	}
}

// WithIsolation represents isolation level option for BeginTx.
func WithIsolation(level sql.IsolationLevel) BeginTxOption {
	return func(txOpts **sql.TxOptions) {
		if *txOpts == nil {
			*txOpts = &sql.TxOptions{}
		}
		(*txOpts).Isolation = level
	}
}

// WrapTx represents wrapper for code that should use transaction.
func WrapTx(
	ctx context.Context,
	b TxBeginner,
	fn func(tx *sql.Tx) error,
	options ...BeginTxOption,
) error {
	var opts *sql.TxOptions
	for _, option := range options {
		option(&opts)
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
	_ Runner     = &DB{}
	_ TxBeginner = &DB{}
)
