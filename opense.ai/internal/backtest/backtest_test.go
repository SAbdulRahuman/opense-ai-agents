package backtest

import (
	"math"
	"testing"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Test Helpers
// ════════════════════════════════════════════════════════════════════

// generateBars creates n daily OHLCV bars with a simple uptrend/downtrend pattern.
// The pattern ramps up for the first half, then down for the second half.
func generateBars(n int, startPrice float64) []models.OHLCV {
	bars := make([]models.OHLCV, n)
	base := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	price := startPrice

	for i := 0; i < n; i++ {
		if i < n/2 {
			price *= 1.01 // +1% per bar
		} else {
			price *= 0.99 // −1% per bar
		}
		bars[i] = models.OHLCV{
			Timestamp: base.AddDate(0, 0, i),
			Open:      price * 0.998,
			High:      price * 1.005,
			Low:       price * 0.995,
			Close:     price,
			Volume:    100000 + int64(i*1000),
		}
	}
	return bars
}

// steadyUptrend generates bars that only go up.
func steadyUptrend(n int, startPrice float64) []models.OHLCV {
	bars := make([]models.OHLCV, n)
	base := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	price := startPrice
	for i := 0; i < n; i++ {
		price *= 1.005
		bars[i] = models.OHLCV{
			Timestamp: base.AddDate(0, 0, i),
			Open:      price * 0.999,
			High:      price * 1.002,
			Low:       price * 0.998,
			Close:     price,
			Volume:    100000,
		}
	}
	return bars
}

// steadyDowntrend generates bars that only go down.
func steadyDowntrend(n int, startPrice float64) []models.OHLCV {
	bars := make([]models.OHLCV, n)
	base := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	price := startPrice
	for i := 0; i < n; i++ {
		price *= 0.995
		bars[i] = models.OHLCV{
			Timestamp: base.AddDate(0, 0, i),
			Open:      price * 1.001,
			High:      price * 1.002,
			Low:       price * 0.998,
			Close:     price,
			Volume:    100000,
		}
	}
	return bars
}

// simpleTestStrategy is a minimal strategy for testing the engine.
type simpleTestStrategy struct {
	name    string
	onBar   func(ctx *StrategyContext, bar models.OHLCV)
}

func (s *simpleTestStrategy) Name() string                                { return s.name }
func (s *simpleTestStrategy) Init(_ *StrategyContext)                     {}
func (s *simpleTestStrategy) OnBar(ctx *StrategyContext, bar models.OHLCV) {
	if s.onBar != nil {
		s.onBar(ctx, bar)
	}
}

// ════════════════════════════════════════════════════════════════════
// Engine Tests
// ════════════════════════════════════════════════════════════════════

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.InitialCapital != 1000000 {
		t.Errorf("expected InitialCapital=1000000, got %f", cfg.InitialCapital)
	}
	if cfg.SlippagePct != 0.001 {
		t.Errorf("expected SlippagePct=0.001, got %f", cfg.SlippagePct)
	}
	if cfg.Product != models.CNC {
		t.Errorf("expected Product=CNC, got %s", cfg.Product)
	}
	if cfg.RiskFreeRate != 0.065 {
		t.Errorf("expected RiskFreeRate=0.065, got %f", cfg.RiskFreeRate)
	}
}

func TestNewEngine_defaults(t *testing.T) {
	e := NewEngine(Config{})
	if e.cfg.InitialCapital != 1000000 {
		t.Error("expected default initial capital")
	}
	if e.cfg.Product != models.CNC {
		t.Error("expected default product CNC")
	}
	if e.cfg.RiskFreeRate != 0.065 {
		t.Error("expected default risk-free rate")
	}
}

func TestEngine_NilStrategy(t *testing.T) {
	e := NewEngine(DefaultConfig())
	bars := generateBars(10, 100)
	_, err := e.Run(nil, "TEST", bars)
	if err == nil {
		t.Error("expected error for nil strategy")
	}
}

