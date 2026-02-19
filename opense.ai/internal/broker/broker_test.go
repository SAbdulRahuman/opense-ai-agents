package broker

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Broker Interface Tests
// ════════════════════════════════════════════════════════════════════

func TestBrokerInterfaceCompliance(t *testing.T) {
	// Verify that all broker implementations satisfy the Broker interface
	var _ Broker = (*PaperBroker)(nil)
	var _ Broker = (*ZerodhaBroker)(nil)
	var _ Broker = (*IBKRBroker)(nil)
	var _ Broker = (*RiskManager)(nil)
}

// ════════════════════════════════════════════════════════════════════
// TradeLogger Tests
// ════════════════════════════════════════════════════════════════════

func TestTradeLogger_Log(t *testing.T) {
	logger := NewTradeLogger()

	logger.Log(models.TradeLog{
		OrderRequest: models.OrderRequest{Ticker: "RELIANCE"},
		AgentName:    "test-agent",
	})

	if logger.Count() != 1 {
		t.Errorf("expected 1 log, got %d", logger.Count())
	}

	logs := logger.Logs()
	if logs[0].OrderRequest.Ticker != "RELIANCE" {
		t.Errorf("expected ticker RELIANCE, got %s", logs[0].OrderRequest.Ticker)
	}
	if logs[0].ID == "" {
		t.Error("expected non-empty ID")
	}
	if logs[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestTradeLogger_RecentLogs(t *testing.T) {
	logger := NewTradeLogger()

	for i := 0; i < 10; i++ {
		logger.Log(models.TradeLog{
			OrderRequest: models.OrderRequest{Ticker: fmt.Sprintf("STOCK%d", i)},
		})
	}

	recent := logger.RecentLogs(3)
	if len(recent) != 3 {
		t.Errorf("expected 3 recent logs, got %d", len(recent))
	}
	// Most recent should be last appended
	if recent[0].OrderRequest.Ticker != "STOCK7" {
		t.Errorf("expected STOCK7, got %s", recent[0].OrderRequest.Ticker)
	}
}

func TestTradeLogger_DayLogs(t *testing.T) {
	logger := NewTradeLogger()

	logger.Log(models.TradeLog{
		OrderRequest: models.OrderRequest{Ticker: "INFY"},
	})

	today := time.Now()
	dayLogs := logger.DayLogs(today)
	if len(dayLogs) != 1 {
		t.Errorf("expected 1 day log, got %d", len(dayLogs))
	}

	past, _ := time.Parse("2006-01-02", "2000-01-01")
	pastLogs := logger.DayLogs(past)
	if len(pastLogs) != 0 {
		t.Errorf("expected 0 past logs, got %d", len(pastLogs))
	}
}

// ════════════════════════════════════════════════════════════════════
// Brokerage Calculation Tests
// ════════════════════════════════════════════════════════════════════

func TestCalculateBrokerage_CNC(t *testing.T) {
	charges := CalculateBrokerage(100, 110, 100, models.CNC)

	if charges.Brokerage != 0 {
		t.Errorf("CNC brokerage should be 0, got %f", charges.Brokerage)
	}
	if charges.NetPnL <= 0 {
		t.Logf("net P&L for profitable CNC trade: %f", charges.NetPnL)
	}
	// Total charges should be positive
	if charges.Total <= 0 {
		t.Error("total charges should be positive")
	}
}

func TestCalculateBrokerage_MIS(t *testing.T) {
	charges := CalculateBrokerage(500, 510, 50, models.MIS)

	// MIS has capped brokerage of ₹20 or 0.03%
	if charges.Brokerage < 0 {
		t.Error("brokerage should not be negative")
	}
	if charges.Total <= 0 {
		t.Error("total charges should be positive")
	}
	grossPnL := (510 - 500) * 50.0
	if charges.NetPnL > grossPnL {
		t.Error("net P&L should be less than gross due to charges")
	}
}

func TestCalculateBrokerage_NRML(t *testing.T) {
	charges := CalculateBrokerage(200, 205, 100, models.NRML)

	if charges.Total <= 0 {
		t.Error("total charges should be positive")
	}
	grossPnL := (205 - 200) * 100.0
	if charges.NetPnL > grossPnL {
		t.Error("net P&L should be less than gross due to charges")
	}
}

// ════════════════════════════════════════════════════════════════════
// Order Validation Tests
// ════════════════════════════════════════════════════════════════════

func TestValidateOrder_ValidMarket(t *testing.T) {
	req := models.OrderRequest{
		Ticker:       "RELIANCE",
		Exchange:     "NSE",
		Side:         models.Buy,
		OrderType:    models.Market,
		Product:      models.CNC,
		Quantity:     10,
		TriggerPrice: 100,
	}
	result := ValidateOrder(req)
	if !result.IsValid() {
		t.Errorf("expected valid, got errors: %s", result.ErrorString())
	}
}

func TestValidateOrder_ValidLimit(t *testing.T) {
	req := models.OrderRequest{
		Ticker:    "INFY",
		Exchange:  "NSE",
		Side:      models.Sell,
		OrderType: models.Limit,
		Product:   models.MIS,
		Quantity:  5,
		Price:     1500.50,
	}
	result := ValidateOrder(req)
	if !result.IsValid() {
		t.Errorf("expected valid, got errors: %s", result.ErrorString())
	}
}

func TestValidateOrder_ValidSL(t *testing.T) {
	req := models.OrderRequest{
		Ticker:       "TCS",
		Exchange:     "NSE",
		Side:         models.Buy,
		OrderType:    models.SL,
		Product:      models.CNC,
		Quantity:     1,
		Price:        3500,
		TriggerPrice: 3490,
	}
	result := ValidateOrder(req)
	if !result.IsValid() {
		t.Errorf("expected valid, got errors: %s", result.ErrorString())
	}
}

func TestValidateOrder_MissingTicker(t *testing.T) {
	req := models.OrderRequest{
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Market,
		Product:   models.CNC,
		Quantity:  10,
	}
	result := ValidateOrder(req)
	if result.IsValid() {
		t.Error("expected invalid for missing ticker")
	}
}

func TestValidateOrder_InvalidExchange(t *testing.T) {
	req := models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NASDAQ",
		Side:      models.Buy,
		OrderType: models.Market,
		Product:   models.CNC,
		Quantity:  10,
	}
	result := ValidateOrder(req)
	if result.IsValid() {
		t.Error("expected invalid for NASDAQ exchange")
	}
}

func TestValidateOrder_ZeroQuantity(t *testing.T) {
	req := models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Market,
		Product:   models.CNC,
		Quantity:  0,
	}
	result := ValidateOrder(req)
	if result.IsValid() {
		t.Error("expected invalid for zero quantity")
	}
}

