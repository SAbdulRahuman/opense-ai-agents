package financeql

import (
	"bytes"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// AST Value Tests
// ════════════════════════════════════════════════════════════════════

func TestValueConstructors(t *testing.T) {
	t.Run("ScalarValue", func(t *testing.T) {
		v := ScalarValue(42.5)
		assertEqual(t, TypeScalar, v.Type)
		assertFloat(t, 42.5, v.Scalar)
	})
	t.Run("StringValue", func(t *testing.T) {
		v := StringValue("hello")
		assertEqual(t, TypeString, v.Type)
		assertEqual(t, "hello", v.Str)
	})
	t.Run("BoolValue", func(t *testing.T) {
		v := BoolValue(true)
		assertEqual(t, TypeBool, v.Type)
		assertTrue(t, v.Bool)
	})
	t.Run("VectorValue", func(t *testing.T) {
		pts := []TimePoint{{Value: 1}, {Value: 2}, {Value: 3}}
		v := VectorValue(pts)
		assertEqual(t, TypeVector, v.Type)
		assertEqual(t, 3, len(v.Vector))
	})
	t.Run("TableValue", func(t *testing.T) {
		row := map[string]interface{}{"a": 1}
		v := TableValue([]map[string]interface{}{row})
		assertEqual(t, TypeTable, v.Type)
		assertEqual(t, 1, len(v.Table))
	})
	t.Run("NilValue", func(t *testing.T) {
		v := NilValue()
		assertEqual(t, TypeNil, v.Type)
	})
}

func TestValueTypeString(t *testing.T) {
	tests := []struct {
		vt   ValueType
		want string
	}{
		{TypeScalar, "Scalar"},
		{TypeString, "String"},
		{TypeVector, "Vector"},
		{TypeMatrix, "Matrix"},
		{TypeTable, "Table"},
		{TypeBool, "Bool"},
		{TypeNil, "Nil"},
	}
	for _, tt := range tests {
		assertEqual(t, tt.want, tt.vt.String())
	}
}

// ════════════════════════════════════════════════════════════════════
// Lexer Tests
// ════════════════════════════════════════════════════════════════════

func TestLexer_SimpleTokens(t *testing.T) {
	input := "+ - * / ( ) [ ] , |"
	tokens, err := NewLexer(input).Tokenize()
	assertNoErr(t, err)

	expected := []TokenType{
		TokenPlus, TokenMinus, TokenStar, TokenSlash,
		TokenLParen, TokenRParen, TokenLBracket, TokenRBracket,
		TokenComma, TokenPipe, TokenEOF,
	}
	assertEqual(t, len(expected), len(tokens))
	for i, exp := range expected {
		assertEqual(t, exp, tokens[i].Type)
	}
}

func TestLexer_ComparisonOperators(t *testing.T) {
	tests := []struct {
		input string
		want  TokenType
	}{
		{">", TokenGT},
		{"<", TokenLT},
		{">=", TokenGTE},
		{"<=", TokenLTE},
		{"==", TokenEQ},
		{"!=", TokenNEQ},
		{"=", TokenEQ}, // single = treated as ==
	}

	for _, tt := range tests {
		tokens, err := NewLexer(tt.input).Tokenize()
		assertNoErr(t, err)
		assertEqual(t, tt.want, tokens[0].Type)
	}
}

func TestLexer_Numbers(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"42", "42"},
		{"3.14", "3.14"},
		{"10000cr", "10000cr"},
		{"5l", "5l"},
		{"100crore", "100crore"},
		{"50lakh", "50lakh"},
		{".5", ".5"},
	}

	for _, tt := range tests {
		tokens, err := NewLexer(tt.input).Tokenize()
		assertNoErr(t, err)
		assertEqual(t, TokenNumber, tokens[0].Type)
		assertEqual(t, tt.value, tokens[0].Value)
	}
}

func TestLexer_Strings(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
		{`"IT sector"`, "IT sector"},
		{`"with \"escape\""`, `with "escape"`},
		{`"new\nline"`, "new\nline"},
	}

	for _, tt := range tests {
		tokens, err := NewLexer(tt.input).Tokenize()
		assertNoErr(t, err)
		assertEqual(t, TokenString, tokens[0].Type)
		assertEqual(t, tt.value, tokens[0].Value)
	}
}

func TestLexer_Keywords(t *testing.T) {
	input := "AND OR NOT"
	tokens, err := NewLexer(input).Tokenize()
	assertNoErr(t, err)
	assertEqual(t, TokenAND, tokens[0].Type)
	assertEqual(t, TokenOR, tokens[1].Type)
	assertEqual(t, TokenNOT, tokens[2].Type)
}

func TestLexer_KeywordsCaseInsensitive(t *testing.T) {
	input := "and or not And Or Not"
	tokens, err := NewLexer(input).Tokenize()
	assertNoErr(t, err)
	assertEqual(t, TokenAND, tokens[0].Type)
	assertEqual(t, TokenOR, tokens[1].Type)
	assertEqual(t, TokenNOT, tokens[2].Type)
	assertEqual(t, TokenAND, tokens[3].Type)
	assertEqual(t, TokenOR, tokens[4].Type)
	assertEqual(t, TokenNOT, tokens[5].Type)
}

func TestLexer_Identifiers(t *testing.T) {
	input := "RELIANCE sma price_range sector_IT"
	tokens, err := NewLexer(input).Tokenize()
	assertNoErr(t, err)
	assertEqual(t, TokenIdentifier, tokens[0].Type)
	assertEqual(t, "RELIANCE", tokens[0].Value)
	assertEqual(t, "sma", tokens[1].Value)
	assertEqual(t, "price_range", tokens[2].Value)
	assertEqual(t, "sector_IT", tokens[3].Value)
}

func TestLexer_Comments(t *testing.T) {
	input := "42 # this is a comment\n+ 5"
	tokens, err := NewLexer(input).Tokenize()
	assertNoErr(t, err)
	// Should get: 42, +, 5, EOF (comment stripped)
	assertEqual(t, TokenNumber, tokens[0].Type)
	assertEqual(t, TokenPlus, tokens[1].Type)
	assertEqual(t, TokenNumber, tokens[2].Type)
}

