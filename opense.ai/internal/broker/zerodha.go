package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Zerodha Kite Connect v3 Broker
// ════════════════════════════════════════════════════════════════════

// ZerodhaBroker implements the Broker interface using Zerodha's Kite
// Connect v3 REST API. It supports OAuth token-based authentication,
// order placement/modification, position & holdings retrieval, and
// margin checks.
type ZerodhaBroker struct {
	mu sync.RWMutex

	apiKey      string
	apiSecret   string
	accessToken string
	baseURL     string
	httpClient  *http.Client

	connected bool
	logger    *TradeLogger
}

// ZerodhaConfig holds Zerodha connection settings.
type ZerodhaConnectConfig struct {
	APIKey    string
	APISecret string
	BaseURL   string        // defaults to "https://api.kite.trade"
	Timeout   time.Duration // HTTP client timeout (default: 30s)
}

// NewZerodhaBroker creates a new Zerodha broker instance.
func NewZerodhaBroker(cfg *ZerodhaConnectConfig) *ZerodhaBroker {
	if cfg == nil {
		cfg = &ZerodhaConnectConfig{}
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.kite.trade"
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &ZerodhaBroker{
		apiKey:    cfg.APIKey,
		apiSecret: cfg.APISecret,
		baseURL:   baseURL,
		httpClient: &http.Client{Timeout: timeout},
		logger:    NewTradeLogger(),
	}
}

// Name returns "zerodha".
func (zb *ZerodhaBroker) Name() string { return "zerodha" }

// ════════════════════════════════════════════════════════════════════
// Authentication
// ════════════════════════════════════════════════════════════════════

// LoginURL returns the Kite login URL for OAuth flow.
func (zb *ZerodhaBroker) LoginURL() string {
	return fmt.Sprintf("https://kite.zerodha.com/connect/login?v=3&api_key=%s", zb.apiKey)
}

// SetAccessToken sets the access token after OAuth callback.
func (zb *ZerodhaBroker) SetAccessToken(token string) {
	zb.mu.Lock()
	defer zb.mu.Unlock()
	zb.accessToken = token
	zb.connected = true
}

// IsConnected returns whether the broker has a valid access token.
func (zb *ZerodhaBroker) IsConnected() bool {
	zb.mu.RLock()
	defer zb.mu.RUnlock()
	return zb.connected
}

// ════════════════════════════════════════════════════════════════════
// Account
// ════════════════════════════════════════════════════════════════════

// GetMargins returns account margin information from Kite.
func (zb *ZerodhaBroker) GetMargins(ctx context.Context) (*models.Margins, error) {
	if !zb.IsConnected() {
		return nil, ErrNotConnected
	}

	body, err := zb.doGet(ctx, "/user/margins/equity")
	if err != nil {
		return nil, fmt.Errorf("get margins: %w", err)
	}

	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Available struct {
				Cash      float64 `json:"cash"`
				Collateral float64 `json:"collateral"`
			} `json:"available"`
			Utilised struct {
				Debits float64 `json:"debits"`
			} `json:"utilised"`
			Net float64 `json:"net"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse margins: %w", err)
	}

	return &models.Margins{
		AvailableCash:   resp.Data.Available.Cash,
		UsedMargin:      resp.Data.Utilised.Debits,
		AvailableMargin: resp.Data.Net,
		Collateral:      resp.Data.Available.Collateral,
	}, nil
}

// ════════════════════════════════════════════════════════════════════
// Positions & Holdings
// ════════════════════════════════════════════════════════════════════

// GetPositions returns all open positions from Kite.
func (zb *ZerodhaBroker) GetPositions(ctx context.Context) ([]models.Position, error) {
	if !zb.IsConnected() {
		return nil, ErrNotConnected
	}

	body, err := zb.doGet(ctx, "/portfolio/positions")
	if err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}

	var resp struct {
		Data struct {
			Net []struct {
				TradingSymbol string  `json:"tradingsymbol"`
				Exchange      string  `json:"exchange"`
				Product       string  `json:"product"`
				Quantity      int     `json:"quantity"`
				AveragePrice  float64 `json:"average_price"`
				LastPrice     float64 `json:"last_price"`
				PnL           float64 `json:"pnl"`
				Value         float64 `json:"value"`
				Multiplier    int     `json:"multiplier"`
			} `json:"net"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse positions: %w", err)
	}

	positions := make([]models.Position, 0, len(resp.Data.Net))
	for _, p := range resp.Data.Net {
		if p.Quantity == 0 {
			continue
		}
		positions = append(positions, models.Position{
			Ticker:   p.TradingSymbol,
			Exchange: p.Exchange,
			Product:  models.OrderProduct(p.Product),
			Quantity: p.Quantity,
			AvgPrice: p.AveragePrice,
			LTP:      p.LastPrice,
			PnL:      p.PnL,
			Value:    p.Value,
			Multiplier: p.Multiplier,
		})
	}
	return positions, nil
}

