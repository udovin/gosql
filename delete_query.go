package gosql

import (
	"strings"
)

// DeleteQuery represents SQL delete query.
type DeleteQuery interface {
	Query
	Where(where BoolExpr) DeleteQuery
}

type deleteQuery struct {
	builder *builder
	table   string
	where   BoolExpr
	values  []Value
}

func (q deleteQuery) Where(where BoolExpr) DeleteQuery {
	q.where = where
	return q
}

func (q deleteQuery) Build() (string, []interface{}) {
	var query strings.Builder
	var opts []interface{}
	query.WriteString("DELETE")
	q.buildFrom(&query, &opts)
	q.buildWhere(&query, &opts)
	return query.String(), opts
}

func (q deleteQuery) buildValues(
	query *strings.Builder, opts *[]interface{},
) {
	if len(q.values) == 0 {
		query.WriteRune('*')
		return
	}
	for i, value := range q.values {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString(value.build(q.builder, opts))
	}
}

func (q deleteQuery) buildFrom(
	query *strings.Builder, _ *[]interface{},
) {
	query.WriteString(" FROM ")
	query.WriteString(q.builder.buildName(q.table))
}

func (q deleteQuery) buildWhere(
	query *strings.Builder, opts *[]interface{},
) {
	query.WriteString(" WHERE ")
	if q.where == nil {
		query.WriteRune('1')
		return
	}
	query.WriteString(q.where.build(q.builder, opts))
}

func (q deleteQuery) String() string {
	query, _ := q.Build()
	return query
}
