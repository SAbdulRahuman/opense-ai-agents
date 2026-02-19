package models

import "time"

// OrderSide represents buy or sell.
type OrderSide string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

// OrderType represents the type of order.
type OrderType string

const (
	Market OrderType = "MARKET"
	Limit  OrderType = "LIMIT"
	SL     OrderType = "SL"       // Stop-Loss
	SLM    OrderType = "SL-M"     // Stop-Loss Market
)

// OrderProduct represents the product type.
type OrderProduct string

const (
	CNC  OrderProduct = "CNC"  // Cash and Carry (delivery)
	MIS  OrderProduct = "MIS"  // Margin Intraday Square-off
	NRML OrderProduct = "NRML" // Normal (F&O)
)

// OrderStatus represents the current state of an order.
type OrderStatus string

const (
	OrderPending   OrderStatus = "PENDING"
	OrderOpen      OrderStatus = "OPEN"
	OrderComplete  OrderStatus = "COMPLETE"
	OrderCancelled OrderStatus = "CANCELLED"
	OrderRejected  OrderStatus = "REJECTED"
)

// OrderRequest represents a request to place a new order.
type OrderRequest struct {
	Ticker        string       `json:"ticker"`
	Exchange      string       `json:"exchange"`       // "NSE" or "NFO"
	Side          OrderSide    `json:"side"`
	OrderType     OrderType    `json:"order_type"`
	Product       OrderProduct `json:"product"`
	Quantity      int          `json:"quantity"`
	Price         float64      `json:"price,omitempty"`          // for LIMIT orders
	TriggerPrice  float64      `json:"trigger_price,omitempty"`  // for SL/SL-M orders
	StopLoss      float64      `json:"stop_loss,omitempty"`
	Target        float64      `json:"target,omitempty"`
	Tag           string       `json:"tag,omitempty"`            // custom tag for tracking
}

// OrderResponse represents the broker's response to an order placement.
type OrderResponse struct {
	OrderID  string `json:"order_id"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
}

// Order represents a placed/historical order.
type Order struct {
	OrderID       string       `json:"order_id"`
	Ticker        string       `json:"ticker"`
	Exchange      string       `json:"exchange"`
	Side          OrderSide    `json:"side"`
	OrderType     OrderType    `json:"order_type"`
	Product       OrderProduct `json:"product"`
	Quantity      int          `json:"quantity"`
	FilledQty     int          `json:"filled_qty"`
	PendingQty    int          `json:"pending_qty"`
	Price         float64      `json:"price"`
	AvgPrice      float64      `json:"avg_price"`
	TriggerPrice  float64      `json:"trigger_price,omitempty"`
	Status        OrderStatus  `json:"status"`
	StatusMessage string       `json:"status_message,omitempty"`
	PlacedAt      time.Time    `json:"placed_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Tag           string       `json:"tag,omitempty"`
}

// Position represents an open trading position.
type Position struct {
	Ticker        string       `json:"ticker"`
	Exchange      string       `json:"exchange"`
	Product       OrderProduct `json:"product"`
	Quantity      int          `json:"quantity"`      // positive = long, negative = short
	AvgPrice      float64      `json:"avg_price"`
	LTP           float64      `json:"ltp"`
	PnL           float64      `json:"pnl"`
	PnLPct        float64      `json:"pnl_pct"`
	DayPnL        float64      `json:"day_pnl"`
	Value         float64      `json:"value"`         // current market value
	Multiplier    int          `json:"multiplier"`    // lot size for F&O
}

// Holding represents a delivery holding (CNC).
type Holding struct {
	Ticker        string    `json:"ticker"`
	Exchange      string    `json:"exchange"`
	ISIN          string    `json:"isin"`
	Quantity      int       `json:"quantity"`
	AvgPrice      float64   `json:"avg_price"`
	LTP           float64   `json:"ltp"`
	PnL           float64   `json:"pnl"`
	PnLPct        float64   `json:"pnl_pct"`
	DayChange     float64   `json:"day_change"`
	DayChangePct  float64   `json:"day_change_pct"`
	CurrentValue  float64   `json:"current_value"`
	InvestedValue float64   `json:"invested_value"`
}

// Margins represents the account margin/funds information.
type Margins struct {
	AvailableCash   float64 `json:"available_cash"`
	UsedMargin      float64 `json:"used_margin"`
	AvailableMargin float64 `json:"available_margin"`
	Collateral      float64 `json:"collateral"`
	OpeningBalance  float64 `json:"opening_balance"`
}

// TradeLog represents a logged trade event for audit trail.
type TradeLog struct {
	ID            string      `json:"id"`
	Timestamp     time.Time   `json:"timestamp"`
	OrderRequest  OrderRequest `json:"order_request"`
	OrderResponse *OrderResponse `json:"order_response,omitempty"`
	Approved      bool        `json:"approved"`        // HITL approval status
	ApprovedAt    *time.Time  `json:"approved_at,omitempty"`
	Reason        string      `json:"reason,omitempty"` // reason for trade / rejection
	AgentName     string      `json:"agent_name"`       // which agent proposed the trade
}
