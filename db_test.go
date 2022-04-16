package gosql

import (
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
