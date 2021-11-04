package gosql

import (
	"strings"
)

// InsertQuery represents SQL insert query.
type InsertQuery interface {
	Query
	Names(name ...string) InsertQuery
	Values(values ...interface{}) InsertQuery
}

type insertQuery struct {
	builder *builder
	table   string
	names   []string
	values  []Value
}

func (q insertQuery) Names(names ...string) InsertQuery {
	q.names = names
	return q
}

func (q insertQuery) Values(values ...interface{}) InsertQuery {
	q.values = nil
	for _, val := range values {
		q.values = append(q.values, wrapValue(val))
	}
	return q
}

func (q insertQuery) Build() (string, []interface{}) {
	var query strings.Builder
	state := buildState{builder: q.builder}
	query.WriteString("INSERT INTO ")
	query.WriteString(q.builder.buildName(q.table))
	q.buildInsert(&query, &state)
	return query.String(), state.Values()
}

func (q insertQuery) buildInsert(
	query *strings.Builder, state *buildState,
) {
	if len(q.names) != len(q.values) {
		panic("amount of names and values differs")
	}
	query.WriteString(" (")
	for i, name := range q.names {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString(q.builder.buildName(name))
	}
	query.WriteString(") VALUES (")
	for i, value := range q.values {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString(value.Build(state))
	}
	query.WriteRune(')')
}

func (q insertQuery) String() string {
	query, _ := q.Build()
	return query
}
