package broker

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Paper Trading Simulator
// ════════════════════════════════════════════════════════════════════

// PaperBroker is an in-memory paper trading simulator that implements the
// Broker interface. It simulates order fills with configurable slippage,
// tracks positions/holdings/margins, and computes Indian brokerage charges.
// This is the DEFAULT broker mode — all new users start with paper trading.
type PaperBroker struct {
	mu sync.RWMutex

	// Account state
	initialCapital float64
	cash           float64
	usedMargin     float64

	// Order management
	orders       map[string]*models.Order
	orderCounter int

	// Position tracking
	positions map[string]*models.Position // key: "TICKER:PRODUCT"
	holdings  map[string]*models.Holding  // key: "TICKER" (delivery only)

	// Configuration
	slippagePct float64 // simulated slippage (default 0.05%)
	fillDelay   time.Duration

	// Trade log
	logger *TradeLogger
}

// PaperBrokerConfig holds configuration for the paper broker.
type PaperBrokerConfig struct {
	InitialCapital float64       // starting capital in INR (default: ₹10,00,000)
	SlippagePct    float64       // simulated slippage percentage (default: 0.05%)
	FillDelay      time.Duration // simulated order fill delay (default: 100ms)
}

// NewPaperBroker creates a new paper trading simulator.
func NewPaperBroker(cfg *PaperBrokerConfig) *PaperBroker {
	if cfg == nil {
		cfg = &PaperBrokerConfig{}
	}

	capital := cfg.InitialCapital
	if capital <= 0 {
		capital = 1_000_000 // ₹10 lakhs default
	}

	slippage := cfg.SlippagePct
	if slippage <= 0 {
		slippage = 0.05
	}

	fillDelay := cfg.FillDelay
	if fillDelay <= 0 {
		fillDelay = 100 * time.Millisecond
	}

	return &PaperBroker{
		initialCapital: capital,
		cash:           capital,
		orders:         make(map[string]*models.Order),
		positions:      make(map[string]*models.Position),
		holdings:       make(map[string]*models.Holding),
		slippagePct:    slippage,
		fillDelay:      fillDelay,
		logger:         NewTradeLogger(),
	}
}

// Name returns "paper".
func (pb *PaperBroker) Name() string { return "paper" }

// ════════════════════════════════════════════════════════════════════
// Account
// ════════════════════════════════════════════════════════════════════

// GetMargins returns the current account margin information.
func (pb *PaperBroker) GetMargins(_ context.Context) (*models.Margins, error) {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	return &models.Margins{
		AvailableCash:   pb.cash,
		UsedMargin:      pb.usedMargin,
		AvailableMargin: pb.cash - pb.usedMargin,
		OpeningBalance:  pb.initialCapital,
	}, nil
}

// ════════════════════════════════════════════════════════════════════
// Positions & Holdings
// ════════════════════════════════════════════════════════════════════

// GetPositions returns all open positions.
func (pb *PaperBroker) GetPositions(_ context.Context) ([]models.Position, error) {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	positions := make([]models.Position, 0, len(pb.positions))
	for _, p := range pb.positions {
		positions = append(positions, *p)
	}
	return positions, nil
}

// GetHoldings returns all delivery holdings.
func (pb *PaperBroker) GetHoldings(_ context.Context) ([]models.Holding, error) {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	holdings := make([]models.Holding, 0, len(pb.holdings))
	for _, h := range pb.holdings {
		holdings = append(holdings, *h)
	}
	return holdings, nil
}

// ════════════════════════════════════════════════════════════════════
// Orders
// ════════════════════════════════════════════════════════════════════

// GetOrders returns all orders.
func (pb *PaperBroker) GetOrders(_ context.Context) ([]models.Order, error) {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	orders := make([]models.Order, 0, len(pb.orders))
	for _, o := range pb.orders {
		orders = append(orders, *o)
	}
	return orders, nil
}

// GetOrderByID returns a specific order.
func (pb *PaperBroker) GetOrderByID(_ context.Context, orderID string) (*models.Order, error) {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	order, ok := pb.orders[orderID]
	if !ok {
		return nil, ErrOrderNotFound
	}
	out := *order
	return &out, nil
}

