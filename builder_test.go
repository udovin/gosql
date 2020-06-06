package gosql

import (
	"reflect"
	"testing"
)

func TestSelectQuery(t *testing.T) {
	b := NewBuilder()
	inputs := []SelectQuery{
		b.Select("t1"),
		b.Select("t1").Where(Column("c1").Equal(123)),
		b.Select("t1").Where(Column("c1").NotEqual(123)),
		b.Select("t1").Where(Column("c2").Equal(nil)),
		b.Select("t1").Where(Column("c2").NotEqual(nil)),
		b.Select("t1").Where(Column("c3").Less(0)),
		b.Select("t1").Where(Column("c3").Greater(0)),
		b.Select("t1").Where(Column("c3").LessEqual(0)),
		b.Select("t1").Where(Column("c3").GreaterEqual(0)),
	}
	outputs := []string{
		`SELECT * FROM "t1" WHERE 1`,
		`SELECT * FROM "t1" WHERE "c1" = $1`,
		`SELECT * FROM "t1" WHERE "c1" <> $1`,
		`SELECT * FROM "t1" WHERE "c2" IS NULL`,
		`SELECT * FROM "t1" WHERE "c2" IS NOT NULL`,
		`SELECT * FROM "t1" WHERE "c3" < $1`,
		`SELECT * FROM "t1" WHERE "c3" > $1`,
		`SELECT * FROM "t1" WHERE "c3" <= $1`,
		`SELECT * FROM "t1" WHERE "c3" >= $1`,
	}
	for i, input := range inputs {
		query := input.String()
		if query != outputs[i] {
			t.Fatalf("Expecte %q, got %q", outputs[i], query)
		}
	}
}

func TestUpdateQuery(t *testing.T) {
	b := NewBuilder()
	q1 := b.Update("t1").
		Where(Column("c1").Equal(123)).
		Values(map[string]interface{}{"c2": "test"})
	s1 := `UPDATE "t1" SET "c2" = $1 WHERE "c1" = $2`
	v1 := []interface{}{"test", 123}
	if s := q1.String(); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	if s, v := q1.Build(); s != s1 || !reflect.DeepEqual(v, v1) {
		t.Fatalf("Expected %q got %q", s1, s)
	}
}

func TestDeleteQuery(t *testing.T) {
	b := NewBuilder()
	q1 := b.Delete("t1").
		Where(Column("c1").Equal(123))
	s1 := `DELETE FROM "t1" WHERE "c1" = $1`
	if s := q1.String(); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
}

func TestInsertQuery(t *testing.T) {
	b := NewBuilder()
	q1 := b.Insert("t1").
		Values(map[string]interface{}{"c2": "test"})
	s1 := `INSERT INTO "t1" ("c2") VALUES ($1)`
	if s := q1.String(); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
}
