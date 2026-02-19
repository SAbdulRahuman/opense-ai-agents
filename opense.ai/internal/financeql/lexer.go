package financeql

import (
	"fmt"
	"strings"
	"unicode"
)

// ════════════════════════════════════════════════════════════════════
// Token Types
// ════════════════════════════════════════════════════════════════════

// TokenType enumerates all token kinds produced by the lexer.
type TokenType int

const (
	// Special
	TokenEOF     TokenType = iota
	TokenIllegal           // unrecognized character

	// Literals
	TokenNumber     // 42, 3.14, 10000cr
	TokenString     // "hello"
	TokenIdentifier // RELIANCE, sma, sector, desc

	// Operators
	TokenPlus     // +
	TokenMinus    // -
	TokenStar     // *
	TokenSlash    // /
	TokenGT       // >
	TokenLT       // <
	TokenGTE      // >=
	TokenLTE      // <=
	TokenEQ       // ==
	TokenNEQ      // !=

	// Delimiters
	TokenLParen   // (
	TokenRParen   // )
	TokenLBracket // [
	TokenRBracket // ]
	TokenComma    // ,
	TokenPipe     // |

	// Keywords (logical)
	TokenAND // AND
	TokenOR  // OR
	TokenNOT // NOT
)

// tokenTypeNames maps token types to human-readable names.
var tokenTypeNames = map[TokenType]string{
	TokenEOF:        "EOF",
	TokenIllegal:    "ILLEGAL",
	TokenNumber:     "NUMBER",
	TokenString:     "STRING",
	TokenIdentifier: "IDENT",
	TokenPlus:       "+",
	TokenMinus:      "-",
	TokenStar:       "*",
	TokenSlash:      "/",
	TokenGT:         ">",
	TokenLT:         "<",
	TokenGTE:        ">=",
	TokenLTE:        "<=",
	TokenEQ:         "==",
	TokenNEQ:        "!=",
	TokenLParen:     "(",
	TokenRParen:     ")",
	TokenLBracket:   "[",
	TokenRBracket:   "]",
	TokenComma:      ",",
	TokenPipe:       "|",
	TokenAND:        "AND",
	TokenOR:         "OR",
	TokenNOT:        "NOT",
}

func (t TokenType) String() string {
	if name, ok := tokenTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Token(%d)", int(t))
}

// ════════════════════════════════════════════════════════════════════
// Token
// ════════════════════════════════════════════════════════════════════

// Token represents a single lexical token from the input.
type Token struct {
	Type     TokenType
	Value    string // literal text
	Position int    // byte offset in source
	Line     int    // 1-based
	Column   int    // 1-based
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q)@%d:%d", t.Type, t.Value, t.Line, t.Column)
}

// ════════════════════════════════════════════════════════════════════
// Lexer
// ════════════════════════════════════════════════════════════════════

// Lexer tokenizes a FinanceQL query string.
type Lexer struct {
	input  []rune
	pos    int // current position
	line   int
	col    int
	tokens []Token
}

// NewLexer creates a new Lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
		pos:   0,
		line:  1,
		col:   1,
	}
}

// Tokenize performs the complete tokenization and returns all tokens.
func (l *Lexer) Tokenize() ([]Token, error) {
	for {
		tok, err := l.nextToken()
		if err != nil {
			return nil, err
		}
		l.tokens = append(l.tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return l.tokens, nil
}

// ────────────────────────────────────────────────────────────────────
// Internal scanning
// ────────────────────────────────────────────────────────────────────

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) advance() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.advance()
	}
}

func (l *Lexer) skipComment() {
	// # line comments
	if l.pos < len(l.input) && l.input[l.pos] == '#' {
		for l.pos < len(l.input) && l.input[l.pos] != '\n' {
			l.advance()
		}
	}
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		l.skipWhitespace()
		if l.pos < len(l.input) && l.input[l.pos] == '#' {
			l.skipComment()
			continue
		}
		break
	}
}

func (l *Lexer) makeToken(typ TokenType, value string, pos, line, col int) Token {
	return Token{Type: typ, Value: value, Position: pos, Line: line, Column: col}
}