func TestLexer_ComplexQuery(t *testing.T) {
	input := `screener(rsi(*, 14) < 30 AND pe(*) < 20)`
	tokens, err := NewLexer(input).Tokenize()
	assertNoErr(t, err)

	// screener ( rsi ( * , 14 ) < 30 AND pe ( * ) < 20 )
	types := []TokenType{
		TokenIdentifier, TokenLParen, TokenIdentifier, TokenLParen,
		TokenStar, TokenComma, TokenNumber, TokenRParen, TokenLT,
		TokenNumber, TokenAND, TokenIdentifier, TokenLParen,
		TokenStar, TokenRParen, TokenLT, TokenNumber, TokenRParen,
		TokenEOF,
	}
	assertEqual(t, len(types), len(tokens))
	for i, exp := range types {
		if tokens[i].Type != exp {
			t.Errorf("token[%d]: want %s, got %s (%q)", i, exp, tokens[i].Type, tokens[i].Value)
		}
	}
}

func TestLexer_UnterminatedString(t *testing.T) {
	_, err := NewLexer(`"unterminated`).Tokenize()
	if err == nil {
		t.Fatal("expected error for unterminated string")
	}
	assertTrue(t, strings.Contains(err.Error(), "unterminated"))
}

func TestLexer_IllegalBang(t *testing.T) {
	_, err := NewLexer("!x").Tokenize()
	if err == nil {
		t.Fatal("expected error for lone !")
	}
}

func TestLexer_LineColTracking(t *testing.T) {
	input := "a\nb"
	tokens, err := NewLexer(input).Tokenize()
	assertNoErr(t, err)
	assertEqual(t, 1, tokens[0].Line) // a is on line 1
	assertEqual(t, 2, tokens[1].Line) // b is on line 2
}

func TestLexer_PipeQuery(t *testing.T) {
	input := "price(RELIANCE) | sma(*, 50)"
	tokens, err := NewLexer(input).Tokenize()
	assertNoErr(t, err)

	// Find pipe token
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenPipe {
			found = true
			break
		}
	}
	assertTrue(t, found)
}

// ════════════════════════════════════════════════════════════════════
// Parser Tests
// ════════════════════════════════════════════════════════════════════

func TestParser_NumberLiteral(t *testing.T) {
	node, err := ParseQuery("42")
	assertNoErr(t, err)
	num, ok := node.(*NumberLiteral)
	assertTrue(t, ok)
	assertFloat(t, 42, num.Value)
}

func TestParser_NumberWithSuffix(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"10cr", 1e8},
		{"5l", 5e5},
		{"2.5crore", 2.5e7},
		{"100lakh", 100e5},
	}
	for _, tt := range tests {
		node, err := ParseQuery(tt.input)
		assertNoErr(t, err)
		num, ok := node.(*NumberLiteral)
		assertTrue(t, ok)
		assertFloat(t, tt.want, num.Value)
	}
}

func TestParser_StringLiteral(t *testing.T) {
	node, err := ParseQuery(`"IT"`)
	assertNoErr(t, err)
	str, ok := node.(*StringLiteral)
	assertTrue(t, ok)
	assertEqual(t, "IT", str.Value)
}

func TestParser_BoolLiteral(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		node, err := ParseQuery("true")
		assertNoErr(t, err)
		b, ok := node.(*BoolLiteral)
		assertTrue(t, ok)
		assertTrue(t, b.Value)
	})
	t.Run("false", func(t *testing.T) {
		node, err := ParseQuery("false")
		assertNoErr(t, err)
		b, ok := node.(*BoolLiteral)
		assertTrue(t, ok)
		assertTrue(t, !b.Value)
	})
}

func TestParser_Identifier(t *testing.T) {
	node, err := ParseQuery("RELIANCE")
	assertNoErr(t, err)
	id, ok := node.(*Identifier)
	assertTrue(t, ok)
	assertEqual(t, "RELIANCE", id.Name)
}

func TestParser_FunctionCall_NoArgs(t *testing.T) {
	node, err := ParseQuery("nifty50()")
	assertNoErr(t, err)
	fn, ok := node.(*FunctionCall)
	assertTrue(t, ok)
	assertEqual(t, "nifty50", fn.Name)
	assertEqual(t, 0, len(fn.Args))
}

func TestParser_FunctionCall_OneArg(t *testing.T) {
	node, err := ParseQuery("price(RELIANCE)")
	assertNoErr(t, err)
	fn, ok := node.(*FunctionCall)
	assertTrue(t, ok)
	assertEqual(t, "price", fn.Name)
	assertEqual(t, 1, len(fn.Args))
	arg, ok := fn.Args[0].(*Identifier)
	assertTrue(t, ok)
	assertEqual(t, "RELIANCE", arg.Name)
}

func TestParser_FunctionCall_MultiArg(t *testing.T) {
	node, err := ParseQuery("sma(TCS, 50)")
	assertNoErr(t, err)
	fn, ok := node.(*FunctionCall)
	assertTrue(t, ok)
	assertEqual(t, "sma", fn.Name)
	assertEqual(t, 2, len(fn.Args))
}

func TestParser_FunctionCall_NestedArgs(t *testing.T) {
	node, err := ParseQuery("crossover(sma(TCS, 50), sma(TCS, 200))")
	assertNoErr(t, err)
	fn, ok := node.(*FunctionCall)
	assertTrue(t, ok)
	assertEqual(t, "crossover", fn.Name)
	assertEqual(t, 2, len(fn.Args))
	inner1, ok := fn.Args[0].(*FunctionCall)
	assertTrue(t, ok)
	assertEqual(t, "sma", inner1.Name)
}

func TestParser_RangeSelector(t *testing.T) {
	node, err := ParseQuery("price(RELIANCE)[30d]")
	assertNoErr(t, err)
	rs, ok := node.(*RangeSelector)
	assertTrue(t, ok)
	assertEqual(t, "30d", rs.Duration)
	assertEqual(t, 30, rs.Days)
	inner, ok := rs.Expr.(*FunctionCall)
	assertTrue(t, ok)
	assertEqual(t, "price", inner.Name)
}

