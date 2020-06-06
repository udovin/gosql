package gosql

import (
	"fmt"
	"strings"
)

// Builder represents SQL query builder.
type Builder interface {
	// Select creates a new select query.
	Select(table string) SelectQuery
	// Update creates a new update query.
	Update(table string) UpdateQuery
	// Delete creates a new delete query.
	Delete(table string) DeleteQuery
	// Insert creates a new insert query.
	Insert(table string) InsertQuery
}

// Query represents SQL query.
type Query interface {
	// Build generates SQL query and values.
	Build() (string, []interface{})
	// String returns SQL query without values.
	String() string
}

// NewBuilder creates a new instance of SQL builder.
func NewBuilder() Builder {
	return &builder{}
}

type builder struct{}

func (b *builder) Select(table string) SelectQuery {
	return selectQuery{builder: b, table: table}
}

func (b *builder) Update(table string) UpdateQuery {
	return updateQuery{builder: b, table: table}
}

func (b *builder) Delete(table string) DeleteQuery {
	return deleteQuery{builder: b, table: table}
}

func (b *builder) Insert(table string) InsertQuery {
	return insertQuery{builder: b, table: table}
}

func (b *builder) buildName(name string) string {
	return fmt.Sprintf("%q", name)
}

func (b *builder) buildOpt(n int) string {
	return fmt.Sprintf("$%d", n)
}

type state struct {
	values []interface{}
}

type BoolExpr interface {
	And(BoolExpr) BoolExpr
	Or(BoolExpr) BoolExpr
	build(*builder, *[]interface{}) string
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

func (e binaryExpr) build(b *builder, opts *[]interface{}) string {
	var builder strings.Builder
	builder.WriteString(e.lhs.build(b, opts))
	switch e.kind {
	case orExpr:
		builder.WriteString(" OR ")
	case andExpr:
		builder.WriteString(" AND ")
	default:
		panic(fmt.Errorf("unsupported binaryExpr %q", e.kind))
	}
	builder.WriteString(e.rhs.build(b, opts))
	return builder.String()
}

type Value interface {
	Equal(interface{}) BoolExpr
	NotEqual(interface{}) BoolExpr
	Less(interface{}) BoolExpr
	Greater(interface{}) BoolExpr
	LessEqual(interface{}) BoolExpr
	GreaterEqual(interface{}) BoolExpr
	build(*builder, *[]interface{}) string
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

func (c Column) build(b *builder, _ *[]interface{}) string {
	return b.buildName(string(c))
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

func (v value) build(b *builder, opts *[]interface{}) string {
	*opts = append(*opts, v.value)
	return b.buildOpt(len(*opts))
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
	kind     cmpKind
	lhs, rhs Value
}

func (c cmp) Or(o BoolExpr) BoolExpr {
	return binaryExpr{kind: orExpr, lhs: c, rhs: o}
}

func (c cmp) And(o BoolExpr) BoolExpr {
	return binaryExpr{kind: andExpr, lhs: c, rhs: o}
}

func (c cmp) build(b *builder, opts *[]interface{}) string {
	var builder strings.Builder
	builder.WriteString(c.lhs.build(b, opts))
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
	builder.WriteString(c.rhs.build(b, opts))
	return builder.String()
}
