package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Pre-Trade Risk Manager
// ════════════════════════════════════════════════════════════════════

// RiskManager wraps a Broker with pre-trade risk checks and guardrails.
// All orders pass through the risk manager before reaching the underlying
// broker. It enforces: max position sizing, daily loss limits, max open
// positions, and human-in-the-loop (HITL) confirmations for live trading.
type RiskManager struct {
	mu sync.RWMutex

	broker Broker
	config RiskConfig

	// Day-level tracking
	dayPnL      float64
	dayDate     string // "2006-01-02" format
	tradeCount  int

	// HITL approval channel
	approvalCh chan ApprovalRequest

	logger *TradeLogger
}

// RiskConfig holds risk management parameters.
type RiskConfig struct {
	MaxPositionPct    float64 // max single position as % of capital (default: 5.0)
	DailyLossLimitPct float64 // max daily loss as % of capital (default: 2.0)
	MaxOpenPositions  int     // max number of concurrent positions (default: 10)
	MaxOrderValuePct  float64 // max single order value as % of capital (default: 10.0)
	RequireApproval   bool    // require HITL approval for live orders
	ApprovalTimeout   time.Duration // timeout for HITL approval (default: 60s)
	InitialCapital    float64 // capital base for % calculations
}

// ApprovalRequest represents a request for human approval before trade execution.
type ApprovalRequest struct {
	OrderRequest models.OrderRequest
	RiskReport   RiskReport
	ResultCh     chan ApprovalResult
}

// ApprovalResult represents the human's decision.
type ApprovalResult struct {
	Approved bool
	Reason   string
}

// RiskReport contains the pre-trade risk assessment results.
type RiskReport struct {
	Passed        bool     `json:"passed"`
	Warnings      []string `json:"warnings,omitempty"`
	Violations    []string `json:"violations,omitempty"`
	OrderValuePct float64  `json:"order_value_pct"` // order value as % of capital
	PositionCount int      `json:"position_count"`
	DayPnL        float64  `json:"day_pnl"`
	DayPnLPct     float64  `json:"day_pnl_pct"`
}

// DefaultRiskConfig returns sensible default risk parameters.
func DefaultRiskConfig() RiskConfig {
	return RiskConfig{
		MaxPositionPct:    5.0,
		DailyLossLimitPct: 2.0,
		MaxOpenPositions:  10,
		MaxOrderValuePct:  10.0,
		RequireApproval:   false,
		ApprovalTimeout:   60 * time.Second,
		InitialCapital:    1_000_000,
	}
}

// NewRiskManager wraps a broker with risk guardrails.
func NewRiskManager(broker Broker, cfg RiskConfig) *RiskManager {
	if cfg.MaxPositionPct <= 0 {
		cfg.MaxPositionPct = 5.0
	}
	if cfg.DailyLossLimitPct <= 0 {
		cfg.DailyLossLimitPct = 2.0
	}
	if cfg.MaxOpenPositions <= 0 {
		cfg.MaxOpenPositions = 10
	}
	if cfg.MaxOrderValuePct <= 0 {
		cfg.MaxOrderValuePct = 10.0
	}
	if cfg.ApprovalTimeout <= 0 {
		cfg.ApprovalTimeout = 60 * time.Second
	}
	if cfg.InitialCapital <= 0 {
		cfg.InitialCapital = 1_000_000
	}

	return &RiskManager{
		broker:     broker,
		config:     cfg,
		approvalCh: make(chan ApprovalRequest, 10),
		logger:     NewTradeLogger(),
	}
}

// Name returns the underlying broker name with a risk prefix.
func (rm *RiskManager) Name() string {
	return fmt.Sprintf("risk-%s", rm.broker.Name())
}

// ════════════════════════════════════════════════════════════════════
// Delegated Methods (pass-through to underlying broker)
// ════════════════════════════════════════════════════════════════════

func (rm *RiskManager) GetMargins(ctx context.Context) (*models.Margins, error) {
	return rm.broker.GetMargins(ctx)
}

func (rm *RiskManager) GetPositions(ctx context.Context) ([]models.Position, error) {
	return rm.broker.GetPositions(ctx)
}

func (rm *RiskManager) GetHoldings(ctx context.Context) ([]models.Holding, error) {
	return rm.broker.GetHoldings(ctx)
}

func (rm *RiskManager) GetOrders(ctx context.Context) ([]models.Order, error) {
	return rm.broker.GetOrders(ctx)
}