func TestParser_RangeSelector_Weeks(t *testing.T) {
	node, err := ParseQuery("rsi(TCS)[2w]")
	assertNoErr(t, err)
	rs, ok := node.(*RangeSelector)
	assertTrue(t, ok)
	assertEqual(t, 14, rs.Days)
}

func TestParser_RangeSelector_Months(t *testing.T) {
	node, err := ParseQuery("price(INFY)[3m]")
	assertNoErr(t, err)
	rs, ok := node.(*RangeSelector)
	assertTrue(t, ok)
	assertEqual(t, 90, rs.Days)
}

func TestParser_RangeSelector_Year(t *testing.T) {
	node, err := ParseQuery("price(INFY)[1y]")
	assertNoErr(t, err)
	rs, ok := node.(*RangeSelector)
	assertTrue(t, ok)
	assertEqual(t, 365, rs.Days)
}

func TestParser_BinaryExpr_Arithmetic(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{"2 + 3", "+"},
		{"10 - 5", "-"},
		{"3 * 4", "*"},
		{"8 / 2", "/"},
	}
	for _, tt := range tests {
		node, err := ParseQuery(tt.input)
		assertNoErr(t, err)
		be, ok := node.(*BinaryExpr)
		assertTrue(t, ok)
		assertEqual(t, tt.op, be.Op)
	}
}

func TestParser_BinaryExpr_Comparison(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{"42 > 10", ">"},
		{"42 < 100", "<"},
		{"42 >= 42", ">="},
		{"42 <= 42", "<="},
		{"42 == 42", "=="},
		{"42 != 7", "!="},
	}
	for _, tt := range tests {
		node, err := ParseQuery(tt.input)
		assertNoErr(t, err)
		be, ok := node.(*BinaryExpr)
		assertTrue(t, ok)
		assertEqual(t, tt.op, be.Op)
	}
}

func TestParser_BinaryExpr_Logical(t *testing.T) {
	node, err := ParseQuery("true AND false")
	assertNoErr(t, err)
	be, ok := node.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "AND", be.Op)

	node, err = ParseQuery("true OR false")
	assertNoErr(t, err)
	be, ok = node.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "OR", be.Op)
}

func TestParser_UnaryExpr_Negate(t *testing.T) {
	node, err := ParseQuery("-42")
	assertNoErr(t, err)
	ue, ok := node.(*UnaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "-", ue.Op)
}

func TestParser_UnaryExpr_NOT(t *testing.T) {
	node, err := ParseQuery("NOT true")
	assertNoErr(t, err)
	ue, ok := node.(*UnaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "NOT", ue.Op)
}

func TestParser_PipeExpr(t *testing.T) {
	node, err := ParseQuery("a | b")
	assertNoErr(t, err)
	pe, ok := node.(*PipeExpr)
	assertTrue(t, ok)
	left, ok := pe.Left.(*Identifier)
	assertTrue(t, ok)
	assertEqual(t, "a", left.Name)
}

func TestParser_ScreenerExpr(t *testing.T) {
	node, err := ParseQuery("screener(42 > 10)")
	assertNoErr(t, err)
	se, ok := node.(*ScreenerExpr)
	assertTrue(t, ok)
	filter, ok := se.Filter.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, ">", filter.Op)
}

func TestParser_AlertExpr(t *testing.T) {
	node, err := ParseQuery(`alert(42 > 10, "high!")`)
	assertNoErr(t, err)
	ae, ok := node.(*AlertExpr)
	assertTrue(t, ok)
	assertEqual(t, "high!", ae.Message)
}

func TestParser_AlertExpr_NoMessage(t *testing.T) {
	node, err := ParseQuery("alert(42 > 10)")
	assertNoErr(t, err)
	ae, ok := node.(*AlertExpr)
	assertTrue(t, ok)
	assertEqual(t, "", ae.Message)
}

func TestParser_Precedence_ArithOverComparison(t *testing.T) {
	// 2 + 3 > 4 should be (2 + 3) > 4
	node, err := ParseQuery("2 + 3 > 4")
	assertNoErr(t, err)
	be, ok := node.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, ">", be.Op)
	left, ok := be.Left.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "+", left.Op)
}

func TestParser_Precedence_MulOverAdd(t *testing.T) {
	// 2 + 3 * 4 should be 2 + (3 * 4)
	node, err := ParseQuery("2 + 3 * 4")
	assertNoErr(t, err)
	be, ok := node.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "+", be.Op)
	right, ok := be.Right.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "*", right.Op)
}

func TestParser_Precedence_Parens(t *testing.T) {
	// (2 + 3) * 4 should be (2+3) * 4
	node, err := ParseQuery("(2 + 3) * 4")
	assertNoErr(t, err)
	be, ok := node.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "*", be.Op)
	left, ok := be.Left.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "+", left.Op)
}

func TestParser_Precedence_PipeLowest(t *testing.T) {
	// a | b AND c should be a | (b AND c)
	node, err := ParseQuery("a | b AND c")
	assertNoErr(t, err)
	pe, ok := node.(*PipeExpr)
	assertTrue(t, ok)
	right, ok := pe.Right.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "AND", right.Op)
}

func TestParser_Wildcard(t *testing.T) {
	node, err := ParseQuery("rsi(*, 14)")
	assertNoErr(t, err)
	fn, ok := node.(*FunctionCall)
	assertTrue(t, ok)
	arg, ok := fn.Args[0].(*Identifier)
	assertTrue(t, ok)
	assertEqual(t, "*", arg.Name)
}

