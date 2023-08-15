package gosql

// UpdateQuery represents SQL update query.
type UpdateQuery interface {
	Query
	SetWhere(where BoolExpr)
	SetNames(names ...string)
	SetValues(values ...any)
}

type updateQuery struct {
	table  string
	where  BoolExpr
	names  []string
	values []Value
}

func (q *updateQuery) SetWhere(where BoolExpr) {
	q.where = where
}

func (q *updateQuery) SetNames(names ...string) {
	q.names = names
}

func (q *updateQuery) SetValues(values ...any) {
	q.values = nil
	for _, val := range values {
		q.values = append(q.values, wrapValue(val))
	}
}

func (q updateQuery) WriteQuery(w Writer) {
	w.WriteString("UPDATE ")
	w.WriteName(q.table)
	q.writeSet(w)
	q.writeWhere(w)
}

func (q updateQuery) writeSet(w Writer) {
	if len(q.names) == 0 {
		panic("list of names can not be empty")
	}
	if len(q.names) != len(q.values) {
		panic("amount of names and values differs")
	}
	w.WriteString(" SET ")
	for i := range q.names {
		if i > 0 {
			w.WriteString(", ")
		}
		w.WriteName(q.names[i])
		w.WriteString(" = ")
		q.values[i].WriteExpr(w)
	}
}

func (q updateQuery) writeWhere(w Writer) {
	w.WriteString(" WHERE ")
	if q.where == nil {
		w.WriteString("1 = 1")
		return
	}
	q.where.WriteExpr(w)
}