func (rm *RiskManager) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	return rm.broker.GetOrderByID(ctx, orderID)
}

func (rm *RiskManager) SubscribeQuotes(ctx context.Context, tickers []string) (<-chan models.Quote, error) {
	return rm.broker.SubscribeQuotes(ctx, tickers)
}

// ════════════════════════════════════════════════════════════════════
// Risk-Gated Methods
// ════════════════════════════════════════════════════════════════════

// PlaceOrder runs pre-trade risk checks before delegating to the underlying broker.
func (rm *RiskManager) PlaceOrder(ctx context.Context, req models.OrderRequest) (*models.OrderResponse, error) {
	// Run risk assessment
	report, err := rm.Assess(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("risk assessment failed: %w", err)
	}

	if !report.Passed {
		rm.logger.Log(models.TradeLog{
			OrderRequest: req,
			Approved:     false,
			AgentName:    rm.Name(),
			Reason:       fmt.Sprintf("risk check failed: %v", report.Violations),
		})
		return &models.OrderResponse{
			Status:  "REJECTED",
			Message: fmt.Sprintf("risk check failed: %v", report.Violations),
		}, ErrTradeBlocked
	}

	// HITL approval if required
	if rm.config.RequireApproval {
		approved, reason, err := rm.requestApproval(ctx, req, *report)
		if err != nil {
			return nil, fmt.Errorf("approval request: %w", err)
		}
		if !approved {
			rm.logger.Log(models.TradeLog{
				OrderRequest: req,
				Approved:     false,
				AgentName:    rm.Name(),
				Reason:       fmt.Sprintf("approval denied: %s", reason),
			})
			return &models.OrderResponse{
				Status:  "REJECTED",
				Message: fmt.Sprintf("human approval denied: %s", reason),
			}, ErrApprovalDenied
		}
	}

	// Delegate to underlying broker
	resp, err := rm.broker.PlaceOrder(ctx, req)

	// Log the trade
	now := time.Now()
	rm.logger.Log(models.TradeLog{
		OrderRequest:  req,
		OrderResponse: resp,
		Approved:      true,
		ApprovedAt:    &now,
		AgentName:     rm.Name(),
	})

	// Update day tracking
	rm.mu.Lock()
	rm.tradeCount++
	rm.mu.Unlock()

	return resp, err
}

// ModifyOrder wraps the modify with basic validation.
func (rm *RiskManager) ModifyOrder(ctx context.Context, orderID string, req models.OrderRequest) (*models.OrderResponse, error) {
	return rm.broker.ModifyOrder(ctx, orderID, req)
}

// CancelOrder delegates to the underlying broker.
func (rm *RiskManager) CancelOrder(ctx context.Context, orderID string) error {
	return rm.broker.CancelOrder(ctx, orderID)
}

// ════════════════════════════════════════════════════════════════════
// Risk Assessment Engine
// ════════════════════════════════════════════════════════════════════

