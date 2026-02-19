package financeql

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Evaluator
// ════════════════════════════════════════════════════════════════════

// BuiltinFunc is the signature for registered FinanceQL functions.
// The evaluator calls it with the already-evaluated arguments.
type BuiltinFunc func(ctx *EvalContext, args []Value) (Value, error)

// EvalContext carries runtime state during expression evaluation.
type EvalContext struct {
	Ctx        context.Context
	Aggregator *datasource.Aggregator // data source
	Functions  map[string]BuiltinFunc // registered functions
	Cache      *EvalCache             // query cache
	PipeInput  *Value                 // upstream value from pipe (nil if none)
}

// NewEvalContext creates an evaluation context with the given aggregator and defaults.
func NewEvalContext(ctx context.Context, agg *datasource.Aggregator) *EvalContext {
	ec := &EvalContext{
		Ctx:        ctx,
		Aggregator: agg,
		Functions:  make(map[string]BuiltinFunc),
		Cache:      NewEvalCache(5 * time.Minute),
	}
	RegisterBuiltins(ec)
	return ec
}

// RegisterFunc registers a function by name (lower-cased).
func (ec *EvalContext) RegisterFunc(name string, fn BuiltinFunc) {
	ec.Functions[strings.ToLower(name)] = fn
}

// ════════════════════════════════════════════════════════════════════
// Cache
// ════════════════════════════════════════════════════════════════════

// EvalCache provides a simple TTL cache for resolved data.
type EvalCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

type cacheEntry struct {
	value     Value
	expiresAt time.Time
}

// NewEvalCache creates a cache with the given TTL.
func NewEvalCache(ttl time.Duration) *EvalCache {
	return &EvalCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cached value. Returns ok=false if missing or expired.
func (c *EvalCache) Get(key string) (Value, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		return NilValue(), false
	}
	return e.value, true
}

// Set stores a value in the cache.
func (c *EvalCache) Set(key string, val Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{value: val, expiresAt: time.Now().Add(c.ttl)}
}

// ════════════════════════════════════════════════════════════════════
// Evaluator — AST Walker
// ════════════════════════════════════════════════════════════════════

// Eval evaluates an AST node and returns a Value.
func Eval(ec *EvalContext, node Node) (Value, error) {
	if node == nil {
		return NilValue(), nil
	}

	switch n := node.(type) {
	case *NumberLiteral:
		return ScalarValue(n.Value), nil

	case *StringLiteral:
		return StringValue(n.Value), nil

	case *BoolLiteral:
		return BoolValue(n.Value), nil

	case *Identifier:
		return evalIdentifier(ec, n)

	case *FunctionCall:
		return evalFunctionCall(ec, n)

	case *RangeSelector:
		return evalRangeSelector(ec, n)

	case *BinaryExpr:
		return evalBinaryExpr(ec, n)

	case *UnaryExpr:
		return evalUnaryExpr(ec, n)

	case *PipeExpr:
		return evalPipeExpr(ec, n)

	case *ScreenerExpr:
		return evalScreenerExpr(ec, n)

	case *AlertExpr:
		return evalAlertExpr(ec, n)

	default:
		return NilValue(), fmt.Errorf("unsupported AST node type: %T", node)
	}
}

// EvalQuery is the top-level convenience function: parse + evaluate.
func EvalQuery(ec *EvalContext, query string) (Value, error) {
	node, err := ParseQuery(query)
	if err != nil {
		return NilValue(), err
	}
	return Eval(ec, node)
}

// ────────────────────────────────────────────────────────────────────
// Node evaluators
// ────────────────────────────────────────────────────────────────────

func evalIdentifier(ec *EvalContext, n *Identifier) (Value, error) {
	// An identifier by itself could be:
	// 1. A ticker symbol — try to resolve latest price
	// 2. A field name in a pipe context
	name := n.Name

	if name == "*" {
		return StringValue("*"), nil
	}

	// In a pipe context, identifiers are field accessors
	if ec.PipeInput != nil {
		return StringValue(name), nil
	}

	// Try treating it as a ticker — resolve price
	if fn, ok := ec.Functions["price"]; ok {
		return fn(ec, []Value{StringValue(name)})
	}

	return StringValue(name), nil
}