// PlaceOrder simulates placing an order with the exchange.
func (pb *PaperBroker) PlaceOrder(ctx context.Context, req models.OrderRequest) (*models.OrderResponse, error) {
	// Validate the order
	validation := ValidateOrder(req)
	if !validation.IsValid() {
		return &models.OrderResponse{
			Status:  "REJECTED",
			Message: validation.ErrorString(),
		}, fmt.Errorf("%w: %s", ErrOrderRejected, validation.ErrorString())
	}

	pb.mu.Lock()
	defer pb.mu.Unlock()

	// Generate order ID
	pb.orderCounter++
	orderID := fmt.Sprintf("PAPER-%d-%d", time.Now().UnixMilli(), pb.orderCounter)

	now := time.Now()
	order := &models.Order{
		OrderID:      orderID,
		Ticker:       req.Ticker,
		Exchange:     req.Exchange,
		Side:         req.Side,
		OrderType:    req.OrderType,
		Product:      req.Product,
		Quantity:     req.Quantity,
		Price:        req.Price,
		TriggerPrice: req.TriggerPrice,
		Status:       models.OrderPending,
		PlacedAt:     now,
		UpdatedAt:    now,
		Tag:          req.Tag,
	}

	// Compute fill price with slippage
	fillPrice := pb.computeFillPrice(req)

	// Check margin
	requiredMargin := pb.computeRequiredMargin(req, fillPrice)
	available := pb.cash - pb.usedMargin
	if requiredMargin > available {
		order.Status = models.OrderRejected
		order.StatusMessage = fmt.Sprintf("insufficient margin: need ₹%.2f, available ₹%.2f", requiredMargin, available)
		pb.orders[orderID] = order

		pb.logger.Log(models.TradeLog{
			OrderRequest: req,
			OrderResponse: &models.OrderResponse{
				OrderID: orderID,
				Status:  "REJECTED",
				Message: order.StatusMessage,
			},
			Approved:  false,
			AgentName: "paper-broker",
			Reason:    order.StatusMessage,
		})

		return &models.OrderResponse{
			OrderID: orderID,
			Status:  "REJECTED",
			Message: order.StatusMessage,
		}, ErrInsufficientMargin
	}

	// Simulate fill
	order.Status = models.OrderComplete
	order.AvgPrice = fillPrice
	order.FilledQty = req.Quantity
	order.PendingQty = 0
	order.UpdatedAt = now

	pb.orders[orderID] = order

	// Update positions/holdings
	pb.updatePositions(order)

	// Log the trade
	pb.logger.Log(models.TradeLog{
		OrderRequest: req,
		OrderResponse: &models.OrderResponse{
			OrderID: orderID,
			Status:  "COMPLETE",
		},
		Approved:  true,
		AgentName: "paper-broker",
	})

	return &models.OrderResponse{
		OrderID: orderID,
		Status:  "COMPLETE",
		Message: fmt.Sprintf("filled at ₹%.2f", fillPrice),
	}, nil
}

// ModifyOrder simulates modifying an existing order.
func (pb *PaperBroker) ModifyOrder(_ context.Context, orderID string, req models.OrderRequest) (*models.OrderResponse, error) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	order, ok := pb.orders[orderID]
	if !ok {
		return nil, ErrOrderNotFound
	}

	if err := ValidateModifyOrder(order, req); err != nil {
		return nil, err
	}

	// Apply modifications
	if req.Quantity > 0 {
		order.Quantity = req.Quantity
		order.PendingQty = req.Quantity - order.FilledQty
	}
	if req.Price > 0 {
		order.Price = req.Price
	}
	if req.TriggerPrice > 0 {
		order.TriggerPrice = req.TriggerPrice
	}
	order.UpdatedAt = time.Now()

	return &models.OrderResponse{
		OrderID: orderID,
		Status:  string(order.Status),
		Message: "order modified",
	}, nil
}

// CancelOrder simulates cancelling an order.
func (pb *PaperBroker) CancelOrder(_ context.Context, orderID string) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	order, ok := pb.orders[orderID]
	if !ok {
		return ErrOrderNotFound
	}

	if order.Status != models.OrderPending && order.Status != models.OrderOpen {
		return fmt.Errorf("%w: order is %s", ErrOrderCantModify, order.Status)
	}

	order.Status = models.OrderCancelled
	order.StatusMessage = "cancelled by user"
	order.UpdatedAt = time.Now()

	return nil
}

