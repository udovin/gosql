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
	SetWhere(where BoolExpression)
}

func testSetWhere[T testSetWhereImpl](query T, where BoolExpression) T {
	query.SetWhere(where)
	return query
}

type testSetOrderByImpl interface {
	SetOrderBy(names ...interface{})
}

func testSetOrderBy[T testSetOrderByImpl](query T, names ...interface{}) T {
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
	SetValues(values ...interface{})
}

func testSetValues[T testSetValuesImpl](query T, values ...interface{}) T {
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
		`SELECT * FROM "t1" WHERE 1 ORDER BY "c1" ASC, "c2" ASC`,
		`SELECT * FROM "t1" WHERE 1 ORDER BY "c1" DESC, "c2" DESC, "c3" ASC`,
		`SELECT * FROM "t1" WHERE ("c1" > $1 AND "c1" <= $2) OR "c1" < $3`,
		`SELECT * FROM "t1" WHERE "c1" > $1 AND "c1" <= $2 AND "c1" < $3`,
		`SELECT * FROM "t1" WHERE "c1" > $1 AND ("c1" <= $2 OR "c1" < $3)`,
		`SELECT * FROM "t1" WHERE 1 ORDER BY "c1" ASC LIMIT 123`,
	}
	for i, input := range inputs {
		query := input.String()
		if query != outputs[i] {
			t.Errorf("Expected %q, got %q", outputs[i], query)
		}
	}
}

func TestUpdateQuery(t *testing.T) {
	b := NewBuilder(SQLiteDialect)
	q1 := testSetValues(testSetNames(testSetWhere(b.Update("t1"), Column("c1").Equal(123)), "c2", "c3"), "test", "test2")
	s1 := `UPDATE "t1" SET "c2" = $1, "c3" = $2 WHERE "c1" = $3`
	v1 := []interface{}{"test", "test2", 123}
	if s, v := q1.Build(); s != s1 || !reflect.DeepEqual(v, v1) {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	if s := q1.String(); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	q2 := testSetValues(testSetNames(b.Update("t2"), "c1", "c2"), "test", "test2")
	s2 := `UPDATE "t2" SET "c1" = $1, "c2" = $2 WHERE 1`
	v2 := []interface{}{"test", "test2"}
	if s, v := q2.Build(); s != s2 || !reflect.DeepEqual(v, v2) {
		t.Fatalf("Expected %q got %q", s2, s)
	}
	testExpectPanic(t, func() {
		testSetValues(testSetNames(b.Update("t1"), "c2", "c3"), "test").Build()
	})
	testExpectPanic(t, func() {
		b.Update("t1").Build()
	})
}

func TestDeleteQuery(t *testing.T) {
	b := NewBuilder(SQLiteDialect)
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
	b := NewBuilder(SQLiteDialect)
	q1 := testSetValues(testSetNames(b.Insert("t1"), "c2", "c3"), "test", "test2")
	s1 := `INSERT INTO "t1" ("c2", "c3") VALUES ($1, $2)`
	v1 := []interface{}{"test", "test2"}
	if s, v := q1.Build(); s != s1 || !reflect.DeepEqual(v, v1) {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	if s := q1.String(); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
	testExpectPanic(t, func() {
		testSetValues(testSetNames(b.Insert("t1"), "c2", "c3"), "test").Build()
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
