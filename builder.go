package gosql

import (
	"fmt"
	"strings"
)

// Builder represents SQL query builder.
type Builder interface {
	// Dialect returns SQL dialect.
	Dialect() Dialect
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
	Build() (string, []any)
	// String returns SQL query without values.
	String() string
}

// Dialect represents kind of SQL dialect.
type Dialect int

const (
	// SQLiteDialect represents SQLite dialect.
	SQLiteDialect Dialect = iota
	// PostgresDialect represents Postgres dialect.
	PostgresDialect
)

// NewBuilder creates a new instance of SQL builder.
func NewBuilder(dialect Dialect) Builder {
	return &builder{dialect: dialect}
}

type builder struct {
	dialect Dialect
}

func (b builder) Dialect() Dialect {
	return b.dialect
}

func (b *builder) Select(table string) SelectQuery {
	return &selectQuery{builder: b, table: table}
}

func (b *builder) Update(table string) UpdateQuery {
	return &updateQuery{builder: b, table: table}
}

func (b *builder) Delete(table string) DeleteQuery {
	return &deleteQuery{builder: b, table: table}
}

func (b *builder) Insert(table string) InsertQuery {
	return &insertQuery{builder: b, table: table}
}

func (b builder) buildName(name string) string {
	return fmt.Sprintf("%q", name)
}

func (b builder) buildOpt(n int) string {
	return fmt.Sprintf("$%d", n)
}

// RawBuilder is used for building query string with specified values.
type RawBuilder interface {
	WriteRune(rune)
	WriteString(string)
	WriteName(string)
	WriteValue(any)
	String() string
	Values() []any
}

