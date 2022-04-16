package gosql

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSQLite(t *testing.T) {
	cfg := SQLiteConfig{Path: ":memory:"}
	db, err := cfg.NewDB()
	if err != nil {
		t.Fatal("Error:", err)
	}
	if db.DB == nil {
		t.Fatal("Writable DB cannot be empty")
	}
	if db.RO == nil {
		t.Fatal("ReadOnly DB cannot be empty")
	}
	if d := db.Builder.Dialect(); d != SQLiteDialect {
		t.Fatalf("Expected %d dialect, but got: %d", SQLiteDialect, d)
	}
}

func TestTxOptions(t *testing.T) {
	{
		var opts *sql.TxOptions
		WithTxOptions(&sql.TxOptions{
			ReadOnly:  true,
			Isolation: sql.LevelRepeatableRead,
		})(&opts)
		if opts == nil {
			t.Fatal("Opts should be initialized")
		}
		if !opts.ReadOnly {
			t.Fatal("Opts should be readonly")
		}
		if opts.Isolation != sql.LevelRepeatableRead {
			t.Fatal("Opts should be repeatable read")
		}
	}
	{
		var opts *sql.TxOptions
		WithReadOnly(true)(&opts)
		if opts == nil {
			t.Fatal("Opts should be initialized")
		}
		if !opts.ReadOnly {
			t.Fatal("Opts should marked readonly")
		}
		if opts.Isolation != sql.LevelDefault {
			t.Fatal("Opts should marked default")
		}
	}
	{
		var opts *sql.TxOptions
		WithIsolation(sql.LevelReadCommitted)(&opts)
		if opts == nil {
			t.Fatal("Opts should be initialized")
		}
		if opts.ReadOnly {
			t.Fatal("Opts should marked writable")
		}
		if opts.Isolation != sql.LevelReadCommitted {
			t.Fatal("Opts should marked read commited")
		}
	}
	{
		var opts *sql.TxOptions
		WithReadOnly(true)(&opts)
		WithIsolation(sql.LevelSerializable)(&opts)
		if opts == nil {
			t.Fatal("Opts should be initialized")
		}
		if !opts.ReadOnly {
			t.Fatal("Opts should marked readonly")
		}
		if opts.Isolation != sql.LevelSerializable {
			t.Fatal("Opts should marked serializable")
		}
		WithReadOnly(false)(&opts)
		if opts.ReadOnly {
			t.Fatal("Opts should marked writable")
		}
	}
}
