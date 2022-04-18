package gosql

// UpdateQuery represents SQL update query.
type UpdateQuery interface {
	Query
	SetWhere(where BoolExpression)
	SetNames(names ...string)
	SetValues(values ...interface{})
}

type updateQuery struct {
	builder *builder
	table   string
	where   BoolExpression
	names   []string
	values  []Value
}

func (q *updateQuery) SetWhere(where BoolExpression) {
	q.where = where
}

func (q *updateQuery) SetNames(names ...string) {
	q.names = names
}

func (q *updateQuery) SetValues(values ...interface{}) {
	q.values = nil
	for _, val := range values {
		q.values = append(q.values, wrapValue(val))
	}
}

func (q updateQuery) Build() (string, []interface{}) {
	builder := rawBuilder{builder: q.builder}
	builder.WriteString("UPDATE ")
	builder.WriteString(q.builder.buildName(q.table))
	q.buildSet(&builder)
	q.buildWhere(&builder)
	return builder.String(), builder.Values()
}

func (q updateQuery) buildSet(builder *rawBuilder) {
	if len(q.names) == 0 {
		panic("list of names can not be empty")
	}
	if len(q.names) != len(q.values) {
		panic("amount of names and values differs")
	}
	builder.WriteString(" SET ")
	for i := range q.names {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteName(q.names[i])
		builder.WriteString(" = ")
		q.values[i].Build(builder)
	}
}

func (q updateQuery) buildWhere(builder *rawBuilder) {
	builder.WriteString(" WHERE ")
	if q.where == nil {
		builder.WriteRune('1')
		return
	}
	q.where.Build(builder)
}

func (q updateQuery) String() string {
	query, _ := q.Build()
	return query
}
