package gosql

// UpdateQuery represents SQL update query.
type UpdateQuery interface {
	Query
	Where(where BoolExpr) UpdateQuery
	Names(names ...string) UpdateQuery
	Values(values ...interface{}) UpdateQuery
}

type updateQuery struct {
	builder *builder
	table   string
	where   BoolExpr
	names   []string
	values  []Value
}

func (q updateQuery) Where(where BoolExpr) UpdateQuery {
	q.where = where
	return q
}

func (q updateQuery) Names(names ...string) UpdateQuery {
	q.names = names
	return q
}

func (q updateQuery) Values(values ...interface{}) UpdateQuery {
	q.values = nil
	for _, val := range values {
		q.values = append(q.values, wrapValue(val))
	}
	return q
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