func TestValidateOrder_LimitWithoutPrice(t *testing.T) {
	req := models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  10,
	}
	result := ValidateOrder(req)
	if result.IsValid() {
		t.Error("expected invalid for LIMIT without price")
	}
}

func TestValidateOrder_SLWithoutTrigger(t *testing.T) {
	req := models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.SL,
		Product:   models.CNC,
		Quantity:  10,
		Price:     2500,
	}
	result := ValidateOrder(req)
	if result.IsValid() {
		t.Error("expected invalid for SL without trigger price")
	}
}

func TestValidateOrder_NRMLOnNSE(t *testing.T) {
	req := models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Market,
		Product:   models.NRML,
		Quantity:  10,
	}
	result := ValidateOrder(req)
	if result.IsValid() {
		t.Error("expected invalid for NRML on NSE (non-NFO)")
	}
}

func TestValidateOrder_NRMLOnNFO(t *testing.T) {
	req := models.OrderRequest{
		Ticker:       "NIFTY24JAN18000CE",
		Exchange:     "NFO",
		Side:         models.Buy,
		OrderType:    models.Market,
		Product:      models.NRML,
		Quantity:     50,
		TriggerPrice: 200,
	}
	result := ValidateOrder(req)
	if !result.IsValid() {
		t.Errorf("expected valid for NRML on NFO, got: %s", result.ErrorString())
	}
}

func TestValidateStopLoss_Buy(t *testing.T) {
	// Buy: stop loss must be below entry
	if err := ValidateStopLoss(models.Buy, 100, 95); err != nil {
		t.Errorf("expected valid SL below entry for buy: %v", err)
	}
	if err := ValidateStopLoss(models.Buy, 100, 105); err == nil {
		t.Error("expected error for SL above entry on buy")
	}
}

func TestValidateStopLoss_Sell(t *testing.T) {
	// Sell: stop loss must be above entry
	if err := ValidateStopLoss(models.Sell, 100, 105); err != nil {
		t.Errorf("expected valid SL above entry for sell: %v", err)
	}
	if err := ValidateStopLoss(models.Sell, 100, 95); err == nil {
		t.Error("expected error for SL below entry on sell")
	}
}

func TestValidateTarget_Buy(t *testing.T) {
	if err := ValidateTarget(models.Buy, 100, 110); err != nil {
		t.Errorf("expected valid target above entry for buy: %v", err)
	}
	if err := ValidateTarget(models.Buy, 100, 90); err == nil {
		t.Error("expected error for target below entry on buy")
	}
}

func TestValidateTarget_Sell(t *testing.T) {
	if err := ValidateTarget(models.Sell, 100, 90); err != nil {
		t.Errorf("expected valid target below entry for sell: %v", err)
	}
	if err := ValidateTarget(models.Sell, 100, 110); err == nil {
		t.Error("expected error for target above entry on sell")
	}
}

func TestValidateModifyOrder_ValidModify(t *testing.T) {
	order := &models.Order{
		Status: models.OrderPending,
		Side:   models.Buy,
	}
	req := models.OrderRequest{
		Side:  models.Buy,
		Price: 100,
	}
	if err := ValidateModifyOrder(order, req); err != nil {
		t.Errorf("expected valid modify: %v", err)
	}
}

func TestValidateModifyOrder_CompletedOrder(t *testing.T) {
	order := &models.Order{
		Status: models.OrderComplete,
		Side:   models.Buy,
	}
	req := models.OrderRequest{Side: models.Buy}
	if err := ValidateModifyOrder(order, req); err == nil {
		t.Error("expected error for completed order")
	}
}

