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

type BuildState interface {
	Builder() *builder
	Values() []interface{}
	AddValue(value interface{})
}

type buildState struct {
	builder *builder
	values  []interface{}
}

func (s buildState) Builder() *builder {
	return s.builder
}

func (s buildState) Values() []interface{} {
	return s.values
}

func (s *buildState) AddValue(value interface{}) {
	s.values = append(s.values, value)
}

type BoolExpr interface {
	And(BoolExpr) BoolExpr
	Or(BoolExpr) BoolExpr
	Build(BuildState) string
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

func (e binaryExpr) Build(state BuildState) string {
	var builder strings.Builder
	builder.WriteString(e.lhs.Build(state))
	switch e.kind {
	case orExpr:
		builder.WriteString(" OR ")
	case andExpr:
		builder.WriteString(" AND ")
	default:
		panic(fmt.Errorf("unsupported binary expr %q", e.kind))
	}
	builder.WriteString(e.rhs.Build(state))
	return builder.String()
}

type Value interface {
	Equal(interface{}) BoolExpr
	NotEqual(interface{}) BoolExpr
	Less(interface{}) BoolExpr
	Greater(interface{}) BoolExpr
	LessEqual(interface{}) BoolExpr
	GreaterEqual(interface{}) BoolExpr
	Build(BuildState) string
}

type Column string

func (c Column) Equal(o interface{}) BoolExpr {
	return cmp{kind: eqCmp, lhs: c, rhs: wrapValue(o)}
}

func (c Column) NotEqual(o interface{}) BoolExpr {
	return cmp{kind: notEqCmp, lhs: c, rhs: wrapValue(o)}
}

func (c Column) Less(o interface{}) BoolExpr {
	return cmp{kind: lessCmp, lhs: c, rhs: wrapValue(o)}
}

func (c Column) Greater(o interface{}) BoolExpr {
	return cmp{kind: greaterCmp, lhs: c, rhs: wrapValue(o)}
}

func (c Column) LessEqual(o interface{}) BoolExpr {
	return cmp{kind: lessEqualCmp, lhs: c, rhs: wrapValue(o)}
}

func (c Column) GreaterEqual(o interface{}) BoolExpr {
	return cmp{kind: greaterEqualCmp, lhs: c, rhs: wrapValue(o)}
}

func (c Column) Build(state BuildState) string {
	return state.Builder().buildName(string(c))
}

type value struct {
	value interface{}
}

func (v value) Equal(o interface{}) BoolExpr {
	return cmp{kind: eqCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) NotEqual(o interface{}) BoolExpr {
	return cmp{kind: notEqCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) Less(o interface{}) BoolExpr {
	return cmp{kind: lessCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) Greater(o interface{}) BoolExpr {
	return cmp{kind: greaterCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) LessEqual(o interface{}) BoolExpr {
	return cmp{kind: lessEqualCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) GreaterEqual(o interface{}) BoolExpr {
	return cmp{kind: greaterEqualCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) Build(state BuildState) string {
	state.AddValue(v.value)
	return state.Builder().buildOpt(len(state.Values()))
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

func (c cmp) Build(state BuildState) string {
	var builder strings.Builder
	builder.WriteString(c.lhs.Build(state))
	switch c.kind {
	case eqCmp:
		if isNullValue(c.rhs) {
			builder.WriteString(" IS NULL")
			return builder.String()
		}
		builder.WriteString(" = ")
	case notEqCmp:
		if isNullValue(c.rhs) {
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
	builder.WriteString(c.rhs.Build(state))
	return builder.String()
}

func wrapValue(val interface{}) Value {
	if v, ok := val.(Value); ok {
		return v
	}
	return value{value: val}
}

func isNullValue(val Value) bool {
	v, ok := val.(value)
	return ok && v.value == nil
}
