package gosql

import (
	"strconv"
)

// SelectQuery represents SQL select query.
type SelectQuery interface {
	Query
	SetTable(table string)
	SetNames(names ...string)
	SetWhere(where BoolExpr)
	SetOrderBy(names ...any)
	SetLimit(limit int)
}

type selectQuery struct {
	table   string
	names   []string
	where   BoolExpr
	orderBy []OrderExpr
	limit   int
}

func (q *selectQuery) SetTable(table string) {
	q.table = table
}

func (q *selectQuery) SetNames(names ...string) {
	q.names = names
}

func (q *selectQuery) SetWhere(where BoolExpr) {
	q.where = where
}

func (q *selectQuery) SetOrderBy(names ...any) {
	q.orderBy = nil
	for _, name := range names {
		q.orderBy = append(q.orderBy, wrapOrderExpression(name))
	}
}

func (q *selectQuery) SetLimit(limit int) {
	q.limit = limit
}

func (q selectQuery) WriteQuery(w Writer) {
	w.WriteString("SELECT ")
	q.writeNames(w)
	w.WriteString(" FROM ")
	w.WriteName(q.table)
	q.writeWhere(w)
	q.writeOrderBy(w)
	if q.limit > 0 {
		w.WriteString(" LIMIT ")
		w.WriteString(strconv.Itoa(q.limit))
	}
}

func (q selectQuery) writeNames(w Writer) {
	if len(q.names) == 0 {
		w.WriteRune('*')
		return
	}
	for i, name := range q.names {
		if i > 0 {
			w.WriteString(", ")
		}
		w.WriteName(name)
	}
}

func (q selectQuery) writeWhere(w Writer) {
	w.WriteString(" WHERE ")
	if q.where == nil {
		w.WriteString("1 = 1")
		return
	}
	q.where.WriteExpr(w)
}

func (q selectQuery) writeOrderBy(w Writer) {
	if len(q.orderBy) == 0 {
		return
	}
	w.WriteString(" ORDER BY ")
	for i, name := range q.orderBy {
		if i > 0 {
			w.WriteString(", ")
		}
		name.WriteExpr(w)
	}
}
