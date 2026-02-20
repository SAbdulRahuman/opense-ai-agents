# OpeNSE.ai — FinanceQL Language Reference

> FinanceQL is a PromQL-inspired domain-specific query language for financial time-series data analysis.

## Overview

FinanceQL lets you query stock data, compute technical indicators, and create screening conditions using a concise, composable syntax.

```
rsi(close("TCS", "1d", "365d"), 14) < 30
```

## Quick Start

```bash
# From CLI
openseai query 'price("RELIANCE")'
openseai query 'rsi(close("TCS", "1d", "365d"), 14)'

# Interactive REPL
openseai chat
> sma(close("INFY", "1d", "200d"), 50) > sma(close("INFY", "1d", "200d"), 200)
```

## Syntax

### Data Functions

| Function | Signature | Returns |
|----------|-----------|---------|
| `price` | `price(ticker)` | Latest quote |
| `close` | `close(ticker, timeframe, range)` | Close price vector |
| `open` | `open(ticker, timeframe, range)` | Open price vector |
| `high` | `high(ticker, timeframe, range)` | High price vector |
| `low` | `low(ticker, timeframe, range)` | Low price vector |
| `volume` | `volume(ticker, timeframe, range)` | Volume vector |
| `ohlcv` | `ohlcv(ticker, timeframe, range)` | Full OHLCV data |

**Parameters**:
- `ticker`: NSE ticker string, e.g., `"TCS"`, `"RELIANCE"`, `"NIFTY"`
- `timeframe`: `"1m"`, `"5m"`, `"15m"`, `"1h"`, `"1d"`, `"1w"`, `"1M"`
- `range`: Duration string, e.g., `"30d"`, `"90d"`, `"365d"`, `"1y"`, `"5y"`

### Technical Indicators

| Function | Signature | Description |
|----------|-----------|-------------|
| `sma` | `sma(vector, period)` | Simple Moving Average |
| `ema` | `ema(vector, period)` | Exponential Moving Average |
| `wma` | `wma(vector, period)` | Weighted Moving Average |
| `rsi` | `rsi(vector, period)` | Relative Strength Index |
| `macd` | `macd(vector, fast, slow, signal)` | MACD (returns object with `.macd`, `.signal`, `.histogram`) |
| `bollinger` | `bollinger(vector, period, mult)` | Bollinger Bands (`.upper`, `.middle`, `.lower`) |
| `atr` | `atr(ohlcv, period)` | Average True Range |
| `supertrend` | `supertrend(ohlcv, period, mult)` | SuperTrend indicator |
| `vwap` | `vwap(ohlcv)` | Volume Weighted Average Price |
| `stdev` | `stdev(vector)` | Standard deviation |

### Aggregation Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `avg` | `avg(vector)` | Mean value |
| `sum` | `sum(vector)` | Sum of all values |
| `min` | `min(vector)` | Minimum value |
| `max` | `max(vector)` | Maximum value |
| `count` | `count(vector)` | Number of elements |
| `last` | `last(vector)` | Most recent value |
| `first` | `first(vector)` | Oldest value |
| `percentile` | `percentile(vector, pct)` | Nth percentile |
| `median` | `median(vector)` | 50th percentile |
| `correlation` | `correlation(v1, v2)` | Pearson correlation |

### Signal Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `crossover` | `crossover(v1, v2)` | True when v1 crosses above v2 |
| `crossunder` | `crossunder(v1, v2)` | True when v1 crosses below v2 |
| `above` | `above(v1, v2)` | True when v1 > v2 |
| `below` | `below(v1, v2)` | True when v1 < v2 |

### Sorting & Filtering

| Function | Signature | Description |
|----------|-----------|-------------|
| `sort` | `sort(vector)` | Sort ascending |
| `sort_desc` | `sort_desc(vector)` | Sort descending |
| `top` | `top(vector, n)` | Top N values |
| `bottom` | `bottom(vector, n)` | Bottom N values |
| `abs` | `abs(vector)` | Absolute values |
| `round` | `round(vector, decimals)` | Round to N decimals |
| `change` | `change(vector)` | Period-over-period change |
| `change_pct` | `change_pct(vector)` | Period-over-period % change |

