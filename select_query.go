package gosql

import (
	"strconv"
)

// SelectQuery represents SQL select query.
type SelectQuery interface {
	Query
	SetNames(names ...string)
	SetWhere(where BoolExpression)
	SetOrderBy(names ...interface{})
	SetLimit(limit int)
}

type selectQuery struct {
	builder *builder
	table   string
	names   []string
	where   BoolExpression
	orderBy []OrderExpression
	limit   int
}

func (q *selectQuery) SetNames(names ...string) {
	q.names = names
}

func (q *selectQuery) SetWhere(where BoolExpression) {
	q.where = where
}

func (q *selectQuery) SetOrderBy(names ...interface{}) {
	q.orderBy = nil
	for _, name := range names {
		q.orderBy = append(q.orderBy, wrapOrderExpression(name))
	}
}

func (q *selectQuery) SetLimit(limit int) {
	q.limit = limit
}

func (q selectQuery) Build() (string, []interface{}) {
	builder := rawBuilder{builder: q.builder}
	builder.WriteString("SELECT ")
	q.buildNames(&builder)
	builder.WriteString(" FROM ")
	builder.WriteString(q.builder.buildName(q.table))
	q.buildWhere(&builder)
	q.buildOrderBy(&builder)
	if q.limit > 0 {
		builder.WriteString(" LIMIT ")
		builder.WriteString(strconv.Itoa(q.limit))
	}
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
