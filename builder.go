package gosql

import (
	"fmt"
	"strings"
)

// Builder represents SQL query builder.
type Builder struct {}

// Select creates a new select query.
func (b *Builder) Select(values... Value) SelectQuery {
	return selectQuery{values: values}
}

// Update creates a new update query.
func (b *Builder) Update(table string) UpdateQuery {
	return updateQuery{table: table}
}

// Delete creates a new delete query.
func (b *Builder) Delete(table string) DeleteQuery {
	return DeleteQuery{table: table}
}

// Insert creates a new insert query.
func (b *Builder) Insert(table string) InsertQuery {
	return InsertQuery{table: table}
}

type Query interface {
	Build() (string, []interface{})
	String() string
}

// DeleteQuery represents SQL delete query.
type DeleteQuery struct {
	table string
	where BoolExpr
}

func (q DeleteQuery) Build() (string, []interface{}) {
	query := fmt.Sprintf(
		"DELETE FROM %q WHERE %s",
		q.table,
		"",
	)
	return query, nil
}

func (q DeleteQuery) String() string {
	query, _ := q.Build()
	return query
}

type InsertQuery struct {
	table string
}

type state struct {
	values []interface{}
}

type BoolExpr interface {
	And(BoolExpr) BoolExpr
	Or(BoolExpr) BoolExpr
	build(*[]interface{}) string
}

type exprKind int

const (
	orExpr exprKind = iota
	andExpr
)

type binaryExpr struct {
	kind     exprKind
	lhs, rhs BoolExpr
}

func (e binaryExpr) Or(o BoolExpr) BoolExpr {
	return binaryExpr{kind: orExpr, lhs: e, rhs: o}
}

func (e binaryExpr) And(o BoolExpr) BoolExpr {
	return binaryExpr{kind: andExpr, lhs: e, rhs: o}
}

func (e binaryExpr) build(opts *[]interface{}) string {
	var builder strings.Builder
	builder.WriteString(e.lhs.build(opts))
	switch e.kind {
	case orExpr:
		builder.WriteString(" OR ")
	case andExpr:
		builder.WriteString(" AND ")
	default:
		panic(fmt.Errorf("unsupported binaryExpr %q", e.kind))
	}
	builder.WriteString(e.rhs.build(opts))
	return builder.String()
}

type Value interface {
	Equal(interface{}) BoolExpr
	NotEqual(interface{}) BoolExpr
	Less(interface{}) BoolExpr
	Greater(interface{}) BoolExpr
	LessEqual(interface{}) BoolExpr
	GreaterEqual(interface{}) BoolExpr
	build(*[]interface{}) string
}

type Column string

func (c Column) Equal(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: eqCmp, lhs: c, rhs: o.(Value)}
}

func (c Column) NotEqual(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: notEqCmp, lhs: c, rhs: o.(Value)}
}

func (c Column) Less(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: lessCmp, lhs: c, rhs: o.(Value)}
}

func (c Column) Greater(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: greaterCmp, lhs: c, rhs: o.(Value)}
}

func (c Column) LessEqual(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: lessEqualCmp, lhs: c, rhs: o.(Value)}
}

func (c Column) GreaterEqual(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: greaterEqualCmp, lhs: c, rhs: o.(Value)}
}

func (c Column) build(*[]interface{}) string {
	return fmt.Sprintf("%q", c)
}

type value struct {
	value interface{}
}

func (v value) Equal(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: eqCmp, lhs: v, rhs: o.(Value)}
}

func (v value) NotEqual(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: notEqCmp, lhs: v, rhs: o.(Value)}
}

func (v value) build(opts *[]interface{}) string {
	*opts = append(*opts, v.value)
	return fmt.Sprintf("$%d", len(*opts))
}

func (v value) Less(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: lessCmp, lhs: v, rhs: o.(Value)}
}

func (v value) Greater(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: greaterCmp, lhs: v, rhs: o.(Value)}
}

func (v value) LessEqual(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: lessEqualCmp, lhs: v, rhs: o.(Value)}
}

func (v value) GreaterEqual(o interface{}) BoolExpr {
	if _, ok := o.(Value); !ok {
		o = value{value: o}
	}
	return cmp{kind: greaterEqualCmp, lhs: v, rhs: o.(Value)}
}

type cmpKind int

const (
	eqCmp cmpKind = iota
	notEqCmp
	lessCmp
	greaterCmp
	lessEqualCmp
	greaterEqualCmp
)

type cmp struct {
	kind cmpKind
	lhs, rhs Value
}

func (c cmp) Or(o BoolExpr) BoolExpr {
	return binaryExpr{kind: orExpr, lhs: c, rhs: o}
}

func (c cmp) And(o BoolExpr) BoolExpr {
	return binaryExpr{kind: andExpr, lhs: c, rhs: o}
}

func (c cmp) build(opts *[]interface{}) string {
	var builder strings.Builder
	builder.WriteString(c.lhs.build(opts))
	switch c.kind {
	case eqCmp:
		if val, ok := c.rhs.(value); ok && val.value == nil {
			builder.WriteString(" IS NULL")
			return builder.String()
		}
		builder.WriteString(" = ")
	case notEqCmp:
		if val, ok := c.rhs.(value); ok && val.value == nil {
			builder.WriteString(" IS NOT NULL")
			return builder.String()
		}
		builder.WriteString(" <> ")
	case lessCmp:
		builder.WriteString(" < ")
	case greaterCmp:
		builder.WriteString(" > ")
	case lessEqualCmp:
		builder.WriteString(" <= ")
	case greaterEqualCmp:
		builder.WriteString(" >= ")
	default:
		panic(fmt.Errorf("unsupported binaryExpr %q", c.kind))
	}
	builder.WriteString(c.rhs.build(opts))
	return builder.String()
}
