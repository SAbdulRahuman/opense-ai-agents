package financeql

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/datasource"
)

// ════════════════════════════════════════════════════════════════════
// Interactive FinanceQL REPL
// ════════════════════════════════════════════════════════════════════

const (
	replBanner = `
╔═══════════════════════════════════════════════════╗
║           FinanceQL Interactive Shell              ║
║  Type queries, e.g. price(RELIANCE) | sma(*, 50)  ║
║  Commands: .help  .functions  .quit                ║
╚═══════════════════════════════════════════════════╝
`
	replPrompt = "fql> "
)

// REPL is the interactive query shell.
type REPL struct {
	ec      *EvalContext
	in      io.Reader
	out     io.Writer
	history []string
}

// NewREPL creates a new REPL with the given aggregator and default I/O.
func NewREPL(agg *datasource.Aggregator) *REPL {
	return &REPL{
		ec:  NewEvalContext(context.Background(), agg),
		in:  os.Stdin,
		out: os.Stdout,
	}
}

// NewREPLWithIO creates a REPL with explicit reader/writer (useful for testing).
func NewREPLWithIO(agg *datasource.Aggregator, in io.Reader, out io.Writer) *REPL {
	return &REPL{
		ec:  NewEvalContext(context.Background(), agg),
		in:  in,
		out: out,
	}
}

// Run starts the interactive loop. Blocks until EOF or .quit.
func (r *REPL) Run() {
	fmt.Fprint(r.out, replBanner)
	scanner := bufio.NewScanner(r.in)
	for {
		fmt.Fprint(r.out, replPrompt)
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Handle dot-commands
		if strings.HasPrefix(line, ".") {
			if r.handleCommand(line) {
				return // .quit
			}
			continue
		}

		r.history = append(r.history, line)
		r.execute(line)
	}
}

// handleCommand processes REPL dot-commands. Returns true if the REPL should exit.
func (r *REPL) handleCommand(cmd string) bool {
	switch strings.ToLower(strings.Fields(cmd)[0]) {
	case ".quit", ".exit", ".q":
		fmt.Fprintln(r.out, "Goodbye!")
		return true

	case ".help":
		r.printHelp()

	case ".functions", ".funcs":
		r.printFunctions()

	case ".history":
		for i, h := range r.history {
			fmt.Fprintf(r.out, "  %d  %s\n", i+1, h)
		}

	case ".clear":
		r.history = nil
		fmt.Fprintln(r.out, "History cleared.")

	default:
		fmt.Fprintf(r.out, "Unknown command: %s  (type .help for help)\n", cmd)
	}
	return false
}

func (r *REPL) printHelp() {
	help := `
FinanceQL Quick Reference
─────────────────────────
  price(RELIANCE)              → Get latest price
  rsi(TCS, 14)                 → 14-period RSI
  price(INFY)[30d]             → 30-day price history
  sma(HDFCBANK, 50)            → 50-day Simple Moving Average
  price(RELIANCE) | sma(*, 20) → Pipe price into SMA
  pe(TCS) > 30 AND rsi(TCS) < 40  → Boolean expression
  screener(rsi(*,14) < 30 AND pe(*) < 20)  → Stock screener
  nifty50() | top(*, 10)       → Top 10 from Nifty 50

Dot-Commands:
  .help        Show this help
  .functions   List all built-in functions
  .history     Show query history
  .clear       Clear history
  .quit        Exit REPL

Number Suffixes: 1cr = 10M, 1l = 100K
Range Suffixes: 7d = 7 days, 2w = 14 days, 3m = 90 days, 1y = 365 days
`
	fmt.Fprint(r.out, help)
}