func TestParser_Error_UnexpectedToken(t *testing.T) {
	_, err := ParseQuery(")")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestParser_Error_UnclosedParen(t *testing.T) {
	_, err := ParseQuery("(42")
	if err == nil {
		t.Fatal("expected parse error for unclosed paren")
	}
}

func TestParser_Error_TrailingToken(t *testing.T) {
	_, err := ParseQuery("42 42")
	if err == nil {
		t.Fatal("expected parse error for trailing token")
	}
}

func TestParser_ComplexScreener(t *testing.T) {
	input := `screener(rsi(*, 14) < 30 AND pe(*) < 20)`
	node, err := ParseQuery(input)
	assertNoErr(t, err)
	se, ok := node.(*ScreenerExpr)
	assertTrue(t, ok)
	filter, ok := se.Filter.(*BinaryExpr)
	assertTrue(t, ok)
	assertEqual(t, "AND", filter.Op)
}

func TestParser_MultiplePipes(t *testing.T) {
	input := "a | b | c"
	node, err := ParseQuery(input)
	assertNoErr(t, err)
	// Should be left-associative: (a | b) | c
	pe, ok := node.(*PipeExpr)
	assertTrue(t, ok)
	left, ok := pe.Left.(*PipeExpr)
	assertTrue(t, ok)
	_ = left
}

// ════════════════════════════════════════════════════════════════════
// AST String() Tests
// ════════════════════════════════════════════════════════════════════

func TestAST_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"number", "42", "42"},
		{"string", `"hello"`, `"hello"`},
		{"identifier", "RELIANCE", "RELIANCE"},
		{"bool_true", "true", "true"},
		{"bool_false", "false", "false"},
		{"func_no_args", "nifty50()", "nifty50()"},
		{"func_with_args", "sma(TCS, 50)", "sma(TCS, 50)"},
		{"range_selector", "price(RELIANCE)[30d]", "price(RELIANCE)[30d]"},
		{"binary_add", "2 + 3", "(2 + 3)"},
		{"unary_neg", "-42", "(- 42)"},
		{"unary_not", "NOT true", "(NOT true)"},
		{"pipe", "a | b", "a | b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseQuery(tt.input)
			assertNoErr(t, err)
			assertEqual(t, tt.want, node.String())
		})
	}
}

// ════════════════════════════════════════════════════════════════════
// Evaluator Tests (pure numeric — no data source)
// ════════════════════════════════════════════════════════════════════

func newTestEvalContext() *EvalContext {
	// Create an EvalContext without a real aggregator—only for pure expressions.
	ec := &EvalContext{
		Ctx:       nil, // no real context needed
		Functions: make(map[string]BuiltinFunc),
		Cache:     NewEvalCache(5 * time.Minute),
	}
	RegisterBuiltins(ec)
	return ec
}

func TestEval_NumberLiteral(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, "42")
	assertNoErr(t, err)
	assertEqual(t, TypeScalar, v.Type)
	assertFloat(t, 42, v.Scalar)
}

func TestEval_StringLiteral(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, `"hello world"`)
	assertNoErr(t, err)
	assertEqual(t, TypeString, v.Type)
	assertEqual(t, "hello world", v.Str)
}

func TestEval_BoolLiteral(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, "true")
	assertNoErr(t, err)
	assertEqual(t, TypeBool, v.Type)
	assertTrue(t, v.Bool)
}

func TestEval_Arithmetic(t *testing.T) {
	ec := newTestEvalContext()
	tests := []struct {
		query string
		want  float64
	}{
		{"2 + 3", 5},
		{"10 - 4", 6},
		{"3 * 7", 21},
		{"20 / 5", 4},
		{"2 + 3 * 4", 14},
		{"(2 + 3) * 4", 20},
		{"-5 + 8", 3},
		{"100 / 0", math.NaN()},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			v, err := EvalQuery(ec, tt.query)
			assertNoErr(t, err)
			if math.IsNaN(tt.want) {
				assertTrue(t, math.IsNaN(v.Scalar))
			} else {
				assertFloat(t, tt.want, v.Scalar)
			}
		})
	}
}

func TestEval_Comparison(t *testing.T) {
	ec := newTestEvalContext()
	tests := []struct {
		query string
		want  bool
	}{
		{"5 > 3", true},
		{"3 > 5", false},
		{"5 < 10", true},
		{"10 < 5", false},
		{"5 >= 5", true},
		{"4 >= 5", false},
		{"5 <= 5", true},
		{"6 <= 5", false},
		{"42 == 42", true},
		{"42 != 43", true},
		{"42 != 42", false},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			v, err := EvalQuery(ec, tt.query)
			assertNoErr(t, err)
			assertEqual(t, TypeBool, v.Type)
			assertEqual(t, tt.want, v.Bool)
		})
	}
}

func TestEval_Logical(t *testing.T) {
	ec := newTestEvalContext()
	tests := []struct {
		query string
		want  bool
	}{
		{"true AND true", true},
		{"true AND false", false},
		{"false OR true", true},
		{"false OR false", false},
		{"NOT true", false},
		{"NOT false", true},
		{"true AND true OR false", true},
		{"true AND (true OR false)", true},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			v, err := EvalQuery(ec, tt.query)
			assertNoErr(t, err)
			assertTrue(t, v.Bool == tt.want)
		})
	}
}

func TestEval_StringEquality(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, `"IT" == "it"`)
	assertNoErr(t, err)
	assertTrue(t, v.Bool) // case-insensitive

	v, err = EvalQuery(ec, `"IT" != "Banking"`)
	assertNoErr(t, err)
	assertTrue(t, v.Bool)
}

func TestEval_UnaryNegate(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, "-42")
	assertNoErr(t, err)
	assertFloat(t, -42, v.Scalar)
}

func TestEval_NumberSuffix(t *testing.T) {
	ec := newTestEvalContext()
	tests := []struct {
		query string
		want  float64
	}{
		{"10cr", 1e8},
		{"5l", 5e5},
		{"2.5crore", 2.5e7},
	}
	for _, tt := range tests {
		v, err := EvalQuery(ec, tt.query)
		assertNoErr(t, err)
		assertFloat(t, tt.want, v.Scalar)
	}
}

// ════════════════════════════════════════════════════════════════════
// Built-in Function Tests (aggregation — no data source needed)
// ════════════════════════════════════════════════════════════════════