func TestValidateModifyOrder_ChangeSide(t *testing.T) {
	order := &models.Order{
		Status: models.OrderPending,
		Side:   models.Buy,
	}
	req := models.OrderRequest{Side: models.Sell}
	if err := ValidateModifyOrder(order, req); err == nil {
		t.Error("expected error for changing side")
	}
}

// ════════════════════════════════════════════════════════════════════
// Paper Broker Tests
// ════════════════════════════════════════════════════════════════════

func TestNewPaperBroker_Defaults(t *testing.T) {
	pb := NewPaperBroker(nil)

	if pb.Name() != "paper" {
		t.Errorf("expected name 'paper', got '%s'", pb.Name())
	}

	ctx := context.Background()
	margins, err := pb.GetMargins(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if margins.AvailableCash != 1_000_000 {
		t.Errorf("expected ₹10L default capital, got %f", margins.AvailableCash)
	}
	if margins.OpeningBalance != 1_000_000 {
		t.Errorf("expected opening balance ₹10L, got %f", margins.OpeningBalance)
	}
}

func TestNewPaperBroker_CustomConfig(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 5_000_000,
		SlippagePct:    0.1,
	})

	ctx := context.Background()
	margins, _ := pb.GetMargins(ctx)
	if margins.AvailableCash != 5_000_000 {
		t.Errorf("expected ₹50L capital, got %f", margins.AvailableCash)
	}
}

func TestPaperBroker_PlaceOrder_Market(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.01, // minimal slippage for testing
	})

	ctx := context.Background()
	resp, err := pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:       "RELIANCE",
		Exchange:     "NSE",
		Side:         models.Buy,
		OrderType:    models.Market,
		Product:      models.CNC,
		Quantity:     10,
		TriggerPrice: 2500,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "COMPLETE" {
		t.Errorf("expected COMPLETE, got %s", resp.Status)
	}
	if resp.OrderID == "" {
		t.Error("expected non-empty order ID")
	}
}

func TestPaperBroker_PlaceOrder_Limit(t *testing.T) {
	pb := NewPaperBroker(nil)
	ctx := context.Background()

	resp, err := pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "INFY",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  5,
		Price:     1500,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "COMPLETE" {
		t.Errorf("expected COMPLETE, got %s", resp.Status)
	}
}

func TestPaperBroker_PlaceOrder_InsufficientMargin(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 10_000, // small capital
	})

	ctx := context.Background()
	_, err := pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  100,
		Price:     2500, // ₹2.5L needed, only ₹10K available
	})

	if err != ErrInsufficientMargin {
		t.Errorf("expected ErrInsufficientMargin, got %v", err)
	}
}

func TestPaperBroker_PlaceOrder_InvalidOrder(t *testing.T) {
	pb := NewPaperBroker(nil)
	ctx := context.Background()

	_, err := pb.PlaceOrder(ctx, models.OrderRequest{
		// Missing required fields
		Quantity: 10,
	})

	if err == nil {
		t.Error("expected error for invalid order")
	}
}

func TestPaperBroker_GetOrders(t *testing.T) {
	pb := NewPaperBroker(nil)
	ctx := context.Background()

	// Initially empty
	orders, _ := pb.GetOrders(ctx)
	if len(orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(orders))
	}

	// Place an order
	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "TCS",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  1,
		Price:     3500,
	})

	orders, _ = pb.GetOrders(ctx)
	if len(orders) != 1 {
		t.Errorf("expected 1 order, got %d", len(orders))
	}
}

func TestPaperBroker_GetOrderByID(t *testing.T) {
	pb := NewPaperBroker(nil)
	ctx := context.Background()

	resp, _ := pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "HDFC",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  1,
		Price:     1600,
	})

	order, err := pb.GetOrderByID(ctx, resp.OrderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Ticker != "HDFC" {
		t.Errorf("expected HDFC, got %s", order.Ticker)
	}

	// Non-existent order
	_, err = pb.GetOrderByID(ctx, "non-existent")
	if err != ErrOrderNotFound {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestPaperBroker_CancelOrder(t *testing.T) {
	pb := NewPaperBroker(nil)
	ctx := context.Background()

	resp, _ := pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "WIPRO",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  1,
		Price:     400,
	})

	// Can't cancel a completed order
	err := pb.CancelOrder(ctx, resp.OrderID)
	if err == nil {
		t.Error("expected error cancelling completed order")
	}

	// Non-existent order
	err = pb.CancelOrder(ctx, "fake-id")
	if err != ErrOrderNotFound {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestPaperBroker_GetPositions_MIS(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	ctx := context.Background()

	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:       "SBIN",
		Exchange:     "NSE",
		Side:         models.Buy,
		OrderType:    models.Market,
		Product:      models.MIS,
		Quantity:     100,
		TriggerPrice: 600,
	})

	positions, err := pb.GetPositions(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions))
	}
	if positions[0].Ticker != "SBIN" {
		t.Errorf("expected SBIN, got %s", positions[0].Ticker)
	}
	if positions[0].Quantity != 100 {
		t.Errorf("expected quantity 100, got %d", positions[0].Quantity)
	}
}

