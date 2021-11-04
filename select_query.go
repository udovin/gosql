package gosql

import (
	"strings"
)

// SelectQuery represents SQL select query.
type SelectQuery interface {
	Query
	Where(where BoolExpr) SelectQuery
	Values(values ...Value) SelectQuery
}

type selectQuery struct {
	builder *builder
	table   string
	where   BoolExpr
	values  []Value
}

func (q selectQuery) Where(where BoolExpr) SelectQuery {
	q.where = where
	return q
}

func (q selectQuery) Values(values ...Value) SelectQuery {
	q.values = values
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
	if len(q.values) == 0 {
		query.WriteRune('*')
		return
	}
	for i, value := range q.values {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString(value.Build(state))
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