func evalFunctionCall(ec *EvalContext, n *FunctionCall) (Value, error) {
	name := strings.ToLower(n.Name)

	fn, ok := ec.Functions[name]
	if !ok {
		return NilValue(), fmt.Errorf("unknown function %q at position %d", name, n.Position)
	}

	// Evaluate arguments
	args := make([]Value, len(n.Args))
	for i, argNode := range n.Args {
		// For function calls that take ticker names, pass identifiers as strings
		if ident, ok := argNode.(*Identifier); ok {
			args[i] = StringValue(ident.Name)
			continue
		}
		val, err := Eval(ec, argNode)
		if err != nil {
			return NilValue(), fmt.Errorf("error evaluating argument %d of %s: %w", i, name, err)
		}
		args[i] = val
	}

	// If we're in a pipe context, prepend the pipe input
	if ec.PipeInput != nil {
		args = append([]Value{*ec.PipeInput}, args...)
	}

	return fn(ec, args)
}

func evalRangeSelector(ec *EvalContext, n *RangeSelector) (Value, error) {
	// A range selector converts an instant query to a range query.
	// E.g., price(RELIANCE)[30d] → 30-day price time-series
	switch inner := n.Expr.(type) {
	case *FunctionCall:
		// Add a range argument
		rangeName := strings.ToLower(inner.Name) + "_range"
		if fn, ok := ec.Functions[rangeName]; ok {
			args := make([]Value, len(inner.Args))
			for i, argNode := range inner.Args {
				if ident, ok := argNode.(*Identifier); ok {
					args[i] = StringValue(ident.Name)
					continue
				}
				val, err := Eval(ec, argNode)
				if err != nil {
					return NilValue(), err
				}
				args[i] = val
			}
			args = append(args, ScalarValue(float64(n.Days)))
			return fn(ec, args)
		}

		// Fallback: evaluate inner as instant, and wrap
		val, err := Eval(ec, inner)
		if err != nil {
			return NilValue(), err
		}
		// If it's already a vector, slice it
		if val.Type == TypeVector && n.Days > 0 && len(val.Vector) > n.Days {
			return VectorValue(val.Vector[len(val.Vector)-n.Days:]), nil
		}
		return val, nil

	case *Identifier:
		// ticker[30d] → price range
		if fn, ok := ec.Functions["price_range"]; ok {
			return fn(ec, []Value{StringValue(inner.Name), ScalarValue(float64(n.Days))})
		}
		return NilValue(), fmt.Errorf("no range function available for identifier %q", inner.Name)

	default:
		val, err := Eval(ec, n.Expr)
		if err != nil {
			return NilValue(), err
		}
		return val, nil
	}
}

func evalBinaryExpr(ec *EvalContext, n *BinaryExpr) (Value, error) {
	left, err := Eval(ec, n.Left)
	if err != nil {
		return NilValue(), err
	}
	right, err := Eval(ec, n.Right)
	if err != nil {
		return NilValue(), err
	}

	switch n.Op {
	// Arithmetic
	case "+":
		return applyArithScalar(left, right, func(a, b float64) float64 { return a + b })
	case "-":
		return applyArithScalar(left, right, func(a, b float64) float64 { return a - b })
	case "*":
		return applyArithScalar(left, right, func(a, b float64) float64 { return a * b })
	case "/":
		return applyArithScalar(left, right, func(a, b float64) float64 {
			if b == 0 {
				return math.NaN()
			}
			return a / b
		})

	// Comparison
	case ">":
		return comparScalar(left, right, func(a, b float64) bool { return a > b })
	case "<":
		return comparScalar(left, right, func(a, b float64) bool { return a < b })
	case ">=":
		return comparScalar(left, right, func(a, b float64) bool { return a >= b })
	case "<=":
		return comparScalar(left, right, func(a, b float64) bool { return a <= b })
	case "==":
		return equalityCheck(left, right, false)
	case "!=":
		return equalityCheck(left, right, true)

	// Logical
	case "AND":
		return BoolValue(toBool(left) && toBool(right)), nil
	case "OR":
		return BoolValue(toBool(left) || toBool(right)), nil

	default:
		return NilValue(), fmt.Errorf("unknown operator %q", n.Op)
	}
}

func evalUnaryExpr(ec *EvalContext, n *UnaryExpr) (Value, error) {
	val, err := Eval(ec, n.Operand)
	if err != nil {
		return NilValue(), err
	}

	switch n.Op {
	case "-":
		if val.Type == TypeScalar {
			return ScalarValue(-val.Scalar), nil
		}
		return NilValue(), fmt.Errorf("cannot negate %s", val.Type)
	case "NOT":
		return BoolValue(!toBool(val)), nil
	default:
		return NilValue(), fmt.Errorf("unknown unary operator %q", n.Op)
	}
}

