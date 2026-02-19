// Package financeql implements FinanceQL — a PromQL-inspired domain-specific
// query language for financial time-series data. It provides a lexer, recursive
// descent parser, AST representation, evaluator, and built-in function library
// for querying stocks, technical indicators, fundamentals, and screening criteria.
package financeql

import (
	"fmt"
	"time"
)

// ════════════════════════════════════════════════════════════════════
// Value Types
// ════════════════════════════════════════════════════════════════════

// ValueType enumerates the possible result types of a FinanceQL expression.
type ValueType int

const (
	TypeScalar  ValueType = iota // single float64
	TypeString                   // single string
	TypeVector                   // time-series []TimePoint
	TypeMatrix                   // multi-stock map[string][]TimePoint
	TypeTable                    // tabular data []map[string]interface{}
	TypeBool                     // boolean
	TypeNil                      // no value / void
)

func (v ValueType) String() string {
	switch v {
	case TypeScalar:
		return "Scalar"
	case TypeString:
		return "String"
	case TypeVector:
		return "Vector"
	case TypeMatrix:
		return "Matrix"
	case TypeTable:
		return "Table"
	case TypeBool:
		return "Bool"
	case TypeNil:
		return "Nil"
	default:
		return "Unknown"
	}
}

// TimePoint is a single data point in a time-series.
type TimePoint struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

// Value is the universal result wrapper for FinanceQL expression evaluation.
type Value struct {
	Type   ValueType                `json:"type"`
	Scalar float64                  `json:"scalar,omitempty"`
	Str    string                   `json:"str,omitempty"`
	Bool   bool                     `json:"bool,omitempty"`
	Vector []TimePoint              `json:"vector,omitempty"`
	Matrix map[string][]TimePoint   `json:"matrix,omitempty"`
	Table  []map[string]interface{} `json:"table,omitempty"`
}

// ScalarValue creates a scalar Value.
func ScalarValue(v float64) Value {
	return Value{Type: TypeScalar, Scalar: v}
}

// StringValue creates a string Value.
func StringValue(s string) Value {
	return Value{Type: TypeString, Str: s}
}

// BoolValue creates a boolean Value.
func BoolValue(b bool) Value {
	return Value{Type: TypeBool, Bool: b}
}

// VectorValue creates a time-series Value.
func VectorValue(v []TimePoint) Value {
	return Value{Type: TypeVector, Vector: v}
}

// MatrixValue creates a multi-stock Value.
func MatrixValue(m map[string][]TimePoint) Value {
	return Value{Type: TypeMatrix, Matrix: m}
}

// TableValue creates a tabular Value.
func TableValue(rows []map[string]interface{}) Value {
	return Value{Type: TypeTable, Table: rows}
}

// NilValue creates a nil/void Value.
func NilValue() Value {
	return Value{Type: TypeNil}
}

// ════════════════════════════════════════════════════════════════════
// AST Node Types
// ════════════════════════════════════════════════════════════════════

// Node is the interface for all AST nodes.
type Node interface {
	nodeType() string
	// Pos returns the position (byte offset) in the original source.
	Pos() int
	String() string
}

// ────────────────────────────────────────────────────────────────────
// Literal Nodes
// ────────────────────────────────────────────────────────────────────

// NumberLiteral represents a numeric constant (e.g. 14, 3.14, 10000cr).
type NumberLiteral struct {
	Position int
	Value    float64
	Raw      string // original text including suffix e.g. "10000cr"
}

func (n *NumberLiteral) nodeType() string { return "NumberLiteral" }
func (n *NumberLiteral) Pos() int         { return n.Position }
func (n *NumberLiteral) String() string   { return n.Raw }

// StringLiteral represents a quoted string (e.g. "IT", "oversold").
type StringLiteral struct {
	Position int
	Value    string
}

func (n *StringLiteral) nodeType() string { return "StringLiteral" }
func (n *StringLiteral) Pos() int         { return n.Position }
func (n *StringLiteral) String() string   { return fmt.Sprintf("%q", n.Value) }

// Identifier represents a bare name — ticker symbol, keyword, field name, or wildcard.
type Identifier struct {
	Position int
	Name     string // e.g. "RELIANCE", "sector", "desc", "*"
}

func (n *Identifier) nodeType() string { return "Identifier" }
func (n *Identifier) Pos() int         { return n.Position }
func (n *Identifier) String() string   { return n.Name }