func (r *REPL) printFunctions() {
	names := make([]string, 0, len(r.ec.Functions))
	for name := range r.ec.Functions {
		if strings.HasPrefix(name, "_") {
			continue // internal functions
		}
		names = append(names, name)
	}
	sort.Strings(names)

	categories := map[string][]string{
		"Price":       {},
		"Technical":   {},
		"Fundamental": {},
		"Aggregation": {},
		"Screening":   {},
		"Utility":     {},
	}

	priceSet := map[string]bool{"price": true, "open": true, "high": true, "low": true, "close": true, "volume": true, "returns": true, "change_pct": true, "vix": true, "price_range": true, "volume_range": true}
	techSet := map[string]bool{"sma": true, "ema": true, "rsi": true, "rsi_range": true, "macd": true, "bollinger": true, "supertrend": true, "atr": true, "vwap": true, "crossover": true, "crossunder": true}
	fundSet := map[string]bool{"pe": true, "pb": true, "roe": true, "roce": true, "debt_equity": true, "market_cap": true, "dividend_yield": true, "promoter_holding": true, "eve_ebitda": true, "eps": true, "book_value": true}
	aggSet := map[string]bool{"avg": true, "sum": true, "min": true, "max": true, "stddev": true, "percentile": true, "correlation": true, "abs": true}
	screenSet := map[string]bool{"nifty50": true, "niftybank": true, "sector": true, "sort": true, "top": true, "bottom": true, "where": true}

	for _, name := range names {
		switch {
		case priceSet[name]:
			categories["Price"] = append(categories["Price"], name)
		case techSet[name]:
			categories["Technical"] = append(categories["Technical"], name)
		case fundSet[name]:
			categories["Fundamental"] = append(categories["Fundamental"], name)
		case aggSet[name]:
			categories["Aggregation"] = append(categories["Aggregation"], name)
		case screenSet[name]:
			categories["Screening"] = append(categories["Screening"], name)
		default:
			categories["Utility"] = append(categories["Utility"], name)
		}
	}

	order := []string{"Price", "Technical", "Fundamental", "Aggregation", "Screening", "Utility"}
	fmt.Fprintln(r.out, "\nBuilt-in Functions")
	fmt.Fprintln(r.out, "──────────────────")
	for _, cat := range order {
		fns := categories[cat]
		if len(fns) == 0 {
			continue
		}
		fmt.Fprintf(r.out, "  %s: %s\n", cat, strings.Join(fns, ", "))
	}
	fmt.Fprintln(r.out)
}

func (r *REPL) execute(query string) {
	start := time.Now()

	node, err := ParseQuery(query)
	if err != nil {
		fmt.Fprintf(r.out, "Parse error: %v\n", err)
		return
	}

	result, err := Eval(r.ec, node)
	if err != nil {
		fmt.Fprintf(r.out, "Eval error: %v\n", err)
		return
	}

	elapsed := time.Since(start)
	r.formatResult(result)
	fmt.Fprintf(r.out, "  (%s)\n", elapsed.Round(time.Millisecond))
}

// formatResult renders a Value to the REPL output.
func (r *REPL) formatResult(v Value) {
	switch v.Type {
	case TypeScalar:
		fmt.Fprintf(r.out, "→ %.4f\n", v.Scalar)

	case TypeString:
		fmt.Fprintf(r.out, "→ %s\n", v.Str)

	case TypeBool:
		if v.Bool {
			fmt.Fprintln(r.out, "→ true ✓")
		} else {
			fmt.Fprintln(r.out, "→ false ✗")
		}

	case TypeVector:
		r.formatVector(v.Vector)

	case TypeTable:
		r.formatTable(v.Table)

	case TypeMatrix:
		for k, vec := range v.Matrix {
			fmt.Fprintf(r.out, "── %s ──\n", k)
			r.formatVector(vec)
		}

	case TypeNil:
		fmt.Fprintln(r.out, "→ nil")
	}
}