func evalPipeExpr(ec *EvalContext, n *PipeExpr) (Value, error) {
	leftVal, err := Eval(ec, n.Left)
	if err != nil {
		return NilValue(), err
	}

	// Create a new context with pipe input set
	pipeCtx := &EvalContext{
		Ctx:        ec.Ctx,
		Aggregator: ec.Aggregator,
		Functions:  ec.Functions,
		Cache:      ec.Cache,
		PipeInput:  &leftVal,
	}

	return Eval(pipeCtx, n.Right)
}

func evalScreenerExpr(ec *EvalContext, n *ScreenerExpr) (Value, error) {
	if fn, ok := ec.Functions["_screener"]; ok {
		// Pass the filter AST as a string and the raw node
		filterStr := n.Filter.String()
		return fn(ec, []Value{StringValue(filterStr)})
	}
	return NilValue(), fmt.Errorf("screener functionality not available")
}

func evalAlertExpr(ec *EvalContext, n *AlertExpr) (Value, error) {
	condVal, err := Eval(ec, n.Condition)
	if err != nil {
		return NilValue(), err
	}

	triggered := toBool(condVal)
	result := map[string]interface{}{
		"triggered": triggered,
		"message":   n.Message,
		"condition": n.Condition.String(),
	}
	return TableValue([]map[string]interface{}{result}), nil
}

// ════════════════════════════════════════════════════════════════════
// Helper functions for binary evaluation
// ════════════════════════════════════════════════════════════════════

func applyArithScalar(left, right Value, op func(float64, float64) float64) (Value, error) {
	a := toScalar(left)
	b := toScalar(right)
	return ScalarValue(op(a, b)), nil
}

func comparScalar(left, right Value, cmp func(float64, float64) bool) (Value, error) {
	a := toScalar(left)
	b := toScalar(right)
	return BoolValue(cmp(a, b)), nil
}

func equalityCheck(left, right Value, negate bool) (Value, error) {
	// String equality
	if left.Type == TypeString && right.Type == TypeString {
		eq := strings.EqualFold(left.Str, right.Str)
		if negate {
			eq = !eq
		}
		return BoolValue(eq), nil
	}
	// Scalar equality
	a := toScalar(left)
	b := toScalar(right)
	eq := a == b
	if negate {
		eq = !eq
	}
	return BoolValue(eq), nil
}

func toScalar(v Value) float64 {
	switch v.Type {
	case TypeScalar:
		return v.Scalar
	case TypeBool:
		if v.Bool {
			return 1
		}
		return 0
	case TypeVector:
		if len(v.Vector) > 0 {
			return v.Vector[len(v.Vector)-1].Value
		}
	}
	return 0
}

func toBool(v Value) bool {
	switch v.Type {
	case TypeBool:
		return v.Bool
	case TypeScalar:
		return v.Scalar != 0
	case TypeString:
		return v.Str != ""
	case TypeVector:
		return len(v.Vector) > 0
	case TypeTable:
		return len(v.Table) > 0
	case TypeNil:
		return false
	}
	return false
}

// ════════════════════════════════════════════════════════════════════
// Data Resolution Helpers
// ════════════════════════════════════════════════════════════════════

// ResolveTicker normalizes a ticker for NSE.
func ResolveTicker(ticker string) string {
	t := strings.TrimSpace(strings.ToUpper(ticker))
	// Remove .NS suffix for internal use
	t = strings.TrimSuffix(t, ".NS")
	return t
}

// FetchHistorical fetches OHLCV data with caching.
func FetchHistorical(ec *EvalContext, ticker string, days int) ([]models.OHLCV, error) {
	key := fmt.Sprintf("hist:%s:%d", ticker, days)
	if v, ok := ec.Cache.Get(key); ok && v.Type == TypeVector {
		// Convert back from vector to OHLCV — simplified, return from real source
		_ = v
	}

	to := time.Now()
	from := to.AddDate(0, 0, -days)

	data, err := ec.Aggregator.YFinance().GetHistoricalData(ec.Ctx, ticker, from, to, models.Timeframe1Day)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data for %s: %w", ticker, err)
	}
	return data, nil
}

// OHLCVToVector converts OHLCV data to a vector of closing prices.
func OHLCVToVector(data []models.OHLCV) []TimePoint {
	pts := make([]TimePoint, len(data))
	for i, d := range data {
		pts[i] = TimePoint{Time: d.Timestamp, Value: d.Close}
	}
	return pts
}
