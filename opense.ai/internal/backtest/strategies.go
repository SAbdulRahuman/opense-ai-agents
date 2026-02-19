package backtest

import (
	"github.com/seenimoa/openseai/internal/analysis/technical"
	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Built-in Strategies
// ════════════════════════════════════════════════════════════════════

// BuiltinStrategies returns all built-in strategies with default parameters.
func BuiltinStrategies() []Strategy {
	return []Strategy{
		NewSMACrossover(20, 50),
		NewRSIMeanReversion(14, 30, 70),
		NewSuperTrendStrategy(7, 3.0),
		NewVWAPBreakout(20),
		NewMACDCrossover(12, 26, 9),
	}
}

// ────────────────────────────────────────────────────────────────────
// 1. SMA Crossover Strategy
// ────────────────────────────────────────────────────────────────────

// SMACrossover implements a dual-SMA crossover strategy.
// Buys when faster SMA crosses above slower SMA, sells when it crosses below.
type SMACrossover struct {
	FastPeriod int
	SlowPeriod int
}

// NewSMACrossover creates a new SMA Crossover strategy.
func NewSMACrossover(fast, slow int) *SMACrossover {
	return &SMACrossover{FastPeriod: fast, SlowPeriod: slow}
}

func (s *SMACrossover) Name() string { return "SMA Crossover" }
func (s *SMACrossover) Init(_ *StrategyContext) {}

func (s *SMACrossover) OnBar(ctx *StrategyContext, bar models.OHLCV) {
	if ctx.CurrentBar < s.SlowPeriod+1 {
		return
	}

	closes := ctx.Closes()
	fastSMA := technical.SMA(closes, s.FastPeriod)
	slowSMA := technical.SMA(closes, s.SlowPeriod)
	if fastSMA == nil || slowSMA == nil {
		return
	}

	idx := ctx.CurrentBar
	prev := idx - 1

	fastNow := fastSMA[idx]
	slowNow := slowSMA[idx]
	fastPrev := fastSMA[prev]
	slowPrev := slowSMA[prev]

	// Bullish crossover — fast crosses above slow
	if fastPrev <= slowPrev && fastNow > slowNow {
		if ctx.Position <= 0 {
			if ctx.Position < 0 {
				ctx.ClosePosition("SMA bearish exit")
			}
			qty := maxShares(ctx.Cash, bar.Close)
			if qty > 0 {
				ctx.Buy(qty, "SMA bullish crossover")
			}
		}
	}

	// Bearish crossover — fast crosses below slow
	if fastPrev >= slowPrev && fastNow < slowNow {
		if ctx.Position > 0 {
			ctx.ClosePosition("SMA bearish crossover")
		}
	}
}

// ────────────────────────────────────────────────────────────────────
// 2. RSI Mean Reversion Strategy
// ────────────────────────────────────────────────────────────────────

// RSIMeanReversion buys when RSI drops below oversold, sells when above overbought.
type RSIMeanReversion struct {
	Period     int
	Oversold   float64
	Overbought float64
}

// NewRSIMeanReversion creates a new RSI Mean Reversion strategy.
func NewRSIMeanReversion(period int, oversold, overbought float64) *RSIMeanReversion {
	return &RSIMeanReversion{Period: period, Oversold: oversold, Overbought: overbought}
}

func (s *RSIMeanReversion) Name() string { return "RSI Mean Reversion" }
func (s *RSIMeanReversion) Init(_ *StrategyContext) {}

func (s *RSIMeanReversion) OnBar(ctx *StrategyContext, bar models.OHLCV) {
	if ctx.CurrentBar < s.Period+2 {
		return
	}

	bars := ctx.HistoricalBars()
	rsiVals := technical.RSI(bars[:ctx.CurrentBar+1], s.Period)
	if rsiVals == nil {
		return
	}

	rsi := rsiVals[ctx.CurrentBar]
	prevRSI := rsiVals[ctx.CurrentBar-1]

	// Enter long when RSI crosses above oversold
	if prevRSI <= s.Oversold && rsi > s.Oversold && ctx.Position <= 0 {
		if ctx.Position < 0 {
			ctx.ClosePosition("RSI exit short")
		}
		qty := maxShares(ctx.Cash, bar.Close)
		if qty > 0 {
			ctx.Buy(qty, "RSI oversold bounce")
		}
	}

	// Exit long when RSI crosses above overbought
	if prevRSI <= s.Overbought && rsi > s.Overbought && ctx.Position > 0 {
		ctx.ClosePosition("RSI overbought exit")
	}
}

// ────────────────────────────────────────────────────────────────────
// 3. SuperTrend Strategy
// ────────────────────────────────────────────────────────────────────

// SuperTrendStrategy follows the SuperTrend indicator trend.
type SuperTrendStrategy struct {
	Period     int
	Multiplier float64
}

// NewSuperTrendStrategy creates a new SuperTrend strategy.
func NewSuperTrendStrategy(period int, mult float64) *SuperTrendStrategy {
	return &SuperTrendStrategy{Period: period, Multiplier: mult}
}

func (s *SuperTrendStrategy) Name() string { return "SuperTrend" }
func (s *SuperTrendStrategy) Init(_ *StrategyContext) {}

func (s *SuperTrendStrategy) OnBar(ctx *StrategyContext, bar models.OHLCV) {
	if ctx.CurrentBar < s.Period+1 {
		return
	}

	bars := ctx.HistoricalBars()
	stData := technical.SuperTrend(bars[:ctx.CurrentBar+1], s.Period, s.Multiplier)
	if stData == nil || len(stData) < ctx.CurrentBar+1 {
		return
	}

	curr := stData[ctx.CurrentBar]
	prev := stData[ctx.CurrentBar-1]

	// Trend flip to UP — go long
	if prev.Trend == "DOWN" && curr.Trend == "UP" {
		if ctx.Position <= 0 {
			if ctx.Position < 0 {
				ctx.ClosePosition("SuperTrend trend flip exit short")
			}
			qty := maxShares(ctx.Cash, bar.Close)
			if qty > 0 {
				ctx.Buy(qty, "SuperTrend UP signal")
			}
		}
	}

	// Trend flip to DOWN — exit long
	if prev.Trend == "UP" && curr.Trend == "DOWN" {
		if ctx.Position > 0 {
			ctx.ClosePosition("SuperTrend DOWN signal")
		}
	}
}

// ────────────────────────────────────────────────────────────────────
// 4. VWAP Breakout Strategy
// ────────────────────────────────────────────────────────────────────

// VWAPBreakout trades breakouts above/below VWAP with SMA confirmation.
type VWAPBreakout struct {
	SMAPeriod int // SMA period for trend confirmation
}

// NewVWAPBreakout creates a new VWAP Breakout strategy.
func NewVWAPBreakout(smaPeriod int) *VWAPBreakout {
	return &VWAPBreakout{SMAPeriod: smaPeriod}
}

func (s *VWAPBreakout) Name() string { return "VWAP Breakout" }
func (s *VWAPBreakout) Init(_ *StrategyContext) {}

func (s *VWAPBreakout) OnBar(ctx *StrategyContext, bar models.OHLCV) {
	if ctx.CurrentBar < s.SMAPeriod+1 {
		return
	}

	bars := ctx.HistoricalBars()
	vwapVals := technical.VWAP(bars[:ctx.CurrentBar+1])
	if vwapVals == nil {
		return
	}

	closes := ctx.Closes()
	smaVals := technical.SMA(closes, s.SMAPeriod)
	if smaVals == nil {
		return
	}

	vwap := vwapVals[ctx.CurrentBar]
	sma := smaVals[ctx.CurrentBar]
	prevClose := bars[ctx.CurrentBar-1].Close

	// Long: price breaks above VWAP and SMA confirms uptrend
	if prevClose <= vwap && bar.Close > vwap && bar.Close > sma {
		if ctx.Position <= 0 {
			if ctx.Position < 0 {
				ctx.ClosePosition("VWAP breakout exit short")
			}
			qty := maxShares(ctx.Cash, bar.Close)
			if qty > 0 {
				ctx.Buy(qty, "VWAP breakout long")
			}
		}
	}

	// Exit: price drops below VWAP
	if prevClose >= vwap && bar.Close < vwap && ctx.Position > 0 {
		ctx.ClosePosition("VWAP breakdown exit")
	}
}

// ────────────────────────────────────────────────────────────────────
// 5. MACD Crossover Strategy
// ────────────────────────────────────────────────────────────────────

// MACDCrossover trades MACD line / signal line crossovers.
type MACDCrossover struct {
	FastPeriod   int
	SlowPeriod   int
	SignalPeriod int
}

// NewMACDCrossover creates a new MACD Crossover strategy.
func NewMACDCrossover(fast, slow, signal int) *MACDCrossover {
	return &MACDCrossover{FastPeriod: fast, SlowPeriod: slow, SignalPeriod: signal}
}

func (s *MACDCrossover) Name() string { return "MACD Crossover" }
func (s *MACDCrossover) Init(_ *StrategyContext) {}

func (s *MACDCrossover) OnBar(ctx *StrategyContext, bar models.OHLCV) {
	if ctx.CurrentBar < s.SlowPeriod+s.SignalPeriod+1 {
		return
	}

	bars := ctx.HistoricalBars()
	macdResults := technical.MACD(bars[:ctx.CurrentBar+1], s.FastPeriod, s.SlowPeriod, s.SignalPeriod)
	if macdResults == nil || len(macdResults) < 2 {
		return
	}

	curr := macdResults[ctx.CurrentBar]
	prev := macdResults[ctx.CurrentBar-1]

	// Bullish: MACD line crosses above signal line
	if prev.MACD <= prev.Signal && curr.MACD > curr.Signal {
		if ctx.Position <= 0 {
			if ctx.Position < 0 {
				ctx.ClosePosition("MACD bearish exit")
			}
			qty := maxShares(ctx.Cash, bar.Close)
			if qty > 0 {
				ctx.Buy(qty, "MACD bullish crossover")
			}
		}
	}

	// Bearish: MACD line crosses below signal line
	if prev.MACD >= prev.Signal && curr.MACD < curr.Signal {
		if ctx.Position > 0 {
			ctx.ClosePosition("MACD bearish crossover")
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// Helpers
// ════════════════════════════════════════════════════════════════════

// maxShares returns the maximum number of whole shares purchasable at given price.
func maxShares(cash, price float64) int {
	if price <= 0 {
		return 0
	}
	return int(cash / price)
}