func TestBuiltin_Avg(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 10}, {Value: 20}, {Value: 30}}
	v, err := ec.Functions["avg"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertFloat(t, 20, v.Scalar)
}

func TestBuiltin_Sum(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 10}, {Value: 20}, {Value: 30}}
	v, err := ec.Functions["sum"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertFloat(t, 60, v.Scalar)
}

func TestBuiltin_Min(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 5}, {Value: 2}, {Value: 8}}
	v, err := ec.Functions["min"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertFloat(t, 2, v.Scalar)
}

func TestBuiltin_Max(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 5}, {Value: 2}, {Value: 8}}
	v, err := ec.Functions["max"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertFloat(t, 8, v.Scalar)
}

func TestBuiltin_Stddev(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 2}, {Value: 4}, {Value: 4}, {Value: 4}, {Value: 5}, {Value: 5}, {Value: 7}, {Value: 9}}
	v, err := ec.Functions["stddev"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	// Sample stddev of {2,4,4,4,5,5,7,9} ~= 2.138
	if v.Scalar < 2.0 || v.Scalar > 2.2 {
		t.Errorf("stddev = %f, want ~2.138", v.Scalar)
	}
}

func TestBuiltin_Percentile(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5}}
	v, err := ec.Functions["percentile"](ec, []Value{VectorValue(pts), ScalarValue(50)})
	assertNoErr(t, err)
	assertFloat(t, 3, v.Scalar) // Median
}

func TestBuiltin_Abs(t *testing.T) {
	ec := newTestEvalContext()
	v, err := ec.Functions["abs"](ec, []Value{ScalarValue(-42)})
	assertNoErr(t, err)
	assertFloat(t, 42, v.Scalar)
}

func TestBuiltin_Correlation(t *testing.T) {
	ec := newTestEvalContext()
	a := []TimePoint{{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5}}
	b := []TimePoint{{Value: 2}, {Value: 4}, {Value: 6}, {Value: 8}, {Value: 10}}
	v, err := ec.Functions["correlation"](ec, []Value{VectorValue(a), VectorValue(b)})
	assertNoErr(t, err)
	// Perfect positive correlation
	if v.Scalar < 0.99 {
		t.Errorf("correlation = %f, want ~1.0", v.Scalar)
	}
}

func TestBuiltin_Count(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 1}, {Value: 2}, {Value: 3}}
	v, err := ec.Functions["count"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertFloat(t, 3, v.Scalar)
}

func TestBuiltin_First(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 10}, {Value: 20}, {Value: 30}}
	v, err := ec.Functions["first"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertFloat(t, 10, v.Scalar)
}

func TestBuiltin_Last(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 10}, {Value: 20}, {Value: 30}}
	v, err := ec.Functions["last"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertFloat(t, 30, v.Scalar)
}

func TestBuiltin_Crossover(t *testing.T) {
	ec := newTestEvalContext()
	v, err := ec.Functions["crossover"](ec, []Value{ScalarValue(60), ScalarValue(40)})
	assertNoErr(t, err)
	assertTrue(t, v.Bool)

	v, err = ec.Functions["crossover"](ec, []Value{ScalarValue(30), ScalarValue(40)})
	assertNoErr(t, err)
	assertTrue(t, !v.Bool)
}

func TestBuiltin_Crossunder(t *testing.T) {
	ec := newTestEvalContext()
	v, err := ec.Functions["crossunder"](ec, []Value{ScalarValue(30), ScalarValue(40)})
	assertNoErr(t, err)
	assertTrue(t, v.Bool)
}

func TestBuiltin_Trend(t *testing.T) {
	ec := newTestEvalContext()

	// Uptrend
	pts := []TimePoint{{Value: 100}, {Value: 110}, {Value: 120}}
	v, err := ec.Functions["trend"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertEqual(t, "UPTREND", v.Str)

	// Downtrend
	pts = []TimePoint{{Value: 120}, {Value: 110}, {Value: 100}}
	v, err = ec.Functions["trend"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertEqual(t, "DOWNTREND", v.Str)

	// Sideways
	pts = []TimePoint{{Value: 100}, {Value: 101}, {Value: 100.5}}
	v, err = ec.Functions["trend"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertEqual(t, "SIDEWAYS", v.Str)
}

func TestBuiltin_Returns(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 100}, {Value: 110}, {Value: 121}}
	v, err := ec.Functions["returns"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertEqual(t, TypeVector, v.Type)
	assertEqual(t, 2, len(v.Vector))
	assertFloat(t, 0.1, v.Vector[0].Value)
	assertFloat(t, 0.1, v.Vector[1].Value)
}

func TestBuiltin_ChangePct(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 100}, {Value: 150}}
	v, err := ec.Functions["change_pct"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertFloat(t, 50, v.Scalar) // 50% change
}

func TestBuiltin_Sort(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 3}, {Value: 1}, {Value: 2}}
	v, err := ec.Functions["sort"](ec, []Value{VectorValue(pts)})
	assertNoErr(t, err)
	assertEqual(t, TypeVector, v.Type)
	assertFloat(t, 1, v.Vector[0].Value)
	assertFloat(t, 2, v.Vector[1].Value)
	assertFloat(t, 3, v.Vector[2].Value)
}

func TestBuiltin_SortDesc(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 3}, {Value: 1}, {Value: 2}}
	v, err := ec.Functions["sort"](ec, []Value{VectorValue(pts), StringValue("desc")})
	assertNoErr(t, err)
	assertFloat(t, 3, v.Vector[0].Value)
	assertFloat(t, 2, v.Vector[1].Value)
	assertFloat(t, 1, v.Vector[2].Value)
}

func TestBuiltin_Top(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5}}
	v, err := ec.Functions["top"](ec, []Value{VectorValue(pts), ScalarValue(3)})
	assertNoErr(t, err)
	assertEqual(t, 3, len(v.Vector))
	assertFloat(t, 1, v.Vector[0].Value)
}

