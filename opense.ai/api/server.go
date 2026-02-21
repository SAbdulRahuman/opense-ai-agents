// Package api provides the HTTP REST API server for OpeNSE.ai.
//
// It exposes endpoints for stock analysis, quotes, backtesting,
// portfolio management, chat, FinanceQL queries, and WebSocket streaming.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
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
	"github.com/seenimoa/openseai/web"
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
	serveUI  bool // when true, serve the embedded web UI at /
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
		serveUI: true, // serve embedded web UI by default
	}

	srv.router = srv.buildRouter()
	return srv, nil
}

// SetServeUI controls whether the embedded web UI is served.
// Must be called before ListenAndServe.
func (s *Server) SetServeUI(enabled bool) {
	s.serveUI = enabled
	s.router = s.buildRouter()
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
		// Health (also available at /health)
		r.Get("/health", s.handleHealth)

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
		r.Post("/alerts", s.handleCreateAlert)
		r.Delete("/alerts/{id}", s.handleDeleteAlert)

		// Orders
		r.Get("/orders", s.handleGetOrders)
		r.Get("/orders/{id}", s.handleGetOrderByID)
		r.Post("/orders", s.handlePlaceOrder)
		r.Put("/orders/{id}", s.handleModifyOrder)
		r.Delete("/orders/{id}", s.handleCancelOrder)

		// Positions
		r.Get("/positions", s.handleGetPositions)

		// Funds / Margins
		r.Get("/funds", s.handleGetFunds)

		// Market data
		r.Get("/ohlcv/{ticker}", s.handleOHLCV)
		r.Get("/market/indices", s.handleMarketIndices)
		r.Get("/market/movers", s.handleTopMovers)
		r.Get("/market/fiidii", s.handleFIIDII)

		// Screener
		r.Post("/screener", s.handleScreener)

		// Ticker search
		r.Get("/search/tickers", s.handleSearchTickers)

		// Trade confirmation (HITL)
		r.Post("/trade/confirm", s.handleTradeConfirm)

		// Configuration
		r.Get("/config", s.handleGetConfig)
		r.Put("/config", s.handleUpdateConfig)
		r.Get("/config/keys", s.handleGetConfigKeys)

		// WebSocket (unified + channel sub-paths)
		r.Get("/ws", s.handleWebSocket)
		r.Get("/ws/market", s.handleWebSocket)
		r.Get("/ws/chat", s.handleWebSocket)
		r.Get("/ws/alerts", s.handleWebSocket)
	})

	// Serve embedded web UI (SPA with fallback to index.html)
	if s.serveUI {
		s.mountSPA(r, web.DistFS())
	}

	return r
}

// mountSPA serves the embedded Next.js static export as a single-page app.
// Static assets (_next/*, favicon.ico, etc.) are served directly with caching.
// All other paths fall back to index.html for client-side routing.
func (s *Server) mountSPA(r chi.Router, distFS fs.FS) {
	fileServer := http.FileServerFS(distFS)

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		rPath := strings.TrimPrefix(r.URL.Path, "/")
		if rPath == "" {
			rPath = "index.html"
		}

		// Try to open the requested file from the embedded FS
		f, err := distFS.Open(rPath)
		if err != nil {
			// File not found — serve index.html for SPA client-side routing
			serveIndexHTML(w, r, distFS)
			return
		}
		f.Close()

		// Set cache headers for immutable assets (_next/static/*)
		if strings.HasPrefix(rPath, "_next/static/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if rPath == "index.html" || strings.HasSuffix(rPath, ".html") {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}

		fileServer.ServeHTTP(w, r)
	})
}