func (r *REPL) formatVector(pts []TimePoint) {
	if len(pts) == 0 {
		fmt.Fprintln(r.out, "→ [] (empty)")
		return
	}

	// Print summary
	fmt.Fprintf(r.out, "→ Vector[%d points]\n", len(pts))

	// Show first/last + sparkline
	first := pts[0]
	last := pts[len(pts)-1]

	if !first.Time.IsZero() {
		fmt.Fprintf(r.out, "  First: %.4f (%s)\n", first.Value, first.Time.Format("2006-01-02"))
		fmt.Fprintf(r.out, "  Last:  %.4f (%s)\n", last.Value, last.Time.Format("2006-01-02"))
	} else {
		fmt.Fprintf(r.out, "  First: %.4f\n", first.Value)
		fmt.Fprintf(r.out, "  Last:  %.4f\n", last.Value)
	}

	// Compute stats
	mn, mx, sum := pts[0].Value, pts[0].Value, 0.0
	for _, p := range pts {
		sum += p.Value
		if p.Value < mn {
			mn = p.Value
		}
		if p.Value > mx {
			mx = p.Value
		}
	}
	avg := sum / float64(len(pts))
	fmt.Fprintf(r.out, "  Min:   %.4f  Max: %.4f  Avg: %.4f\n", mn, mx, avg)

	// Sparkline
	if len(pts) > 1 {
		fmt.Fprintf(r.out, "  %s\n", sparkline(pts))
	}
}

func (r *REPL) formatTable(rows []map[string]interface{}) {
	if len(rows) == 0 {
		fmt.Fprintln(r.out, "→ (empty table)")
		return
	}

	// Gather column names
	colSet := make(map[string]bool)
	var cols []string
	for _, row := range rows {
		for k := range row {
			if !colSet[k] {
				colSet[k] = true
				cols = append(cols, k)
			}
		}
	}
	sort.Strings(cols)

	// Compute column widths
	widths := make(map[string]int)
	for _, c := range cols {
		widths[c] = len(c)
	}
	for _, row := range rows {
		for _, c := range cols {
			s := fmt.Sprintf("%v", row[c])
			if len(s) > widths[c] {
				widths[c] = len(s)
			}
		}
	}

	// Print header
	fmt.Fprintf(r.out, "→ Table[%d rows]\n", len(rows))
	var header, sep strings.Builder
	for i, c := range cols {
		if i > 0 {
			header.WriteString(" │ ")
			sep.WriteString("─┼─")
		}
		header.WriteString(padRight(c, widths[c]))
		sep.WriteString(strings.Repeat("─", widths[c]))
	}
	fmt.Fprintf(r.out, "  %s\n", header.String())
	fmt.Fprintf(r.out, "  %s\n", sep.String())

	// Print rows (cap at 50)
	limit := len(rows)
	if limit > 50 {
		limit = 50
	}
	for _, row := range rows[:limit] {
		var line strings.Builder
		for i, c := range cols {
			if i > 0 {
				line.WriteString(" │ ")
			}
			s := fmt.Sprintf("%v", row[c])
			line.WriteString(padRight(s, widths[c]))
		}
		fmt.Fprintf(r.out, "  %s\n", line.String())
	}
	if len(rows) > 50 {
		fmt.Fprintf(r.out, "  ... and %d more rows\n", len(rows)-50)
	}
}

// ════════════════════════════════════════════════════════════════════
// Formatting Helpers
// ════════════════════════════════════════════════════════════════════

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// sparkline renders an ASCII sparkline for a time-series.
func sparkline(pts []TimePoint) string {
	if len(pts) == 0 {
		return ""
	}
	// Determine range
	mn, mx := pts[0].Value, pts[0].Value
	for _, p := range pts {
		if p.Value < mn {
			mn = p.Value
		}
		if p.Value > mx {
			mx = p.Value
		}
	}
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	span := mx - mn
	if span == 0 {
		span = 1
	}

	// Resample to max 60 chars
	width := len(pts)
	if width > 60 {
		width = 60
	}

	var sb strings.Builder
	for i := 0; i < width; i++ {
		idx := i * len(pts) / width
		norm := (pts[idx].Value - mn) / span
		bi := int(norm * float64(len(blocks)-1))
		if bi < 0 {
			bi = 0
		}
		if bi >= len(blocks) {
			bi = len(blocks) - 1
		}
		sb.WriteRune(blocks[bi])
	}
	return sb.String()
}

// GetFunctionNames returns all registered function names (for tab-completion).
func (r *REPL) GetFunctionNames() []string {
	names := make([]string, 0, len(r.ec.Functions))
	for name := range r.ec.Functions {
		if !strings.HasPrefix(name, "_") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// History returns the REPL's query history.
func (r *REPL) History() []string {
	return r.history
}
