package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Interactive Brokers Client Portal REST API Broker
// ════════════════════════════════════════════════════════════════════

// IBKRBroker implements the Broker interface using Interactive Brokers'
// Client Portal REST API. The IB Gateway or TWS must be running locally
// and the Client Portal API authenticated before using this broker.
type IBKRBroker struct {
	mu sync.RWMutex

	baseURL    string
	httpClient *http.Client

	accountID string
	connected bool
	logger    *TradeLogger
}

// IBKRConnectConfig holds IBKR connection settings.
type IBKRConnectConfig struct {
	Host      string        // Gateway host (default: "localhost")
	Port      int           // Gateway port (default: 5000)
	AccountID string        // IB account ID
	Timeout   time.Duration // HTTP timeout (default: 30s)
}

// NewIBKRBroker creates a new IBKR broker instance.
func NewIBKRBroker(cfg *IBKRConnectConfig) *IBKRBroker {
	if cfg == nil {
		cfg = &IBKRConnectConfig{}
	}

	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Port
	if port <= 0 {
		port = 5000
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &IBKRBroker{
		baseURL:   fmt.Sprintf("https://%s:%d/v1/api", host, port),
		accountID: cfg.AccountID,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: NewTradeLogger(),
	}
}

// Name returns "ibkr".
func (ib *IBKRBroker) Name() string { return "ibkr" }

// ════════════════════════════════════════════════════════════════════
// Connection
// ════════════════════════════════════════════════════════════════════

// Connect verifies the Client Portal API connection and discovers the account ID.
func (ib *IBKRBroker) Connect(ctx context.Context) error {
	ib.mu.Lock()
	defer ib.mu.Unlock()

	// Check authentication status
	body, err := ib.doGet(ctx, "/iserver/auth/status")
	if err != nil {
		return fmt.Errorf("ibkr auth check: %w", err)
	}

	var status struct {
		Authenticated bool   `json:"authenticated"`
		Connected     bool   `json:"connected"`
		Competing     bool   `json:"competing"`
	}
	if err := json.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("parse auth status: %w", err)
	}

	if !status.Authenticated {
		return fmt.Errorf("%w: IBKR gateway not authenticated — login via Client Portal first", ErrNotConnected)
	}

	// Get accounts if not specified
	if ib.accountID == "" {
		acctBody, err := ib.doGet(ctx, "/portfolio/accounts")
		if err != nil {
			return fmt.Errorf("get accounts: %w", err)
		}
		var accounts []struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(acctBody, &accounts); err != nil {
			return fmt.Errorf("parse accounts: %w", err)
		}
		if len(accounts) == 0 {
			return fmt.Errorf("no IBKR accounts found")
		}
		ib.accountID = accounts[0].ID
	}

	ib.connected = true
	return nil
}

// IsConnected returns whether the IBKR gateway is connected.
func (ib *IBKRBroker) IsConnected() bool {
	ib.mu.RLock()
	defer ib.mu.RUnlock()
	return ib.connected
}

// ════════════════════════════════════════════════════════════════════
// Account
// ════════════════════════════════════════════════════════════════════

// GetMargins returns account balance/margin information.
func (ib *IBKRBroker) GetMargins(ctx context.Context) (*models.Margins, error) {
	if !ib.IsConnected() {
		return nil, ErrNotConnected
	}

	body, err := ib.doGet(ctx, fmt.Sprintf("/portfolio/%s/summary", ib.accountID))
	if err != nil {
		return nil, fmt.Errorf("get margins: %w", err)
	}

	var summary map[string]struct {
		Amount float64 `json:"amount"`
	}
	if err := json.Unmarshal(body, &summary); err != nil {
		return nil, fmt.Errorf("parse margins: %w", err)
	}

	return &models.Margins{
		AvailableCash:   summary["totalcashvalue"].Amount,
		UsedMargin:      summary["initmarginreq"].Amount,
		AvailableMargin: summary["buyingpower"].Amount,
	}, nil
}

// ════════════════════════════════════════════════════════════════════
// Positions & Holdings
// ════════════════════════════════════════════════════════════════════

