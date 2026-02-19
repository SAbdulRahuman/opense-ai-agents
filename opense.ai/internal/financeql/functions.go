package financeql

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/seenimoa/openseai/internal/analysis/technical"
	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Built-in Function Registration
// ════════════════════════════════════════════════════════════════════

// RegisterBuiltins registers all built-in FinanceQL functions on the given EvalContext.
func RegisterBuiltins(ec *EvalContext) {
	// ── Price & Market ──────────────────────────────────────────
	ec.RegisterFunc("price", fnPrice)
	ec.RegisterFunc("price_range", fnPriceRange)
	ec.RegisterFunc("open", fnOpen)
	ec.RegisterFunc("high", fnHigh)
	ec.RegisterFunc("low", fnLow)
	ec.RegisterFunc("close", fnClose)
	ec.RegisterFunc("volume", fnVolume)
	ec.RegisterFunc("volume_range", fnVolumeRange)
	ec.RegisterFunc("returns", fnReturns)
	ec.RegisterFunc("change_pct", fnChangePct)
	ec.RegisterFunc("vix", fnVIX)

	// ── Technical Indicator Functions ────────────────────────────
	ec.RegisterFunc("sma", fnSMA)
	ec.RegisterFunc("ema", fnEMA)
	ec.RegisterFunc("rsi", fnRSI)
	ec.RegisterFunc("rsi_range", fnRSIRange)
	ec.RegisterFunc("macd", fnMACD)
	ec.RegisterFunc("bollinger", fnBollinger)
	ec.RegisterFunc("supertrend", fnSuperTrend)
	ec.RegisterFunc("atr", fnATR)
	ec.RegisterFunc("vwap", fnVWAP)
	ec.RegisterFunc("crossover", fnCrossover)
	ec.RegisterFunc("crossunder", fnCrossunder)

	// ── Fundamental Functions ────────────────────────────────────
	ec.RegisterFunc("pe", fnPE)
	ec.RegisterFunc("pb", fnPB)
	ec.RegisterFunc("roe", fnROE)
	ec.RegisterFunc("roce", fnROCE)
	ec.RegisterFunc("debt_equity", fnDebtEquity)
	ec.RegisterFunc("market_cap", fnMarketCap)
	ec.RegisterFunc("dividend_yield", fnDividendYield)
	ec.RegisterFunc("promoter_holding", fnPromoterHolding)
	ec.RegisterFunc("eve_ebitda", fnEVEBITDA)
	ec.RegisterFunc("eps", fnEPS)
	ec.RegisterFunc("book_value", fnBookValue)

	// ── Aggregation & Math Functions ─────────────────────────────
	ec.RegisterFunc("avg", fnAvg)
	ec.RegisterFunc("sum", fnSum)
	ec.RegisterFunc("min", fnMin)
	ec.RegisterFunc("max", fnMax)
	ec.RegisterFunc("stddev", fnStddev)
	ec.RegisterFunc("percentile", fnPercentile)
	ec.RegisterFunc("correlation", fnCorrelation)
	ec.RegisterFunc("abs", fnAbs)

	// ── Screening & Filtering ────────────────────────────────────
	ec.RegisterFunc("nifty50", fnNifty50)
	ec.RegisterFunc("niftybank", fnNiftyBank)
	ec.RegisterFunc("sector", fnSector)
	ec.RegisterFunc("sort", fnSort)
	ec.RegisterFunc("top", fnTop)
	ec.RegisterFunc("bottom", fnBottom)
	ec.RegisterFunc("where", fnWhere)

	// ── Utility / Display ────────────────────────────────────────
	ec.RegisterFunc("trend", fnTrend)
	ec.RegisterFunc("count", fnCount)
	ec.RegisterFunc("last", fnLast)
	ec.RegisterFunc("first", fnFirst)

	// ── Screener internal ────────────────────────────────────────
	ec.RegisterFunc("_screener", fnScreenerInternal)
}

// ════════════════════════════════════════════════════════════════════
// Argument helpers
// ════════════════════════════════════════════════════════════════════

func requireTicker(args []Value, pos int) (string, error) {
	if pos >= len(args) {
		return "", fmt.Errorf("missing ticker argument at position %d", pos)
	}
	v := args[pos]
	if v.Type == TypeString {
		return ResolveTicker(v.Str), nil
	}
	return "", fmt.Errorf("expected ticker string at position %d, got %s", pos, v.Type)
}

