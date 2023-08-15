package gosql

// InsertQuery represents SQL insert query.
type InsertQuery interface {
	Query
	SetTable(table string)
	SetNames(name ...string)
	SetValues(values ...any)
}

type insertQuery struct {
	table  string
	names  []string
	values []Value
}

func (q *insertQuery) SetTable(table string) {
	q.table = table
}

func (q *insertQuery) SetNames(names ...string) {
	q.names = names
}

func (q *insertQuery) SetValues(values ...any) {
	q.values = nil
	for _, val := range values {
		q.values = append(q.values, wrapValue(val))
	}
}

func (q insertQuery) WriteQuery(w Writer) {
	w.WriteString("INSERT INTO ")
	w.WriteName(q.table)
	q.writeInsert(w)
}

func (q insertQuery) writeInsert(w Writer) {
	if len(q.names) == 0 {
		panic("list of names can not be empty")
	}
	if len(q.names) != len(q.values) {
		panic("amount of names and values differs")
	}
	w.WriteString(" (")
	for i, name := range q.names {
		if i > 0 {
			w.WriteString(", ")
		}
		w.WriteName(name)
	}
	w.WriteString(") VALUES (")
	for i, value := range q.values {
		if i > 0 {
			w.WriteString(", ")
		}
		value.WriteExpr(w)
	}
	w.WriteRune(')')
}

type PostgresInsertQuery struct {
	insertQuery
	returning []string
}

func (q *PostgresInsertQuery) SetReturning(names ...string) {
	q.returning = names
}

func (q PostgresInsertQuery) WriteQuery(w Writer) {
	q.insertQuery.WriteQuery(w)
	q.writeReturning(w)
}

func (q PostgresInsertQuery) writeReturning(w Writer) {
	if len(q.returning) > 0 {
		w.WriteString(" RETURNING ")
		for i, name := range q.returning {
			if i > 0 {
				w.WriteString(", ")
			}
			w.WriteName(name)
		}
	}
}
