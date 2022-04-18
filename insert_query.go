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
	state := rawBuilder{builder: q.builder}
	state.WriteString("INSERT INTO ")
	state.WriteString(q.builder.buildName(q.table))
	q.buildInsert(&state)
	return state.String(), state.Values()
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