func optionalInt(args []Value, pos int, def int) int {
	if pos >= len(args) {
		return def
	}
	if args[pos].Type == TypeScalar {
		return int(args[pos].Scalar)
	}
	return def
}

func optionalFloat(args []Value, pos int, def float64) float64 {
	if pos >= len(args) {
		return def
	}
	if args[pos].Type == TypeScalar {
		return args[pos].Scalar
	}
	return def
}

func extractVector(args []Value) []float64 {
	for _, a := range args {
		if a.Type == TypeVector {
			v := make([]float64, len(a.Vector))
			for i, p := range a.Vector {
				v[i] = p.Value
			}
			return v
		}
	}
	return nil
}

// fetchCandles fetches historical candles for a ticker.
func fetchCandles(ec *EvalContext, ticker string, days int) ([]models.OHLCV, error) {
	return FetchHistorical(ec, ticker, days)
}

// ════════════════════════════════════════════════════════════════════
// Price & Market Functions
// ════════════════════════════════════════════════════════════════════

// price(TICKER) → latest closing price
func fnPrice(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}

	cacheKey := "price:" + ticker
	if v, ok := ec.Cache.Get(cacheKey); ok {
		return v, nil
	}

	quote, err := ec.Aggregator.YFinance().GetQuote(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), fmt.Errorf("failed to get quote for %s: %w", ticker, err)
	}
	val := ScalarValue(quote.LastPrice)
	ec.Cache.Set(cacheKey, val)
	return val, nil
}

// price_range(TICKER, days) → price time-series
func fnPriceRange(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	days := optionalInt(args, 1, 30)

	data, err := fetchCandles(ec, ticker, days)
	if err != nil {
		return NilValue(), err
	}
	return VectorValue(OHLCVToVector(data)), nil
}

func fnOpen(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	quote, err := ec.Aggregator.YFinance().GetQuote(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), err
	}
	return ScalarValue(quote.Open), nil
}

func fnHigh(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	quote, err := ec.Aggregator.YFinance().GetQuote(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), err
	}
	return ScalarValue(quote.High), nil
}

func fnLow(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	quote, err := ec.Aggregator.YFinance().GetQuote(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), err
	}
	return ScalarValue(quote.Low), nil
}

func fnClose(ec *EvalContext, args []Value) (Value, error) {
	return fnPrice(ec, args)
}

func fnVolume(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	quote, err := ec.Aggregator.YFinance().GetQuote(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), err
	}
	return ScalarValue(float64(quote.Volume)), nil
}

func fnVolumeRange(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	days := optionalInt(args, 1, 30)
	data, err := fetchCandles(ec, ticker, days)
	if err != nil {
		return NilValue(), err
	}
	pts := make([]TimePoint, len(data))
	for i, d := range data {
		pts[i] = TimePoint{Time: d.Timestamp, Value: float64(d.Volume)}
	}
	return VectorValue(pts), nil
}

func fnReturns(ec *EvalContext, args []Value) (Value, error) {
	// If pipe input is a vector, compute returns
	if len(args) > 0 && args[0].Type == TypeVector {
		vec := args[0].Vector
		if len(vec) < 2 {
			return VectorValue(nil), nil
		}
		ret := make([]TimePoint, len(vec)-1)
		for i := 1; i < len(vec); i++ {
			if vec[i-1].Value != 0 {
				ret[i-1] = TimePoint{
					Time:  vec[i].Time,
					Value: (vec[i].Value - vec[i-1].Value) / vec[i-1].Value,
				}
			}
		}
		return VectorValue(ret), nil
	}

	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	days := optionalInt(args, 1, 252)
	data, err := fetchCandles(ec, ticker, days)
	if err != nil {
		return NilValue(), err
	}
	if len(data) < 2 {
		return VectorValue(nil), nil
	}
	ret := make([]TimePoint, len(data)-1)
	for i := 1; i < len(data); i++ {
		if data[i-1].Close != 0 {
			ret[i-1] = TimePoint{
				Time:  data[i].Timestamp,
				Value: (data[i].Close - data[i-1].Close) / data[i-1].Close,
			}
		}
	}
	return VectorValue(ret), nil
}

