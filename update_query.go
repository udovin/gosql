package gosql

import (
	"strings"
)

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
	var query strings.Builder
	state := buildState{builder: q.builder}
	query.WriteString("UPDATE ")
	query.WriteString(q.builder.buildName(q.table))
	q.buildSet(&query, &state)
	q.buildWhere(&query, &state)
	return query.String(), state.Values()
}

func (q updateQuery) buildSet(
	query *strings.Builder, state *buildState,
) {
	if len(q.values) == 0 {
		return
	}
	if len(q.names) != len(q.values) {
		panic("amount of names and values differs")
	}
	query.WriteString(" SET ")
	first := true
	for i := range q.names {
		if !first {
			query.WriteString(", ")
			first = false
		}
		query.WriteString(q.builder.buildName(q.names[i]))
		query.WriteString(" = ")
		query.WriteString(q.values[i].Build(state))
	}
}

func (q updateQuery) buildWhere(
	query *strings.Builder, state *buildState,
) {
	query.WriteString(" WHERE ")
	if q.where == nil {
		query.WriteRune('1')
		return
	}
	query.WriteString(q.where.Build(state))
}

func (q updateQuery) String() string {
	query, _ := q.Build()
	return query
}
