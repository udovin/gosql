package gosql

import (
	"fmt"
	"strings"
)

// SelectQuery represents SQL select query.
type SelectQuery interface {
	Query
	From(from string) SelectQuery
	Where(where BoolExpr) SelectQuery
}

type selectQuery struct {
	values []Value
	from   string
	where  BoolExpr
}

func (q selectQuery) From(from string) SelectQuery {
	q.from = from
	return q
}

func (q selectQuery) Where(where BoolExpr) SelectQuery {
	q.where = where
	return q
}

func (q selectQuery) Build() (string, []interface{}) {
	var query strings.Builder
	var opts []interface{}
	query.WriteString("SELECT ")
	q.buildValues(&query, &opts)
	q.buildFrom(&query, &opts)
	q.buildWhere(&query, &opts)
	return query.String(), opts
}

func (q selectQuery) buildValues(
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
		query.WriteString(value.build(opts))
	}
}

func (q selectQuery) buildFrom(
	query *strings.Builder, _ *[]interface{},
) {
	query.WriteString(" FROM ")
	query.WriteString(fmt.Sprintf("%q", q.from))
}

func (q selectQuery) buildWhere(
	query *strings.Builder, opts *[]interface{},
) {
	query.WriteString(" WHERE ")
	if q.where == nil {
		query.WriteRune('1')
		return
	}
	query.WriteString(q.where.build(opts))
}

func (q selectQuery) String() string {
	query, _ := q.Build()
	return query
}
