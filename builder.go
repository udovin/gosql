package gosql

import (
	"fmt"
	"strings"
)

// Query represents SQL query.
type Query interface {
	WriteQuery(w Writer)
}

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
	// Build renders query string and values.
	Build(query Query) (string, []any)
	// BuildString formats query string.
	BuildString(query Query) string
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
	return &selectQuery{table: table}
}

func (b *builder) Update(table string) UpdateQuery {
	return &updateQuery{table: table}
}

func (b *builder) Delete(table string) DeleteQuery {
	return &deleteQuery{table: table}
}

func (b *builder) Insert(table string) InsertQuery {
	switch b.dialect {
	case PostgresDialect:
		return &PostgresInsertQuery{
			insertQuery: insertQuery{table: table},
		}
	default:
		return &insertQuery{table: table}
	}
}

func (b *builder) Build(query Query) (string, []any) {
	builder := &writer{builder: b}
	query.WriteQuery(builder)
	return builder.String(), builder.Values()
}

func (b *builder) BuildString(query Query) string {
	str, _ := b.Build(query)
	return str
}

func (b builder) formatName(name string) string {
	return fmt.Sprintf("%q", name)
}

func (b builder) formatOpt(n int) string {
	return fmt.Sprintf("$%d", n)
}

// Writer is used for building query string with specified values.
type Writer interface {
	WriteRune(r rune)
	WriteString(s string)
	WriteName(n string)
	WriteValue(v any)
	String() string
	Values() []any
}

type writer struct {
	builder *builder
	query   strings.Builder
	values  []any
}

func (w *writer) WriteRune(r rune) {
	w.query.WriteRune(r)
}

func (w *writer) WriteString(str string) {
	w.query.WriteString(str)
}

func (w *writer) WriteName(name string) {
	w.query.WriteString(w.builder.formatName(name))
}

func (w *writer) WriteValue(value any) {
	w.values = append(w.values, value)
	w.query.WriteString(w.builder.formatOpt(len(w.values)))
}

func (w *writer) String() string {
	return w.query.String()
}

func (w *writer) Values() []any {
	return w.values
}

// Expr represents buildable expression.
type Expr interface {
	WriteExpr(Writer)
}