func TestEngine_InsufficientBars(t *testing.T) {
	e := NewEngine(DefaultConfig())
	s := &simpleTestStrategy{name: "Empty"}
	_, err := e.Run(s, "TEST", []models.OHLCV{})
	if err == nil {
		t.Error("expected error for empty bars")
	}
	_, err = e.Run(s, "TEST", []models.OHLCV{{Close: 100}})
	if err == nil {
		t.Error("expected error for 1 bar")
	}
}

func TestEngine_DoNothing(t *testing.T) {
	e := NewEngine(DefaultConfig())
	s := &simpleTestStrategy{name: "DoNothing"}
	bars := generateBars(50, 100)

	result, err := e.Run(s, "INFY", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StrategyName != "DoNothing" {
		t.Errorf("expected strategy name DoNothing, got %s", result.StrategyName)
	}
	if result.Ticker != "INFY" {
		t.Errorf("expected ticker INFY, got %s", result.Ticker)
	}
	if result.TotalTrades != 0 {
		t.Errorf("expected 0 trades, got %d", result.TotalTrades)
	}
	if result.FinalCapital != result.InitialCapital {
		t.Errorf("expected final=initial when no trades, got final=%f", result.FinalCapital)
	}
	if len(result.EquityCurve) != len(bars) {
		t.Errorf("expected %d equity points, got %d", len(bars), len(result.EquityCurve))
	}
}

func TestEngine_BuyAndHold(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0 // zero slippage for deterministic test
	e := NewEngine(cfg)

	bars := steadyUptrend(20, 100)

	s := &simpleTestStrategy{
		name: "BuyBar0",
		onBar: func(ctx *StrategyContext, bar models.OHLCV) {
			if ctx.CurrentBar == 0 && ctx.Position == 0 {
				// Use 90% of cash to account for slippage & brokerage on fill
				qty := int(ctx.Cash * 0.9 / bar.Close)
				if qty > 0 {
					ctx.Buy(qty, "buy_all")
				}
			}
		},
	}

	result, err := e.Run(s, "TCS", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have at least 1 trade (the forced close at end)
	if result.TotalTrades < 1 {
		t.Errorf("expected at least 1 trade, got %d", result.TotalTrades)
	}
	// With uptrend, final should be > initial (roughly)
	if result.FinalCapital <= cfg.InitialCapital*0.95 {
		t.Errorf("expected capital increase during uptrend, got %f", result.FinalCapital)
	}
}

func TestEngine_SortsBars(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	bars := generateBars(10, 100)
	// Reverse the order
	reversed := make([]models.OHLCV, len(bars))
	for i, b := range bars {
		reversed[len(bars)-1-i] = b
	}

	s := &simpleTestStrategy{name: "Check"}
	result, err := e.Run(s, "TEST", reversed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be sorted — first equity date < last
	if !result.From.Before(result.To) {
		t.Error("expected From < To when bars are sorted")
	}
}

func TestEngine_MarketOrder_Fill(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	bars := generateBars(10, 100)
	buyBar := -1
	sellBar := -1

	s := &simpleTestStrategy{
		name: "BuySell",
		onBar: func(ctx *StrategyContext, bar models.OHLCV) {
			if ctx.CurrentBar == 2 && ctx.Position == 0 {
				ctx.Buy(10, "test_buy")
				buyBar = ctx.CurrentBar
			}
			if ctx.CurrentBar == 5 && ctx.Position > 0 {
				ctx.ClosePosition("test_sell")
				sellBar = ctx.CurrentBar
			}
		},
	}

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = buyBar
	_ = sellBar
	// Should have a trade from the explicit close
	if len(result.Trades) < 1 {
		t.Error("expected at least 1 completed trade")
	}
}

func TestEngine_LimitOrder(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	bars := generateBars(20, 100)

	s := &simpleTestStrategy{
		name: "LimitTest",
		onBar: func(ctx *StrategyContext, bar models.OHLCV) {
			if ctx.CurrentBar == 0 {
				// Place a limit buy well below market — may not fill
				ctx.BuyLimit(10, bar.Close*0.5, "limit_low")
			}
			if ctx.CurrentBar == 2 {
				// Place limit buy at market — should fill
				ctx.BuyLimit(10, bar.Close*1.1, "limit_fill")
			}
		},
	}

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// At least the end-of-backtest force close should create a trade
	if result.TotalTrades < 1 {
		t.Errorf("expected at least 1 trade")
	}
}

func TestEngine_Slippage(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0.01 // 1% slippage
	e := NewEngine(cfg)

	bars := steadyUptrend(10, 100)

	s := &simpleTestStrategy{
		name: "Slippage",
		onBar: func(ctx *StrategyContext, bar models.OHLCV) {
			if ctx.CurrentBar == 0 {
				ctx.Buy(10, "buy")
			}
		},
	}

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Slippage should reduce returns vs zero slippage
	if result.FinalCapital >= cfg.InitialCapital*1.20 {
		t.Error("1% slippage should materially impact returns")
	}
}

func TestEngine_BenchmarkReturn(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Benchmark = []models.OHLCV{
		{Close: 100},
		{Close: 120},
	}
	e := NewEngine(cfg)

	bars := generateBars(10, 100)
	s := &simpleTestStrategy{name: "Bench"}
	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 20.0 // (120-100)/100 * 100
	if math.Abs(result.BenchmarkReturn-expected) > 0.01 {
		t.Errorf("expected benchmark return %f, got %f", expected, result.BenchmarkReturn)
	}
}

func TestEngine_ForceClose(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	bars := generateBars(10, 100)

	s := &simpleTestStrategy{
		name: "HoldForever",
		onBar: func(ctx *StrategyContext, bar models.OHLCV) {
			if ctx.CurrentBar == 0 && ctx.Position == 0 {
				ctx.Buy(10, "hold_forever")
			}
			// never sell
		},
	}

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Engine should have forced close at end
	found := false
	for _, trade := range result.Trades {
		if trade.Reason == "backtest_end_close" {
			found = true
		}
	}
	if !found {
		t.Error("expected forced close trade at backtest end")
	}
}

func TestEngine_EquityCurve(t *testing.T) {
	e := NewEngine(DefaultConfig())
	bars := generateBars(30, 100)
	s := &simpleTestStrategy{name: "EC"}

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.EquityCurve) != 30 {
		t.Errorf("expected 30 equity points, got %d", len(result.EquityCurve))
	}
	// Equity should be constant when no trades
	for _, ep := range result.EquityCurve {
		if math.Abs(ep.Value-1000000) > 0.01 {
			t.Errorf("equity should be constant at 1M, got %f at %v", ep.Value, ep.Date)
			break
		}
	}
}

func TestEngine_CancelPending(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)
	bars := generateBars(10, 100)

	s := &simpleTestStrategy{
		name: "CancelTest",
		onBar: func(ctx *StrategyContext, bar models.OHLCV) {
			if ctx.CurrentBar == 0 {
				ctx.BuyLimit(100, bar.Close*0.5, "will_cancel")
			}
			if ctx.CurrentBar == 1 {
				ctx.CancelPending()
			}
		},
	}

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalTrades != 0 {
		t.Errorf("expected 0 trades after cancel, got %d", result.TotalTrades)
	}
}

// ════════════════════════════════════════════════════════════════════
// Strategy Context Tests
// ════════════════════════════════════════════════════════════════════

func TestContext_HistoricalBars(t *testing.T) {
	bars := generateBars(10, 100)
	ctx := &StrategyContext{
		Bars:       bars,
		CurrentBar: 5,
	}
	hist := ctx.HistoricalBars()
	if len(hist) != 6 { // 0..5 inclusive
		t.Errorf("expected 6 bars, got %d", len(hist))
	}
}

func TestContext_Closes(t *testing.T) {
	bars := generateBars(10, 100)
	ctx := &StrategyContext{
		Bars:       bars,
		CurrentBar: 4,
	}
	closes := ctx.Closes()
	if len(closes) != 5 {
		t.Errorf("expected 5 closes, got %d", len(closes))
	}
	for i, c := range closes {
		if c != bars[i].Close {
			t.Errorf("close mismatch at %d: expected %f, got %f", i, bars[i].Close, c)
		}
	}
}

func TestContext_LookBack(t *testing.T) {
	bars := generateBars(10, 100)
	ctx := &StrategyContext{
		Bars:       bars,
		CurrentBar: 5,
	}

	lb := ctx.LookBack(0)
	if lb.Close != bars[5].Close {
		t.Error("LookBack(0) should return current bar")
	}
	lb = ctx.LookBack(3)
	if lb.Close != bars[2].Close {
		t.Error("LookBack(3) should return bar 2")
	}
	lb = ctx.LookBack(100)
	if lb.Close != 0 {
		t.Error("LookBack out of range should return zero OHLCV")
	}
}

func TestContext_PortfolioValue(t *testing.T) {
	ctx := &StrategyContext{
		Cash:        500000,
		Position:    100,
		CurrentOHLCV: models.OHLCV{Close: 1000},
	}
	pv := ctx.PortfolioValue()
	expected := 500000.0 + 100*1000.0
	if pv != expected {
		t.Errorf("expected %f, got %f", expected, pv)
	}
}

func TestContext_PositionValue(t *testing.T) {
	ctx := &StrategyContext{
		Position:    50,
		CurrentOHLCV: models.OHLCV{Close: 200},
	}
	pv := ctx.PositionValue()
	if pv != 10000 {
		t.Errorf("expected 10000, got %f", pv)
	}
}

func TestContext_UnrealizedPnL_Long(t *testing.T) {
	ctx := &StrategyContext{
		Position:    10,
		AvgPrice:    100,
		CurrentOHLCV: models.OHLCV{Close: 120},
	}
	pnl := ctx.UnrealizedPnL()
	expected := 10.0 * (120 - 100)
	if pnl != expected {
		t.Errorf("expected %f, got %f", expected, pnl)
	}
}

func TestContext_UnrealizedPnL_Short(t *testing.T) {
	ctx := &StrategyContext{
		Position:    -10,
		AvgPrice:    100,
		CurrentOHLCV: models.OHLCV{Close: 80},
	}
	pnl := ctx.UnrealizedPnL()
	expected := 10.0 * (100 - 80)
	if pnl != expected {
		t.Errorf("expected %f, got %f", expected, pnl)
	}
}

func TestContext_UnrealizedPnL_Flat(t *testing.T) {
	ctx := &StrategyContext{
		Position: 0,
	}
	if ctx.UnrealizedPnL() != 0 {
		t.Error("expected 0 PnL when flat")
	}
}

func TestContext_StateStore(t *testing.T) {
	ctx := &StrategyContext{}

	ctx.Set("key1", "hello")
	ctx.Set("key2", 42.5)
	ctx.Set("key3", 7)

	v, ok := ctx.Get("key1")
	if !ok || v.(string) != "hello" {
		t.Error("expected key1=hello")
	}

	f := ctx.GetFloat64("key2")
	if f != 42.5 {
		t.Errorf("expected 42.5, got %f", f)
	}

	i := ctx.GetInt("key3")
	if i != 7 {
		t.Errorf("expected 7, got %d", i)
	}

	_, ok = ctx.Get("missing")
	if ok {
		t.Error("expected missing key to return false")
	}

	f = ctx.GetFloat64("missing")
	if f != 0 {
		t.Error("GetFloat64 for missing key should return 0")
	}

	i = ctx.GetInt("missing")
	if i != 0 {
		t.Error("GetInt for missing key should return 0")
	}
}

// ════════════════════════════════════════════════════════════════════
// Metrics Tests
// ════════════════════════════════════════════════════════════════════

func TestComputeMetrics_Nil(t *testing.T) {
	// Should not panic
	ComputeMetrics(nil, 0.065)
}

func TestComputeMetrics_NoTrades(t *testing.T) {
	r := &models.BacktestResult{
		InitialCapital: 1000000,
		FinalCapital:   1000000,
	}
	ComputeMetrics(r, 0.065)
	if r.TotalTrades != 0 {
		t.Error("expected 0 trades")
	}
	if r.WinRate != 0 {
		t.Error("expected 0 win rate")
	}
}

func TestComputeTradeStats(t *testing.T) {
	r := &models.BacktestResult{
		Trades: []models.BacktestTrade{
			{PnL: 100},
			{PnL: 200},
			{PnL: -50},
			{PnL: -30},
			{PnL: 150},
		},
	}
	computeTradeStats(r)

	if r.TotalTrades != 5 {
		t.Errorf("expected 5 trades, got %d", r.TotalTrades)
	}
	if r.WinningTrades != 3 {
		t.Errorf("expected 3 wins, got %d", r.WinningTrades)
	}
	if r.LosingTrades != 2 {
		t.Errorf("expected 2 losses, got %d", r.LosingTrades)
	}
	if r.WinRate != 60 {
		t.Errorf("expected 60%% win rate, got %f", r.WinRate)
	}
	expectedAvgWin := (100.0 + 200 + 150) / 3
	if math.Abs(r.AvgWin-expectedAvgWin) > 0.01 {
		t.Errorf("expected AvgWin=%f, got %f", expectedAvgWin, r.AvgWin)
	}
	expectedAvgLoss := (50.0 + 30) / 2
	if math.Abs(r.AvgLoss-expectedAvgLoss) > 0.01 {
		t.Errorf("expected AvgLoss=%f, got %f", expectedAvgLoss, r.AvgLoss)
	}
	expectedPF := (100.0 + 200 + 150) / (50.0 + 30)
	if math.Abs(r.ProfitFactor-expectedPF) > 0.01 {
		t.Errorf("expected ProfitFactor=%f, got %f", expectedPF, r.ProfitFactor)
	}
}

func TestComputeTradeStats_AllWins(t *testing.T) {
	r := &models.BacktestResult{
		Trades: []models.BacktestTrade{
			{PnL: 100},
			{PnL: 200},
		},
	}
	computeTradeStats(r)
	if r.WinRate != 100 {
		t.Errorf("expected 100%% win rate, got %f", r.WinRate)
	}
	if !math.IsInf(r.ProfitFactor, 1) {
		t.Errorf("expected infinite profit factor, got %f", r.ProfitFactor)
	}
}

func TestComputeCAGR(t *testing.T) {
	r := &models.BacktestResult{
		InitialCapital: 100000,
		FinalCapital:   200000,
		From:           time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		To:             time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	computeCAGR(r)
	// CAGR ≈ (2^(1/3) - 1) * 100 ≈ 26%
	if r.CAGR < 20 || r.CAGR > 30 {
		t.Errorf("expected CAGR ~26%%, got %f", r.CAGR)
	}
}

func TestComputeCAGR_ZeroDays(t *testing.T) {
	now := time.Now()
	r := &models.BacktestResult{
		InitialCapital: 100000,
		FinalCapital:   200000,
		From:           now,
		To:             now,
	}
	computeCAGR(r)
	if r.CAGR != 0 {
		t.Errorf("expected 0 CAGR for zero duration, got %f", r.CAGR)
	}
}

func TestComputeDrawdown(t *testing.T) {
	// Equity: 100, 120, 90, 110, 80
	r := &models.BacktestResult{
		EquityCurve: []models.EquityPoint{
			{Value: 100},
			{Value: 120},
			{Value: 90},
			{Value: 110},
			{Value: 80},
		},
	}
	computeDrawdown(r)
	// Peak=120, lowest after peak=80, drawdown=40, drawdownPct=(40/120)*100=33.33
	if math.Abs(r.MaxDrawdown-40) > 0.01 {
		t.Errorf("expected MaxDrawdown=40, got %f", r.MaxDrawdown)
	}
	if math.Abs(r.MaxDrawdownPct-33.333) > 0.5 {
		t.Errorf("expected MaxDrawdownPct≈33.33%%, got %f", r.MaxDrawdownPct)
	}
}

func TestComputeDrawdown_NoDrawdown(t *testing.T) {
	// Monotonically increasing
	r := &models.BacktestResult{
		EquityCurve: []models.EquityPoint{
			{Value: 100},
			{Value: 110},
			{Value: 120},
		},
	}
	computeDrawdown(r)
	if r.MaxDrawdown != 0 {
		t.Errorf("expected 0 drawdown, got %f", r.MaxDrawdown)
	}
}

func TestComputeDrawdown_Empty(t *testing.T) {
	r := &models.BacktestResult{}
	computeDrawdown(r)
	if r.MaxDrawdown != 0 {
		t.Errorf("expected 0 drawdown for empty curve, got %f", r.MaxDrawdown)
	}
}

func TestComputeSharpe(t *testing.T) {
	curve := make([]models.EquityPoint, 253)
	for i := range curve {
		curve[i] = models.EquityPoint{Value: 100 * (1 + float64(i)*0.001)}
	}
	r := &models.BacktestResult{EquityCurve: curve}
	computeSharpe(r, 0.065)
	// With steady daily return ~0.1%, Sharpe should be positive
	if r.SharpeRatio <= 0 {
		t.Errorf("expected positive Sharpe, got %f", r.SharpeRatio)
	}
}

func TestComputeSharpe_FewPoints(t *testing.T) {
	r := &models.BacktestResult{
		EquityCurve: []models.EquityPoint{{Value: 100}},
	}
	computeSharpe(r, 0.065)
	if r.SharpeRatio != 0 {
		t.Error("Sharpe should be 0 with insufficient data")
	}
}

func TestComputeSortino(t *testing.T) {
	// Create some variance with downside
	curve := make([]models.EquityPoint, 100)
	curve[0] = models.EquityPoint{Value: 100}
	for i := 1; i < 100; i++ {
		if i%5 == 0 {
			curve[i] = models.EquityPoint{Value: curve[i-1].Value * 0.99} // down day
		} else {
			curve[i] = models.EquityPoint{Value: curve[i-1].Value * 1.005} // up day
		}
	}
	r := &models.BacktestResult{EquityCurve: curve}
	computeSortino(r, 0.065)
	// Should be positive since more up than down
	if r.SortinoRatio <= 0 {
		t.Errorf("expected positive Sortino, got %f", r.SortinoRatio)
	}
}

func TestDailyReturns(t *testing.T) {
	curve := []models.EquityPoint{
		{Value: 100},
		{Value: 110},
		{Value: 105},
	}
	ret := dailyReturns(curve)
	if len(ret) != 2 {
		t.Fatalf("expected 2 returns, got %d", len(ret))
	}
	if math.Abs(ret[0]-0.1) > 1e-9 {
		t.Errorf("expected return[0]=0.1, got %f", ret[0])
	}
	if math.Abs(ret[1]-(-5.0/110.0)) > 1e-9 {
		t.Errorf("expected return[1]=%f, got %f", -5.0/110.0, ret[1])
	}
}

func TestDailyReturns_Empty(t *testing.T) {
	ret := dailyReturns(nil)
	if ret != nil {
		t.Error("expected nil for empty curve")
	}
	ret = dailyReturns([]models.EquityPoint{{Value: 100}})
	if ret != nil {
		t.Error("expected nil for single point")
	}
}

func TestMean(t *testing.T) {
	m := mean([]float64{10, 20, 30})
	if m != 20 {
		t.Errorf("expected 20, got %f", m)
	}
	if mean(nil) != 0 {
		t.Error("mean of nil should be 0")
	}
}

func TestStddev(t *testing.T) {
	sd := stddev([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	// Expected sample stddev ≈ 2.138
	if math.Abs(sd-2.138) > 0.01 {
		t.Errorf("expected stddev ≈ 2.138, got %f", sd)
	}
	if stddev(nil) != 0 {
		t.Error("stddev of nil should be 0")
	}
	if stddev([]float64{42}) != 0 {
		t.Error("stddev of single element should be 0")
	}
}

// ════════════════════════════════════════════════════════════════════
// Utility Functions Tests
// ════════════════════════════════════════════════════════════════════

func TestMaxConsecutiveWins(t *testing.T) {
	trades := []models.BacktestTrade{
		{PnL: 10}, {PnL: 20}, {PnL: -5}, {PnL: 15}, {PnL: 25}, {PnL: 30},
	}
	w := MaxConsecutiveWins(trades)
	if w != 3 {
		t.Errorf("expected 3 consecutive wins, got %d", w)
	}
}

func TestMaxConsecutiveLosses(t *testing.T) {
	trades := []models.BacktestTrade{
		{PnL: -10}, {PnL: -20}, {PnL: -5}, {PnL: 15}, {PnL: -1},
	}
	l := MaxConsecutiveLosses(trades)
	if l != 3 {
		t.Errorf("expected 3 consecutive losses, got %d", l)
	}
}

func TestExpectancyPerTrade(t *testing.T) {
	trades := []models.BacktestTrade{
		{PnL: 100}, {PnL: -50}, {PnL: 200},
	}
	e := ExpectancyPerTrade(trades)
	expected := (100.0 - 50 + 200) / 3
	if math.Abs(e-expected) > 0.01 {
		t.Errorf("expected %f, got %f", expected, e)
	}
	if ExpectancyPerTrade(nil) != 0 {
		t.Error("expected 0 for nil trades")
	}
}

func TestMedianTradePnL(t *testing.T) {
	// Odd count
	trades := []models.BacktestTrade{{PnL: 30}, {PnL: 10}, {PnL: 20}}
	m := MedianTradePnL(trades)
	if m != 20 {
		t.Errorf("expected 20, got %f", m)
	}

	// Even count
	trades = []models.BacktestTrade{{PnL: 10}, {PnL: 20}, {PnL: 30}, {PnL: 40}}
	m = MedianTradePnL(trades)
	if m != 25 {
		t.Errorf("expected 25, got %f", m)
	}

	if MedianTradePnL(nil) != 0 {
		t.Error("expected 0 for nil trades")
	}
}

func TestAverageHoldingPeriod(t *testing.T) {
	now := time.Now()
	trades := []models.BacktestTrade{
		{EntryDate: now, ExitDate: now.AddDate(0, 0, 10)},
		{EntryDate: now, ExitDate: now.AddDate(0, 0, 20)},
	}
	avg := AverageHoldingPeriod(trades)
	if math.Abs(avg-15) > 0.01 {
		t.Errorf("expected 15 days, got %f", avg)
	}
	if AverageHoldingPeriod(nil) != 0 {
		t.Error("expected 0 for nil trades")
	}
}

func TestMaxShares(t *testing.T) {
	if maxShares(10000, 100) != 100 {
		t.Error("expected 100 shares")
	}
	if maxShares(10000, 0) != 0 {
		t.Error("expected 0 for zero price")
	}
	if maxShares(10000, -5) != 0 {
		t.Error("expected 0 for negative price")
	}
	if maxShares(150, 100) != 1 {
		t.Error("expected 1 share")
	}
}

// ════════════════════════════════════════════════════════════════════
// Built-in Strategy Tests
// ════════════════════════════════════════════════════════════════════

func TestSMACrossover_Name(t *testing.T) {
	s := NewSMACrossover(20, 50)
	if s.Name() != "SMA Crossover" {
		t.Errorf("unexpected name: %s", s.Name())
	}
}

func TestSMACrossover_Run(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	bars := generateBars(120, 100)
	s := NewSMACrossover(10, 20)

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StrategyName != "SMA Crossover" {
		t.Error("wrong strategy name")
	}
	if len(result.EquityCurve) != len(bars) {
		t.Errorf("expected %d equity points", len(bars))
	}
}

func TestRSIMeanReversion_Name(t *testing.T) {
	s := NewRSIMeanReversion(14, 30, 70)
	if s.Name() != "RSI Mean Reversion" {
		t.Errorf("unexpected name: %s", s.Name())
	}
}

func TestRSIMeanReversion_Run(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	// Use larger dataset for RSI to have good values
	bars := generateBars(100, 100)
	s := NewRSIMeanReversion(14, 30, 70)

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StrategyName != "RSI Mean Reversion" {
		t.Error("wrong strategy name")
	}
}

func TestSuperTrendStrategy_Name(t *testing.T) {
	s := NewSuperTrendStrategy(7, 3)
	if s.Name() != "SuperTrend" {
		t.Errorf("unexpected name: %s", s.Name())
	}
}

func TestSuperTrendStrategy_Run(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	bars := generateBars(100, 100)
	s := NewSuperTrendStrategy(7, 3)

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StrategyName != "SuperTrend" {
		t.Error("wrong strategy name")
	}
}

func TestVWAPBreakout_Name(t *testing.T) {
	s := NewVWAPBreakout(20)
	if s.Name() != "VWAP Breakout" {
		t.Errorf("unexpected name: %s", s.Name())
	}
}

func TestVWAPBreakout_Run(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	bars := generateBars(100, 100)
	s := NewVWAPBreakout(10)

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StrategyName != "VWAP Breakout" {
		t.Error("wrong strategy name")
	}
}

func TestMACDCrossover_Name(t *testing.T) {
	s := NewMACDCrossover(12, 26, 9)
	if s.Name() != "MACD Crossover" {
		t.Errorf("unexpected name: %s", s.Name())
	}
}

func TestMACDCrossover_Run(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	bars := generateBars(120, 100)
	s := NewMACDCrossover(12, 26, 9)

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StrategyName != "MACD Crossover" {
		t.Error("wrong strategy name")
	}
}

func TestBuiltinStrategies(t *testing.T) {
	strategies := BuiltinStrategies()
	if len(strategies) != 5 {
		t.Errorf("expected 5 built-in strategies, got %d", len(strategies))
	}

	names := make(map[string]bool)
	for _, s := range strategies {
		names[s.Name()] = true
	}
	expected := []string{"SMA Crossover", "RSI Mean Reversion", "SuperTrend", "VWAP Breakout", "MACD Crossover"}
	for _, n := range expected {
		if !names[n] {
			t.Errorf("missing built-in strategy: %s", n)
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// Integration Test — Full Pipeline
// ════════════════════════════════════════════════════════════════════

func TestIntegration_AllStrategiesOnSameData(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0.001
	e := NewEngine(cfg)

	bars := generateBars(200, 500) // 200 bars from ₹500

	strategies := BuiltinStrategies()
	for _, strat := range strategies {
		result, err := e.Run(strat, "RELIANCE", bars)
		if err != nil {
			t.Errorf("strategy %s failed: %v", strat.Name(), err)
			continue
		}
		if result.InitialCapital != cfg.InitialCapital {
			t.Errorf("[%s] initial capital mismatch", strat.Name())
		}
		if result.FinalCapital <= 0 {
			t.Errorf("[%s] final capital should be positive", strat.Name())
		}
		if len(result.EquityCurve) != len(bars) {
			t.Errorf("[%s] equity curve length mismatch", strat.Name())
		}
	}
}

func TestIntegration_MetricsComputed(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	bars := generateBars(200, 100)
	s := NewSMACrossover(10, 30)

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With 200 bars of up/down pattern, should have trades and metrics
	if result.TotalReturn == 0 && result.TotalTrades > 0 {
		// Possible but unlikely — just verify metrics are computed
		t.Log("TotalReturn is 0 but trades exist — verify manually")
	}

	// CAGR should be computed (non-zero unless returns are exactly 0)
	// MaxDrawdown should be >= 0
	if result.MaxDrawdown < 0 {
		t.Error("MaxDrawdown should be >= 0")
	}
	if result.MaxDrawdownPct < 0 {
		t.Error("MaxDrawdownPct should be >= 0")
	}
	// WinRate should be 0-100
	if result.WinRate < 0 || result.WinRate > 100 {
		t.Errorf("WinRate out of bounds: %f", result.WinRate)
	}
}

func TestIntegration_UptrendProfit(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SlippagePct = 0
	e := NewEngine(cfg)

	// Pure uptrend — SMA crossover should profit eventually
	bars := steadyUptrend(200, 100)
	s := NewSMACrossover(5, 15)

	result, err := e.Run(s, "TEST", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// In a pure uptrend, a trend-following strategy should be profitable
	if result.TotalReturnPct < -10 {
		t.Errorf("expected near-positive returns in uptrend, got %f%%", result.TotalReturnPct)
	}
}
