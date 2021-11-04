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
}

func (q deleteQuery) Where(where BoolExpr) DeleteQuery {
	q.where = where
	return q
}

func (q deleteQuery) Build() (string, []interface{}) {
	var query strings.Builder
	state := buildState{builder: q.builder}
	query.WriteString("DELETE")
	q.buildFrom(&query, &state)
	q.buildWhere(&query, &state)
	return query.String(), state.Values()
}

func (q deleteQuery) buildFrom(
	query *strings.Builder, state *buildState,
) {
	query.WriteString(" FROM ")
	query.WriteString(q.builder.buildName(q.table))
}

func (q deleteQuery) buildWhere(
	query *strings.Builder, state *buildState,
) {
	query.WriteString(" WHERE ")
	if q.where == nil {
		query.WriteRune('1')
		return
	}
	query.WriteString(q.where.Build(state))
}

func (q deleteQuery) String() string {
	query, _ := q.Build()
	return query
}
