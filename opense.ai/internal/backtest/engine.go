// Package backtest provides an event-driven backtesting engine for evaluating
// trading strategies against historical OHLCV data with realistic simulation
// of fills, slippage, and Indian brokerage charges.
package backtest

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/seenimoa/openseai/internal/broker"
	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Engine Configuration
// ════════════════════════════════════════════════════════════════════

// Config holds all parameters for a backtest run.
type Config struct {
	InitialCapital float64          // starting capital in INR (default: 1,000,000)
	SlippagePct    float64          // slippage per trade as fraction (default: 0.001 = 0.1%)
	Product        models.OrderProduct // CNC, MIS, NRML (default: CNC)
	Benchmark      []models.OHLCV  // optional benchmark data (e.g., Nifty 50) for comparison
	BenchmarkName  string           // benchmark name (default: "NIFTY 50")
	RiskFreeRate   float64          // annual risk-free rate for Sharpe (default: 0.065 = 6.5% India)
}

// DefaultConfig returns sensible defaults for Indian markets.
func DefaultConfig() Config {
	return Config{
		InitialCapital: 1000000,
		SlippagePct:    0.001,
		Product:        models.CNC,
		BenchmarkName:  "NIFTY 50",
		RiskFreeRate:   0.065,
	}
}

// ════════════════════════════════════════════════════════════════════
// Engine — Event-Driven Backtesting
// ════════════════════════════════════════════════════════════════════

// Engine runs a Strategy against historical data bar-by-bar.
type Engine struct {
	cfg   Config
	mu    sync.Mutex
}

// NewEngine creates a new backtesting engine with the given config.
func NewEngine(cfg Config) *Engine {
	if cfg.InitialCapital <= 0 {
		cfg.InitialCapital = DefaultConfig().InitialCapital
	}
	if cfg.SlippagePct < 0 {
		cfg.SlippagePct = 0
	}
	if cfg.Product == "" {
		cfg.Product = models.CNC
	}
	if cfg.RiskFreeRate <= 0 {
		cfg.RiskFreeRate = 0.065
	}
	return &Engine{cfg: cfg}
}

// Run executes the strategy against the provided OHLCV bars and returns
// a BacktestResult with full trade log, equity curve, and performance metrics.
func (e *Engine) Run(strategy Strategy, ticker string, bars []models.OHLCV) (*models.BacktestResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if strategy == nil {
		return nil, fmt.Errorf("strategy is nil")
	}
	if len(bars) < 2 {
		return nil, fmt.Errorf("insufficient data: need at least 2 bars, got %d", len(bars))
	}

	// Sort bars by timestamp
	sorted := make([]models.OHLCV, len(bars))
	copy(sorted, bars)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	// Initialize context
	ctx := &StrategyContext{
		Ticker:     ticker,
		Capital:    e.cfg.InitialCapital,
		Cash:       e.cfg.InitialCapital,
		Position:   0,
		AvgPrice:   0,
		Bars:       sorted,
		CurrentBar: 0,
		orders:     make([]pendingOrder, 0),
		trades:     make([]models.BacktestTrade, 0),
		equity:     make([]models.EquityPoint, 0, len(sorted)),
		slippage:   e.cfg.SlippagePct,
		product:    e.cfg.Product,
	}

	// Let strategy initialize
	strategy.Init(ctx)

	// Process bars one by one
	for i := 0; i < len(sorted); i++ {
		ctx.CurrentBar = i
		ctx.CurrentOHLCV = sorted[i]

		// Process pending orders at current bar's open
		e.processPendingOrders(ctx, sorted[i])

		// Call strategy
		strategy.OnBar(ctx, sorted[i])

		// Record equity
		equity := ctx.Cash
		if ctx.Position > 0 {
			equity += float64(ctx.Position) * sorted[i].Close
		} else if ctx.Position < 0 {
			equity += float64(ctx.Position) * sorted[i].Close // negative for shorts
		}
		ctx.equity = append(ctx.equity, models.EquityPoint{
			Date:  sorted[i].Timestamp,
			Value: equity,
		})
	}

	// Close any open position at last bar's close
	lastBar := sorted[len(sorted)-1]
	if ctx.Position != 0 {
		e.forceClose(ctx, lastBar)
	}

	// Build result
	result := e.buildResult(strategy, ticker, sorted, ctx)
	return result, nil
}