// GetHoldings returns all delivery holdings from Kite.
func (zb *ZerodhaBroker) GetHoldings(ctx context.Context) ([]models.Holding, error) {
	if !zb.IsConnected() {
		return nil, ErrNotConnected
	}

	body, err := zb.doGet(ctx, "/portfolio/holdings")
	if err != nil {
		return nil, fmt.Errorf("get holdings: %w", err)
	}

	var resp struct {
		Data []struct {
			TradingSymbol string  `json:"tradingsymbol"`
			Exchange      string  `json:"exchange"`
			ISIN          string  `json:"isin"`
			Quantity      int     `json:"quantity"`
			AveragePrice  float64 `json:"average_price"`
			LastPrice     float64 `json:"last_price"`
			PnL           float64 `json:"pnl"`
			DayChange     float64 `json:"day_change"`
			DayChangePct  float64 `json:"day_change_percentage"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse holdings: %w", err)
	}

	holdings := make([]models.Holding, 0, len(resp.Data))
	for _, h := range resp.Data {
		invested := h.AveragePrice * float64(h.Quantity)
		holdings = append(holdings, models.Holding{
			Ticker:        h.TradingSymbol,
			Exchange:      h.Exchange,
			ISIN:          h.ISIN,
			Quantity:      h.Quantity,
			AvgPrice:      h.AveragePrice,
			LTP:           h.LastPrice,
			PnL:           h.PnL,
			CurrentValue:  h.LastPrice * float64(h.Quantity),
			InvestedValue: invested,
			DayChange:     h.DayChange,
			DayChangePct:  h.DayChangePct,
		})
	}
	return holdings, nil
}

// ════════════════════════════════════════════════════════════════════
// Orders
// ════════════════════════════════════════════════════════════════════

// GetOrders returns all orders for the day from Kite.
func (zb *ZerodhaBroker) GetOrders(ctx context.Context) ([]models.Order, error) {
	if !zb.IsConnected() {
		return nil, ErrNotConnected
	}

	body, err := zb.doGet(ctx, "/orders")
	if err != nil {
		return nil, fmt.Errorf("get orders: %w", err)
	}

	var resp struct {
		Data []struct {
			OrderID       string  `json:"order_id"`
			TradingSymbol string  `json:"tradingsymbol"`
			Exchange      string  `json:"exchange"`
			TransType     string  `json:"transaction_type"`
			OrderType     string  `json:"order_type"`
			Product       string  `json:"product"`
			Quantity      int     `json:"quantity"`
			FilledQty     int     `json:"filled_quantity"`
			PendingQty    int     `json:"pending_quantity"`
			Price         float64 `json:"price"`
			AvgPrice      float64 `json:"average_price"`
			TriggerPrice  float64 `json:"trigger_price"`
			Status        string  `json:"status"`
			StatusMessage string  `json:"status_message"`
			Tag           string  `json:"tag"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse orders: %w", err)
	}

	orders := make([]models.Order, 0, len(resp.Data))
	for _, o := range resp.Data {
		orders = append(orders, models.Order{
			OrderID:       o.OrderID,
			Ticker:        o.TradingSymbol,
			Exchange:      o.Exchange,
			Side:          models.OrderSide(o.TransType),
			OrderType:     models.OrderType(o.OrderType),
			Product:       models.OrderProduct(o.Product),
			Quantity:      o.Quantity,
			FilledQty:     o.FilledQty,
			PendingQty:    o.PendingQty,
			Price:         o.Price,
			AvgPrice:      o.AvgPrice,
			TriggerPrice:  o.TriggerPrice,
			Status:        mapKiteStatus(o.Status),
			StatusMessage: o.StatusMessage,
			Tag:           o.Tag,
		})
	}
	return orders, nil
}

// GetOrderByID returns a specific order from Kite.
func (zb *ZerodhaBroker) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	if !zb.IsConnected() {
		return nil, ErrNotConnected
	}

	body, err := zb.doGet(ctx, fmt.Sprintf("/orders/%s", orderID))
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	var resp struct {
		Data []struct {
			OrderID       string  `json:"order_id"`
			TradingSymbol string  `json:"tradingsymbol"`
			Exchange      string  `json:"exchange"`
			TransType     string  `json:"transaction_type"`
			OrderType     string  `json:"order_type"`
			Product       string  `json:"product"`
			Quantity      int     `json:"quantity"`
			FilledQty     int     `json:"filled_quantity"`
			PendingQty    int     `json:"pending_quantity"`
			Price         float64 `json:"price"`
			AvgPrice      float64 `json:"average_price"`
			TriggerPrice  float64 `json:"trigger_price"`
			Status        string  `json:"status"`
			StatusMessage string  `json:"status_message"`
			Tag           string  `json:"tag"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse order: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, ErrOrderNotFound
	}

	// Return the latest state (last entry in the order history)
	o := resp.Data[len(resp.Data)-1]
	return &models.Order{
		OrderID:       o.OrderID,
		Ticker:        o.TradingSymbol,
		Exchange:      o.Exchange,
		Side:          models.OrderSide(o.TransType),
		OrderType:     models.OrderType(o.OrderType),
		Product:       models.OrderProduct(o.Product),
		Quantity:      o.Quantity,
		FilledQty:     o.FilledQty,
		PendingQty:    o.PendingQty,
		Price:         o.Price,
		AvgPrice:      o.AvgPrice,
		TriggerPrice:  o.TriggerPrice,
		Status:        mapKiteStatus(o.Status),
		StatusMessage: o.StatusMessage,
		Tag:           o.Tag,
	}, nil
}

// PlaceOrder places a new order via Kite API.
func (zb *ZerodhaBroker) PlaceOrder(ctx context.Context, req models.OrderRequest) (*models.OrderResponse, error) {
	if !zb.IsConnected() {
		return nil, ErrNotConnected
	}

	validation := ValidateOrder(req)
	if !validation.IsValid() {
		return &models.OrderResponse{
			Status:  "REJECTED",
			Message: validation.ErrorString(),
		}, fmt.Errorf("%w: %s", ErrOrderRejected, validation.ErrorString())
	}

	// Prepare Kite API form parameters
	params := url.Values{}
	params.Set("tradingsymbol", req.Ticker)
	params.Set("exchange", req.Exchange)
	params.Set("transaction_type", string(req.Side))
	params.Set("order_type", string(req.OrderType))
	params.Set("product", string(req.Product))
	params.Set("quantity", fmt.Sprintf("%d", req.Quantity))
	if req.Price > 0 {
		params.Set("price", fmt.Sprintf("%.2f", req.Price))
	}
	if req.TriggerPrice > 0 {
		params.Set("trigger_price", fmt.Sprintf("%.2f", req.TriggerPrice))
	}
	if req.Tag != "" {
		params.Set("tag", req.Tag)
	}

	body, err := zb.doPost(ctx, "/orders/regular", params)
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}

	var resp struct {
		Data struct {
			OrderID string `json:"order_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse place order response: %w", err)
	}

	result := &models.OrderResponse{
		OrderID: resp.Data.OrderID,
		Status:  "PLACED",
		Message: "order placed successfully",
	}

	zb.logger.Log(models.TradeLog{
		OrderRequest:  req,
		OrderResponse: result,
		Approved:      true,
		AgentName:     "zerodha-broker",
	})

	return result, nil
}