// Assess runs all pre-trade risk checks and returns a risk report.
func (rm *RiskManager) Assess(ctx context.Context, req models.OrderRequest) (*RiskReport, error) {
	report := &RiskReport{
		Passed: true,
	}

	capital := rm.config.InitialCapital

	// ── Check 1: Order value vs max position size ──
	orderPrice := req.Price
	if orderPrice <= 0 {
		orderPrice = req.TriggerPrice
	}
	if orderPrice <= 0 {
		orderPrice = 100 // fallback for market orders without price hint
	}

	orderValue := orderPrice * float64(req.Quantity)
	orderValuePct := (orderValue / capital) * 100
	report.OrderValuePct = orderValuePct

	if orderValuePct > rm.config.MaxPositionPct {
		report.Passed = false
		report.Violations = append(report.Violations,
			fmt.Sprintf("order size %.1f%% exceeds max position size %.1f%% of capital",
				orderValuePct, rm.config.MaxPositionPct))
	} else if orderValuePct > rm.config.MaxPositionPct*0.8 {
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("order size %.1f%% approaching position limit %.1f%%",
				orderValuePct, rm.config.MaxPositionPct))
	}

	// ── Check 2: Max order value ──
	if orderValuePct > rm.config.MaxOrderValuePct {
		report.Passed = false
		report.Violations = append(report.Violations,
			fmt.Sprintf("order value %.1f%% exceeds max order value %.1f%% of capital",
				orderValuePct, rm.config.MaxOrderValuePct))
	}

	// ── Check 3: Max open positions ──
	positions, err := rm.broker.GetPositions(ctx)
	if err == nil {
		report.PositionCount = len(positions)
		if len(positions) >= rm.config.MaxOpenPositions {
			report.Passed = false
			report.Violations = append(report.Violations,
				fmt.Sprintf("position count %d has reached max %d",
					len(positions), rm.config.MaxOpenPositions))
		} else if len(positions) >= rm.config.MaxOpenPositions-2 {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("approaching position limit: %d/%d",
					len(positions), rm.config.MaxOpenPositions))
		}
	}

	// ── Check 4: Daily loss limit ──
	rm.refreshDayPnL(ctx)

	rm.mu.RLock()
	dayPnL := rm.dayPnL
	rm.mu.RUnlock()

	dayPnLPct := (dayPnL / capital) * 100
	report.DayPnL = dayPnL
	report.DayPnLPct = dayPnLPct

	if dayPnLPct < -rm.config.DailyLossLimitPct {
		report.Passed = false
		report.Violations = append(report.Violations,
			fmt.Sprintf("daily loss %.2f%% exceeds limit %.1f%%",
				dayPnLPct, rm.config.DailyLossLimitPct))
	} else if dayPnLPct < -rm.config.DailyLossLimitPct*0.8 {
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("approaching daily loss limit: %.2f%% (limit: %.1f%%)",
				dayPnLPct, rm.config.DailyLossLimitPct))
	}

	// ── Check 5: Margin availability ──
	margins, err := rm.broker.GetMargins(ctx)
	if err == nil {
		if orderValue > margins.AvailableMargin {
			report.Passed = false
			report.Violations = append(report.Violations,
				fmt.Sprintf("order value ₹%.0f exceeds available margin ₹%.0f",
					orderValue, margins.AvailableMargin))
		}
	}

	return report, nil
}

// refreshDayPnL recalculates the day's P&L from positions.
func (rm *RiskManager) refreshDayPnL(ctx context.Context) {
	today := time.Now().Format("2006-01-02")

	rm.mu.Lock()
	if rm.dayDate != today {
		rm.dayPnL = 0
		rm.dayDate = today
		rm.tradeCount = 0
	}
	rm.mu.Unlock()

	positions, err := rm.broker.GetPositions(ctx)
	if err != nil {
		return
	}

	var totalPnL float64
	for _, p := range positions {
		totalPnL += p.PnL
	}

	rm.mu.Lock()
	rm.dayPnL = totalPnL
	rm.mu.Unlock()
}

// ════════════════════════════════════════════════════════════════════
// HITL Approval
// ════════════════════════════════════════════════════════════════════

// ApprovalChannel returns the channel for receiving approval requests.
// External systems (CLI, web UI) should listen on this channel and
// send results back via the embedded ResultCh.
func (rm *RiskManager) ApprovalChannel() <-chan ApprovalRequest {
	return rm.approvalCh
}

// requestApproval sends an approval request and waits for the response.
func (rm *RiskManager) requestApproval(ctx context.Context, req models.OrderRequest, report RiskReport) (bool, string, error) {
	resultCh := make(chan ApprovalResult, 1)

	approvalReq := ApprovalRequest{
		OrderRequest: req,
		RiskReport:   report,
		ResultCh:     resultCh,
	}

	// Send to approval channel (non-blocking)
	select {
	case rm.approvalCh <- approvalReq:
	default:
		return false, "approval queue full", ErrApprovalTimeout
	}

	// Wait for response with timeout
	timeout := time.After(rm.config.ApprovalTimeout)
	select {
	case result := <-resultCh:
		return result.Approved, result.Reason, nil
	case <-timeout:
		return false, "approval timed out", ErrApprovalTimeout
	case <-ctx.Done():
		return false, "context cancelled", ctx.Err()
	}
}

// ════════════════════════════════════════════════════════════════════
// Accessors
// ════════════════════════════════════════════════════════════════════

// Logger returns the risk manager's trade logger.
func (rm *RiskManager) Logger() *TradeLogger {
	return rm.logger
}

// Config returns the current risk configuration.
func (rm *RiskManager) Config() RiskConfig {
	return rm.config
}

// UpdateConfig updates risk parameters at runtime.
func (rm *RiskManager) UpdateConfig(cfg RiskConfig) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.config = cfg
}

// TradeCount returns the number of trades executed today.
func (rm *RiskManager) TradeCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.tradeCount
}

// DayPnL returns the current day's P&L.
func (rm *RiskManager) DayPnL() float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.dayPnL
}