// ════════════════════════════════════════════════════════════════════
// Pending Order — Internal
// ════════════════════════════════════════════════════════════════════

type pendingOrder struct {
	Side         models.OrderSide
	OrderType    models.OrderType
	Quantity     int
	Price        float64 // limit price
	TriggerPrice float64 // SL trigger
	Reason       string
}

func (e *Engine) processPendingOrders(ctx *StrategyContext, bar models.OHLCV) {
	remaining := make([]pendingOrder, 0)

	for _, o := range ctx.orders {
		filled, fillPrice := e.tryFill(o, bar)
		if filled {
			// Apply slippage
			if o.Side == models.Buy {
				fillPrice *= (1 + ctx.slippage)
			} else {
				fillPrice *= (1 - ctx.slippage)
			}
			e.executeFill(ctx, o, fillPrice, bar.Timestamp)
		} else {
			remaining = append(remaining, o)
		}
	}

	ctx.orders = remaining
}

func (e *Engine) tryFill(o pendingOrder, bar models.OHLCV) (bool, float64) {
	switch o.OrderType {
	case models.Market:
		return true, bar.Open
	case models.Limit:
		if o.Side == models.Buy && bar.Low <= o.Price {
			return true, math.Min(bar.Open, o.Price)
		}
		if o.Side == models.Sell && bar.High >= o.Price {
			return true, math.Max(bar.Open, o.Price)
		}
	case models.SL, models.SLM:
		if o.Side == models.Buy && bar.High >= o.TriggerPrice {
			if o.OrderType == models.SLM {
				return true, o.TriggerPrice
			}
			return true, math.Max(o.TriggerPrice, o.Price)
		}
		if o.Side == models.Sell && bar.Low <= o.TriggerPrice {
			if o.OrderType == models.SLM {
				return true, o.TriggerPrice
			}
			return true, math.Min(o.TriggerPrice, o.Price)
		}
	}
	return false, 0
}

