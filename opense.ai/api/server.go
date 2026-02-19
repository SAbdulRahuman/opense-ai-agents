// Package api provides the HTTP REST API server for OpeNSE.ai.
//
// It exposes endpoints for stock analysis, quotes, backtesting,
// portfolio management, chat, FinanceQL queries, and WebSocket streaming.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/seenimoa/openseai/internal/agent"
	"github.com/seenimoa/openseai/internal/backtest"
	"github.com/seenimoa/openseai/internal/broker"
	"github.com/seenimoa/openseai/internal/config"
	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/internal/financeql"
	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/pkg/models"
	"github.com/seenimoa/openseai/pkg/utils"
)

// Server is the HTTP API server.
type Server struct {
	router   chi.Router
	cfg      *config.Config
	orch     *agent.Orchestrator
	agg      *datasource.Aggregator
	broker   broker.Broker
	riskMgr  *broker.RiskManager
	wsHub    *WSHub
}

// NewServer creates a configured API server with all routes and middleware.
func NewServer(cfg *config.Config) (*Server, error) {
	agg := datasource.NewAggregator()

	router, err := llm.NewRouterFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("LLM setup failed: %w", err)
	}

	opts := &llm.ChatOptions{
		Model:       cfg.LLM.Model,
		Temperature: cfg.LLM.Temperature,
		MaxTokens:   cfg.LLM.MaxTokens,
	}

	orch := agent.NewOrchestrator(agent.OrchestratorConfig{
		Provider:    router,
		Aggregator:  agg,
		ChatOptions: opts,
		DefaultMode: agent.ModeSingle,
		Capital:     cfg.Trading.InitialCapital,
	})

	b := broker.NewPaperBroker(nil)
	riskCfg := broker.DefaultRiskConfig()
	riskCfg.MaxPositionPct = cfg.Trading.MaxPositionPct
	riskCfg.DailyLossLimitPct = cfg.Trading.DailyLossLimitPct
	riskCfg.MaxOpenPositions = cfg.Trading.MaxOpenPositions
	rm := broker.NewRiskManager(b, riskCfg)

	srv := &Server{
		cfg:     cfg,
		orch:    orch,
		agg:     agg,
		broker:  b,
		riskMgr: rm,
		wsHub:   NewWSHub(),
	}

	srv.router = srv.buildRouter()
	return srv, nil
}

// Router returns the chi router for testing.
func (s *Server) Router() chi.Router {
	return s.router
}

// ListenAndServe starts the HTTP server with graceful shutdown.
func (s *Server) ListenAndServe(addr string) error {
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start WebSocket hub
	go s.wsHub.Run()

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return httpSrv.Shutdown(ctx)
}

// buildRouter configures all routes and middleware.
func (s *Server) buildRouter() chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(120 * time.Second))

	// CORS
	origins := []string{"*"}
	if len(s.cfg.API.CORSOrigins) > 0 {
		origins = s.cfg.API.CORSOrigins
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", s.handleHealth)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Analysis
		r.Post("/analyze", s.handleAnalyze)

		// Quotes
		r.Get("/quote/{ticker}", s.handleQuote)

		// Backtest
		r.Post("/backtest", s.handleBacktest)

		// Portfolio
		r.Get("/portfolio", s.handlePortfolio)

		// Chat
		r.Post("/chat", s.handleChat)

		// FinanceQL
		r.Post("/query", s.handleQuery)
		r.Post("/query/explain", s.handleQueryExplain)
		r.Post("/query/nl", s.handleQueryNL)

		// Alerts
		r.Get("/alerts", s.handleAlerts)

		// WebSocket
		r.Get("/ws", s.handleWebSocket)
	})

	return r
}

// ============================================================
// Request / Response types
// ============================================================

// APIResponse is the standard JSON envelope.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// AnalyzeRequest is the body for POST /api/v1/analyze.
type AnalyzeRequest struct {
	Ticker string `json:"ticker"`
	Deep   bool   `json:"deep,omitempty"`
}

// BacktestRequest is the body for POST /api/v1/backtest.
type BacktestRequest struct {
	Strategy string  `json:"strategy"`
	Ticker   string  `json:"ticker"`
	From     string  `json:"from"`                   // YYYY-MM-DD
	To       string  `json:"to,omitempty"`            // YYYY-MM-DD, default today
	Capital  float64 `json:"capital,omitempty"`
}

