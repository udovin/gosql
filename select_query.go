package gosql

// SelectQuery represents SQL select query.
type SelectQuery interface {
	Query
	Names(names ...string) SelectQuery
	Where(where BoolExpression) SelectQuery
	OrderBy(names ...interface{}) SelectQuery
}

type selectQuery struct {
	builder *builder
	table   string
	names   []string
	where   BoolExpression
	orderBy []OrderExpression
}

func (q selectQuery) Names(names ...string) SelectQuery {
	q.names = names
	return q
}

func (q selectQuery) Where(where BoolExpression) SelectQuery {
	q.where = where
	return q
}

func (q selectQuery) OrderBy(names ...interface{}) SelectQuery {
	q.orderBy = nil
	for _, name := range names {
		q.orderBy = append(q.orderBy, wrapOrderExpression(name))
	}
	return q
}

func (q selectQuery) Build() (string, []interface{}) {
	builder := rawBuilder{builder: q.builder}
	builder.WriteString("SELECT ")
	q.buildNames(&builder)
	builder.WriteString(" FROM ")
	builder.WriteString(q.builder.buildName(q.table))
	q.buildWhere(&builder)
	q.buildOrderBy(&builder)
	return builder.String(), builder.Values()
}

func (q selectQuery) buildNames(builder *rawBuilder) {
	if len(q.names) == 0 {
		builder.WriteRune('*')
		return
	}
	for i, name := range q.names {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteName(name)
	}
}

func (q selectQuery) buildWhere(builder *rawBuilder) {
	builder.WriteString(" WHERE ")
	if q.where == nil {
		builder.WriteRune('1')
		return
	}
	q.where.Build(builder)
}

func (q selectQuery) buildOrderBy(builder *rawBuilder) {
	if len(q.orderBy) == 0 {
		return
	}
	builder.WriteString(" ORDER BY ")
	for i, name := range q.orderBy {
		if i > 0 {
			builder.WriteString(", ")
		}
		name.Build(builder)
	}
}

func (q selectQuery) String() string {
	query, _ := q.Build()
	return query
}
