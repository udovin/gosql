package gosql

import (
	"strings"
)

// UpdateQuery represents SQL update query.
type UpdateQuery interface {
	Query
	Where(where BoolExpr) UpdateQuery
	Values(values map[string]interface{}) UpdateQuery
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

func (q updateQuery) Values(values map[string]interface{}) UpdateQuery {
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

func (q updateQuery) Build() (string, []interface{}) {
	var query strings.Builder
	var opts []interface{}
	query.WriteString("UPDATE ")
	query.WriteString(q.builder.buildName(q.table))
	q.buildSet(&query, &opts)
	q.buildWhere(&query, &opts)
	return query.String(), opts
}

func (q updateQuery) buildSet(
	query *strings.Builder, opts *[]interface{},
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
		query.WriteString(q.values[i].build(q.builder, opts))
	}
}

func (q updateQuery) buildWhere(
	query *strings.Builder, opts *[]interface{},
) {
	query.WriteString(" WHERE ")
	if q.where == nil {
		query.WriteRune('1')
		return
	}
	query.WriteString(q.where.build(q.builder, opts))
}

func (q updateQuery) String() string {
	query, _ := q.Build()
	return query
}
