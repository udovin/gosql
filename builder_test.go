package gosql

import (
	"reflect"
	"testing"
)

type testSetNamesImpl interface {
	SetNames(names ...string)
}

func testSetNames[T testSetNamesImpl](query T, names ...string) T {
	query.SetNames(names...)
	return query
}

type testSetWhereImpl interface {
	SetWhere(where BoolExpr)
}

func testSetWhere[T testSetWhereImpl](query T, where BoolExpr) T {
	query.SetWhere(where)
	return query
}

type testSetOrderByImpl interface {
	SetOrderBy(names ...any)
}

func testSetOrderBy[T testSetOrderByImpl](query T, names ...any) T {
	query.SetOrderBy(names...)
	return query
}

type testSetLimitImpl interface {
	SetLimit(limit int)
}

func testSetLimit[T testSetLimitImpl](query T, limit int) T {
	query.SetLimit(limit)
	return query
}

type testSetValuesImpl interface {
	SetValues(values ...any)
}

func testSetValues[T testSetValuesImpl](query T, values ...any) T {
	query.SetValues(values...)
	return query
}

func TestSelectQuery(t *testing.T) {
	b := NewBuilder(SQLiteDialect)
	inputs := []SelectQuery{
		b.Select("t1"),
		testSetNames(b.Select("t1"), "c1", "c2", "c3"),
		testSetWhere(b.Select("t1"), Column("c1").Equal(123)),
		testSetWhere(b.Select("t1"), Column("c1").NotEqual(123)),
		testSetWhere(b.Select("t1"), Column("c2").Equal(nil)),
		testSetWhere(b.Select("t1"), Column("c2").NotEqual(nil)),
		testSetWhere(b.Select("t1"), Column("c3").Less(0)),
		testSetWhere(b.Select("t1"), Column("c3").Greater(0)),
		testSetWhere(b.Select("t1"), Column("c3").LessEqual(0)),
		testSetWhere(b.Select("t1"), Column("c3").GreaterEqual(0)),
		testSetWhere(b.Select("t1"), Column("c1").Greater(0).And(Column("c1").LessEqual(100))),
		testSetWhere(b.Select("t1"), Column("c1").Greater(0).Or(Column("c1").LessEqual(100))),
		testSetOrderBy(b.Select("t1"), "c1", "c2"),
		testSetOrderBy(b.Select("t1"), Descending("c1"), Descending(Ascending("c2")), Ascending(Descending("c3"))),
		testSetWhere(b.Select("t1"), Column("c1").Greater(0).And(Column("c1").LessEqual(100)).Or(Column("c1").Less(-10))),
		testSetWhere(b.Select("t1"), Column("c1").Greater(0).And(Column("c1").LessEqual(100)).And(Column("c1").Less(10))),
		testSetWhere(b.Select("t1"), Column("c1").Greater(0).And(Column("c1").LessEqual(100).Or(Column("c1").Less(-10)))),
		testSetLimit(testSetOrderBy(b.Select("t1"), "c1"), 123),
	}
	outputs := []string{
		`SELECT * FROM "t1" WHERE 1 = 1`,
		`SELECT "c1", "c2", "c3" FROM "t1" WHERE 1 = 1`,
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
		`SELECT * FROM "t1" WHERE 1 = 1 ORDER BY "c1" ASC, "c2" ASC`,
		`SELECT * FROM "t1" WHERE 1 = 1 ORDER BY "c1" DESC, "c2" DESC, "c3" ASC`,
		`SELECT * FROM "t1" WHERE ("c1" > $1 AND "c1" <= $2) OR "c1" < $3`,
		`SELECT * FROM "t1" WHERE "c1" > $1 AND "c1" <= $2 AND "c1" < $3`,
		`SELECT * FROM "t1" WHERE "c1" > $1 AND ("c1" <= $2 OR "c1" < $3)`,
		`SELECT * FROM "t1" WHERE 1 = 1 ORDER BY "c1" ASC LIMIT 123`,
	}
	for i, input := range inputs {
		query := b.BuildString(input)
		if query != outputs[i] {
			t.Errorf("Expected %q, got %q", outputs[i], query)
		}
	}
}

func TestUpdateQuery(t *testing.T) {
	b := NewBuilder(SQLiteDialect)
	q1 := testSetValues(testSetNames(testSetWhere(b.Update("t1"), Column("c1").Equal(123)), "c2", "c3"), "test", "test2")
	s1 := `UPDATE "t1" SET "c2" = $1, "c3" = $2 WHERE "c1" = $3`
	v1 := []any{"test", "test2", 123}
	if s, v := b.Build(q1); s != s1 || !reflect.DeepEqual(v, v1) {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	if s := b.BuildString(q1); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	q2 := testSetValues(testSetNames(b.Update("t2"), "c1", "c2"), "test", "test2")
	s2 := `UPDATE "t2" SET "c1" = $1, "c2" = $2 WHERE 1 = 1`
	v2 := []any{"test", "test2"}
	if s, v := b.Build(q2); s != s2 || !reflect.DeepEqual(v, v2) {
		t.Fatalf("Expected %q got %q", s2, s)
	}
	testExpectPanic(t, func() {
		b.Build(testSetValues(testSetNames(b.Update("t1"), "c2", "c3"), "test"))
	})
	testExpectPanic(t, func() {
		b.Build(b.Update("t1"))
	})
}

func TestDeleteQuery(t *testing.T) {
	b := NewBuilder(SQLiteDialect)
	q1 := testSetWhere(b.Delete("t1"), Column("c1").Equal(123))
	s1 := `DELETE FROM "t1" WHERE "c1" = $1`
	if s := b.BuildString(q1); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	q2 := b.Delete("t2")
	s2 := `DELETE FROM "t2" WHERE 1 = 1`
	if s := b.BuildString(q2); s != s2 {
		t.Fatalf("Expected %q got %q", s2, s)
	}
}

func TestInsertQuery(t *testing.T) {
	b := NewBuilder(SQLiteDialect)
	q1 := testSetValues(testSetNames(b.Insert("t1"), "c2", "c3"), "test", "test2")
	s1 := `INSERT INTO "t1" ("c2", "c3") VALUES ($1, $2)`
	v1 := []any{"test", "test2"}
	if s, v := b.Build(q1); s != s1 || !reflect.DeepEqual(v, v1) {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	if s := b.BuildString(q1); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	testExpectPanic(t, func() {
		b.Build(testSetValues(testSetNames(b.Insert("t1"), "c2", "c3"), "test"))
	})
	testExpectPanic(t, func() {
		b.Build(b.Insert("t1"))
	})
}

func TestPostgresInsertQuery(t *testing.T) {
	b := NewBuilder(PostgresDialect)
	q1 := testSetValues(testSetNames(b.Insert("t1"), "c2", "c3"), "test", "test2")
	q1.(*PostgresInsertQuery).SetReturning("id")
	s1 := `INSERT INTO "t1" ("c2", "c3") VALUES ($1, $2) RETURNING "id"`
	v1 := []any{"test", "test2"}
	if s, v := b.Build(q1); s != s1 || !reflect.DeepEqual(v, v1) {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	if s := b.BuildString(q1); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	testExpectPanic(t, func() {
		b2 := NewBuilder(SQLiteDialect)
		b2.Build(q1)
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
