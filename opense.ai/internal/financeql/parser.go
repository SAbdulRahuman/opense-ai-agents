package financeql

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ════════════════════════════════════════════════════════════════════
// Parser — Recursive Descent
// ════════════════════════════════════════════════════════════════════

// Parser transforms a token stream into an AST.
type Parser struct {
	tokens []Token
	pos    int
	source string // original source for error context
}

// NewParser creates a parser from a token slice.
func NewParser(tokens []Token, source string) *Parser {
	return &Parser{tokens: tokens, source: source}
}

// Parse parses the full expression.
func (p *Parser) Parse() (Node, error) {
	node, err := p.parsePipeExpr()
	if err != nil {
		return nil, err
	}
	if !p.atEnd() && p.peek().Type != TokenEOF {
		tok := p.peek()
		return nil, p.errorf(tok, "unexpected token %s after expression", tok.Value)
	}
	return node, nil
}

// ParseQuery is the top-level public function to parse a FinanceQL query string.
func ParseQuery(input string) (Node, error) {
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}
	parser := NewParser(tokens, input)
	return parser.Parse()
}

// ────────────────────────────────────────────────────────────────────
// Token helpers
// ────────────────────────────────────────────────────────────────────

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.peek()
	if tok.Type != TokenEOF {
		p.pos++
	}
	return tok
}

func (p *Parser) atEnd() bool {
	return p.pos >= len(p.tokens) || p.tokens[p.pos].Type == TokenEOF
}

func (p *Parser) expect(typ TokenType) (Token, error) {
	tok := p.peek()
	if tok.Type != typ {
		return tok, p.errorf(tok, "expected %s, got %s (%q)", typ, tok.Type, tok.Value)
	}
	return p.advance(), nil
}

func (p *Parser) errorf(tok Token, format string, args ...interface{}) error {
	return &ParseError{
		Position: tok.Position,
		Line:     tok.Line,
		Column:   tok.Column,
		Message:  fmt.Sprintf(format, args...),
	}
}

// ────────────────────────────────────────────────────────────────────
// Grammar (precedence from lowest to highest):
//   PipeExpr       → OrExpr ( '|' OrExpr )*
//   OrExpr         → AndExpr ( 'OR' AndExpr )*
//   AndExpr        → NotExpr ( 'AND' NotExpr )*
//   NotExpr        → 'NOT' NotExpr | Comparison
//   Comparison     → Addition ( ('>'|'<'|'>='|'<='|'=='|'!=') Addition )?
//   Addition       → Multiplication ( ('+'|'-') Multiplication )*
//   Multiplication → Unary ( ('*'|'/') Unary )*
//   Unary          → '-' Unary | Postfix
//   Postfix        → Primary ( '[' range ']' )*
//   Primary        → Number | String | Bool | '(' Expr ')' | FunctionCall | Identifier
// ────────────────────────────────────────────────────────────────────