func (e *Engine) executeFill(ctx *StrategyContext, o pendingOrder, fillPrice float64, ts time.Time) {
	qty := o.Quantity
	if qty <= 0 {
		qty = 1
	}

	if o.Side == models.Buy {
		cost := fillPrice * float64(qty)
		if cost > ctx.Cash {
			return // insufficient funds
		}

		// Calculate brokerage
		charges := broker.CalculateBrokerage(fillPrice, fillPrice, qty, ctx.product)
		totalCost := cost + charges.Total

		if totalCost > ctx.Cash {
			return
		}

		if ctx.Position < 0 {
			// Closing short
			entryPrice := ctx.AvgPrice
			pnl := (entryPrice - fillPrice) * float64(qty)
			pnl -= charges.Total
			ctx.Cash += pnl + entryPrice*float64(qty) // return margin
			ctx.Position += qty
			if ctx.Position == 0 {
				ctx.AvgPrice = 0
			}
			// Record trade
			trade := models.BacktestTrade{
				EntryDate:  ctx.entryTime,
				ExitDate:   ts,
				Side:       models.Sell, // was a short
				EntryPrice: entryPrice,
				ExitPrice:  fillPrice,
				Quantity:   qty,
				PnL:        pnl,
				PnLPct:     (pnl / (entryPrice * float64(qty))) * 100,
				Reason:     o.Reason,
			}
			ctx.trades = append(ctx.trades, trade)
		} else {
			// Opening/adding to long
			totalQty := ctx.Position + qty
			if ctx.Position > 0 {
				ctx.AvgPrice = (ctx.AvgPrice*float64(ctx.Position) + fillPrice*float64(qty)) / float64(totalQty)
			} else {
				ctx.AvgPrice = fillPrice
				ctx.entryTime = ts
			}
			ctx.Position = totalQty
			ctx.Cash -= totalCost
		}
	} else {
		// SELL
		if ctx.Position > 0 {
			// Closing long
			entryPrice := ctx.AvgPrice
			revenue := fillPrice * float64(qty)
			charges := broker.CalculateBrokerage(entryPrice, fillPrice, qty, ctx.product)
			pnl := revenue - entryPrice*float64(qty) - charges.Total

			ctx.Cash += revenue - charges.Total
			ctx.Position -= qty
			if ctx.Position == 0 {
				ctx.AvgPrice = 0
			}

			trade := models.BacktestTrade{
				EntryDate:  ctx.entryTime,
				ExitDate:   ts,
				Side:       models.Buy, // was a long
				EntryPrice: entryPrice,
				ExitPrice:  fillPrice,
				Quantity:   qty,
				PnL:        pnl,
				PnLPct:     (pnl / (entryPrice * float64(qty))) * 100,
				Reason:     o.Reason,
			}
			ctx.trades = append(ctx.trades, trade)
		} else {
			// Opening short (if MIS/NRML)
			if ctx.product == models.CNC {
				return // can't short in CNC
			}
			marginReq := fillPrice * float64(qty) * 0.2 // ~20% margin for futures
			if marginReq > ctx.Cash {
				return
			}
			if ctx.Position == 0 {
				ctx.AvgPrice = fillPrice
				ctx.entryTime = ts
			}
			ctx.Position -= qty
			ctx.Cash -= marginReq
		}
	}
}

func (e *Engine) forceClose(ctx *StrategyContext, bar models.OHLCV) {
	if ctx.Position > 0 {
		o := pendingOrder{
			Side:      models.Sell,
			OrderType: models.Market,
			Quantity:  ctx.Position,
			Reason:    "backtest_end_close",
		}
		e.executeFill(ctx, o, bar.Close*(1-ctx.slippage), bar.Timestamp)
	} else if ctx.Position < 0 {
		o := pendingOrder{
			Side:      models.Buy,
			OrderType: models.Market,
			Quantity:  -ctx.Position,
			Reason:    "backtest_end_close",
		}
		e.executeFill(ctx, o, bar.Close*(1+ctx.slippage), bar.Timestamp)
	}
}

func (e *Engine) buildResult(strategy Strategy, ticker string, bars []models.OHLCV, ctx *StrategyContext) *models.BacktestResult {
	finalEquity := ctx.Cash
	if ctx.Position > 0 {
		finalEquity += float64(ctx.Position) * bars[len(bars)-1].Close
	}

	result := &models.BacktestResult{
		StrategyName:   strategy.Name(),
		Ticker:         ticker,
		From:           bars[0].Timestamp,
		To:             bars[len(bars)-1].Timestamp,
		InitialCapital: e.cfg.InitialCapital,
		FinalCapital:   finalEquity,
		TotalReturn:    finalEquity - e.cfg.InitialCapital,
		TotalReturnPct: ((finalEquity - e.cfg.InitialCapital) / e.cfg.InitialCapital) * 100,
		Trades:         ctx.trades,
		EquityCurve:    ctx.equity,
	}

	// Compute metrics
	ComputeMetrics(result, e.cfg.RiskFreeRate)

	// Benchmark return
	if len(e.cfg.Benchmark) >= 2 {
		first := e.cfg.Benchmark[0].Close
		last := e.cfg.Benchmark[len(e.cfg.Benchmark)-1].Close
		if first > 0 {
			result.BenchmarkReturn = ((last - first) / first) * 100
		}
	}

	return result
}
