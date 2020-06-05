package gosql

import (
	"testing"
)

func TestSelectQuery(t *testing.T) {
	b := Builder{}
	inputs := []SelectQuery{
		b.Select().From("t1").Where(Column("c1").Equal(123)),
		b.Select().From("t1").Where(Column("c1").NotEqual(123)),
		b.Select().From("t1").Where(Column("c2").Equal(nil)),
		b.Select().From("t1").Where(Column("c2").NotEqual(nil)),
		b.Select().From("t1").Where(Column("c3").Less(0)),
		b.Select().From("t1").Where(Column("c3").Greater(0)),
		b.Select().From("t1").Where(Column("c3").LessEqual(0)),
		b.Select().From("t1").Where(Column("c3").GreaterEqual(0)),
	}
	outputs := []string{
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
	b := Builder{}
	q1 := b.Update("users").
		Set(map[string]interface{}{"login": "test"}).
		Where(Column("id").Equal(123))
	s1 := `UPDATE "users" SET "login" = $1 WHERE "id" = $2`
	if s := q1.String(); s != s1 {
		t.Fatalf("Expected %q got %q", s1, s)
	}
}