func fnChangePct(ec *EvalContext, args []Value) (Value, error) {
	// change_pct(price(X), 30d) or change_pct(TICKER, 30d)
	if len(args) > 0 && args[0].Type == TypeVector {
		vec := args[0].Vector
		if len(vec) < 2 {
			return ScalarValue(0), nil
		}
		first := vec[0].Value
		last := vec[len(vec)-1].Value
		if first != 0 {
			return ScalarValue((last - first) / first * 100), nil
		}
		return ScalarValue(0), nil
	}
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	days := optionalInt(args, 1, 30)
	data, err := fetchCandles(ec, ticker, days)
	if err != nil {
		return NilValue(), err
	}
	if len(data) < 2 {
		return ScalarValue(0), nil
	}
	first := data[0].Close
	last := data[len(data)-1].Close
	if first != 0 {
		return ScalarValue((last - first) / first * 100), nil
	}
	return ScalarValue(0), nil
}

func fnVIX(ec *EvalContext, args []Value) (Value, error) {
	quote, err := ec.Aggregator.YFinance().GetQuote(ec.Ctx, "^INDIAVIX")
	if err != nil {
		// Fallback ticker
		quote, err = ec.Aggregator.YFinance().GetQuote(ec.Ctx, "INDIAVIX")
		if err != nil {
			return NilValue(), fmt.Errorf("failed to get India VIX: %w", err)
		}
	}
	return ScalarValue(quote.LastPrice), nil
}

// ════════════════════════════════════════════════════════════════════
// Technical Indicator Functions
// ════════════════════════════════════════════════════════════════════

func fnSMA(ec *EvalContext, args []Value) (Value, error) {
	// sma(vector, period) — pipe mode
	if len(args) > 0 && args[0].Type == TypeVector {
		data := vectorToFloat64(args[0].Vector)
		period := optionalInt(args, 1, 20)
		result := technical.SMA(data, period)
		if result == nil {
			return NilValue(), nil
		}
		return ScalarValue(result[len(result)-1]), nil
	}

	// sma(TICKER, period)
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	period := optionalInt(args, 1, 20)
	candles, err := fetchCandles(ec, ticker, period*3)
	if err != nil {
		return NilValue(), err
	}
	closes := ohlcvCloses(candles)
	val := technical.SMALatest(closes, period)
	return ScalarValue(val), nil
}

func fnEMA(ec *EvalContext, args []Value) (Value, error) {
	if len(args) > 0 && args[0].Type == TypeVector {
		data := vectorToFloat64(args[0].Vector)
		period := optionalInt(args, 1, 21)
		result := technical.EMA(data, period)
		if result == nil {
			return NilValue(), nil
		}
		return ScalarValue(result[len(result)-1]), nil
	}

	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	period := optionalInt(args, 1, 21)
	candles, err := fetchCandles(ec, ticker, period*3)
	if err != nil {
		return NilValue(), err
	}
	closes := ohlcvCloses(candles)
	val := technical.EMALatest(closes, period)
	return ScalarValue(val), nil
}

func fnRSI(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	period := optionalInt(args, 1, 14)
	candles, err := fetchCandles(ec, ticker, period*5)
	if err != nil {
		return NilValue(), err
	}
	val := technical.RSILatest(candles, period)
	return ScalarValue(val), nil
}

func fnRSIRange(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	period := optionalInt(args, 1, 14)
	days := optionalInt(args, 2, 90)
	candles, err := fetchCandles(ec, ticker, days+period*2)
	if err != nil {
		return NilValue(), err
	}
	rsiVals := technical.RSI(candles, period)
	if rsiVals == nil {
		return VectorValue(nil), nil
	}
	pts := make([]TimePoint, 0, len(rsiVals))
	for i, v := range rsiVals {
		if v != 0 || i > period {
			pts = append(pts, TimePoint{Time: candles[i].Timestamp, Value: v})
		}
	}
	return VectorValue(pts), nil
}