// ModifyOrder modifies an existing order via Kite API.
func (zb *ZerodhaBroker) ModifyOrder(ctx context.Context, orderID string, req models.OrderRequest) (*models.OrderResponse, error) {
	if !zb.IsConnected() {
		return nil, ErrNotConnected
	}

	params := url.Values{}
	if req.Quantity > 0 {
		params.Set("quantity", fmt.Sprintf("%d", req.Quantity))
	}
	if req.Price > 0 {
		params.Set("price", fmt.Sprintf("%.2f", req.Price))
	}
	if req.TriggerPrice > 0 {
		params.Set("trigger_price", fmt.Sprintf("%.2f", req.TriggerPrice))
	}

	_, err := zb.doPut(ctx, fmt.Sprintf("/orders/regular/%s", orderID), params)
	if err != nil {
		return nil, fmt.Errorf("modify order: %w", err)
	}

	return &models.OrderResponse{
		OrderID: orderID,
		Status:  "MODIFIED",
		Message: "order modified",
	}, nil
}

// CancelOrder cancels an order via Kite API.
func (zb *ZerodhaBroker) CancelOrder(ctx context.Context, orderID string) error {
	if !zb.IsConnected() {
		return ErrNotConnected
	}

	_, err := zb.doDelete(ctx, fmt.Sprintf("/orders/regular/%s", orderID))
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	return nil
}

