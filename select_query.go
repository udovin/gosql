package gosql

import (
	"strings"
)

// SelectQuery represents SQL select query.
type SelectQuery interface {
	Query
	Names(names ...string) SelectQuery
	Where(where BoolExpr) SelectQuery
}

type selectQuery struct {
	builder *builder
	table   string
	names   []string
	where   BoolExpr
}

func (q selectQuery) Names(names ...string) SelectQuery {
	q.names = names
	return q
}

func (q selectQuery) Where(where BoolExpr) SelectQuery {
	q.where = where
	return q
}

func (q selectQuery) Build() (string, []interface{}) {
	var query strings.Builder
	state := buildState{builder: q.builder}
	query.WriteString("SELECT ")
	q.buildValues(&query, &state)
	query.WriteString(" FROM ")
	query.WriteString(q.builder.buildName(q.table))
	q.buildWhere(&query, &state)
	return query.String(), state.Values()
}

func (q selectQuery) buildValues(
	query *strings.Builder, state *buildState,
) {
	if len(q.names) == 0 {
		query.WriteRune('*')
		return
	}
	for i, name := range q.names {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString(q.builder.buildName(name))
	}
}

func (q selectQuery) buildWhere(
	query *strings.Builder, state *buildState,
) {
	query.WriteString(" WHERE ")
	if q.where == nil {
		query.WriteRune('1')
		return
	}
	query.WriteString(q.where.Build(state))
}

func (q selectQuery) String() string {
	query, _ := q.Build()
	return query
}