// BoolExpr represents boolean expression.
type BoolExpr interface {
	Expr
	And(BoolExpr) BoolExpr
	Or(BoolExpr) BoolExpr
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

func (e binaryExpr) formatPart(w Writer, expr BoolExpr) {
	if part, ok := expr.(binaryExpr); ok && part.kind != e.kind {
		w.WriteRune('(')
		expr.WriteExpr(w)
		w.WriteRune(')')
	} else {
		expr.WriteExpr(w)
	}
}

func (e binaryExpr) WriteExpr(w Writer) {
	e.formatPart(w, e.lhs)
	switch e.kind {
	case orExpr:
		w.WriteString(" OR ")
	case andExpr:
		w.WriteString(" AND ")
	default:
		panic(fmt.Errorf("unsupported binary expression: %d", e.kind))
	}
	e.formatPart(w, e.rhs)
}

// Value represents comparable value.
type Value interface {
	Expr
	Equal(any) BoolExpr
	NotEqual(any) BoolExpr
	Less(any) BoolExpr
	Greater(any) BoolExpr
	LessEqual(any) BoolExpr
	GreaterEqual(any) BoolExpr
}

// Column represents comparable table column.
type Column string

// Equal build boolean expression: "column = value".
func (c Column) Equal(o any) BoolExpr {
	return cmp{kind: eqCmp, lhs: c, rhs: wrapValue(o)}
}

// NotEqual build boolean expression: "column <> value".
func (c Column) NotEqual(o any) BoolExpr {
	return cmp{kind: notEqCmp, lhs: c, rhs: wrapValue(o)}
}

// Less build boolean expression: "column < value".
func (c Column) Less(o any) BoolExpr {
	return cmp{kind: lessCmp, lhs: c, rhs: wrapValue(o)}
}

// Greater build boolean expression: "column > value".
func (c Column) Greater(o any) BoolExpr {
	return cmp{kind: greaterCmp, lhs: c, rhs: wrapValue(o)}
}

// LessEqual build boolean expression: "column <= value".
func (c Column) LessEqual(o any) BoolExpr {
	return cmp{kind: lessEqualCmp, lhs: c, rhs: wrapValue(o)}
}

// GreaterEqual build boolean expression: "column >= value".
func (c Column) GreaterEqual(o any) BoolExpr {
	return cmp{kind: greaterEqualCmp, lhs: c, rhs: wrapValue(o)}
}

func (c Column) WriteExpr(w Writer) {
	w.WriteName(string(c))
}

type value struct {
	value any
}

func (v value) Equal(o any) BoolExpr {
	return cmp{kind: eqCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) NotEqual(o any) BoolExpr {
	return cmp{kind: notEqCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) Less(o any) BoolExpr {
	return cmp{kind: lessCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) Greater(o any) BoolExpr {
	return cmp{kind: greaterCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) LessEqual(o any) BoolExpr {
	return cmp{kind: lessEqualCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) GreaterEqual(o any) BoolExpr {
	return cmp{kind: greaterEqualCmp, lhs: v, rhs: wrapValue(o)}
}

func (v value) WriteExpr(w Writer) {
	w.WriteValue(v.value)
}

type Order int

const (
	AscendingOrder Order = iota
	DescendingOrder
)

type OrderExpr interface {
	Expr
	Order() Order
	Expr() Expr
}

type order struct {
	kind Order
	expr Expr
}

// Order returns order of expression.
func (e order) Order() Order {
	return e.kind
}

// Expr returns wrapped expression.
func (e order) Expr() Expr {
	return e.expr
}

func (e order) WriteExpr(w Writer) {
	e.expr.WriteExpr(w)
	switch e.kind {
	case AscendingOrder:
		w.WriteString(" ASC")
	case DescendingOrder:
		w.WriteString(" DESC")
	default:
		panic(fmt.Errorf("unsupported order: %d", e.kind))
	}
}

// Ascending represents ascending order of sorting.
func Ascending(val any) OrderExpr {
	switch v := val.(type) {
	case OrderExpr:
		return order{kind: AscendingOrder, expr: v.Expr()}
	default:
		return order{kind: AscendingOrder, expr: wrapExpression(v)}
	}
}

// Descending represents descending order of sorting.
func Descending(val any) OrderExpr {
	switch v := val.(type) {
	case OrderExpr:
		return order{kind: DescendingOrder, expr: v.Expr()}
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

func (c cmp) Or(o BoolExpr) BoolExpr {
	return binaryExpr{kind: orExpr, lhs: c, rhs: o}
}

func (c cmp) And(o BoolExpr) BoolExpr {
	return binaryExpr{kind: andExpr, lhs: c, rhs: o}
}

func (c cmp) WriteExpr(w Writer) {
	c.lhs.WriteExpr(w)
	switch c.kind {
	case eqCmp:
		if isNullValue(c.rhs) {
			w.WriteString(" IS NULL")
			return
		}
		w.WriteString(" = ")
	case notEqCmp:
		if isNullValue(c.rhs) {
			w.WriteString(" IS NOT NULL")
			return
		}
		w.WriteString(" <> ")
	case lessCmp:
		w.WriteString(" < ")
	case greaterCmp:
		w.WriteString(" > ")
	case lessEqualCmp:
		w.WriteString(" <= ")
	case greaterEqualCmp:
		w.WriteString(" >= ")
	default:
		panic(fmt.Errorf("unsupported binaryExpr %q", c.kind))
	}
	c.rhs.WriteExpr(w)
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

func wrapExpression(val any) Expr {
	switch v := val.(type) {
	case Expr:
		return v
	case string:
		return Column(v)
	default:
		panic(fmt.Errorf("unsupported type: %T", v))
	}
}

func wrapOrderExpression(val any) OrderExpr {
	switch v := val.(type) {
	case OrderExpr:
		return v
	default:
		return Ascending(wrapExpression(v))
	}
}