## Operators

### Arithmetic
```
+  -  *  /  %  ^
```

### Comparison
```
>  <  >=  <=  ==  !=
```

### Logical
```
AND  OR  NOT
```

### Pipe (Pipeline)
```
|    # Chain operations left-to-right
```

## Pipe Syntax

Use `|` to chain operations for readable queries:

```
close("RELIANCE", "1d", "365d") | sma(50) | crossover(ema(20))
```

Equivalent to:
```
crossover(sma(close("RELIANCE", "1d", "365d"), 50), ema(close("RELIANCE", "1d", "365d"), 20))
```

## Examples

### Basic Queries

```bash
# Current price
price("TCS")

# 14-day RSI
rsi(close("TCS", "1d", "365d"), 14)

# 50-day SMA
sma(close("RELIANCE", "1d", "200d"), 50)

# MACD histogram
macd(close("TCS", "1d", "365d"), 12, 26, 9).histogram
```

### Screening Conditions

```bash
# Oversold RSI
rsi(close("TCS", "1d", "365d"), 14) < 30

# Golden Cross (50 SMA > 200 SMA)
sma(close("INFY", "1d", "365d"), 50) > sma(close("INFY", "1d", "365d"), 200)

# Combined condition
rsi(close("TCS", "1d", "365d"), 14) < 30 AND
macd(close("TCS", "1d", "365d"), 12, 26, 9).histogram > 0
```

### Advanced Queries

```bash
# Volatility analysis
stdev(change_pct(close("RELIANCE", "1d", "90d")))

# Relative strength between two stocks
correlation(close("TCS", "1d", "365d"), close("INFY", "1d", "365d"))

# Price range percentage
(high("TCS", "1d", "30d") - low("TCS", "1d", "30d")) / close("TCS", "1d", "30d") * 100
```

### Pipeline Examples

```bash
# RSI with pipeline syntax
close("TCS", "1d", "365d") | rsi(14)

# Multi-step pipeline
close("RELIANCE", "1d", "365d") | sma(50) | crossover(ema(20))
```

## Internals

### Processing Pipeline

```
Query String → Lexer → Token Stream → Parser → AST → Evaluator → Result Value
```

1. **Lexer** (`lexer.go`): Rune-by-rune tokenization — identifiers, strings, numbers, operators, pipes
2. **Parser** (`parser.go`): Recursive descent parser — handles precedence, pipes, function calls, member access
3. **AST** (`ast.go`): Node types — `FunctionCall`, `BinaryOp`, `UnaryOp`, `PipeExpr`, `Literal`, `MemberAccess`
4. **Evaluator** (`evaluator.go`): Tree-walking evaluator with `EvalContext` for data resolution
5. **Functions** (`functions.go`): 40+ built-in functions registered via `RegisterBuiltins()`

### Value Types

| Type | Go Type | Description |
|------|---------|-------------|
| Scalar | `float64` | Single numeric value |
| Vector | `[]TimePoint` | Time-series data points |
| Boolean | `bool` | True/false |
| String | `string` | Text value |
| Object | `map[string]Value` | Composite (e.g., MACD result) |

## Configuration

```yaml
financeql:
  cache_ttl: 60              # Query result cache TTL (seconds)
  max_range: "365d"          # Maximum data range per query
  alert_check_interval: 30   # Alert re-evaluation interval (seconds)
  repl_history_file: "~/.openseai/financeql_history"
```

## Error Messages

| Error | Cause | Fix |
|-------|-------|-----|
| `unexpected token` | Syntax error in query | Check parentheses and operator usage |
| `unknown function` | Misspelled function name | Use `help()` to list available functions |
| `type mismatch` | Wrong argument type | Ensure vectors go to vector functions |
| `no data for ticker` | Ticker not found | Verify NSE ticker symbol |
| `range exceeds maximum` | Range > `max_range` | Reduce the date range |
