// Package broker provides a unified interface for broker integrations.
// It supports paper trading (default), Zerodha Kite, and Interactive Brokers.
// All live trading requires human-in-the-loop confirmation.
package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Broker Interface
// ════════════════════════════════════════════════════════════════════

// Broker defines the common interface that all broker implementations must satisfy.
// Methods map closely to standard Indian broker APIs (Zerodha Kite, IBKR, etc.).
type Broker interface {
	// Name returns the broker provider name ("paper", "zerodha", "ibkr").
	Name() string

	// --- Account ---

	// GetMargins returns the account's margin/fund information.
	GetMargins(ctx context.Context) (*models.Margins, error)

	// --- Positions & Holdings ---

	// GetPositions returns all open trading positions (intraday + F&O).
	GetPositions(ctx context.Context) ([]models.Position, error)

	// GetHoldings returns delivery holdings (CNC).
	GetHoldings(ctx context.Context) ([]models.Holding, error)

	// --- Orders ---

	// GetOrders returns all orders for the current day.
	GetOrders(ctx context.Context) ([]models.Order, error)

	// GetOrderByID returns a specific order by its ID.
	GetOrderByID(ctx context.Context, orderID string) (*models.Order, error)

	// PlaceOrder submits a new order to the exchange.
	PlaceOrder(ctx context.Context, req models.OrderRequest) (*models.OrderResponse, error)

	// ModifyOrder modifies an open/pending order.
	ModifyOrder(ctx context.Context, orderID string, req models.OrderRequest) (*models.OrderResponse, error)

	// CancelOrder cancels an open/pending order.
	CancelOrder(ctx context.Context, orderID string) error

	// --- Streaming ---

	// SubscribeQuotes subscribes to live tick data for the given tickers.
	// Returns a channel that receives quote updates.
	SubscribeQuotes(ctx context.Context, tickers []string) (<-chan models.Quote, error)
}

// ════════════════════════════════════════════════════════════════════
// Trade Logger
// ════════════════════════════════════════════════════════════════════

// TradeLogger logs all trade events for audit trail.
type TradeLogger struct {
	mu   sync.Mutex
	logs []models.TradeLog
}

// NewTradeLogger creates a new trade logger.
func NewTradeLogger() *TradeLogger {
	return &TradeLogger{
		logs: make([]models.TradeLog, 0, 100),
	}
}

// Log records a trade event.
func (tl *TradeLogger) Log(log models.TradeLog) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	if log.ID == "" {
		log.ID = fmt.Sprintf("TL-%d", len(tl.logs)+1)
	}
	tl.logs = append(tl.logs, log)
}

// Logs returns all logged trade events.
func (tl *TradeLogger) Logs() []models.TradeLog {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	out := make([]models.TradeLog, len(tl.logs))
	copy(out, tl.logs)
	return out
}

// RecentLogs returns the last n trade events.
func (tl *TradeLogger) RecentLogs(n int) []models.TradeLog {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	if n >= len(tl.logs) {
		out := make([]models.TradeLog, len(tl.logs))
		copy(out, tl.logs)
		return out
	}
	out := make([]models.TradeLog, n)
	copy(out, tl.logs[len(tl.logs)-n:])
	return out
}

// Count returns the total number of logged trades.
func (tl *TradeLogger) Count() int {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	return len(tl.logs)
}

// DayLogs returns trade logs for a specific date.
func (tl *TradeLogger) DayLogs(date time.Time) []models.TradeLog {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	y, m, d := date.Date()
	var out []models.TradeLog
	for _, log := range tl.logs {
		ly, lm, ld := log.Timestamp.Date()
		if ly == y && lm == m && ld == d {
			out = append(out, log)
		}
	}
	return out
}

// ════════════════════════════════════════════════════════════════════
// Common Errors
// ════════════════════════════════════════════════════════════════════