func (l *Lexer) nextToken() (Token, error) {
	l.skipWhitespaceAndComments()

	if l.pos >= len(l.input) {
		return l.makeToken(TokenEOF, "", l.pos, l.line, l.col), nil
	}

	startPos := l.pos
	startLine := l.line
	startCol := l.col
	ch := l.peek()

	// Single character tokens
	switch ch {
	case '(':
		l.advance()
		return l.makeToken(TokenLParen, "(", startPos, startLine, startCol), nil
	case ')':
		l.advance()
		return l.makeToken(TokenRParen, ")", startPos, startLine, startCol), nil
	case '[':
		l.advance()
		return l.makeToken(TokenLBracket, "[", startPos, startLine, startCol), nil
	case ']':
		l.advance()
		return l.makeToken(TokenRBracket, "]", startPos, startLine, startCol), nil
	case ',':
		l.advance()
		return l.makeToken(TokenComma, ",", startPos, startLine, startCol), nil
	case '|':
		l.advance()
		return l.makeToken(TokenPipe, "|", startPos, startLine, startCol), nil
	case '+':
		l.advance()
		return l.makeToken(TokenPlus, "+", startPos, startLine, startCol), nil
	case '-':
		l.advance()
		return l.makeToken(TokenMinus, "-", startPos, startLine, startCol), nil
	case '/':
		l.advance()
		return l.makeToken(TokenSlash, "/", startPos, startLine, startCol), nil
	case '*':
		// Check if it's a standalone wildcard or multiplication
		l.advance()
		return l.makeToken(TokenStar, "*", startPos, startLine, startCol), nil
	}

	// Two-character tokens
	if ch == '>' {
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(TokenGTE, ">=", startPos, startLine, startCol), nil
		}
		return l.makeToken(TokenGT, ">", startPos, startLine, startCol), nil
	}
	if ch == '<' {
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(TokenLTE, "<=", startPos, startLine, startCol), nil
		}
		return l.makeToken(TokenLT, "<", startPos, startLine, startCol), nil
	}
	if ch == '=' {
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(TokenEQ, "==", startPos, startLine, startCol), nil
		}
		// Single = treated as ==
		return l.makeToken(TokenEQ, "==", startPos, startLine, startCol), nil
	}
	if ch == '!' {
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(TokenNEQ, "!=", startPos, startLine, startCol), nil
		}
		return Token{}, &ParseError{
			Position: startPos,
			Line:     startLine,
			Column:   startCol,
			Message:  "unexpected '!', did you mean '!='?",
		}
	}

	// String literals
	if ch == '"' || ch == '\'' {
		return l.readString(ch, startPos, startLine, startCol)
	}

	// Numbers (digits or .digit)
	if unicode.IsDigit(ch) || (ch == '.' && l.pos+1 < len(l.input) && unicode.IsDigit(l.input[l.pos+1])) {
		return l.readNumber(startPos, startLine, startCol)
	}

	// Identifiers and keywords
	if unicode.IsLetter(ch) || ch == '_' {
		return l.readIdentifier(startPos, startLine, startCol)
	}

	l.advance()
	return Token{}, &ParseError{
		Position: startPos,
		Line:     startLine,
		Column:   startCol,
		Message:  fmt.Sprintf("unexpected character %q", ch),
	}
}

func (l *Lexer) readString(quote rune, startPos, startLine, startCol int) (Token, error) {
	l.advance() // consume opening quote
	var sb strings.Builder
	for {
		if l.pos >= len(l.input) {
			return Token{}, &ParseError{
				Position: startPos,
				Line:     startLine,
				Column:   startCol,
				Message:  "unterminated string literal",
			}
		}
		ch := l.advance()
		if ch == quote {
			break
		}
		if ch == '\\' {
			next := l.advance()
			switch next {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case '\'':
				sb.WriteByte('\'')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(next)
			}
			continue
		}
		sb.WriteRune(ch)
	}
	return l.makeToken(TokenString, sb.String(), startPos, startLine, startCol), nil
}

func (l *Lexer) readNumber(startPos, startLine, startCol int) (Token, error) {
	var sb strings.Builder
	hasDot := false

	for l.pos < len(l.input) {
		ch := l.peek()
		if unicode.IsDigit(ch) {
			sb.WriteRune(l.advance())
		} else if ch == '.' && !hasDot {
			hasDot = true
			sb.WriteRune(l.advance())
		} else {
			break
		}
	}

	// Check for suffix like 'cr' (crore) or 'l' (lakh)
	if l.pos < len(l.input) {
		ch := l.peek()
		if ch == 'c' || ch == 'C' || ch == 'l' || ch == 'L' {
			suffixStart := l.pos
			var sfx strings.Builder
			for l.pos < len(l.input) && unicode.IsLetter(l.peek()) {
				sfx.WriteRune(l.advance())
			}
			suffix := strings.ToLower(sfx.String())
			if suffix == "cr" || suffix == "crore" || suffix == "l" || suffix == "lakh" {
				sb.WriteString(sfx.String())
			} else {
				// Not a known suffix — roll back
				l.pos = suffixStart
				l.recalcLineCol()
			}
		}
	}

	return l.makeToken(TokenNumber, sb.String(), startPos, startLine, startCol), nil
}

func (l *Lexer) readIdentifier(startPos, startLine, startCol int) (Token, error) {
	var sb strings.Builder
	for l.pos < len(l.input) {
		ch := l.peek()
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			sb.WriteRune(l.advance())
		} else {
			break
		}
	}

	word := sb.String()
	upper := strings.ToUpper(word)

	// Check for keywords
	switch upper {
	case "AND":
		return l.makeToken(TokenAND, "AND", startPos, startLine, startCol), nil
	case "OR":
		return l.makeToken(TokenOR, "OR", startPos, startLine, startCol), nil
	case "NOT":
		return l.makeToken(TokenNOT, "NOT", startPos, startLine, startCol), nil
	}

	return l.makeToken(TokenIdentifier, word, startPos, startLine, startCol), nil
}

// recalcLineCol recalculates line and column from the beginning.
// Used after rolling back the position.
func (l *Lexer) recalcLineCol() {
	l.line = 1
	l.col = 1
	for i := 0; i < l.pos && i < len(l.input); i++ {
		if l.input[i] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
	}
}
