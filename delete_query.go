package gosql

// DeleteQuery represents SQL delete query.
type DeleteQuery interface {
	Query
	SetTable(table string)
	SetWhere(where BoolExpr)
}

type deleteQuery struct {
	table string
	where BoolExpr
}

func (q *deleteQuery) SetTable(table string) {
	q.table = table
}

func (q *deleteQuery) SetWhere(where BoolExpr) {
	q.where = where
}

func (q deleteQuery) WriteQuery(w Writer) {
	w.WriteString("DELETE FROM ")
	w.WriteName(q.table)
	q.writeWhere(w)
}

func (q deleteQuery) writeWhere(w Writer) {
	w.WriteString(" WHERE ")
	if q.where == nil {
		w.WriteString("1 = 1")
		return
	}
	q.where.WriteExpr(w)
}