// BoolLiteral represents true/false.
type BoolLiteral struct {
	Position int
	Value    bool
}

func (n *BoolLiteral) nodeType() string { return "BoolLiteral" }
func (n *BoolLiteral) Pos() int         { return n.Position }
func (n *BoolLiteral) String() string {
	if n.Value {
		return "true"
	}
	return "false"
}

// ────────────────────────────────────────────────────────────────────
// Expression Nodes
// ────────────────────────────────────────────────────────────────────

// FunctionCall represents a function invocation e.g. rsi(RELIANCE, 14).
type FunctionCall struct {
	Position int
	Name     string // function name, lower-cased
	Args     []Node // arguments
}

func (n *FunctionCall) nodeType() string { return "FunctionCall" }
func (n *FunctionCall) Pos() int         { return n.Position }
func (n *FunctionCall) String() string {
	s := n.Name + "("
	for i, a := range n.Args {
		if i > 0 {
			s += ", "
		}
		s += a.String()
	}
	return s + ")"
}

// RangeSelector represents a time range on an expression e.g. [30d], [90d], [1w].
type RangeSelector struct {
	Position int
	Expr     Node   // the expression being ranged
	Duration string // raw duration string e.g. "30d", "1w", "252d"
	Days     int    // parsed number of calendar days
}

func (n *RangeSelector) nodeType() string { return "RangeSelector" }
func (n *RangeSelector) Pos() int         { return n.Position }
func (n *RangeSelector) String() string   { return fmt.Sprintf("%s[%s]", n.Expr.String(), n.Duration) }

// BinaryExpr represents a binary operation e.g. a + b, pe < 15, a AND b.
type BinaryExpr struct {
	Position int
	Op       string // "+", "-", "*", "/", ">", "<", ">=", "<=", "==", "!=", "AND", "OR"
	Left     Node
	Right    Node
}

func (n *BinaryExpr) nodeType() string { return "BinaryExpr" }
func (n *BinaryExpr) Pos() int         { return n.Position }
func (n *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Op, n.Right.String())
}

// UnaryExpr represents a unary operation e.g. NOT x, -5.
type UnaryExpr struct {
	Position int
	Op       string // "NOT", "-"
	Operand  Node
}

func (n *UnaryExpr) nodeType() string { return "UnaryExpr" }
func (n *UnaryExpr) Pos() int         { return n.Position }
func (n *UnaryExpr) String() string   { return fmt.Sprintf("(%s %s)", n.Op, n.Operand.String()) }

// PipeExpr represents a pipe composition e.g. expr | func(...).
type PipeExpr struct {
	Position int
	Left     Node // upstream expression
	Right    Node // downstream function call (receives left as first arg)
}

func (n *PipeExpr) nodeType() string { return "PipeExpr" }
func (n *PipeExpr) Pos() int         { return n.Position }
func (n *PipeExpr) String() string   { return fmt.Sprintf("%s | %s", n.Left.String(), n.Right.String()) }

// ScreenerExpr represents a screener(...) call with filter predicates.
type ScreenerExpr struct {
	Position int
	Filter   Node // the filter predicate (typically a BinaryExpr tree)
}

func (n *ScreenerExpr) nodeType() string { return "ScreenerExpr" }
func (n *ScreenerExpr) Pos() int         { return n.Position }
func (n *ScreenerExpr) String() string   { return fmt.Sprintf("screener(%s)", n.Filter.String()) }

// AlertExpr represents an alert(...) call.
type AlertExpr struct {
	Position  int
	Condition Node   // the alert condition
	Message   string // alert message
}

func (n *AlertExpr) nodeType() string { return "AlertExpr" }
func (n *AlertExpr) Pos() int         { return n.Position }
func (n *AlertExpr) String() string {
	return fmt.Sprintf("alert(%s, %q)", n.Condition.String(), n.Message)
}

// ════════════════════════════════════════════════════════════════════
// Parse Error
// ════════════════════════════════════════════════════════════════════

// ParseError captures parsing errors with position context.
type ParseError struct {
	Position int
	Line     int
	Column   int
	Message  string
	Hint     string // optional suggestion
}

func (e *ParseError) Error() string {
	loc := fmt.Sprintf("line %d, col %d", e.Line, e.Column)
	msg := fmt.Sprintf("parse error at %s: %s", loc, e.Message)
	if e.Hint != "" {
		msg += " (hint: " + e.Hint + ")"
	}
	return msg
}