// SubscribeQuotes is not supported for paper broker.
func (pb *PaperBroker) SubscribeQuotes(_ context.Context, _ []string) (<-chan models.Quote, error) {
	return nil, ErrNotSupported
}

// ════════════════════════════════════════════════════════════════════
// Paper-specific Methods
// ════════════════════════════════════════════════════════════════════

// Logger returns the trade logger.
func (pb *PaperBroker) Logger() *TradeLogger {
	return pb.logger
}

// Reset resets the paper broker to initial state.
func (pb *PaperBroker) Reset() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.cash = pb.initialCapital
	pb.usedMargin = 0
	pb.orders = make(map[string]*models.Order)
	pb.positions = make(map[string]*models.Position)
	pb.holdings = make(map[string]*models.Holding)
	pb.orderCounter = 0
	pb.logger = NewTradeLogger()
}

// SetPrice simulates updating the LTP (last traded price) for a ticker.
// This is used for P&L calculation in paper mode.
func (pb *PaperBroker) SetPrice(ticker string, price float64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// Update positions
	for key, pos := range pb.positions {
		if pos.Ticker == ticker {
			pos.LTP = price
			pos.Value = price * float64(absInt(pos.Quantity))
			if pos.Quantity > 0 { // long
				pos.PnL = (price - pos.AvgPrice) * float64(pos.Quantity)
			} else { // short
				pos.PnL = (pos.AvgPrice - price) * float64(-pos.Quantity)
			}
			if pos.AvgPrice > 0 {
				pos.PnLPct = (pos.PnL / (pos.AvgPrice * float64(absInt(pos.Quantity)))) * 100
			}
			pb.positions[key] = pos
		}
	}

	// Update holdings
	for key, h := range pb.holdings {
		if h.Ticker == ticker {
			h.LTP = price
			h.CurrentValue = price * float64(h.Quantity)
			h.PnL = h.CurrentValue - h.InvestedValue
			if h.InvestedValue > 0 {
				h.PnLPct = (h.PnL / h.InvestedValue) * 100
			}
			pb.holdings[key] = h
		}
	}
}

// TotalPnL returns the total P&L across all positions and holdings.
func (pb *PaperBroker) TotalPnL() float64 {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	var total float64
	for _, p := range pb.positions {
		total += p.PnL
	}
	for _, h := range pb.holdings {
		total += h.PnL
	}
	return total
}

// PositionCount returns the number of open positions.
func (pb *PaperBroker) PositionCount() int {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	return len(pb.positions)
}

// ════════════════════════════════════════════════════════════════════
// Internal Helpers
// ════════════════════════════════════════════════════════════════════

// computeFillPrice simulates order fill with slippage.
func (pb *PaperBroker) computeFillPrice(req models.OrderRequest) float64 {
	basePrice := req.Price
	if req.OrderType == models.Market || basePrice <= 0 {
		// For market orders, use trigger price as reference or default to 100
		basePrice = req.TriggerPrice
		if basePrice <= 0 {
			basePrice = 100 // shouldn't happen in real usage
		}
	}

	// Apply random slippage
	slippage := basePrice * (pb.slippagePct / 100) * (rand.Float64()*2 - 1) // ±slippage%
	if req.Side == models.Buy {
		// Buyers pay slightly more
		return basePrice + absFloat(slippage)
	}
	// Sellers get slightly less
	return basePrice - absFloat(slippage)
}

// computeRequiredMargin calculates the margin needed for an order.
func (pb *PaperBroker) computeRequiredMargin(req models.OrderRequest, fillPrice float64) float64 {
	value := fillPrice * float64(req.Quantity)

	switch req.Product {
	case models.CNC:
		return value // full value for delivery
	case models.MIS:
		return value * 0.20 // 5x leverage for intraday
	case models.NRML:
		return value * 0.15 // ~6.7x for F&O
	default:
		return value
	}
}

// updatePositions updates positions/holdings based on a filled order.
func (pb *PaperBroker) updatePositions(order *models.Order) {
	if order.Product == models.CNC {
		pb.updateHoldings(order)
	} else {
		pb.updateTradePositions(order)
	}
}