func TestBuiltin_Bottom(t *testing.T) {
	ec := newTestEvalContext()
	pts := []TimePoint{{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5}}
	v, err := ec.Functions["bottom"](ec, []Value{VectorValue(pts), ScalarValue(2)})
	assertNoErr(t, err)
	assertEqual(t, 2, len(v.Vector))
	assertFloat(t, 4, v.Vector[0].Value)
}

func TestBuiltin_Nifty50(t *testing.T) {
	ec := newTestEvalContext()
	v, err := ec.Functions["nifty50"](ec, nil)
	assertNoErr(t, err)
	assertEqual(t, TypeTable, v.Type)
	assertEqual(t, 50, len(v.Table))
	assertEqual(t, "RELIANCE", v.Table[0]["ticker"])
}

func TestBuiltin_NiftyBank(t *testing.T) {
	ec := newTestEvalContext()
	v, err := ec.Functions["niftybank"](ec, nil)
	assertNoErr(t, err)
	assertEqual(t, TypeTable, v.Type)
	assertEqual(t, 12, len(v.Table))
}

func TestBuiltin_SMA_VectorInput(t *testing.T) {
	ec := newTestEvalContext()
	pts := make([]TimePoint, 50)
	for i := range pts {
		pts[i] = TimePoint{Value: float64(i + 1)}
	}
	v, err := ec.Functions["sma"](ec, []Value{VectorValue(pts), ScalarValue(20)})
	assertNoErr(t, err)
	assertEqual(t, TypeScalar, v.Type)
	// SMA of last 20 values (31..50) = 40.5
	assertFloat(t, 40.5, v.Scalar)
}

func TestBuiltin_EMA_VectorInput(t *testing.T) {
	ec := newTestEvalContext()
	pts := make([]TimePoint, 50)
	for i := range pts {
		pts[i] = TimePoint{Value: float64(i + 1)}
	}
	v, err := ec.Functions["ema"](ec, []Value{VectorValue(pts), ScalarValue(21)})
	assertNoErr(t, err)
	assertEqual(t, TypeScalar, v.Type)
	// EMA should be computed; just verify it's a reasonable number
	if v.Scalar <= 0 || v.Scalar > 55 {
		t.Errorf("unexpected EMA value: %f", v.Scalar)
	}
}

// ════════════════════════════════════════════════════════════════════
// Evaluator Pipe Tests
// ════════════════════════════════════════════════════════════════════

func TestEval_Pipe(t *testing.T) {
	ec := newTestEvalContext()
	// Register a test function for pipe
	ec.RegisterFunc("double", func(_ *EvalContext, args []Value) (Value, error) {
		if len(args) > 0 && args[0].Type == TypeScalar {
			return ScalarValue(args[0].Scalar * 2), nil
		}
		return ScalarValue(0), nil
	})
	v, err := EvalQuery(ec, "21 | double(*)")
	assertNoErr(t, err)
	// 21 pipes into double → 42
	assertFloat(t, 42, v.Scalar)
}

// ════════════════════════════════════════════════════════════════════
// Cache Tests
// ════════════════════════════════════════════════════════════════════

func TestEvalCache(t *testing.T) {
	cache := NewEvalCache(100 * time.Millisecond)

	cache.Set("key1", ScalarValue(42))
	v, ok := cache.Get("key1")
	assertTrue(t, ok)
	assertFloat(t, 42, v.Scalar)

	// Missing key
	_, ok = cache.Get("missing")
	assertTrue(t, !ok)
}

func TestEvalCache_Expiry(t *testing.T) {
	cache := NewEvalCache(50 * time.Millisecond)
	cache.Set("temp", ScalarValue(99))

	time.Sleep(100 * time.Millisecond)
	_, ok := cache.Get("temp")
	assertTrue(t, !ok) // expired
}