func (p *Parser) parsePipeExpr() (Node, error) {
	left, err := p.parseOrExpr()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == TokenPipe {
		pipeTok := p.advance()
		right, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		left = &PipeExpr{Position: pipeTok.Position, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseOrExpr() (Node, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == TokenOR {
		opTok := p.advance()
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Position: opTok.Position, Op: "OR", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAndExpr() (Node, error) {
	left, err := p.parseNotExpr()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == TokenAND {
		opTok := p.advance()
		right, err := p.parseNotExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Position: opTok.Position, Op: "AND", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseNotExpr() (Node, error) {
	if p.peek().Type == TokenNOT {
		opTok := p.advance()
		operand, err := p.parseNotExpr()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Position: opTok.Position, Op: "NOT", Operand: operand}, nil
	}
	return p.parseComparison()
}

func (p *Parser) parseComparison() (Node, error) {
	left, err := p.parseAddition()
	if err != nil {
		return nil, err
	}

	tok := p.peek()
	switch tok.Type {
	case TokenGT, TokenLT, TokenGTE, TokenLTE, TokenEQ, TokenNEQ:
		opTok := p.advance()
		right, err := p.parseAddition()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Position: opTok.Position, Op: opTok.Value, Left: left, Right: right}, nil
	}
	return left, nil
}

func (p *Parser) parseAddition() (Node, error) {
	left, err := p.parseMultiplication()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.peek()
		if tok.Type == TokenPlus || tok.Type == TokenMinus {
			opTok := p.advance()
			right, err := p.parseMultiplication()
			if err != nil {
				return nil, err
			}
			left = &BinaryExpr{Position: opTok.Position, Op: opTok.Value, Left: left, Right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parseMultiplication() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.peek()
		if tok.Type == TokenStar || tok.Type == TokenSlash {
			opTok := p.advance()
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			left = &BinaryExpr{Position: opTok.Position, Op: opTok.Value, Left: left, Right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parseUnary() (Node, error) {
	if p.peek().Type == TokenMinus {
		opTok := p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Position: opTok.Position, Op: "-", Operand: operand}, nil
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (Node, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	// Range selector: expr[30d]
	for p.peek().Type == TokenLBracket {
		p.advance() // consume [

		// Range content can be:
		//   [30d] → NUMBER "30" + IDENTIFIER "d"
		//   [1y]  → NUMBER "1" + IDENTIFIER "y"
		//   [30]  → NUMBER "30"
		//   [3m]  → NUMBER "3" + IDENTIFIER "m"  (since 'm' is not a suffix the lexer keeps)
		var rangeStr string
		tok := p.peek()
		switch tok.Type {
		case TokenIdentifier:
			// e.g. [duration_var] — single identifier
			rangeStr = p.advance().Value
		case TokenNumber:
			rangeStr = p.advance().Value
			// Check if followed by an identifier (unit suffix like d, w, m, y)
			if p.peek().Type == TokenIdentifier {
				rangeStr += p.advance().Value
			}
		default:
			return nil, p.errorf(tok, "expected range duration in [...], got %s (%q)", tok.Type, tok.Value)
		}

		_, err := p.expect(TokenRBracket)
		if err != nil {
			return nil, err
		}
		days := parseDuration(rangeStr)
		expr = &RangeSelector{
			Position: tok.Position,
			Expr:     expr,
			Duration: rangeStr,
			Days:     days,
		}
	}

	return expr, nil
}

func (p *Parser) parsePrimary() (Node, error) {
	tok := p.peek()

	switch tok.Type {
	case TokenNumber:
		return p.parseNumberLiteral()

	case TokenString:
		p.advance()
		return &StringLiteral{Position: tok.Position, Value: tok.Value}, nil

	case TokenLParen:
		p.advance() // consume (
		inner, err := p.parsePipeExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(TokenRParen)
		if err != nil {
			return nil, err
		}
		return inner, nil

	case TokenStar:
		p.advance()
		return &Identifier{Position: tok.Position, Name: "*"}, nil

	case TokenIdentifier:
		return p.parseIdentifierOrCall()

	default:
		return nil, p.errorf(tok, "unexpected token %s (%q)", tok.Type, tok.Value)
	}
}

func (p *Parser) parseNumberLiteral() (Node, error) {
	tok := p.advance()
	val, err := parseNumber(tok.Value)
	if err != nil {
		return nil, p.errorf(tok, "invalid number %q: %v", tok.Value, err)
	}
	return &NumberLiteral{Position: tok.Position, Value: val, Raw: tok.Value}, nil
}

func (p *Parser) parseIdentifierOrCall() (Node, error) {
	tok := p.advance()
	name := tok.Value
	nameLower := strings.ToLower(name)

	// Check for boolean literals
	if nameLower == "true" {
		return &BoolLiteral{Position: tok.Position, Value: true}, nil
	}
	if nameLower == "false" {
		return &BoolLiteral{Position: tok.Position, Value: false}, nil
	}

	// Function call: name(...)
	if p.peek().Type == TokenLParen {
		return p.parseFunctionCall(tok)
	}

	// Plain identifier (ticker, field name, etc.)
	return &Identifier{Position: tok.Position, Name: name}, nil
}

func (p *Parser) parseFunctionCall(nameTok Token) (Node, error) {
	name := strings.ToLower(nameTok.Value)

	// Special handling for screener(...)
	if name == "screener" {
		return p.parseScreenerCall(nameTok)
	}

	// Special handling for alert(...)
	if name == "alert" {
		return p.parseAlertCall(nameTok)
	}

	p.advance() // consume (

	var args []Node
	if p.peek().Type != TokenRParen {
		for {
			arg, err := p.parsePipeExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			if p.peek().Type != TokenComma {
				break
			}
			p.advance() // consume ,
		}
	}

	_, err := p.expect(TokenRParen)
	if err != nil {
		return nil, err
	}

	return &FunctionCall{Position: nameTok.Position, Name: name, Args: args}, nil
}

func (p *Parser) parseScreenerCall(nameTok Token) (Node, error) {
	p.advance() // consume (

	filter, err := p.parsePipeExpr()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(TokenRParen)
	if err != nil {
		return nil, err
	}

	return &ScreenerExpr{Position: nameTok.Position, Filter: filter}, nil
}

func (p *Parser) parseAlertCall(nameTok Token) (Node, error) {
	p.advance() // consume (

	condition, err := p.parsePipeExpr()
	if err != nil {
		return nil, err
	}

	message := ""
	if p.peek().Type == TokenComma {
		p.advance() // consume ,
		msgTok, err := p.expect(TokenString)
		if err != nil {
			return nil, err
		}
		message = msgTok.Value
	}

	_, err = p.expect(TokenRParen)
	if err != nil {
		return nil, err
	}

	return &AlertExpr{Position: nameTok.Position, Condition: condition, Message: message}, nil
}

// ════════════════════════════════════════════════════════════════════
// Number & Duration Parsing Helpers
// ════════════════════════════════════════════════════════════════════

// parseNumber parses a number string with optional Indian suffix.
// Examples: "42", "3.14", "10000cr" → 100000000000, "5l" → 500000
func parseNumber(s string) (float64, error) {
	lower := strings.ToLower(s)

	multiplier := 1.0
	numStr := lower

	if strings.HasSuffix(lower, "crore") {
		multiplier = 1e7
		numStr = lower[:len(lower)-5]
	} else if strings.HasSuffix(lower, "cr") {
		multiplier = 1e7
		numStr = lower[:len(lower)-2]
	} else if strings.HasSuffix(lower, "lakh") {
		multiplier = 1e5
		numStr = lower[:len(lower)-4]
	} else if strings.HasSuffix(lower, "l") {
		multiplier = 1e5
		numStr = lower[:len(lower)-1]
	}

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}
	return val * multiplier, nil
}

// parseDuration parses a duration string into calendar days.
// Examples: "30d" → 30, "1w" → 7, "3m" → 90, "1y" → 365, "252d" → 252
func parseDuration(s string) int {
	lower := strings.ToLower(strings.TrimSpace(s))
	if len(lower) == 0 {
		return 0
	}

	// Extract numeric part and unit
	i := 0
	for i < len(lower) && (lower[i] >= '0' && lower[i] <= '9' || lower[i] == '.') {
		i++
	}
	numStr := lower[:i]
	unit := lower[i:]

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	switch unit {
	case "d", "day", "days":
		return int(num)
	case "w", "week", "weeks":
		return int(num * 7)
	case "m", "mo", "month", "months":
		return int(num * 30)
	case "y", "yr", "year", "years":
		return int(num * 365)
	case "": // bare number = days
		return int(num)
	default:
		return int(math.Round(num))
	}
}