func fnMACD(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	fast := optionalInt(args, 1, 12)
	slow := optionalInt(args, 2, 26)
	signal := optionalInt(args, 3, 9)

	candles, err := fetchCandles(ec, ticker, slow*5)
	if err != nil {
		return NilValue(), err
	}
	macd := technical.MACDLatest(candles, fast, slow, signal)
	// Return as table row with MACD, Signal, Histogram
	row := map[string]interface{}{
		"macd_line": macd.MACDLine,
		"signal":    macd.SignalLine,
		"histogram": macd.Histogram,
	}
	return TableValue([]map[string]interface{}{row}), nil
}

func fnBollinger(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	period := optionalInt(args, 1, 20)
	mult := optionalFloat(args, 2, 2.0)

	candles, err := fetchCandles(ec, ticker, period*5)
	if err != nil {
		return NilValue(), err
	}
	bb := technical.BollingerLatest(candles, period, mult)
	row := map[string]interface{}{
		"upper":  bb.Upper,
		"middle": bb.Middle,
		"lower":  bb.Lower,
	}
	return TableValue([]map[string]interface{}{row}), nil
}

func fnSuperTrend(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	period := optionalInt(args, 1, 7)
	mult := optionalFloat(args, 2, 3.0)

	candles, err := fetchCandles(ec, ticker, period*10)
	if err != nil {
		return NilValue(), err
	}
	st := technical.SuperTrendLatest(candles, period, mult)
	row := map[string]interface{}{
		"value": st.Value,
		"trend": st.Trend,
	}
	return TableValue([]map[string]interface{}{row}), nil
}

func fnATR(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	period := optionalInt(args, 1, 14)

	candles, err := fetchCandles(ec, ticker, period*5)
	if err != nil {
		return NilValue(), err
	}
	val := technical.ATRLatest(candles, period)
	return ScalarValue(val), nil
}

func fnVWAP(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}

	candles, err := fetchCandles(ec, ticker, 30)
	if err != nil {
		return NilValue(), err
	}
	val := technical.VWAPLatest(candles)
	return ScalarValue(val), nil
}

func fnCrossover(ec *EvalContext, args []Value) (Value, error) {
	// crossover(sma(X, 50), sma(X, 200)) — checks if first > second (simplified)
	if len(args) < 2 {
		return BoolValue(false), nil
	}
	a := toScalar(args[0])
	b := toScalar(args[1])
	return BoolValue(a > b), nil
}

func fnCrossunder(ec *EvalContext, args []Value) (Value, error) {
	if len(args) < 2 {
		return BoolValue(false), nil
	}
	a := toScalar(args[0])
	b := toScalar(args[1])
	return BoolValue(a < b), nil
}

// ════════════════════════════════════════════════════════════════════
// Fundamental Functions
// ════════════════════════════════════════════════════════════════════

func fnPE(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	quote, err := ec.Aggregator.YFinance().GetQuote(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), err
	}
	return ScalarValue(quote.PE), nil
}

func fnPB(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	quote, err := ec.Aggregator.YFinance().GetQuote(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), err
	}
	return ScalarValue(quote.PB), nil
}

func fnROE(ec *EvalContext, args []Value) (Value, error) {
	return fetchRatioField(ec, args, func(r *models.FinancialRatios) float64 { return r.ROE })
}

func fnROCE(ec *EvalContext, args []Value) (Value, error) {
	return fetchRatioField(ec, args, func(r *models.FinancialRatios) float64 { return r.ROCE })
}

func fnDebtEquity(ec *EvalContext, args []Value) (Value, error) {
	return fetchRatioField(ec, args, func(r *models.FinancialRatios) float64 { return r.DebtEquity })
}

func fnMarketCap(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	quote, err := ec.Aggregator.YFinance().GetQuote(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), err
	}
	return ScalarValue(quote.MarketCap), nil
}

func fnDividendYield(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	quote, err := ec.Aggregator.YFinance().GetQuote(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), err
	}
	return ScalarValue(quote.DividendYield), nil
}

func fnPromoterHolding(ec *EvalContext, args []Value) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	profile, err := ec.Aggregator.Screener().GetStockProfile(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), err
	}
	if profile.Promoter != nil {
		return ScalarValue(profile.Promoter.PromoterHolding), nil
	}
	return ScalarValue(0), nil
}

func fnEVEBITDA(ec *EvalContext, args []Value) (Value, error) {
	return fetchRatioField(ec, args, func(r *models.FinancialRatios) float64 { return r.EVBITDA })
}