// ════════════════════════════════════════════════════════════════════
// Duration & Number Parsing Tests
// ════════════════════════════════════════════════════════════════════

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"30d", 30},
		{"1w", 7},
		{"2w", 14},
		{"3m", 90},
		{"1y", 365},
		{"252d", 252},
		{"30", 30},
		{"", 0},
	}
	for _, tt := range tests {
		got := parseDuration(tt.input)
		if got != tt.want {
			t.Errorf("parseDuration(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseNumber(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"42", 42},
		{"3.14", 3.14},
		{"10cr", 1e8},
		{"5l", 5e5},
		{"2.5crore", 2.5e7},
		{"100lakh", 1e7},
	}
	for _, tt := range tests {
		got, err := parseNumber(tt.input)
		assertNoErr(t, err)
		assertFloat(t, tt.want, got)
	}
}

// ════════════════════════════════════════════════════════════════════
// REPL Tests
// ════════════════════════════════════════════════════════════════════

func TestREPL_DotHelp(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader(".help\n.quit\n")
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()

	output := out.String()
	assertTrue(t, strings.Contains(output, "Quick Reference"))
	assertTrue(t, strings.Contains(output, "Goodbye!"))
}

func TestREPL_DotFunctions(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader(".functions\n.quit\n")
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()

	output := out.String()
	assertTrue(t, strings.Contains(output, "Built-in Functions"))
	assertTrue(t, strings.Contains(output, "Price"))
	assertTrue(t, strings.Contains(output, "Technical"))
}

func TestREPL_DotHistory(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("42\n.history\n.quit\n")
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()

	output := out.String()
	assertTrue(t, strings.Contains(output, "42"))
}

func TestREPL_DotClear(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("42\n.clear\n.quit\n")
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()

	assertEqual(t, 0, len(repl.History()))
}

func TestREPL_ArithmeticQuery(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("2 + 3\n.quit\n")
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()

	output := out.String()
	assertTrue(t, strings.Contains(output, "5.0000"))
}

func TestREPL_BoolResult(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("5 > 3\n.quit\n")
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()

	output := out.String()
	assertTrue(t, strings.Contains(output, "true"))
}

func TestREPL_StringResult(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader(`"hello"` + "\n.quit\n")
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()

	output := out.String()
	assertTrue(t, strings.Contains(output, "hello"))
}

func TestREPL_ParseError(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader(")\n.quit\n")
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()

	output := out.String()
	assertTrue(t, strings.Contains(output, "Parse error"))
}

func TestREPL_UnknownCommand(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader(".unknowncmd\n.quit\n")
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()

	output := out.String()
	assertTrue(t, strings.Contains(output, "Unknown command"))
}

func TestREPL_EmptyLine(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("\n\n.quit\n")
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()

	// Just verifies no crash on empty input
	assertTrue(t, strings.Contains(out.String(), "Goodbye!"))
}

func TestREPL_GetFunctionNames(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader(".quit\n")
	repl := NewREPLWithIO(nil, in, &out)

	names := repl.GetFunctionNames()
	// Verify some expected functions are present
	found := make(map[string]bool)
	for _, n := range names {
		found[n] = true
	}
	assertTrue(t, found["price"])
	assertTrue(t, found["sma"])
	assertTrue(t, found["rsi"])
	assertTrue(t, found["avg"])
	assertTrue(t, found["nifty50"])
	// Internal functions should be excluded
	assertTrue(t, !found["_screener"])
}

func TestREPL_EOF(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("") // EOF immediately
	repl := NewREPLWithIO(nil, in, &out)
	repl.Run()
	// Should exit gracefully without crash
}

// ════════════════════════════════════════════════════════════════════
// Format & Display Tests
// ════════════════════════════════════════════════════════════════════

func TestFormatResult_Scalar(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader(".quit\n")
	repl := NewREPLWithIO(nil, in, &out)

	repl.formatResult(ScalarValue(42.5))
	assertTrue(t, strings.Contains(out.String(), "42.5000"))
}

func TestFormatResult_Vector(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader(".quit\n")
	repl := NewREPLWithIO(nil, in, &out)

	pts := []TimePoint{{Value: 10}, {Value: 20}, {Value: 30}}
	repl.formatResult(VectorValue(pts))
	output := out.String()
	assertTrue(t, strings.Contains(output, "Vector[3 points]"))
	assertTrue(t, strings.Contains(output, "Min:"))
}

func TestFormatResult_Table(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader(".quit\n")
	repl := NewREPLWithIO(nil, in, &out)

	rows := []map[string]interface{}{
		{"ticker": "TCS", "pe": 30.5},
		{"ticker": "INFY", "pe": 25.2},
	}
	repl.formatResult(TableValue(rows))
	output := out.String()
	assertTrue(t, strings.Contains(output, "Table[2 rows]"))
	assertTrue(t, strings.Contains(output, "TCS"))
}

func TestFormatResult_Nil(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader(".quit\n")
	repl := NewREPLWithIO(nil, in, &out)

	repl.formatResult(NilValue())
	assertTrue(t, strings.Contains(out.String(), "nil"))
}

func TestSparkline(t *testing.T) {
	pts := []TimePoint{{Value: 1}, {Value: 5}, {Value: 3}, {Value: 9}, {Value: 2}}
	s := sparkline(pts)
	// Should produce some block characters
	assertTrue(t, len(s) > 0)
	assertTrue(t, len(s) <= 60)
}

func TestSparkline_Empty(t *testing.T) {
	s := sparkline(nil)
	assertEqual(t, "", s)
}

func TestSparkline_AllSame(t *testing.T) {
	pts := []TimePoint{{Value: 5}, {Value: 5}, {Value: 5}}
	s := sparkline(pts)
	assertTrue(t, len(s) > 0)
}

func TestPadRight(t *testing.T) {
	assertEqual(t, "hi   ", padRight("hi", 5))
	assertEqual(t, "hello", padRight("hello", 3)) // longer than width
	assertEqual(t, "abc", padRight("abc", 3))
}

// ════════════════════════════════════════════════════════════════════
// Integration Evaluation Tests (complex expressions)
// ════════════════════════════════════════════════════════════════════

func TestEval_ComplexArithmetic(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, "(10 + 5) * 2 - 3")
	assertNoErr(t, err)
	assertFloat(t, 27, v.Scalar)
}

func TestEval_NestedComparison(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, "10 > 5 AND 3 < 7")
	assertNoErr(t, err)
	assertTrue(t, v.Bool)
}

func TestEval_ComplexLogical(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, "NOT (5 > 10) AND 3 < 7")
	assertNoErr(t, err)
	assertTrue(t, v.Bool)
}

func TestEval_ScreenerExpr(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, "screener(42 > 10)")
	assertNoErr(t, err)
	assertEqual(t, TypeTable, v.Type)
}

func TestEval_AlertExpr_Triggered(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, `alert(5 > 3, "high!")`)
	assertNoErr(t, err)
	assertEqual(t, TypeTable, v.Type)
	assertTrue(t, len(v.Table) > 0)
	assertEqual(t, true, v.Table[0]["triggered"])
}

func TestEval_AlertExpr_NotTriggered(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, `alert(3 > 5, "low!")`)
	assertNoErr(t, err)
	assertEqual(t, TypeTable, v.Type)
	assertEqual(t, false, v.Table[0]["triggered"])
}

func TestEval_Nifty50Function(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, "nifty50()")
	assertNoErr(t, err)
	assertEqual(t, TypeTable, v.Type)
	assertEqual(t, 50, len(v.Table))
}

func TestEval_NiftyBankFunction(t *testing.T) {
	ec := newTestEvalContext()
	v, err := EvalQuery(ec, "niftybank()")
	assertNoErr(t, err)
	assertEqual(t, TypeTable, v.Type)
	assertEqual(t, 12, len(v.Table))
}

func TestEval_UnknownFunction(t *testing.T) {
	ec := newTestEvalContext()
	_, err := EvalQuery(ec, "unknown_func(42)")
	if err == nil {
		t.Fatal("expected error for unknown function")
	}
	assertTrue(t, strings.Contains(err.Error(), "unknown function"))
}

func TestEval_Nil(t *testing.T) {
	ec := newTestEvalContext()
	v, err := Eval(ec, nil)
	assertNoErr(t, err)
	assertEqual(t, TypeNil, v.Type)
}