// SubscribeQuotes is not yet implemented for Zerodha.
func (zb *ZerodhaBroker) SubscribeQuotes(_ context.Context, _ []string) (<-chan models.Quote, error) {
	return nil, ErrNotSupported
}

// Logger returns the trade logger.
func (zb *ZerodhaBroker) Logger() *TradeLogger {
	return zb.logger
}

// ════════════════════════════════════════════════════════════════════
// HTTP Helpers
// ════════════════════════════════════════════════════════════════════

func (zb *ZerodhaBroker) doGet(ctx context.Context, path string) ([]byte, error) {
	return zb.doRequest(ctx, http.MethodGet, path, nil)
}

func (zb *ZerodhaBroker) doPost(ctx context.Context, path string, params url.Values) ([]byte, error) {
	return zb.doRequest(ctx, http.MethodPost, path, strings.NewReader(params.Encode()))
}

func (zb *ZerodhaBroker) doPut(ctx context.Context, path string, params url.Values) ([]byte, error) {
	return zb.doRequest(ctx, http.MethodPut, path, strings.NewReader(params.Encode()))
}

func (zb *ZerodhaBroker) doDelete(ctx context.Context, path string) ([]byte, error) {
	return zb.doRequest(ctx, http.MethodDelete, path, nil)
}

func (zb *ZerodhaBroker) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	zb.mu.RLock()
	token := zb.accessToken
	key := zb.apiKey
	zb.mu.RUnlock()

	reqURL := fmt.Sprintf("%s%s", zb.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s:%s", key, token))
	req.Header.Set("X-Kite-Version", "3")
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := zb.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kite api error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ════════════════════════════════════════════════════════════════════
// Internal Utilities
// ════════════════════════════════════════════════════════════════════

// mapKiteStatus maps Kite order status strings to models.OrderStatus.
func mapKiteStatus(status string) models.OrderStatus {
	switch strings.ToUpper(status) {
	case "COMPLETE":
		return models.OrderComplete
	case "CANCELLED":
		return models.OrderCancelled
	case "REJECTED":
		return models.OrderRejected
	case "OPEN":
		return models.OrderOpen
	case "TRIGGER PENDING":
		return models.OrderPending
	default:
		return models.OrderPending
	}
}