// ChatRequest is the body for POST /api/v1/chat.
type ChatRequest struct {
	Message string        `json:"message"`
	Deep    bool          `json:"deep,omitempty"`
	History []ChatMessage `json:"history,omitempty"`
}

// ChatMessage represents a single chat message in history.
type ChatMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

// QueryRequest is the body for POST /api/v1/query.
type QueryRequest struct {
	Expression string `json:"expression"`
}

// QueryNLRequest is the body for POST /api/v1/query/nl.
type QueryNLRequest struct {
	Query string `json:"query"`
}

// QueryExplainResponse describes a parsed FinanceQL expression.
type QueryExplainResponse struct {
	Expression string `json:"expression"`
	AST        string `json:"ast"`
	Valid      bool   `json:"valid"`
	Error      string `json:"error,omitempty"`
}

// QueryResult represents a FinanceQL evaluation result.
type QueryResult struct {
	Type   string      `json:"type"`
	Value  interface{} `json:"value"`
}

// AlertInfo represents an active alert.
type AlertInfo struct {
	ID         string `json:"id"`
	Expression string `json:"expression"`
	Status     string `json:"status"`
}

// ============================================================
// Handlers
// ============================================================

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":        "ok",
			"version":       "dev",
			"market_status": utils.MarketStatus(),
			"time_ist":      utils.FormatDateTimeIST(utils.NowIST()),
		},
	})
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Ticker == "" {
		writeError(w, http.StatusBadRequest, "ticker is required")
		return
	}

	ticker := utils.NormalizeTicker(req.Ticker)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	var result *agent.AgentResult
	var err error
	if req.Deep {
		result, err = s.orch.FullAnalysis(ctx, ticker)
	} else {
		result, err = s.orch.QuickQuery(ctx, fmt.Sprintf("Analyze %s stock", ticker))
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Broadcast to WebSocket clients
	s.wsHub.Broadcast(WSMessage{
		Type: "analysis_complete",
		Data: map[string]interface{}{
			"ticker": ticker,
			"agent":  result.AgentName,
		},
	})

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result,
	})
}

func (s *Server) handleQuote(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	if ticker == "" {
		writeError(w, http.StatusBadRequest, "ticker is required")
		return
	}

	ticker = utils.NormalizeTicker(ticker)
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	quote, err := s.agg.YFinance().GetQuote(ctx, ticker)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    quote,
	})
}

func (s *Server) handleBacktest(w http.ResponseWriter, r *http.Request) {
	var req BacktestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Strategy == "" || req.Ticker == "" {
		writeError(w, http.StatusBadRequest, "strategy and ticker are required")
		return
	}

	ticker := utils.NormalizeTicker(req.Ticker)

	from, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from date; use YYYY-MM-DD")
		return
	}
	to := time.Now()
	if req.To != "" {
		to, err = time.Parse("2006-01-02", req.To)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid to date; use YYYY-MM-DD")
			return
		}
	}

	// Find strategy
	strategy := findStrategy(req.Strategy)
	if strategy == nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown strategy: %s", req.Strategy))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	bars, err := s.agg.FetchHistoricalData(ctx, ticker, from, to, models.Timeframe1Day)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to fetch data: %v", err))
		return
	}

	if len(bars) < 50 {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("insufficient data: %d bars", len(bars)))
		return
	}

	btCfg := backtest.DefaultConfig()
	if req.Capital > 0 {
		btCfg.InitialCapital = req.Capital
	} else if s.cfg.Trading.InitialCapital > 0 {
		btCfg.InitialCapital = s.cfg.Trading.InitialCapital
	}

	engine := backtest.NewEngine(btCfg)
	result, err := engine.Run(strategy, ticker, bars)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result,
	})
}

func (s *Server) handlePortfolio(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	margins, err := s.broker.GetMargins(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	positions, err := s.broker.GetPositions(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	holdings, err := s.broker.GetHoldings(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	orders, err := s.broker.GetOrders(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"margins":   margins,
			"positions": positions,
			"holdings":  holdings,
			"orders":    orders,
		},
	})
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	// Convert history
	var history []llm.Message
	for _, m := range req.History {
		switch m.Role {
		case "user":
			history = append(history, llm.UserMessage(m.Content))
		case "assistant":
			history = append(history, llm.AssistantMessage(m.Content))
		}
	}

	if req.Deep {
		s.orch.SetMode(agent.ModeMulti)
	} else {
		s.orch.SetMode(agent.ModeSingle)
	}

	result, err := s.orch.Chat(ctx, req.Message, history)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"agent":   result.AgentName,
			"role":    result.Role,
			"content": result.Content,
			"tokens":  result.Tokens,
		},
	})
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Expression == "" {
		writeError(w, http.StatusBadRequest, "expression is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	ec := financeql.NewEvalContext(ctx, s.agg)
	financeql.RegisterBuiltins(ec)

	val, err := financeql.EvalQuery(ec, req.Expression)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    valueToQueryResult(val),
	})
}

