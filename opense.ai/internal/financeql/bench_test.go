package financeql

import (
	"testing"
)

// ── Lexer Benchmarks ──

func BenchmarkTokenizeSimple(b *testing.B) {
	input := `price("TCS")`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		l.Tokenize()
	}
}

func BenchmarkTokenizeMedium(b *testing.B) {
	input := `sma(close("RELIANCE", "1d", "200d"), 50) > sma(close("RELIANCE", "1d", "200d"), 200)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		l.Tokenize()
	}
}

func BenchmarkTokenizeComplex(b *testing.B) {
	input := `rsi(close("TCS", "1d", "365d"), 14) < 30 AND macd(close("TCS", "1d", "365d"), 12, 26, 9).histogram > 0 AND sma(close("TCS", "1d", "365d"), 50) > sma(close("TCS", "1d", "365d"), 200)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		l.Tokenize()
	}
}

func BenchmarkTokenizePipeline(b *testing.B) {
	input := `close("RELIANCE", "1d", "365d") | sma(50) | crossover(ema(20))`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		l.Tokenize()
	}
}

// ── Parser Benchmarks ──

func BenchmarkParseSimple(b *testing.B) {
	input := `price("TCS")`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseQuery(input)
	}
}

func BenchmarkParseMedium(b *testing.B) {
	input := `sma(close("RELIANCE", "1d", "200d"), 50) > sma(close("RELIANCE", "1d", "200d"), 200)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseQuery(input)
	}
}

func BenchmarkParseComplex(b *testing.B) {
	input := `rsi(close("TCS", "1d", "365d"), 14) < 30 AND macd(close("TCS", "1d", "365d"), 12, 26, 9).histogram > 0`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseQuery(input)
	}
}

func BenchmarkParsePipeline(b *testing.B) {
	input := `close("RELIANCE", "1d", "365d") | sma(50) | crossover(ema(20))`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseQuery(input)
	}
}

func BenchmarkParseNestedFunctions(b *testing.B) {
	input := `max(sma(close("TCS", "1d", "365d"), 20), ema(close("TCS", "1d", "365d"), 20))`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseQuery(input)
	}
}

func BenchmarkParseArithmeticExpression(b *testing.B) {
	input := `(high("TCS", "1d", "30d") - low("TCS", "1d", "30d")) / close("TCS", "1d", "30d") * 100`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseQuery(input)
	}
}
