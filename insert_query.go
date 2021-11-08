package gosql

// InsertQuery represents SQL insert query.
type InsertQuery interface {
	Query
	Names(name ...string) InsertQuery
	Values(values ...interface{}) InsertQuery
}

type insertQuery struct {
	builder *builder
	table   string
	names   []string
	values  []Value
}

func (q insertQuery) Names(names ...string) InsertQuery {
	q.names = names
	return q
}

func (q insertQuery) Values(values ...interface{}) InsertQuery {
	q.values = nil
	for _, val := range values {
		q.values = append(q.values, wrapValue(val))
	}
	return q
}

func (q insertQuery) Build() (string, []interface{}) {
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
