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
	values  map[string]Value
}

func (q insertQuery) Values(values map[string]interface{}) InsertQuery {
	q.values = map[string]Value{}
	for key, val := range values {
		if _, ok := val.(Value); !ok {
			val = value{value: val}
		}
		q.values[key] = val.(Value)
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
	var names []string
	var values []Value
	for name, value := range q.values {
		names = append(names, name)
		values = append(values, value)
	}
	query.WriteString(" (")
	for i, name := range names {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString(q.builder.buildName(name))
	}
	query.WriteString(") VALUES (")
	for i, value := range values {
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
