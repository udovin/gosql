package gosql

// SelectQuery represents SQL select query.
type SelectQuery interface {
	Query
	Names(names ...string) SelectQuery
	Where(where BoolExpr) SelectQuery
	OrderBy(names ...string) SelectQuery
}

type selectQuery struct {
	builder *builder
	table   string
	names   []string
	where   BoolExpr
	orderBy []string
}

func (q selectQuery) Names(names ...string) SelectQuery {
	q.names = names
	return q
}

func (q selectQuery) Where(where BoolExpr) SelectQuery {
	q.where = where
	return q
}

func (q selectQuery) OrderBy(names ...string) SelectQuery {
	q.orderBy = names
	return q
}

func (q selectQuery) Build() (string, []interface{}) {
	builder := rawBuilder{builder: q.builder}
	builder.WriteString("SELECT ")
	q.buildValues(&builder)
	builder.WriteString(" FROM ")
	builder.WriteString(q.builder.buildName(q.table))
	q.buildWhere(&builder)
	return builder.String(), builder.Values()
}

func (q selectQuery) buildValues(builder *rawBuilder) {
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

func (q selectQuery) String() string {
	query, _ := q.Build()
	return query
}
