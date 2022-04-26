package gosql

// InsertQuery represents SQL insert query.
type InsertQuery interface {
	Query
	SetNames(name ...string)
	SetValues(values ...any)
}

type insertQuery struct {
	builder *builder
	table   string
	names   []string
	values  []Value
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

func (q insertQuery) Build() (string, []any) {
	builder := rawBuilder{builder: q.builder}
	builder.WriteString("INSERT INTO ")
	builder.WriteString(q.builder.buildName(q.table))
	q.buildInsert(&builder)
	return builder.String(), builder.Values()
}

func (q insertQuery) buildInsert(builder *rawBuilder) {
	if len(q.names) == 0 {
		panic("list of names can not be empty")
	}
	if len(q.names) != len(q.values) {
		panic("amount of names and values differs")
	}
	builder.WriteString(" (")
	for i, name := range q.names {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteName(name)
	}
	builder.WriteString(") VALUES (")
	for i, value := range q.values {
		if i > 0 {
			builder.WriteString(", ")
		}
		value.Build(builder)
	}
	builder.WriteRune(')')
}

func (q insertQuery) String() string {
	query, _ := q.Build()
	return query
}

type PostgresInsertQuery struct {
	insertQuery
	returning []string
}

func (q *PostgresInsertQuery) SetReturning(names ...string) {
	q.returning = names
}

func (q PostgresInsertQuery) buildReturning(builder *rawBuilder) {
	if len(q.returning) > 0 {
		builder.WriteString(" RETURNING ")
		for i, name := range q.returning {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteName(name)
		}
	}
}

func (q PostgresInsertQuery) Build() (string, []any) {
	builder := rawBuilder{builder: q.builder}
	builder.WriteString("INSERT INTO ")
	builder.WriteString(q.builder.buildName(q.table))
	q.buildInsert(&builder)
	q.buildReturning(&builder)
	return builder.String(), builder.Values()
}

func (q PostgresInsertQuery) String() string {
	query, _ := q.Build()
	return query
}