func TestPaperBroker_GetHoldings_CNC(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	ctx := context.Background()

	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "ITC",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  100,
		Price:     450,
	})

	holdings, err := pb.GetHoldings(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(holdings) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(holdings))
	}
	if holdings[0].Ticker != "ITC" {
		t.Errorf("expected ITC, got %s", holdings[0].Ticker)
	}
	if holdings[0].Quantity != 100 {
		t.Errorf("expected quantity 100, got %d", holdings[0].Quantity)
	}
}

func TestPaperBroker_SellHolding(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	ctx := context.Background()

	// Buy 100 shares
	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "HDFCBANK",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  100,
		Price:     1600,
	})

	// Sell all 100
	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "HDFCBANK",
		Exchange:  "NSE",
		Side:      models.Sell,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  100,
		Price:     1650,
	})

	holdings, _ := pb.GetHoldings(ctx)
	if len(holdings) != 0 {
		t.Errorf("expected 0 holdings after selling all, got %d", len(holdings))
	}
}

func TestPaperBroker_ClosePosition_MIS(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	ctx := context.Background()

	// Open long position
	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:       "TATAMOTORS",
		Exchange:     "NSE",
		Side:         models.Buy,
		OrderType:    models.Market,
		Product:      models.MIS,
		Quantity:     50,
		TriggerPrice: 700,
	})

	// Close position
	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:       "TATAMOTORS",
		Exchange:     "NSE",
		Side:         models.Sell,
		OrderType:    models.Market,
		Product:      models.MIS,
		Quantity:     50,
		TriggerPrice: 710,
	})

	positions, _ := pb.GetPositions(ctx)
	if len(positions) != 0 {
		t.Errorf("expected 0 positions after closing, got %d", len(positions))
	}
}

func TestPaperBroker_SetPrice(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	ctx := context.Background()

	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "BAJFINANCE",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  10,
		Price:     7000,
	})

	pb.SetPrice("BAJFINANCE", 7500)

	holdings, _ := pb.GetHoldings(ctx)
	if len(holdings) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(holdings))
	}
	if holdings[0].LTP != 7500 {
		t.Errorf("expected LTP 7500, got %f", holdings[0].LTP)
	}
	if holdings[0].PnL <= 0 {
		t.Error("expected positive PnL after price increase")
	}
}

func TestPaperBroker_TotalPnL(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	ctx := context.Background()

	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "MARUTI",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  1,
		Price:     10000,
	})

	pb.SetPrice("MARUTI", 10500)

	pnl := pb.TotalPnL()
	if pnl <= 0 {
		t.Error("expected positive total PnL")
	}
}

func TestPaperBroker_PositionCount(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	ctx := context.Background()

	if pb.PositionCount() != 0 {
		t.Error("expected 0 positions initially")
	}

	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:       "ACC",
		Exchange:     "NSE",
		Side:         models.Buy,
		OrderType:    models.Market,
		Product:      models.MIS,
		Quantity:     10,
		TriggerPrice: 2000,
	})

	if pb.PositionCount() != 1 {
		t.Errorf("expected 1 position, got %d", pb.PositionCount())
	}
}

func TestPaperBroker_Reset(t *testing.T) {
	pb := NewPaperBroker(nil)
	ctx := context.Background()

	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "INFY",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  10,
		Price:     1500,
	})

	pb.Reset()

	margins, _ := pb.GetMargins(ctx)
	if margins.AvailableCash != 1_000_000 {
		t.Errorf("expected reset capital, got %f", margins.AvailableCash)
	}

	orders, _ := pb.GetOrders(ctx)
	if len(orders) != 0 {
		t.Errorf("expected 0 orders after reset, got %d", len(orders))
	}
}

func TestPaperBroker_SubscribeQuotes(t *testing.T) {
	pb := NewPaperBroker(nil)
	_, err := pb.SubscribeQuotes(context.Background(), []string{"RELIANCE"})
	if err != ErrNotSupported {
		t.Errorf("expected ErrNotSupported, got %v", err)
	}
}

func TestPaperBroker_Logger(t *testing.T) {
	pb := NewPaperBroker(nil)
	ctx := context.Background()

	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "SUNPHARMA",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  1,
		Price:     1200,
	})

	if pb.Logger().Count() != 1 {
		t.Errorf("expected 1 log, got %d", pb.Logger().Count())
	}
}

func TestPaperBroker_ModifyOrder(t *testing.T) {
	pb := NewPaperBroker(nil)
	ctx := context.Background()

	resp, _ := pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "ICICIBANK",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  10,
		Price:     1000,
	})

	// Paper orders are immediately filled (COMPLETE), so modify should fail
	_, err := pb.ModifyOrder(ctx, resp.OrderID, models.OrderRequest{
		Side:  models.Buy,
		Price: 1010,
	})
	if err == nil {
		t.Error("expected error modifying completed order")
	}
}

// ════════════════════════════════════════════════════════════════════
// Zerodha Broker Tests
// ════════════════════════════════════════════════════════════════════