// serveIndexHTML reads and serves the embedded index.html for SPA fallback.
func serveIndexHTML(w http.ResponseWriter, r *http.Request, distFS fs.FS) {
	data, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		http.Error(w, "web UI not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write(data) //nolint:errcheck
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

// MoverEntry represents a top mover stock.
type MoverEntry struct {
	Ticker        string  `json:"ticker"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	ChangePercent float64 `json:"changePercent"`
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

// CreateAlertRequest is the body for POST /api/v1/alerts.
type CreateAlertRequest struct {
	Expression string `json:"expression"`
}

func (s *Server) handleCreateAlert(w http.ResponseWriter, r *http.Request) {
	var req CreateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Expression == "" {
		writeError(w, http.StatusBadRequest, "expression is required")
		return
	}
	// TODO: persist alerts; for now return a stub
	alert := AlertInfo{
		ID:         fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		Expression: req.Expression,
		Status:     "pending",
	}
	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    alert,
	})
}

func (s *Server) handleDeleteAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "alert id is required")
		return
	}
	// TODO: actually remove the alert from storage
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"deleted": id},
	})
}

// ============================================================
// Order handlers
// ============================================================

func (s *Server) handleGetOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	orders, err := s.broker.GetOrders(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    orders,
	})
}

func (s *Server) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "order id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	order, err := s.broker.GetOrderByID(ctx, orderID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    order,
	})
}

func (s *Server) handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req models.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Ticker == "" {
		writeError(w, http.StatusBadRequest, "ticker is required")
		return
	}
	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "quantity must be positive")
		return
	}

	req.Ticker = utils.NormalizeTicker(req.Ticker)

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	resp, err := s.broker.PlaceOrder(ctx, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Broadcast order event via WebSocket
	s.wsHub.Broadcast(WSMessage{
		Type: "order_placed",
		Data: map[string]interface{}{
			"order_id": resp.OrderID,
			"ticker":   req.Ticker,
			"side":     req.Side,
			"status":   resp.Status,
		},
	})

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (s *Server) handleModifyOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "order id is required")
		return
	}

	var req models.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	resp, err := s.broker.ModifyOrder(ctx, orderID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "order id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := s.broker.CancelOrder(ctx, orderID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"cancelled": orderID},
	})
}

// ============================================================
// Position & Funds handlers
// ============================================================

func (s *Server) handleGetPositions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	positions, err := s.broker.GetPositions(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    positions,
	})
}

func (s *Server) handleGetFunds(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	margins, err := s.broker.GetMargins(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    margins,
	})
}

// ============================================================
// Market data handlers
// ============================================================

func (s *Server) handleOHLCV(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	if ticker == "" {
		writeError(w, http.StatusBadRequest, "ticker is required")
		return
	}

	ticker = utils.NormalizeTicker(ticker)

	// Parse query params
	tfStr := r.URL.Query().Get("timeframe")
	if tfStr == "" {
		tfStr = "1D"
	}

	daysStr := r.URL.Query().Get("days")
	days := 365
	if daysStr != "" {
		if d, err := fmt.Sscanf(daysStr, "%d", &days); d < 1 || err != nil {
			days = 365
		}
	}

	tf := models.Timeframe1Day
	switch strings.ToUpper(tfStr) {
	case "1M", "1MIN":
		tf = models.Timeframe1Min
	case "5M", "5MIN":
		tf = models.Timeframe5Min
	case "15M", "15MIN":
		tf = models.Timeframe15Min
	case "1H", "1HOUR":
		tf = models.Timeframe1Hour
	case "1D", "1DAY", "DAILY":
		tf = models.Timeframe1Day
	case "1W", "1WEEK", "WEEKLY":
		tf = models.Timeframe1Week
	}

	to := time.Now()
	from := to.AddDate(0, 0, -days)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	candles, err := s.agg.FetchHistoricalData(ctx, ticker, from, to, tf)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    candles,
	})
}

func (s *Server) handleMarketIndices(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	overview, err := s.agg.FetchMarketOverview(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type IndexData struct {
		Name          string  `json:"name"`
		Value         float64 `json:"value"`
		Change        float64 `json:"change"`
		ChangePercent float64 `json:"changePercent"`
	}

	var indices []IndexData
	if overview.Nifty50 != nil {
		indices = append(indices, IndexData{
			Name:          "NIFTY 50",
			Value:         overview.Nifty50.LastPrice,
			Change:        overview.Nifty50.Change,
			ChangePercent: overview.Nifty50.ChangePct,
		})
	}
	if overview.BankNifty != nil {
		indices = append(indices, IndexData{
			Name:          "BANK NIFTY",
			Value:         overview.BankNifty.LastPrice,
			Change:        overview.BankNifty.Change,
			ChangePercent: overview.BankNifty.ChangePct,
		})
	}
	if overview.IndiaVIX != nil {
		indices = append(indices, IndexData{
			Name:          "INDIA VIX",
			Value:         overview.IndiaVIX.Value,
			Change:        overview.IndiaVIX.Change,
			ChangePercent: overview.IndiaVIX.ChangePct,
		})
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    indices,
	})
}

func (s *Server) handleTopMovers(w http.ResponseWriter, r *http.Request) {
	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "gainers"
	}

	// Use a set of popular tickers to compute top movers
	tickers := []string{"RELIANCE", "TCS", "INFY", "HDFCBANK", "ICICIBANK",
		"HINDUNILVR", "BHARTIARTL", "ITC", "SBIN", "BAJFINANCE",
		"LT", "KOTAKBANK", "AXISBANK", "ASIANPAINT", "MARUTI"}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var movers []MoverEntry
	results := make(chan MoverEntry, len(tickers))
	var wg sync.WaitGroup

	for _, t := range tickers {
		wg.Add(1)
		go func(ticker string) {
			defer wg.Done()
			q, err := s.agg.YFinance().GetQuote(ctx, ticker)
			if err != nil || q == nil {
				return
			}
			results <- MoverEntry{
				Ticker:        q.Ticker,
				Name:          q.Name,
				Price:         q.LastPrice,
				ChangePercent: q.ChangePct,
			}
		}(t)
	}

	wg.Wait()
	close(results)

	for m := range results {
		movers = append(movers, m)
	}

	// Sort by changePercent
	if direction == "gainers" {
		sortMovers(movers, false) // descending
	} else {
		sortMovers(movers, true) // ascending
	}

	// Return top 10
	if len(movers) > 10 {
		movers = movers[:10]
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    movers,
	})
}

func (s *Server) handleFIIDII(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	data, err := s.agg.FIIDII().GetFIIDIIActivity(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// ============================================================
// Screener & Search handlers
// ============================================================

// ScreenerRequest is the body for POST /api/v1/screener.
type ScreenerRequest struct {
	Query string `json:"query"`
}

func (s *Server) handleScreener(w http.ResponseWriter, r *http.Request) {
	var req ScreenerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// TODO: implement real screener logic based on query filters
	// For now, return empty results
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    []interface{}{},
	})
}

func (s *Server) handleSearchTickers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, APIResponse{
			Success: true,
			Data:    []interface{}{},
		})
		return
	}

	q = strings.ToUpper(q)

	// Use a static list of well-known NSE tickers for autocomplete
	knownTickers := map[string]string{
		"RELIANCE":    "Reliance Industries Ltd",
		"TCS":         "Tata Consultancy Services Ltd",
		"INFY":        "Infosys Ltd",
		"HDFCBANK":    "HDFC Bank Ltd",
		"ICICIBANK":   "ICICI Bank Ltd",
		"HINDUNILVR":  "Hindustan Unilever Ltd",
		"BHARTIARTL":  "Bharti Airtel Ltd",
		"ITC":         "ITC Ltd",
		"SBIN":        "State Bank of India",
		"BAJFINANCE":  "Bajaj Finance Ltd",
		"LT":          "Larsen & Toubro Ltd",
		"KOTAKBANK":   "Kotak Mahindra Bank Ltd",
		"AXISBANK":    "Axis Bank Ltd",
		"ASIANPAINT":  "Asian Paints Ltd",
		"MARUTI":      "Maruti Suzuki India Ltd",
		"WIPRO":       "Wipro Ltd",
		"TITAN":       "Titan Company Ltd",
		"HCLTECH":     "HCL Technologies Ltd",
		"SUNPHARMA":   "Sun Pharmaceutical Industries Ltd",
		"TATAMOTORS":  "Tata Motors Ltd",
		"ONGC":        "Oil and Natural Gas Corporation Ltd",
		"NTPC":        "NTPC Ltd",
		"POWERGRID":   "Power Grid Corporation of India Ltd",
		"ULTRACEMCO":  "UltraTech Cement Ltd",
		"TECHM":       "Tech Mahindra Ltd",
		"TATASTEEL":   "Tata Steel Ltd",
		"BAJAJFINSV":  "Bajaj Finserv Ltd",
		"NESTLEIND":   "Nestle India Ltd",
		"INDUSINDBK":  "IndusInd Bank Ltd",
		"HDFCLIFE":    "HDFC Life Insurance Co Ltd",
	}

	type TickerResult struct {
		Ticker string `json:"ticker"`
		Name   string `json:"name"`
	}

	var results []TickerResult
	for ticker, name := range knownTickers {
		if strings.Contains(ticker, q) || strings.Contains(strings.ToUpper(name), q) {
			results = append(results, TickerResult{Ticker: ticker, Name: name})
			if len(results) >= 10 {
				break
			}
		}
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    results,
	})
}

// ============================================================
// Trade confirmation handler (HITL)
// ============================================================

// TradeConfirmRequest is the body for POST /api/v1/trade/confirm.
type TradeConfirmRequest struct {
	ProposalID    string                 `json:"proposalId"`
	Action        string                 `json:"action"` // "approve", "reject", "modify"
	Modifications map[string]interface{} `json:"modifications,omitempty"`
}

func (s *Server) handleTradeConfirm(w http.ResponseWriter, r *http.Request) {
	var req TradeConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ProposalID == "" || req.Action == "" {
		writeError(w, http.StatusBadRequest, "proposalId and action are required")
		return
	}

	// TODO: locate the pending trade proposal and apply the action
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]string{
			"proposal_id": req.ProposalID,
			"status":      req.Action + "d",
		},
	})
}

// ============================================================
// Helpers
// ============================================================

// sortMovers sorts MoverEntry slice by changePercent.
// If ascending is true, sort lowest first (losers); otherwise highest first (gainers).
func sortMovers(movers []MoverEntry, ascending bool) {
	sort.Slice(movers, func(i, j int) bool {
		if ascending {
			return movers[i].ChangePercent < movers[j].ChangePercent
		}
		return movers[i].ChangePercent > movers[j].ChangePercent
	})
}

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