// updateHoldings updates delivery holdings (CNC).
func (pb *PaperBroker) updateHoldings(order *models.Order) {
	key := order.Ticker

	existing, exists := pb.holdings[key]

	if order.Side == models.Buy {
		cost := order.AvgPrice * float64(order.FilledQty)
		pb.cash -= cost

		if exists {
			// Average up/down
			totalQty := existing.Quantity + order.FilledQty
			totalInvested := existing.InvestedValue + cost
			existing.Quantity = totalQty
			existing.AvgPrice = totalInvested / float64(totalQty)
			existing.InvestedValue = totalInvested
			existing.CurrentValue = existing.LTP * float64(totalQty)
		} else {
			pb.holdings[key] = &models.Holding{
				Ticker:        order.Ticker,
				Exchange:      order.Exchange,
				Quantity:      order.FilledQty,
				AvgPrice:      order.AvgPrice,
				LTP:           order.AvgPrice,
				InvestedValue: cost,
				CurrentValue:  cost,
			}
		}
	} else { // SELL
		if !exists || existing.Quantity < order.FilledQty {
			// Short sell not allowed in CNC — handled as best effort
			return
		}
		proceeds := order.AvgPrice * float64(order.FilledQty)
		pb.cash += proceeds

		existing.Quantity -= order.FilledQty
		existing.InvestedValue = existing.AvgPrice * float64(existing.Quantity)
		existing.CurrentValue = existing.LTP * float64(existing.Quantity)

		if existing.Quantity == 0 {
			delete(pb.holdings, key)
		}
	}
}

// updateTradePositions updates intraday/F&O positions (MIS/NRML).
func (pb *PaperBroker) updateTradePositions(order *models.Order) {
	key := fmt.Sprintf("%s:%s", order.Ticker, order.Product)

	existing, exists := pb.positions[key]

	qty := order.FilledQty
	if order.Side == models.Sell {
		qty = -qty // negative for shorts
	}

	margin := pb.computeRequiredMargin(models.OrderRequest{
		Product:  order.Product,
		Quantity: order.FilledQty,
	}, order.AvgPrice)

	if !exists {
		pb.positions[key] = &models.Position{
			Ticker:   order.Ticker,
			Exchange: order.Exchange,
			Product:  order.Product,
			Quantity: qty,
			AvgPrice: order.AvgPrice,
			LTP:      order.AvgPrice,
			Value:    order.AvgPrice * float64(absInt(qty)),
		}
		pb.usedMargin += margin
		pb.cash -= margin
		return
	}

	// Position exists — adjust
	oldQty := existing.Quantity
	newQty := oldQty + qty

	if newQty == 0 {
		// Position closed
		if oldQty > 0 {
			existing.PnL = (order.AvgPrice - existing.AvgPrice) * float64(oldQty)
		} else {
			existing.PnL = (existing.AvgPrice - order.AvgPrice) * float64(-oldQty)
		}
		pb.cash += margin + existing.PnL
		pb.usedMargin -= margin
		delete(pb.positions, key)
		return
	}

	// Position increased or partially closed
	if (oldQty > 0 && qty > 0) || (oldQty < 0 && qty < 0) {
		// Adding to position — average price
		totalValue := existing.AvgPrice*float64(absInt(oldQty)) + order.AvgPrice*float64(absInt(qty))
		existing.AvgPrice = totalValue / float64(absInt(newQty))
		pb.usedMargin += margin
		pb.cash -= margin
	} else {
		// Partial close — realize partial P&L
		closedQty := absInt(qty)
		if closedQty > absInt(oldQty) {
			closedQty = absInt(oldQty)
		}
		if oldQty > 0 {
			existing.PnL += (order.AvgPrice - existing.AvgPrice) * float64(closedQty)
		} else {
			existing.PnL += (existing.AvgPrice - order.AvgPrice) * float64(closedQty)
		}
		partialMargin := margin * (float64(closedQty) / float64(absInt(qty)))
		pb.cash += partialMargin + existing.PnL
		pb.usedMargin -= partialMargin
	}

	existing.Quantity = newQty
	existing.Value = existing.LTP * float64(absInt(newQty))
}

// absInt returns absolute value of int.
func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// absFloat returns absolute value of float64.
func absFloat(n float64) float64 {
	if n < 0 {
		return -n
	}
	return n
}
