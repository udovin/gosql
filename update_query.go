package gosql

import (
	"fmt"
	"strings"
)

// UpdateQuery represents SQL update query.
type UpdateQuery interface {
	Query
	Set(values map[string]interface{}) UpdateQuery
	Where(where BoolExpr) UpdateQuery
}

type updateQuery struct {
	table  string
	values map[string]Value
	where  BoolExpr
}

func (q updateQuery) Set(values map[string]interface{}) UpdateQuery {
	q.values = map[string]Value{}
	for key, val := range values {
		if _, ok := val.(Value); !ok {
			val = value{value: val}
		}
		q.values[key] = val.(Value)
	}
	return q
}

func (q updateQuery) Where(where BoolExpr) UpdateQuery {
	q.where = where
	return q
}

func (q updateQuery) Build() (string, []interface{}) {
	var opts []interface{}
	values := q.buildValues(&opts)
	where := q.buildWhere(&opts)
	query := fmt.Sprintf(
		"UPDATE %q SET %s WHERE %s",
		q.table, values, where,
	)
	return query, opts
}

func (q updateQuery) buildValues(opts *[]interface{}) string {
	if len(q.values) == 0 {
		return ""
	}
	var builder strings.Builder
	for key, value := range q.values {
		if builder.Len() > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%q", key))
		builder.WriteString(" = ")
		builder.WriteString(value.build(opts))
	}
	return builder.String()
}

func (q updateQuery) buildWhere(opts *[]interface{}) string {
	if q.where == nil {
		return "1"
	}
	return q.where.build(opts)
}

func (q updateQuery) String() string {
	query, _ := q.Build()
	return query
}
