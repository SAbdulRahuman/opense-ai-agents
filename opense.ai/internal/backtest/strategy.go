package backtest

import (
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Strategy Interface
// ════════════════════════════════════════════════════════════════════

// Strategy defines the interface for a backtestable trading strategy.
type Strategy interface {
	// Name returns the human-readable strategy name.
	Name() string

	// Init is called once before the first bar. Use it to set up state.
	Init(ctx *StrategyContext)

	// OnBar is called for each OHLCV bar during the backtest.
	// The strategy should analyze the bar and use ctx methods to place orders.
	OnBar(ctx *StrategyContext, bar models.OHLCV)
}

// ════════════════════════════════════════════════════════════════════
// Strategy Context — The strategy's view of the world
// ════════════════════════════════════════════════════════════════════

// StrategyContext provides the strategy with market data, position state,
// and order placement methods. Passed by the engine on each bar.
type StrategyContext struct {
	// Public state — read by strategies
	Ticker      string           // ticker being tested
	Capital     float64          // initial capital
	Cash        float64          // current available cash
	Position    int              // current position (positive=long, negative=short, 0=flat)
	AvgPrice    float64          // average entry price of current position
	Bars        []models.OHLCV   // all bars in the backtest
	CurrentBar  int              // index of the current bar being processed
	CurrentOHLCV models.OHLCV    // the current bar

	// Private state — managed by engine
	orders    []pendingOrder
	trades    []models.BacktestTrade
	equity    []models.EquityPoint
	slippage  float64
	product   models.OrderProduct
	entryTime time.Time
	state     map[string]interface{} // strategy-local key/value store
}

// ════════════════════════════════════════════════════════════════════
// Order Placement Methods
// ════════════════════════════════════════════════════════════════════

// Buy places a market buy order for the given quantity.
func (ctx *StrategyContext) Buy(qty int, reason string) {
	ctx.orders = append(ctx.orders, pendingOrder{
		Side:      models.Buy,
		OrderType: models.Market,
		Quantity:  qty,
		Reason:    reason,
	})
}

// Sell places a market sell order for the given quantity.
func (ctx *StrategyContext) Sell(qty int, reason string) {
	ctx.orders = append(ctx.orders, pendingOrder{
		Side:      models.Sell,
		OrderType: models.Market,
		Quantity:  qty,
		Reason:    reason,
	})
}

// BuyLimit places a limit buy order.
func (ctx *StrategyContext) BuyLimit(qty int, price float64, reason string) {
	ctx.orders = append(ctx.orders, pendingOrder{
		Side:      models.Buy,
		OrderType: models.Limit,
		Quantity:  qty,
		Price:     price,
		Reason:    reason,
	})
}

// SellLimit places a limit sell order.
func (ctx *StrategyContext) SellLimit(qty int, price float64, reason string) {
	ctx.orders = append(ctx.orders, pendingOrder{
		Side:      models.Sell,
		OrderType: models.Limit,
		Quantity:  qty,
		Price:     price,
		Reason:    reason,
	})
}

// BuyStop places a stop-loss market buy order (triggers at triggerPrice).
func (ctx *StrategyContext) BuyStop(qty int, triggerPrice float64, reason string) {
	ctx.orders = append(ctx.orders, pendingOrder{
		Side:         models.Buy,
		OrderType:    models.SLM,
		Quantity:     qty,
		TriggerPrice: triggerPrice,
		Reason:       reason,
	})
}

// SellStop places a stop-loss market sell order (triggers at triggerPrice).
func (ctx *StrategyContext) SellStop(qty int, triggerPrice float64, reason string) {
	ctx.orders = append(ctx.orders, pendingOrder{
		Side:         models.Sell,
		OrderType:    models.SLM,
		Quantity:     qty,
		TriggerPrice: triggerPrice,
		Reason:       reason,
	})
}

// ClosePosition closes the entire current position at market.
func (ctx *StrategyContext) ClosePosition(reason string) {
	if ctx.Position > 0 {
		ctx.Sell(ctx.Position, reason)
	} else if ctx.Position < 0 {
		ctx.Buy(-ctx.Position, reason)
	}
}

// CancelPending removes all pending (unfilled) orders.
func (ctx *StrategyContext) CancelPending() {
	ctx.orders = ctx.orders[:0]
}

// ════════════════════════════════════════════════════════════════════
// Data Access Helpers
// ════════════════════════════════════════════════════════════════════

// HistoricalBars returns bars from the start up to and including the current bar.
// Useful for computing indicators on the available data.
func (ctx *StrategyContext) HistoricalBars() []models.OHLCV {
	if ctx.CurrentBar >= len(ctx.Bars) {
		return ctx.Bars
	}
	return ctx.Bars[:ctx.CurrentBar+1]
}

// Closes returns closing prices up to and including the current bar.
func (ctx *StrategyContext) Closes() []float64 {
	bars := ctx.HistoricalBars()
	closes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
	}
	return closes
}