func TestNewZerodhaBroker(t *testing.T) {
	zb := NewZerodhaBroker(&ZerodhaConnectConfig{
		APIKey:    "test-key",
		APISecret: "test-secret",
	})

	if zb.Name() != "zerodha" {
		t.Errorf("expected name 'zerodha', got '%s'", zb.Name())
	}
	if zb.IsConnected() {
		t.Error("should not be connected initially")
	}
}

func TestZerodhaBroker_LoginURL(t *testing.T) {
	zb := NewZerodhaBroker(&ZerodhaConnectConfig{
		APIKey: "my-api-key",
	})

	url := zb.LoginURL()
	expected := "https://kite.zerodha.com/connect/login?v=3&api_key=my-api-key"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestZerodhaBroker_SetAccessToken(t *testing.T) {
	zb := NewZerodhaBroker(nil)
	zb.SetAccessToken("test-token-123")

	if !zb.IsConnected() {
		t.Error("expected connected after setting token")
	}
}

func TestZerodhaBroker_NotConnectedErrors(t *testing.T) {
	zb := NewZerodhaBroker(nil)
	ctx := context.Background()

	_, err := zb.GetMargins(ctx)
	if err != ErrNotConnected {
		t.Errorf("GetMargins: expected ErrNotConnected, got %v", err)
	}

	_, err = zb.GetPositions(ctx)
	if err != ErrNotConnected {
		t.Errorf("GetPositions: expected ErrNotConnected, got %v", err)
	}

	_, err = zb.GetHoldings(ctx)
	if err != ErrNotConnected {
		t.Errorf("GetHoldings: expected ErrNotConnected, got %v", err)
	}

	_, err = zb.GetOrders(ctx)
	if err != ErrNotConnected {
		t.Errorf("GetOrders: expected ErrNotConnected, got %v", err)
	}

	_, err = zb.GetOrderByID(ctx, "123")
	if err != ErrNotConnected {
		t.Errorf("GetOrderByID: expected ErrNotConnected, got %v", err)
	}

	_, err = zb.PlaceOrder(ctx, models.OrderRequest{})
	if err != ErrNotConnected {
		t.Errorf("PlaceOrder: expected ErrNotConnected, got %v", err)
	}

	_, err = zb.ModifyOrder(ctx, "123", models.OrderRequest{})
	if err != ErrNotConnected {
		t.Errorf("ModifyOrder: expected ErrNotConnected, got %v", err)
	}

	err = zb.CancelOrder(ctx, "123")
	if err != ErrNotConnected {
		t.Errorf("CancelOrder: expected ErrNotConnected, got %v", err)
	}
}

func TestZerodhaBroker_SubscribeQuotes(t *testing.T) {
	zb := NewZerodhaBroker(nil)
	_, err := zb.SubscribeQuotes(context.Background(), nil)
	if err != ErrNotSupported {
		t.Errorf("expected ErrNotSupported, got %v", err)
	}
}

func TestMapKiteStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected models.OrderStatus
	}{
		{"COMPLETE", models.OrderComplete},
		{"CANCELLED", models.OrderCancelled},
		{"REJECTED", models.OrderRejected},
		{"OPEN", models.OrderOpen},
		{"TRIGGER PENDING", models.OrderPending},
		{"UNKNOWN", models.OrderPending},
	}

	for _, tc := range tests {
		got := mapKiteStatus(tc.input)
		if got != tc.expected {
			t.Errorf("mapKiteStatus(%q) = %s, want %s", tc.input, got, tc.expected)
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// IBKR Broker Tests
// ════════════════════════════════════════════════════════════════════

func TestNewIBKRBroker(t *testing.T) {
	ib := NewIBKRBroker(nil)

	if ib.Name() != "ibkr" {
		t.Errorf("expected name 'ibkr', got '%s'", ib.Name())
	}
	if ib.IsConnected() {
		t.Error("should not be connected initially")
	}
}

func TestNewIBKRBroker_CustomConfig(t *testing.T) {
	ib := NewIBKRBroker(&IBKRConnectConfig{
		Host:      "192.168.1.100",
		Port:      5001,
		AccountID: "U12345",
	})

	if ib.baseURL != "https://192.168.1.100:5001/v1/api" {
		t.Errorf("unexpected base URL: %s", ib.baseURL)
	}
	if ib.accountID != "U12345" {
		t.Errorf("unexpected account ID: %s", ib.accountID)
	}
}

func TestIBKRBroker_NotConnectedErrors(t *testing.T) {
	ib := NewIBKRBroker(nil)
	ctx := context.Background()

	_, err := ib.GetMargins(ctx)
	if err != ErrNotConnected {
		t.Errorf("GetMargins: expected ErrNotConnected, got %v", err)
	}

	_, err = ib.GetPositions(ctx)
	if err != ErrNotConnected {
		t.Errorf("GetPositions: expected ErrNotConnected, got %v", err)
	}

	_, err = ib.GetHoldings(ctx)
	if err != ErrNotConnected {
		t.Errorf("GetHoldings: expected ErrNotConnected, got %v", err)
	}

	_, err = ib.GetOrders(ctx)
	if err != ErrNotConnected {
		t.Errorf("GetOrders: expected ErrNotConnected, got %v", err)
	}

	_, err = ib.PlaceOrder(ctx, models.OrderRequest{})
	if err != ErrNotConnected {
		t.Errorf("PlaceOrder: expected ErrNotConnected, got %v", err)
	}

	_, err = ib.ModifyOrder(ctx, "123", models.OrderRequest{})
	if err != ErrNotConnected {
		t.Errorf("ModifyOrder: expected ErrNotConnected, got %v", err)
	}

	err = ib.CancelOrder(ctx, "123")
	if err != ErrNotConnected {
		t.Errorf("CancelOrder: expected ErrNotConnected, got %v", err)
	}
}

func TestIBKRBroker_SubscribeQuotes(t *testing.T) {
	ib := NewIBKRBroker(nil)
	_, err := ib.SubscribeQuotes(context.Background(), nil)
	if err != ErrNotSupported {
		t.Errorf("expected ErrNotSupported, got %v", err)
	}
}

func TestMapIBKRStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected models.OrderStatus
	}{
		{"FILLED", models.OrderComplete},
		{"CANCELLED", models.OrderCancelled},
		{"INACTIVE", models.OrderRejected},
		{"SUBMITTED", models.OrderOpen},
		{"PRESUBMITTED", models.OrderOpen},
		{"UNKNOWN", models.OrderPending},
	}

	for _, tc := range tests {
		got := mapIBKRStatus(tc.input)
		if got != tc.expected {
			t.Errorf("mapIBKRStatus(%q) = %s, want %s", tc.input, got, tc.expected)
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// Risk Manager Tests
// ════════════════════════════════════════════════════════════════════

func TestNewRiskManager_Defaults(t *testing.T) {
	pb := NewPaperBroker(nil)
	rm := NewRiskManager(pb, DefaultRiskConfig())

	if rm.Name() != "risk-paper" {
		t.Errorf("expected name 'risk-paper', got '%s'", rm.Name())
	}

	cfg := rm.Config()
	if cfg.MaxPositionPct != 5.0 {
		t.Errorf("expected max position 5%%, got %f%%", cfg.MaxPositionPct)
	}
	if cfg.DailyLossLimitPct != 2.0 {
		t.Errorf("expected daily loss limit 2%%, got %f%%", cfg.DailyLossLimitPct)
	}
	if cfg.MaxOpenPositions != 10 {
		t.Errorf("expected max open positions 10, got %d", cfg.MaxOpenPositions)
	}
}

func TestRiskManager_AssessOrder_Passes(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{InitialCapital: 1_000_000})
	rm := NewRiskManager(pb, RiskConfig{
		MaxPositionPct:    5.0,
		DailyLossLimitPct: 2.0,
		MaxOpenPositions:  10,
		MaxOrderValuePct:  10.0,
		InitialCapital:    1_000_000,
	})

	ctx := context.Background()
	report, err := rm.Assess(ctx, models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  10,
		Price:     2500, // ₹25,000 = 2.5% of ₹10L — within limits
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.Passed {
		t.Errorf("expected risk check to pass, violations: %v", report.Violations)
	}
	if report.OrderValuePct != 2.5 {
		t.Errorf("expected order value 2.5%%, got %f%%", report.OrderValuePct)
	}
}

func TestRiskManager_AssessOrder_ExceedsPositionSize(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{InitialCapital: 1_000_000})
	rm := NewRiskManager(pb, RiskConfig{
		MaxPositionPct:    5.0,
		MaxOrderValuePct:  10.0,
		DailyLossLimitPct: 2.0,
		MaxOpenPositions:  10,
		InitialCapital:    1_000_000,
	})

	ctx := context.Background()
	report, _ := rm.Assess(ctx, models.OrderRequest{
		Ticker:    "RELIANCE",
		Price:     2500,
		Quantity:  30, // ₹75,000 = 7.5% — exceeds 5%
	})

	if report.Passed {
		t.Error("expected risk check to fail for oversized position")
	}
	if len(report.Violations) == 0 {
		t.Error("expected at least one violation")
	}
}

func TestRiskManager_AssessOrder_ExceedsMaxPositions(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 10_000_000,
		SlippagePct:    0.001,
	})
	rm := NewRiskManager(pb, RiskConfig{
		MaxPositionPct:    50.0,
		MaxOrderValuePct:  50.0,
		DailyLossLimitPct: 50.0,
		MaxOpenPositions:  3, // low limit for testing
		InitialCapital:    10_000_000,
	})

	ctx := context.Background()
	tickers := []string{"STOCK1", "STOCK2", "STOCK3"}
	for _, ticker := range tickers {
		pb.PlaceOrder(ctx, models.OrderRequest{
			Ticker:       ticker,
			Exchange:     "NSE",
			Side:         models.Buy,
			OrderType:    models.Market,
			Product:      models.MIS,
			Quantity:     10,
			TriggerPrice: 100,
		})
	}

	// 4th order should be blocked
	report, _ := rm.Assess(ctx, models.OrderRequest{
		Ticker:   "STOCK4",
		Price:    100,
		Quantity: 10,
	})

	if report.Passed {
		t.Error("expected risk check to fail — max positions reached")
	}
}

