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

// RawBuilder is used for building query string with specified values.
type RawBuilder interface {
	WriteRune(rune)
	WriteString(string)
	WriteName(string)
	WriteValue(interface{})
	String() string
	Values() []interface{}
}

type rawBuilder struct {
	builder *builder
	query   strings.Builder
	values  []interface{}
}

func (s *rawBuilder) WriteRune(r rune) {
	s.query.WriteRune(r)
}

func (s *rawBuilder) WriteString(str string) {
	s.query.WriteString(str)
}

func (s *rawBuilder) WriteName(name string) {
	s.query.WriteString(s.builder.buildName(name))
}

func (s *rawBuilder) WriteValue(value interface{}) {
	s.values = append(s.values, value)
	s.query.WriteString(s.builder.buildOpt(len(s.values)))
}

func (s rawBuilder) String() string {
	return s.query.String()
}

func (s rawBuilder) Values() []interface{} {
	return s.values
}

// BoolExpr represents boolean expression.
type BoolExpr interface {
	And(BoolExpr) BoolExpr
	Or(BoolExpr) BoolExpr
	Build(RawBuilder)
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

func (e binaryExpr) Build(state RawBuilder) {
	e.lhs.Build(state)
	switch e.kind {
	case orExpr:
		state.WriteString(" OR ")
	case andExpr:
		state.WriteString(" AND ")
	default:
		panic(fmt.Errorf("unsupported binary expr %q", e.kind))
	}
	e.rhs.Build(state)
}

// Value represents comparable value.
type Value interface {
	Equal(interface{}) BoolExpr
	NotEqual(interface{}) BoolExpr
	Less(interface{}) BoolExpr
	Greater(interface{}) BoolExpr
	LessEqual(interface{}) BoolExpr
	GreaterEqual(interface{}) BoolExpr
	Build(RawBuilder)
}

// Column represents comparable table column.
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

func (c Column) Build(state RawBuilder) {
	state.WriteName(string(c))
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

func (v value) Build(state RawBuilder) {
	state.WriteValue(v.value)
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

func (c cmp) Build(state RawBuilder) {
	c.lhs.Build(state)
	switch c.kind {
	case eqCmp:
		if isNullValue(c.rhs) {
			state.WriteString(" IS NULL")
			return
		}
		state.WriteString(" = ")
	case notEqCmp:
		if isNullValue(c.rhs) {
			state.WriteString(" IS NOT NULL")
			return
		}
		state.WriteString(" <> ")
	case lessCmp:
		state.WriteString(" < ")
	case greaterCmp:
		state.WriteString(" > ")
	case lessEqualCmp:
		state.WriteString(" <= ")
	case greaterEqualCmp:
		state.WriteString(" >= ")
	default:
		panic(fmt.Errorf("unsupported binaryExpr %q", c.kind))
	}
	c.rhs.Build(state)
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