type rawBuilder struct {
	builder *builder
	query   strings.Builder
	values  []any
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

func (s *rawBuilder) WriteValue(value any) {
	s.values = append(s.values, value)
	s.query.WriteString(s.builder.buildOpt(len(s.values)))
}

func (s rawBuilder) String() string {
	return s.query.String()
}

func (s rawBuilder) Values() []any {
	return s.values
}

// Expression represents buildable expression.
type Expression interface {
	Build(RawBuilder)
}

// BoolExpression represents boolean expression.
type BoolExpression interface {
	Expression
	And(BoolExpression) BoolExpression
	Or(BoolExpression) BoolExpression
}

type exprKind int

const (
	orExpr exprKind = iota
	andExpr
)

type binaryExpr struct {
	kind     exprKind
	lhs, rhs BoolExpression
}

func (e binaryExpr) Or(o BoolExpression) BoolExpression {
	return binaryExpr{kind: orExpr, lhs: e, rhs: o}
}

func (e binaryExpr) And(o BoolExpression) BoolExpression {
	return binaryExpr{kind: andExpr, lhs: e, rhs: o}
}

func (e binaryExpr) buildPart(builder RawBuilder, expr BoolExpression) {
	if part, ok := expr.(binaryExpr); ok && part.kind != e.kind {
		builder.WriteRune('(')
		expr.Build(builder)
		builder.WriteRune(')')
	} else {
		expr.Build(builder)
	}
}

func (e binaryExpr) Build(builder RawBuilder) {
	e.buildPart(builder, e.lhs)
	switch e.kind {
	case orExpr:
		builder.WriteString(" OR ")
	case andExpr:
		builder.WriteString(" AND ")
	default:
		panic(fmt.Errorf("unsupported binary expression: %d", e.kind))
	}
	e.buildPart(builder, e.rhs)
}

// Value represents comparable value.
type Value interface {
	Expression
	Equal(any) BoolExpression
	NotEqual(any) BoolExpression
	Less(any) BoolExpression
	Greater(any) BoolExpression
	LessEqual(any) BoolExpression
	GreaterEqual(any) BoolExpression
}

// Column represents comparable table column.
type Column string

// Equal build boolean expression: "column = value".
func (c Column) Equal(o any) BoolExpression {
	return cmp{kind: eqCmp, lhs: c, rhs: wrapValue(o)}
}

// NotEqual build boolean expression: "column <> value".
func (c Column) NotEqual(o any) BoolExpression {
	return cmp{kind: notEqCmp, lhs: c, rhs: wrapValue(o)}
}

// Less build boolean expression: "column < value".
func (c Column) Less(o any) BoolExpression {
	return cmp{kind: lessCmp, lhs: c, rhs: wrapValue(o)}
}

// Greater build boolean expression: "column > value".
func (c Column) Greater(o any) BoolExpression {
	return cmp{kind: greaterCmp, lhs: c, rhs: wrapValue(o)}
}

// LessEqual build boolean expression: "column <= value".
func (c Column) LessEqual(o any) BoolExpression {
	return cmp{kind: lessEqualCmp, lhs: c, rhs: wrapValue(o)}
}

// GreaterEqual build boolean expression: "column >= value".
func (c Column) GreaterEqual(o any) BoolExpression {
	return cmp{kind: greaterEqualCmp, lhs: c, rhs: wrapValue(o)}
}

func (c Column) Build(builder RawBuilder) {
	builder.WriteName(string(c))
}

type value struct {
	value any
}

func (v value) Equal(o any) BoolExpression {
	return cmp{kind: eqCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) NotEqual(o any) BoolExpression {
	return cmp{kind: notEqCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) Less(o any) BoolExpression {
	return cmp{kind: lessCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) Greater(o any) BoolExpression {
	return cmp{kind: greaterCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) LessEqual(o any) BoolExpression {
	return cmp{kind: lessEqualCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) GreaterEqual(o any) BoolExpression {
	return cmp{kind: greaterEqualCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) Build(builder RawBuilder) {
	builder.WriteValue(v.value)
}

type Order int

const (
	AscendingOrder Order = iota
	DescendingOrder
)

type OrderExpression interface {
	Expression
	Order() Order
	Expression() Expression
}

type order struct {
	kind Order
	expr Expression
}

// Order returns order of expression.
func (e order) Order() Order {
	return e.kind
}

// Expression returns wrapped expression.
func (e order) Expression() Expression {
	return e.expr
}

func (e order) Build(builder RawBuilder) {
	e.expr.Build(builder)
	switch e.kind {
	case AscendingOrder:
		builder.WriteString(" ASC")
	case DescendingOrder:
		builder.WriteString(" DESC")
	default:
		panic(fmt.Errorf("unsupported order: %d", e.kind))
	}
}

// Ascending represents ascending order of sorting.
func Ascending(val any) OrderExpression {
	switch v := val.(type) {
	case OrderExpression:
		return order{kind: AscendingOrder, expr: v.Expression()}
	default:
		return order{kind: AscendingOrder, expr: wrapExpression(v)}
	}
}

// Descending represents descending order of sorting.
func Descending(val any) OrderExpression {
	switch v := val.(type) {
	case OrderExpression:
		return order{kind: DescendingOrder, expr: v.Expression()}
	default:
		return order{kind: DescendingOrder, expr: wrapExpression(v)}
	}
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

func (c cmp) Or(o BoolExpression) BoolExpression {
	return binaryExpr{kind: orExpr, lhs: c, rhs: o}
}

func (c cmp) And(o BoolExpression) BoolExpression {
	return binaryExpr{kind: andExpr, lhs: c, rhs: o}
}

func (c cmp) Build(builder RawBuilder) {
	c.lhs.Build(builder)
	switch c.kind {
	case eqCmp:
		if isNullValue(c.rhs) {
			builder.WriteString(" IS NULL")
			return
		}
		builder.WriteString(" = ")
	case notEqCmp:
		if isNullValue(c.rhs) {
			builder.WriteString(" IS NOT NULL")
			return
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
	c.rhs.Build(builder)
}

func isNullValue(val Value) bool {
	v, ok := val.(value)
	return ok && v.value == nil
}

func wrapValue(val any) Value {
	if v, ok := val.(Value); ok {
		return v
	}
	return value{value: val}
}

func wrapExpression(val any) Expression {
	switch v := val.(type) {
	case Expression:
		return v
	case string:
		return Column(v)
	default:
		panic(fmt.Errorf("unsupported type: %T", v))
	}
}

func wrapOrderExpression(val any) OrderExpression {
	switch v := val.(type) {
	case OrderExpression:
		return v
	default:
		return Ascending(wrapExpression(v))
	}
}