var (
	// ErrNotConnected is returned when the broker connection is not established.
	ErrNotConnected = fmt.Errorf("broker not connected")

	// ErrInsufficientMargin is returned when there isn't enough margin/funds.
	ErrInsufficientMargin = fmt.Errorf("insufficient margin")

	// ErrOrderNotFound is returned when an order ID doesn't exist.
	ErrOrderNotFound = fmt.Errorf("order not found")

	// ErrOrderCantModify is returned when an order can't be modified (already executed/cancelled).
	ErrOrderCantModify = fmt.Errorf("order cannot be modified")

	// ErrOrderRejected is returned when the exchange rejects an order.
	ErrOrderRejected = fmt.Errorf("order rejected")

	// ErrTradeBlocked is returned when risk checks block a trade.
	ErrTradeBlocked = fmt.Errorf("trade blocked by risk manager")

	// ErrApprovalDenied is returned when a human denies trade approval.
	ErrApprovalDenied = fmt.Errorf("trade approval denied")

	// ErrApprovalTimeout is returned when human approval times out.
	ErrApprovalTimeout = fmt.Errorf("trade approval timed out")

	// ErrNotSupported is returned for unimplemented broker features.
	ErrNotSupported = fmt.Errorf("operation not supported by this broker")
)

// ════════════════════════════════════════════════════════════════════
// Brokerage Calculator
// ════════════════════════════════════════════════════════════════════

// BrokerageCharges represents the breakdown of Indian brokerage charges.
type BrokerageCharges struct {
	Brokerage     float64 `json:"brokerage"`
	STT           float64 `json:"stt"`
	ExchangeTxn   float64 `json:"exchange_txn"`
	SEBICharges   float64 `json:"sebi_charges"`
	StampDuty     float64 `json:"stamp_duty"`
	GST           float64 `json:"gst"`
	Total         float64 `json:"total"`
	NetPnL        float64 `json:"net_pnl,omitempty"` // PnL after charges
}

// CalculateBrokerage computes Indian brokerage charges for a trade.
// Rates are based on standard discount broker (Zerodha-like) fee structure.
func CalculateBrokerage(buyPrice, sellPrice float64, qty int, product models.OrderProduct) BrokerageCharges {
	buyValue := buyPrice * float64(qty)
	sellValue := sellPrice * float64(qty)
	turnover := buyValue + sellValue

	var charges BrokerageCharges

	switch product {
	case models.CNC: // Delivery
		charges.Brokerage = 0 // zero brokerage for delivery
		charges.STT = turnover * 0.001 // 0.1% on buy + sell
		charges.StampDuty = buyValue * 0.00015 // 0.015% on buy side

	case models.MIS: // Intraday
		// ₹20 per order or 0.03%, whichever is lower
		buyBrok := min(buyValue*0.0003, 20.0)
		sellBrok := min(sellValue*0.0003, 20.0)
		charges.Brokerage = buyBrok + sellBrok
		charges.STT = sellValue * 0.00025 // 0.025% on sell side only
		charges.StampDuty = buyValue * 0.00003 // 0.003% on buy side

	case models.NRML: // F&O
		buyBrok := min(buyValue*0.0003, 20.0)
		sellBrok := min(sellValue*0.0003, 20.0)
		charges.Brokerage = buyBrok + sellBrok
		charges.STT = sellValue * 0.000625 // 0.0625% on sell (futures)
		charges.StampDuty = buyValue * 0.00003

	default: // default to delivery
		charges.STT = turnover * 0.001
		charges.StampDuty = buyValue * 0.00015
	}

	charges.ExchangeTxn = turnover * 0.0000345 // NSE transaction charges
	charges.SEBICharges = turnover * 0.000001   // ₹10 per crore
	charges.GST = (charges.Brokerage + charges.ExchangeTxn + charges.SEBICharges) * 0.18

	charges.Total = charges.Brokerage + charges.STT + charges.ExchangeTxn +
		charges.SEBICharges + charges.StampDuty + charges.GST

	grossPnL := (sellPrice - buyPrice) * float64(qty)
	charges.NetPnL = grossPnL - charges.Total

	return charges
}

// min returns the smaller of two float64 values.
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
