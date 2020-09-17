package gosql

import (
	"strings"
)

// InsertQuery represents SQL insert query.
type InsertQuery interface {
	Query
	Values(values map[string]interface{}) InsertQuery
}

type insertQuery struct {
	builder *builder
	table   string
	names   []string
	values  []Value
}

func (q insertQuery) Values(values map[string]interface{}) InsertQuery {
	q.names, q.values = nil, nil
	for name, val := range values {
		if _, ok := val.(Value); !ok {
			val = value{value: val}
		}
		q.names = append(q.names, name)
		q.values = append(q.values, val.(Value))
	}
	return q
}

func (q insertQuery) Build() (string, []interface{}) {
	var query strings.Builder
	var opts []interface{}
	query.WriteString("INSERT INTO ")
	query.WriteString(q.builder.buildName(q.table))
	q.buildInsert(&query, &opts)
	return query.String(), opts
}

func (q insertQuery) buildInsert(
	query *strings.Builder, opts *[]interface{},
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
		query.WriteString(value.build(q.builder, opts))
	}
	query.WriteRune(')')
}

func (q insertQuery) String() string {
	query, _ := q.Build()
	return query
}
