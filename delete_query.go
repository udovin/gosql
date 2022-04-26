package gosql

// DeleteQuery represents SQL delete query.
type DeleteQuery interface {
	Query
	SetWhere(where BoolExpression)
}

type deleteQuery struct {
	builder *builder
	table   string
	where   BoolExpression
}

func (q *deleteQuery) SetWhere(where BoolExpression) {
	q.where = where
}

func (q deleteQuery) Build() (string, []any) {
	state := rawBuilder{builder: q.builder}
	state.WriteString("DELETE")
	q.buildFrom(&state)
	q.buildWhere(&state)
	return state.String(), state.Values()
}

func (q deleteQuery) buildFrom(builder *rawBuilder) {
	builder.WriteString(" FROM ")
	builder.WriteName(q.table)
}

func (q deleteQuery) buildWhere(builder *rawBuilder) {
	builder.WriteString(" WHERE ")
	if q.where == nil {
		builder.WriteString("1 = 1")
		return
	}
	q.where.Build(builder)
}

func (q deleteQuery) String() string {
	query, _ := q.Build()
	return query
}