// GetPositions returns open positions from IBKR.
func (ib *IBKRBroker) GetPositions(ctx context.Context) ([]models.Position, error) {
	if !ib.IsConnected() {
		return nil, ErrNotConnected
	}

	body, err := ib.doGet(ctx, fmt.Sprintf("/portfolio/%s/positions/0", ib.accountID))
	if err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}

	var ibPositions []struct {
		ContractDesc string  `json:"contractDesc"`
		Position     float64 `json:"position"`
		AvgCost      float64 `json:"avgCost"`
		MktPrice     float64 `json:"mktPrice"`
		UnrealizedPnl float64 `json:"unrealizedPnl"`
		MktValue     float64 `json:"mktValue"`
	}
	if err := json.Unmarshal(body, &ibPositions); err != nil {
		return nil, fmt.Errorf("parse positions: %w", err)
	}

	positions := make([]models.Position, 0, len(ibPositions))
	for _, p := range ibPositions {
		if p.Position == 0 {
			continue
		}
		positions = append(positions, models.Position{
			Ticker:   p.ContractDesc,
			Quantity: int(p.Position),
			AvgPrice: p.AvgCost,
			LTP:      p.MktPrice,
			PnL:      p.UnrealizedPnl,
			Value:    p.MktValue,
		})
	}
	return positions, nil
}

// GetHoldings returns holdings — maps to IBKR positions as IBKR doesn't
// distinguish between holdings and positions.
func (ib *IBKRBroker) GetHoldings(ctx context.Context) ([]models.Holding, error) {
	if !ib.IsConnected() {
		return nil, ErrNotConnected
	}

	// IBKR doesn't have a distinct holdings concept — return empty
	return []models.Holding{}, nil
}

// ════════════════════════════════════════════════════════════════════
// Orders
// ════════════════════════════════════════════════════════════════════

// GetOrders returns all live orders from IBKR.
func (ib *IBKRBroker) GetOrders(ctx context.Context) ([]models.Order, error) {
	if !ib.IsConnected() {
		return nil, ErrNotConnected
	}

	body, err := ib.doGet(ctx, "/iserver/account/orders")
	if err != nil {
		return nil, fmt.Errorf("get orders: %w", err)
	}

	var resp struct {
		Orders []struct {
			OrderID    int     `json:"orderId"`
			Symbol     string  `json:"ticker"`
			Side       string  `json:"side"`
			OrderType  string  `json:"orderType"`
			TotalQty   float64 `json:"totalSize"`
			FilledQty  float64 `json:"filledQuantity"`
			Price      float64 `json:"price"`
			AvgPrice   float64 `json:"avgPrice"`
			Status     string  `json:"status"`
		} `json:"orders"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse orders: %w", err)
	}

	orders := make([]models.Order, 0, len(resp.Orders))
	for _, o := range resp.Orders {
		side := models.Buy
		if strings.ToUpper(o.Side) == "SELL" {
			side = models.Sell
		}

		orders = append(orders, models.Order{
			OrderID:  fmt.Sprintf("%d", o.OrderID),
			Ticker:   o.Symbol,
			Side:     side,
			Quantity: int(o.TotalQty),
			FilledQty: int(o.FilledQty),
			PendingQty: int(o.TotalQty) - int(o.FilledQty),
			Price:    o.Price,
			AvgPrice: o.AvgPrice,
			Status:   mapIBKRStatus(o.Status),
		})
	}
	return orders, nil
}

// GetOrderByID returns a specific order — IBKR doesn't have a direct
// single-order endpoint, so we filter from all orders.
func (ib *IBKRBroker) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	orders, err := ib.GetOrders(ctx)
	if err != nil {
		return nil, err
	}

	for _, o := range orders {
		if o.OrderID == orderID {
			return &o, nil
		}
	}
	return nil, ErrOrderNotFound
}

// PlaceOrder places a new order via IBKR Client Portal API.
func (ib *IBKRBroker) PlaceOrder(ctx context.Context, req models.OrderRequest) (*models.OrderResponse, error) {
	if !ib.IsConnected() {
		return nil, ErrNotConnected
	}

	validation := ValidateOrder(req)
	if !validation.IsValid() {
		return &models.OrderResponse{
			Status:  "REJECTED",
			Message: validation.ErrorString(),
		}, fmt.Errorf("%w: %s", ErrOrderRejected, validation.ErrorString())
	}

	orderType := "LMT"
	switch req.OrderType {
	case models.Market:
		orderType = "MKT"
	case models.SL:
		orderType = "STP LMT"
	case models.SLM:
		orderType = "STP"
	}

	side := "BUY"
	if req.Side == models.Sell {
		side = "SELL"
	}

	payload := map[string]interface{}{
		"orders": []map[string]interface{}{
			{
				"acctId":    ib.accountID,
				"conid":     0, // Would need symbol resolution in production
				"orderType": orderType,
				"side":      side,
				"quantity":  req.Quantity,
				"price":     req.Price,
				"tif":       "DAY",
			},
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	body, err := ib.doPost(ctx, fmt.Sprintf("/iserver/account/%s/orders", ib.accountID), payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}

	var resp []struct {
		OrderID   string `json:"order_id"`
		OrderStatus string `json:"order_status"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse order response: %w", err)
	}

	orderID := ""
	status := "PLACED"
	if len(resp) > 0 {
		orderID = resp[0].OrderID
		if resp[0].OrderStatus != "" {
			status = resp[0].OrderStatus
		}
	}

	result := &models.OrderResponse{
		OrderID: orderID,
		Status:  status,
		Message: "order placed via IBKR",
	}

	ib.logger.Log(models.TradeLog{
		OrderRequest:  req,
		OrderResponse: result,
		Approved:      true,
		AgentName:     "ibkr-broker",
	})

	return result, nil
}

