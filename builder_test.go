package gosql

import (
	"reflect"
	"testing"
)

func TestSelectQuery(t *testing.T) {
	b := NewBuilder()
	inputs := []SelectQuery{
		b.Select("t1"),
		b.Select("t1").Names("c1", "c2", "c3"),
		b.Select("t1").Where(Column("c1").Equal(123)),
		b.Select("t1").Where(Column("c1").NotEqual(123)),
		b.Select("t1").Where(Column("c2").Equal(nil)),
		b.Select("t1").Where(Column("c2").NotEqual(nil)),
		b.Select("t1").Where(Column("c3").Less(0)),
		b.Select("t1").Where(Column("c3").Greater(0)),
		b.Select("t1").Where(Column("c3").LessEqual(0)),
		b.Select("t1").Where(Column("c3").GreaterEqual(0)),
		b.Select("t1").Where(Column("c1").Greater(0).And(Column("c1").LessEqual(100))),
		b.Select("t1").Where(Column("c1").Greater(0).Or(Column("c1").LessEqual(100))),
	}
	outputs := []string{
		`SELECT * FROM "t1" WHERE 1`,
		`SELECT "c1", "c2", "c3" FROM "t1" WHERE 1`,
		`SELECT * FROM "t1" WHERE "c1" = $1`,
		`SELECT * FROM "t1" WHERE "c1" <> $1`,
		`SELECT * FROM "t1" WHERE "c2" IS NULL`,
		`SELECT * FROM "t1" WHERE "c2" IS NOT NULL`,
		`SELECT * FROM "t1" WHERE "c3" < $1`,
		`SELECT * FROM "t1" WHERE "c3" > $1`,
		`SELECT * FROM "t1" WHERE "c3" <= $1`,
		`SELECT * FROM "t1" WHERE "c3" >= $1`,
		`SELECT * FROM "t1" WHERE "c1" > $1 AND "c1" <= $2`,
		`SELECT * FROM "t1" WHERE "c1" > $1 OR "c1" <= $2`,
	}
	for i, input := range inputs {
		query := input.String()
		if query != outputs[i] {
			t.Errorf("Expected %q, got %q", outputs[i], query)
		}
	}
}

func TestUpdateQuery(t *testing.T) {
	b := NewBuilder()
	q1 := b.Update("t1").
		Where(Column("c1").Equal(123)).
		Names("c2", "c3").Values("test", "test2")
	s1 := `UPDATE "t1" SET "c2" = $1, "c3" = $2 WHERE "c1" = $3`
	v1 := []interface{}{"test", "test2", 123}
	if s, v := q1.Build(); s != s1 || !reflect.DeepEqual(v, v1) {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	if s := q1.String(); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	q2 := b.Update("t2").
		Names("c1", "c2").Values("test", "test2")
	s2 := `UPDATE "t2" SET "c1" = $1, "c2" = $2 WHERE 1`
	v2 := []interface{}{"test", "test2"}
	if s, v := q2.Build(); s != s2 || !reflect.DeepEqual(v, v2) {
		t.Fatalf("Expected %q got %q", s2, s)
	}
	testExpectPanic(t, func() {
		b.Update("t1").Names("c2", "c3").Values("test").Build()
	})
	testExpectPanic(t, func() {
		b.Update("t1").Build()
	})
}

func TestDeleteQuery(t *testing.T) {
	b := NewBuilder()
	q1 := b.Delete("t1").
		Where(Column("c1").Equal(123))
	s1 := `DELETE FROM "t1" WHERE "c1" = $1`
	if s := q1.String(); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	q2 := b.Delete("t2")
	s2 := `DELETE FROM "t2" WHERE 1`
	if s := q2.String(); s != s2 {
		t.Fatalf("Expected %q got %q", s2, s)
	}
}

func TestInsertQuery(t *testing.T) {
	b := NewBuilder()
	q1 := b.Insert("t1").Names("c2", "c3").Values("test", "test2")
	s1 := `INSERT INTO "t1" ("c2", "c3") VALUES ($1, $2)`
	v1 := []interface{}{"test", "test2"}
	if s, v := q1.Build(); s != s1 || !reflect.DeepEqual(v, v1) {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	if s := q1.String(); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	testExpectPanic(t, func() {
		b.Insert("t1").Names("c2", "c3").Values("test").Build()
	})
	testExpectPanic(t, func() {
		b.Insert("t1").Build()
	})
}

func testExpectPanic(tb testing.TB, fn func()) {
	defer func() {
		if r := recover(); r == nil {
			tb.Fatalf("Expected panic")
		}
	}()
	fn()
}