func TestRiskManager_PlaceOrder_PassesRisk(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	rm := NewRiskManager(pb, RiskConfig{
		MaxPositionPct:    10.0,
		MaxOrderValuePct:  20.0,
		DailyLossLimitPct: 5.0,
		MaxOpenPositions:  10,
		InitialCapital:    1_000_000,
	})

	ctx := context.Background()
	resp, err := rm.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  10,
		Price:     2500,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "COMPLETE" {
		t.Errorf("expected COMPLETE, got %s", resp.Status)
	}
	if rm.TradeCount() != 1 {
		t.Errorf("expected trade count 1, got %d", rm.TradeCount())
	}
}

func TestRiskManager_PlaceOrder_BlockedByRisk(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{InitialCapital: 100_000})
	rm := NewRiskManager(pb, RiskConfig{
		MaxPositionPct:    1.0, // very restrictive
		MaxOrderValuePct:  1.0,
		DailyLossLimitPct: 0.5,
		MaxOpenPositions:  10,
		InitialCapital:    100_000,
	})

	ctx := context.Background()
	_, err := rm.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  10,
		Price:     2500, // ₹25K = 25% of ₹1L
	})

	if err != ErrTradeBlocked {
		t.Errorf("expected ErrTradeBlocked, got %v", err)
	}
}