func fnEPS(ec *EvalContext, args []Value) (Value, error) {
	return fetchRatioField(ec, args, func(r *models.FinancialRatios) float64 { return r.EPS })
}

func fnBookValue(ec *EvalContext, args []Value) (Value, error) {
	return fetchRatioField(ec, args, func(r *models.FinancialRatios) float64 { return r.BookValue })
}

// fetchRatioField fetches a stock profile and extracts a ratio field.
func fetchRatioField(ec *EvalContext, args []Value, extract func(*models.FinancialRatios) float64) (Value, error) {
	ticker, err := requireTicker(args, 0)
	if err != nil {
		return NilValue(), err
	}
	profile, err := ec.Aggregator.Screener().GetStockProfile(ec.Ctx, ticker)
	if err != nil {
		return NilValue(), err
	}
	if profile.Ratios != nil {
		return ScalarValue(extract(profile.Ratios)), nil
	}
	return ScalarValue(0), nil
}

// ════════════════════════════════════════════════════════════════════
// Aggregation & Math Functions
// ════════════════════════════════════════════════════════════════════

func fnAvg(ec *EvalContext, args []Value) (Value, error) {
	vals := collectFloats(args)
	if len(vals) == 0 {
		return ScalarValue(0), nil
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return ScalarValue(sum / float64(len(vals))), nil
}

func fnSum(ec *EvalContext, args []Value) (Value, error) {
	vals := collectFloats(args)
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return ScalarValue(sum), nil
}

func fnMin(ec *EvalContext, args []Value) (Value, error) {
	vals := collectFloats(args)
	if len(vals) == 0 {
		return ScalarValue(0), nil
	}
	mn := vals[0]
	for _, v := range vals[1:] {
		if v < mn {
			mn = v
		}
	}
	return ScalarValue(mn), nil
}

func fnMax(ec *EvalContext, args []Value) (Value, error) {
	vals := collectFloats(args)
	if len(vals) == 0 {
		return ScalarValue(0), nil
	}
	mx := vals[0]
	for _, v := range vals[1:] {
		if v > mx {
			mx = v
		}
	}
	return ScalarValue(mx), nil
}

func fnStddev(ec *EvalContext, args []Value) (Value, error) {
	vals := collectFloats(args)
	if len(vals) < 2 {
		return ScalarValue(0), nil
	}
	mean := 0.0
	for _, v := range vals {
		mean += v
	}
	mean /= float64(len(vals))

	sumSq := 0.0
	for _, v := range vals {
		d := v - mean
		sumSq += d * d
	}
	return ScalarValue(math.Sqrt(sumSq / float64(len(vals)-1))), nil
}

func fnPercentile(ec *EvalContext, args []Value) (Value, error) {
	vals := collectFloats(args[:len(args)-1])
	pct := 50.0
	if len(args) > 0 && args[len(args)-1].Type == TypeScalar {
		pct = args[len(args)-1].Scalar
	}
	if len(vals) == 0 {
		return ScalarValue(0), nil
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)

	idx := (pct / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(idx))
	upper := lower + 1
	if upper >= len(sorted) {
		return ScalarValue(sorted[len(sorted)-1]), nil
	}
	frac := idx - float64(lower)
	return ScalarValue(sorted[lower] + frac*(sorted[upper]-sorted[lower])), nil
}

func fnCorrelation(ec *EvalContext, args []Value) (Value, error) {
	// correlation(vec1, vec2) or correlation(TICKER1, TICKER2, days)
	if len(args) >= 2 && args[0].Type == TypeVector && args[1].Type == TypeVector {
		a := vectorToFloat64(args[0].Vector)
		b := vectorToFloat64(args[1].Vector)
		return ScalarValue(pearson(a, b)), nil
	}
	// Ticker-based
	if len(args) >= 2 && args[0].Type == TypeString && args[1].Type == TypeString {
		days := optionalInt(args, 2, 90)
		dataA, err := fetchCandles(ec, ResolveTicker(args[0].Str), days)
		if err != nil {
			return NilValue(), err
		}
		dataB, err := fetchCandles(ec, ResolveTicker(args[1].Str), days)
		if err != nil {
			return NilValue(), err
		}
		a := ohlcvCloses(dataA)
		b := ohlcvCloses(dataB)
		return ScalarValue(pearson(a, b)), nil
	}
	return ScalarValue(0), nil
}

func fnAbs(ec *EvalContext, args []Value) (Value, error) {
	if len(args) > 0 && args[0].Type == TypeScalar {
		return ScalarValue(math.Abs(args[0].Scalar)), nil
	}
	return ScalarValue(0), nil
}

// ════════════════════════════════════════════════════════════════════
// Screening & Filtering Functions
// ════════════════════════════════════════════════════════════════════

// Nifty 50 constituent symbols (representative subset).
var nifty50Symbols = []string{
	"RELIANCE", "TCS", "HDFCBANK", "INFY", "ICICIBANK",
	"HINDUNILVR", "ITC", "SBIN", "BHARTIARTL", "KOTAKBANK",
	"LT", "AXISBANK", "BAJFINANCE", "ASIANPAINT", "MARUTI",
	"TITAN", "SUNPHARMA", "HCLTECH", "NTPC", "TATAMOTORS",
	"ULTRACEMCO", "WIPRO", "POWERGRID", "NESTLEIND", "ONGC",
	"JSWSTEEL", "ADANIENT", "ADANIPORTS", "TECHM", "TATASTEEL",
	"M_M", "BAJAJFINSV", "HDFCLIFE", "DIVISLAB", "DRREDDY",
	"SBILIFE", "BRITANNIA", "CIPLA", "COALINDIA", "INDUSINDBK",
	"GRASIM", "EICHERMOT", "APOLLOHOSP", "HEROMOTOCO", "TATACONSUM",
	"BPCL", "UPL", "BAJAJ_AUTO", "HINDALCO", "LTIM",
}

var niftyBankSymbols = []string{
	"HDFCBANK", "ICICIBANK", "KOTAKBANK", "AXISBANK", "SBIN",
	"INDUSINDBK", "BANDHANBNK", "FEDERALBNK", "IDFCFIRSTB", "PNB",
	"AUBANK", "BANKBARODA",
}

func fnNifty50(_ *EvalContext, _ []Value) (Value, error) {
	rows := make([]map[string]interface{}, len(nifty50Symbols))
	for i, s := range nifty50Symbols {
		rows[i] = map[string]interface{}{"ticker": s, "index": "NIFTY 50"}
	}
	return TableValue(rows), nil
}

func fnNiftyBank(_ *EvalContext, _ []Value) (Value, error) {
	rows := make([]map[string]interface{}, len(niftyBankSymbols))
	for i, s := range niftyBankSymbols {
		rows[i] = map[string]interface{}{"ticker": s, "index": "NIFTY BANK"}
	}
	return TableValue(rows), nil
}

func fnSector(_ *EvalContext, args []Value) (Value, error) {
	sector := ""
	if len(args) > 0 && args[0].Type == TypeString {
		sector = args[0].Str
	}
	row := map[string]interface{}{"sector": sector}
	return TableValue([]map[string]interface{}{row}), nil
}

func fnSort(_ *EvalContext, args []Value) (Value, error) {
	if len(args) > 0 && args[0].Type == TypeVector {
		vec := make([]TimePoint, len(args[0].Vector))
		copy(vec, args[0].Vector)
		desc := false
		if len(args) > 1 && args[1].Type == TypeString && strings.EqualFold(args[1].Str, "desc") {
			desc = true
		}
		sort.Slice(vec, func(i, j int) bool {
			if desc {
				return vec[i].Value > vec[j].Value
			}
			return vec[i].Value < vec[j].Value
		})
		return VectorValue(vec), nil
	}
	if len(args) > 0 && args[0].Type == TypeTable {
		// Already sorted, noop for now
		return args[0], nil
	}
	if len(args) > 0 {
		return args[0], nil
	}
	return NilValue(), nil
}

func fnTop(_ *EvalContext, args []Value) (Value, error) {
	n := optionalInt(args, 1, 10)
	if len(args) > 0 && args[0].Type == TypeVector {
		vec := args[0].Vector
		if n > len(vec) {
			n = len(vec)
		}
		return VectorValue(vec[:n]), nil
	}
	if len(args) > 0 && args[0].Type == TypeTable {
		table := args[0].Table
		if n > len(table) {
			n = len(table)
		}
		return TableValue(table[:n]), nil
	}
	return NilValue(), nil
}

func fnBottom(_ *EvalContext, args []Value) (Value, error) {
	n := optionalInt(args, 1, 10)
	if len(args) > 0 && args[0].Type == TypeVector {
		vec := args[0].Vector
		if n > len(vec) {
			n = len(vec)
		}
		return VectorValue(vec[len(vec)-n:]), nil
	}
	if len(args) > 0 && args[0].Type == TypeTable {
		table := args[0].Table
		if n > len(table) {
			n = len(table)
		}
		return TableValue(table[len(table)-n:]), nil
	}
	return NilValue(), nil
}

func fnWhere(_ *EvalContext, args []Value) (Value, error) {
	// Simplified: filter table rows based on condition value
	if len(args) > 0 && args[0].Type == TypeTable {
		return args[0], nil
	}
	return NilValue(), nil
}

func fnScreenerInternal(_ *EvalContext, args []Value) (Value, error) {
	// Placeholder — actual implementation would evaluate filter against stock universe
	filter := ""
	if len(args) > 0 && args[0].Type == TypeString {
		filter = args[0].Str
	}
	row := map[string]interface{}{
		"filter": filter,
		"status": "screener_pending",
		"note":   "screener requires live data connection",
	}
	return TableValue([]map[string]interface{}{row}), nil
}

// ════════════════════════════════════════════════════════════════════
// Utility / Display Functions
// ════════════════════════════════════════════════════════════════════

func fnTrend(_ *EvalContext, args []Value) (Value, error) {
	if len(args) > 0 && args[0].Type == TypeVector {
		vec := args[0].Vector
		if len(vec) < 2 {
			return StringValue("FLAT"), nil
		}
		first := vec[0].Value
		last := vec[len(vec)-1].Value
		if last > first*1.02 {
			return StringValue("UPTREND"), nil
		} else if last < first*0.98 {
			return StringValue("DOWNTREND"), nil
		}
		return StringValue("SIDEWAYS"), nil
	}
	return StringValue("UNKNOWN"), nil
}

func fnCount(_ *EvalContext, args []Value) (Value, error) {
	if len(args) > 0 {
		switch args[0].Type {
		case TypeVector:
			return ScalarValue(float64(len(args[0].Vector))), nil
		case TypeTable:
			return ScalarValue(float64(len(args[0].Table))), nil
		}
	}
	return ScalarValue(0), nil
}

func fnLast(_ *EvalContext, args []Value) (Value, error) {
	if len(args) > 0 && args[0].Type == TypeVector {
		vec := args[0].Vector
		if len(vec) > 0 {
			return ScalarValue(vec[len(vec)-1].Value), nil
		}
	}
	return ScalarValue(0), nil
}

func fnFirst(_ *EvalContext, args []Value) (Value, error) {
	if len(args) > 0 && args[0].Type == TypeVector {
		vec := args[0].Vector
		if len(vec) > 0 {
			return ScalarValue(vec[0].Value), nil
		}
	}
	return ScalarValue(0), nil
}

// ════════════════════════════════════════════════════════════════════
// Internal Helpers
// ════════════════════════════════════════════════════════════════════

func ohlcvCloses(candles []models.OHLCV) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	return closes
}

func vectorToFloat64(pts []TimePoint) []float64 {
	out := make([]float64, len(pts))
	for i, p := range pts {
		out[i] = p.Value
	}
	return out
}

func collectFloats(args []Value) []float64 {
	var vals []float64
	for _, a := range args {
		switch a.Type {
		case TypeScalar:
			vals = append(vals, a.Scalar)
		case TypeVector:
			for _, p := range a.Vector {
				vals = append(vals, p.Value)
			}
		}
	}
	return vals
}

func pearson(a, b []float64) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n < 2 {
		return 0
	}

	var sumA, sumB, sumAB, sumA2, sumB2 float64
	for i := 0; i < n; i++ {
		sumA += a[i]
		sumB += b[i]
		sumAB += a[i] * b[i]
		sumA2 += a[i] * a[i]
		sumB2 += b[i] * b[i]
	}

	fn := float64(n)
	num := fn*sumAB - sumA*sumB
	den := math.Sqrt((fn*sumA2 - sumA*sumA) * (fn*sumB2 - sumB*sumB))
	if den == 0 {
		return 0
	}
	return num / den
}