func (s *Server) handleQueryExplain(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Expression == "" {
		writeError(w, http.StatusBadRequest, "expression is required")
		return
	}

	node, err := financeql.ParseQuery(req.Expression)

	resp := QueryExplainResponse{
		Expression: req.Expression,
	}
	if err != nil {
		resp.Valid = false
		resp.Error = err.Error()
	} else {
		resp.Valid = true
		resp.AST = fmt.Sprintf("%v", node)
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (s *Server) handleQueryNL(w http.ResponseWriter, r *http.Request) {
	var req QueryNLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Translate NL to FinanceQL via LLM
	prompt := fmt.Sprintf("Translate this natural language query to a FinanceQL expression. "+
		"Only return the FinanceQL expression, nothing else: %s", req.Query)
	result, err := s.orch.QuickQuery(ctx, prompt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("translation failed: %v", err))
		return
	}

	fqlExpr := strings.TrimSpace(result.Content)

	// Execute the translated expression
	ec := financeql.NewEvalContext(ctx, s.agg)
	financeql.RegisterBuiltins(ec)
	val, err := financeql.EvalQuery(ec, fqlExpr)
	if err != nil {
		writeJSON(w, http.StatusOK, APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"original_query": req.Query,
				"translated":     fqlExpr,
				"error":          err.Error(),
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"original_query": req.Query,
			"translated":     fqlExpr,
			"result":         valueToQueryResult(val),
		},
	})
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	// Placeholder: alerts are managed by FinanceQL REPL in-memory
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    []AlertInfo{},
	})
}

// ============================================================
// Helpers
// ============================================================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, APIResponse{
		Success: false,
		Error:   msg,
	})
}

func valueToQueryResult(val financeql.Value) QueryResult {
	switch val.Type {
	case financeql.TypeScalar:
		return QueryResult{Type: "scalar", Value: val.Scalar}
	case financeql.TypeString:
		return QueryResult{Type: "string", Value: val.Str}
	case financeql.TypeBool:
		return QueryResult{Type: "bool", Value: val.Bool}
	case financeql.TypeVector:
		return QueryResult{Type: "vector", Value: val.Vector}
	case financeql.TypeMatrix:
		return QueryResult{Type: "matrix", Value: val.Matrix}
	case financeql.TypeTable:
		return QueryResult{Type: "table", Value: val.Table}
	default:
		return QueryResult{Type: "nil", Value: nil}
	}
}

func findStrategy(name string) backtest.Strategy {
	name = strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	for _, s := range backtest.BuiltinStrategies() {
		sName := strings.ToLower(strings.ReplaceAll(s.Name(), " ", "_"))
		if sName == name || strings.Contains(sName, name) {
			return s
		}
	}
	return nil
}

// ============================================================
// WebSocket Hub
// ============================================================

// WSMessage is a message sent over WebSocket connections.
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// WSHub manages WebSocket connections and message broadcasting.
type WSHub struct {
	mu      sync.RWMutex
	clients map[*WSClient]bool
	broadcast chan WSMessage
	register  chan *WSClient
	unregister chan *WSClient
}

// WSClient represents a single WebSocket connection.
type WSClient struct {
	hub  *WSHub
	send chan WSMessage
}

// NewWSHub creates a new WebSocket hub.
func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan WSMessage, 256),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

// Run starts the hub event loop.
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// Slow client; disconnect
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					close(client.send)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected WebSocket clients.
func (h *WSHub) Broadcast(msg WSMessage) {
	select {
	case h.broadcast <- msg:
	default:
		// Drop message if broadcast channel is full
	}
}

// ClientCount returns the number of connected WebSocket clients.
func (h *WSHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Register adds a client to the hub.
func (h *WSHub) Register(client *WSClient) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *WSHub) Unregister(client *WSClient) {
	h.unregister <- client
}