func TestRiskManager_Delegated_Methods(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	rm := NewRiskManager(pb, DefaultRiskConfig())
	ctx := context.Background()

	// Margins
	margins, err := rm.GetMargins(ctx)
	if err != nil {
		t.Fatalf("GetMargins: %v", err)
	}
	if margins.AvailableCash != 1_000_000 {
		t.Errorf("unexpected margin: %f", margins.AvailableCash)
	}

	// Positions
	positions, err := rm.GetPositions(ctx)
	if err != nil {
		t.Fatalf("GetPositions: %v", err)
	}
	if len(positions) != 0 {
		t.Errorf("expected 0 positions, got %d", len(positions))
	}

	// Holdings
	holdings, err := rm.GetHoldings(ctx)
	if err != nil {
		t.Fatalf("GetHoldings: %v", err)
	}
	if len(holdings) != 0 {
		t.Errorf("expected 0 holdings, got %d", len(holdings))
	}

	// Orders
	orders, err := rm.GetOrders(ctx)
	if err != nil {
		t.Fatalf("GetOrders: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(orders))
	}
}

func TestRiskManager_UpdateConfig(t *testing.T) {
	pb := NewPaperBroker(nil)
	rm := NewRiskManager(pb, DefaultRiskConfig())

	newCfg := RiskConfig{
		MaxPositionPct:    10.0,
		DailyLossLimitPct: 3.0,
		MaxOpenPositions:  20,
		MaxOrderValuePct:  15.0,
		ApprovalTimeout:   30 * time.Second,
		InitialCapital:    2_000_000,
	}
	rm.UpdateConfig(newCfg)

	cfg := rm.Config()
	if cfg.MaxPositionPct != 10.0 {
		t.Errorf("expected 10%%, got %f%%", cfg.MaxPositionPct)
	}
	if cfg.InitialCapital != 2_000_000 {
		t.Errorf("expected ₹20L, got %f", cfg.InitialCapital)
	}
}

func TestRiskManager_Logger(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	rm := NewRiskManager(pb, RiskConfig{
		MaxPositionPct:    10.0,
		MaxOrderValuePct:  20.0,
		DailyLossLimitPct: 5.0,
		MaxOpenPositions:  10,
		InitialCapital:    1_000_000,
	})

	ctx := context.Background()
	rm.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "WIPRO",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  5,
		Price:     400,
	})

	if rm.Logger().Count() != 1 {
		t.Errorf("expected 1 log, got %d", rm.Logger().Count())
	}
}

func TestRiskManager_Approval_Denied(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	rm := NewRiskManager(pb, RiskConfig{
		MaxPositionPct:    10.0,
		MaxOrderValuePct:  20.0,
		DailyLossLimitPct: 5.0,
		MaxOpenPositions:  10,
		RequireApproval:   true,
		ApprovalTimeout:   1 * time.Second, // short timeout
		InitialCapital:    1_000_000,
	})

	// No one will respond to the approval — should timeout
	ctx := context.Background()
	_, err := rm.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "TCS",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  5,
		Price:     3500,
	})

	if err == nil || (!errors.Is(err, ErrApprovalDenied) && !errors.Is(err, ErrApprovalTimeout)) {
		t.Errorf("expected ErrApprovalDenied or ErrApprovalTimeout, got %v", err)
	}
}