// ModifyOrder modifies an existing IBKR order.
func (ib *IBKRBroker) ModifyOrder(ctx context.Context, orderID string, req models.OrderRequest) (*models.OrderResponse, error) {
	if !ib.IsConnected() {
		return nil, ErrNotConnected
	}

	payload := map[string]interface{}{
		"quantity": req.Quantity,
		"price":   req.Price,
	}
	payloadBytes, _ := json.Marshal(payload)

	_, err := ib.doPost(ctx, fmt.Sprintf("/iserver/account/%s/order/%s", ib.accountID, orderID), payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("modify order: %w", err)
	}

	return &models.OrderResponse{
		OrderID: orderID,
		Status:  "MODIFIED",
		Message: "order modified via IBKR",
	}, nil
}

// CancelOrder cancels an IBKR order.
func (ib *IBKRBroker) CancelOrder(ctx context.Context, orderID string) error {
	if !ib.IsConnected() {
		return ErrNotConnected
	}

	_, err := ib.doDelete(ctx, fmt.Sprintf("/iserver/account/%s/order/%s", ib.accountID, orderID))
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	return nil
}

// SubscribeQuotes is not yet implemented for IBKR.
func (ib *IBKRBroker) SubscribeQuotes(_ context.Context, _ []string) (<-chan models.Quote, error) {
	return nil, ErrNotSupported
}

// Logger returns the trade logger.
func (ib *IBKRBroker) Logger() *TradeLogger {
	return ib.logger
}

// ════════════════════════════════════════════════════════════════════
// HTTP Helpers
// ════════════════════════════════════════════════════════════════════

func (ib *IBKRBroker) doGet(ctx context.Context, path string) ([]byte, error) {
	return ib.doRequest(ctx, http.MethodGet, path, nil)
}

func (ib *IBKRBroker) doPost(ctx context.Context, path string, body []byte) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		reader = strings.NewReader(string(body))
	}
	return ib.doRequest(ctx, http.MethodPost, path, reader)
}

func (ib *IBKRBroker) doDelete(ctx context.Context, path string) ([]byte, error) {
	return ib.doRequest(ctx, http.MethodDelete, path, nil)
}

func (ib *IBKRBroker) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	reqURL := fmt.Sprintf("%s%s", ib.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ib.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ibkr api error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ════════════════════════════════════════════════════════════════════
// Internal Utilities
// ════════════════════════════════════════════════════════════════════

// mapIBKRStatus maps IBKR order status strings to models.OrderStatus.
func mapIBKRStatus(status string) models.OrderStatus {
	switch strings.ToUpper(status) {
	case "FILLED":
		return models.OrderComplete
	case "CANCELLED":
		return models.OrderCancelled
	case "INACTIVE":
		return models.OrderRejected
	case "SUBMITTED", "PRESUBMITTED":
		return models.OrderOpen
	default:
		return models.OrderPending
	}
}