// LookBack returns the bar that is `n` bars before the current bar.
// Returns zero-value OHLCV if n is out of range.
func (ctx *StrategyContext) LookBack(n int) models.OHLCV {
	idx := ctx.CurrentBar - n
	if idx < 0 || idx >= len(ctx.Bars) {
		return models.OHLCV{}
	}
	return ctx.Bars[idx]
}

// BarsSince returns the number of bars since the position was entered.
// Returns 0 if flat.
func (ctx *StrategyContext) BarsSince() int {
	if ctx.Position == 0 {
		return 0
	}
	for i := ctx.CurrentBar; i >= 0; i-- {
		if ctx.Bars[i].Timestamp.Equal(ctx.entryTime) || ctx.Bars[i].Timestamp.Before(ctx.entryTime) {
			return ctx.CurrentBar - i
		}
	}
	return 0
}

// PortfolioValue returns the current total portfolio value (cash + position).
func (ctx *StrategyContext) PortfolioValue() float64 {
	val := ctx.Cash
	if ctx.Position != 0 {
		val += float64(ctx.Position) * ctx.CurrentOHLCV.Close
	}
	return val
}

// PositionValue returns the current market value of the position.
func (ctx *StrategyContext) PositionValue() float64 {
	return float64(ctx.Position) * ctx.CurrentOHLCV.Close
}

// UnrealizedPnL returns the current unrealized P&L of the open position.
func (ctx *StrategyContext) UnrealizedPnL() float64 {
	if ctx.Position == 0 {
		return 0
	}
	if ctx.Position > 0 {
		return float64(ctx.Position) * (ctx.CurrentOHLCV.Close - ctx.AvgPrice)
	}
	return float64(-ctx.Position) * (ctx.AvgPrice - ctx.CurrentOHLCV.Close)
}

// ════════════════════════════════════════════════════════════════════
// Strategy-local state store
// ════════════════════════════════════════════════════════════════════

// Set stores a value in the strategy's local state.
func (ctx *StrategyContext) Set(key string, value interface{}) {
	if ctx.state == nil {
		ctx.state = make(map[string]interface{})
	}
	ctx.state[key] = value
}

// Get retrieves a value from the strategy's local state.
func (ctx *StrategyContext) Get(key string) (interface{}, bool) {
	if ctx.state == nil {
		return nil, false
	}
	v, ok := ctx.state[key]
	return v, ok
}

// GetFloat64 retrieves a float64 value from state, returns 0 if not found.
func (ctx *StrategyContext) GetFloat64(key string) float64 {
	v, ok := ctx.Get(key)
	if !ok {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

// GetInt retrieves an int value from state, returns 0 if not found.
func (ctx *StrategyContext) GetInt(key string) int {
	v, ok := ctx.Get(key)
	if !ok {
		return 0
	}
	if i, ok := v.(int); ok {
		return i
	}
	return 0
}