func TestRiskManager_Approval_Approved(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	rm := NewRiskManager(pb, RiskConfig{
		MaxPositionPct:    10.0,
		MaxOrderValuePct:  20.0,
		DailyLossLimitPct: 5.0,
		MaxOpenPositions:  10,
		RequireApproval:   true,
		ApprovalTimeout:   5 * time.Second,
		InitialCapital:    1_000_000,
	})

	// Launch approval handler in background
	go func() {
		req := <-rm.ApprovalChannel()
		req.ResultCh <- ApprovalResult{Approved: true, Reason: "looks good"}
	}()

	ctx := context.Background()
	resp, err := rm.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "HCLTECH",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  5,
		Price:     1500,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "COMPLETE" {
		t.Errorf("expected COMPLETE, got %s", resp.Status)
	}
}

func TestRiskManager_DayPnL(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	rm := NewRiskManager(pb, DefaultRiskConfig())

	if rm.DayPnL() != 0 {
		t.Errorf("expected 0 day PnL initially, got %f", rm.DayPnL())
	}
}

// ════════════════════════════════════════════════════════════════════
// Edge Case & Integration Tests
// ════════════════════════════════════════════════════════════════════

func TestPaperBroker_MultipleOrders(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 10_000_000,
		SlippagePct:    0.001,
	})
	ctx := context.Background()

	tickers := []string{"RELIANCE", "TCS", "INFY", "HDFC", "ICICIBANK"}
	for _, ticker := range tickers {
		resp, err := pb.PlaceOrder(ctx, models.OrderRequest{
			Ticker:    ticker,
			Exchange:  "NSE",
			Side:      models.Buy,
			OrderType: models.Limit,
			Product:   models.CNC,
			Quantity:  10,
			Price:     1000,
		})
		if err != nil {
			t.Fatalf("failed to place order for %s: %v", ticker, err)
		}
		if resp.Status != "COMPLETE" {
			t.Errorf("expected COMPLETE for %s, got %s", ticker, resp.Status)
		}
	}

	holdings, _ := pb.GetHoldings(ctx)
	if len(holdings) != 5 {
		t.Errorf("expected 5 holdings, got %d", len(holdings))
	}

	orders, _ := pb.GetOrders(ctx)
	if len(orders) != 5 {
		t.Errorf("expected 5 orders, got %d", len(orders))
	}
}

func TestPaperBroker_PartialSell(t *testing.T) {
	pb := NewPaperBroker(&PaperBrokerConfig{
		InitialCapital: 1_000_000,
		SlippagePct:    0.001,
	})
	ctx := context.Background()

	// Buy 100 shares
	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NSE",
		Side:      models.Buy,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  100,
		Price:     2500,
	})

	// Sell 40 shares
	pb.PlaceOrder(ctx, models.OrderRequest{
		Ticker:    "RELIANCE",
		Exchange:  "NSE",
		Side:      models.Sell,
		OrderType: models.Limit,
		Product:   models.CNC,
		Quantity:  40,
		Price:     2550,
	})

	holdings, _ := pb.GetHoldings(ctx)
	if len(holdings) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(holdings))
	}
	if holdings[0].Quantity != 60 {
		t.Errorf("expected 60 remaining shares, got %d", holdings[0].Quantity)
	}
}

func TestValidateOrder_MultipleErrors(t *testing.T) {
	req := models.OrderRequest{
		// Everything is wrong
		Ticker:    "",
		Exchange:  "INVALID",
		Side:      "MAYBE",
		OrderType: "WEIRD",
		Product:   "UNKNOWN",
		Quantity:  -5,
	}
	result := ValidateOrder(req)
	if result.IsValid() {
		t.Error("expected invalid")
	}
	if len(result.Errors) < 3 {
		t.Errorf("expected multiple errors, got %d", len(result.Errors))
	}
}

func TestAbsInt(t *testing.T) {
	if absInt(-5) != 5 {
		t.Error("absInt(-5) should be 5")
	}
	if absInt(5) != 5 {
		t.Error("absInt(5) should be 5")
	}
	if absInt(0) != 0 {
		t.Error("absInt(0) should be 0")
	}
}

func TestAbsFloat(t *testing.T) {
	if absFloat(-3.14) != 3.14 {
		t.Error("absFloat(-3.14) should be 3.14")
	}
	if absFloat(3.14) != 3.14 {
		t.Error("absFloat(3.14) should be 3.14")
	}
}

func TestBrokerageCharges_StructFields(t *testing.T) {
	charges := CalculateBrokerage(100, 110, 100, models.MIS)

	if charges.STT < 0 {
		t.Error("STT should not be negative")
	}
	if charges.ExchangeTxn < 0 {
		t.Error("exchange txn charge should not be negative")
	}
	if charges.GST < 0 {
		t.Error("GST should not be negative")
	}
	if charges.SEBICharges < 0 {
		t.Error("SEBI charge should not be negative")
	}
	if charges.StampDuty < 0 {
		t.Error("stamp duty should not be negative")
	}
	componentSum := charges.Brokerage + charges.STT + charges.ExchangeTxn + charges.GST + charges.SEBICharges + charges.StampDuty
	if math.Abs(charges.Total-componentSum) > 0.01 {
		t.Error("total charges should equal sum of components")
	}
	grossPnL := (110 - 100) * 100.0
	if math.Abs(charges.NetPnL-(grossPnL-charges.Total)) > 0.01 {
		t.Error("net P&L should equal gross - total charges")
	}
}
