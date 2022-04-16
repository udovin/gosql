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
		option := WithTxOptions(&sql.TxOptions{
			ReadOnly:  true,
			Isolation: sql.LevelRepeatableRead,
		})
		option(&opts)
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
		option := WithReadOnly(true)
		option(&opts)
		if opts == nil {
			t.Fatal("Opts should be initialized")
		}
		if !opts.ReadOnly {
			t.Fatal("Opts should be readonly")
		}
		if opts.Isolation != sql.LevelDefault {
			t.Fatal("Opts should be default")
		}
	}
	{
		var opts *sql.TxOptions
		option := WithIsolation(sql.LevelReadCommitted)
		option(&opts)
		if opts == nil {
			t.Fatal("Opts should be initialized")
		}
		if opts.ReadOnly {
			t.Fatal("Opts should be writable")
		}
		if opts.Isolation != sql.LevelReadCommitted {
			t.Fatal("Opts should be read commited")
		}
	}
	{
		var opts *sql.TxOptions
		option1 := WithReadOnly(true)
		option2 := WithIsolation(sql.LevelSerializable)
		option1(&opts)
		option2(&opts)
		if opts == nil {
			t.Fatal("Opts should be initialized")
		}
		if !opts.ReadOnly {
			t.Fatal("Opts should be readonly")
		}
		if opts.Isolation != sql.LevelSerializable {
			t.Fatal("Opts should be read commited")
		}
	}
}