// ════════════════════════════════════════════════════════════════════
// ToBool / ToScalar Tests
// ════════════════════════════════════════════════════════════════════

func TestToBool(t *testing.T) {
	assertTrue(t, toBool(BoolValue(true)))
	assertTrue(t, !toBool(BoolValue(false)))
	assertTrue(t, toBool(ScalarValue(1)))
	assertTrue(t, !toBool(ScalarValue(0)))
	assertTrue(t, toBool(StringValue("x")))
	assertTrue(t, !toBool(StringValue("")))
	assertTrue(t, toBool(VectorValue([]TimePoint{{Value: 1}})))
	assertTrue(t, !toBool(VectorValue(nil)))
	assertTrue(t, !toBool(NilValue()))
	assertTrue(t, toBool(TableValue([]map[string]interface{}{{"a": 1}})))
	assertTrue(t, !toBool(TableValue(nil)))
}

func TestToScalar(t *testing.T) {
	assertFloat(t, 42, toScalar(ScalarValue(42)))
	assertFloat(t, 1, toScalar(BoolValue(true)))
	assertFloat(t, 0, toScalar(BoolValue(false)))
	assertFloat(t, 5, toScalar(VectorValue([]TimePoint{{Value: 3}, {Value: 5}})))
	assertFloat(t, 0, toScalar(NilValue()))
}

// ════════════════════════════════════════════════════════════════════
// ResolveTicker Tests
// ════════════════════════════════════════════════════════════════════

func TestResolveTicker(t *testing.T) {
	assertEqual(t, "RELIANCE", ResolveTicker("RELIANCE"))
	assertEqual(t, "RELIANCE", ResolveTicker("reliance"))
	assertEqual(t, "RELIANCE", ResolveTicker("RELIANCE.NS"))
	assertEqual(t, "INFY", ResolveTicker(" infy "))
}

// ════════════════════════════════════════════════════════════════════
// Helper Utilities Tests
// ════════════════════════════════════════════════════════════════════

func TestCollectFloats(t *testing.T) {
	vals := collectFloats([]Value{
		ScalarValue(1),
		VectorValue([]TimePoint{{Value: 2}, {Value: 3}}),
	})
	assertEqual(t, 3, len(vals))
	assertFloat(t, 1, vals[0])
	assertFloat(t, 2, vals[1])
	assertFloat(t, 3, vals[2])
}

func TestPearson(t *testing.T) {
	// Perfect correlation
	a := []float64{1, 2, 3, 4, 5}
	b := []float64{2, 4, 6, 8, 10}
	r := pearson(a, b)
	if r < 0.99 {
		t.Errorf("pearson = %f, want ~1.0", r)
	}

	// Too few points
	r = pearson([]float64{1}, []float64{2})
	assertFloat(t, 0, r)
}

func TestVectorToFloat64(t *testing.T) {
	pts := []TimePoint{{Value: 1.5}, {Value: 2.5}, {Value: 3.5}}
	vals := vectorToFloat64(pts)
	assertEqual(t, 3, len(vals))
	assertFloat(t, 2.5, vals[1])
}

func TestOHLCVToVector(t *testing.T) {
	now := time.Now()
	ohlcv := []models.OHLCV{
		{Timestamp: now, Close: 100},
		{Timestamp: now.Add(24 * time.Hour), Close: 110},
	}
	pts := OHLCVToVector(ohlcv)
	assertEqual(t, 2, len(pts))
	assertFloat(t, 100, pts[0].Value)
	assertFloat(t, 110, pts[1].Value)
}

// ════════════════════════════════════════════════════════════════════
// RegisterFunc & Custom Functions
// ════════════════════════════════════════════════════════════════════

func TestRegisterFunc(t *testing.T) {
	ec := newTestEvalContext()
	ec.RegisterFunc("double", func(_ *EvalContext, args []Value) (Value, error) {
		if len(args) > 0 {
			return ScalarValue(args[0].Scalar * 2), nil
		}
		return ScalarValue(0), nil
	})

	v, err := EvalQuery(ec, "double(21)")
	assertNoErr(t, err)
	assertFloat(t, 42, v.Scalar)
}

func TestRegisterFunc_CaseInsensitive(t *testing.T) {
	ec := newTestEvalContext()
	ec.RegisterFunc("MyFunc", func(_ *EvalContext, _ []Value) (Value, error) {
		return ScalarValue(99), nil
	})

	// Functions are stored lower-cased, and parsed lower-cased
	v, err := EvalQuery(ec, "myfunc()")
	assertNoErr(t, err)
	assertFloat(t, 99, v.Scalar)
}

// ════════════════════════════════════════════════════════════════════
// Wildcard / Pipe Integration
// ════════════════════════════════════════════════════════════════════

func TestEval_WildcardIdentifier(t *testing.T) {
	ec := newTestEvalContext()
	node, err := ParseQuery("*")
	assertNoErr(t, err)
	v, err := Eval(ec, node)
	assertNoErr(t, err)
	assertEqual(t, TypeString, v.Type)
	assertEqual(t, "*", v.Str)
}

// ════════════════════════════════════════════════════════════════════
// ParseError Tests
// ════════════════════════════════════════════════════════════════════

func TestParseError_Error(t *testing.T) {
	pe := &ParseError{Line: 1, Column: 5, Message: "oops"}
	s := pe.Error()
	assertTrue(t, strings.Contains(s, "oops"))
	assertTrue(t, strings.Contains(s, "line 1"))
	assertTrue(t, strings.Contains(s, "col 5"))
}

// ════════════════════════════════════════════════════════════════════
// Test Helpers
// ════════════════════════════════════════════════════════════════════

func assertEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

func assertFloat(t *testing.T, want, got float64) {
	t.Helper()
	if math.Abs(want-got) > 1e-6 {
		t.Errorf("want %f, got %f", want, got)
	}
}

func assertTrue(t *testing.T, v bool) {
	t.Helper()
	if !v {
		t.Error("expected true, got false")
	}
}

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
