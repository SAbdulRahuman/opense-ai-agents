package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/seenimoa/openseai/internal/backtest"
	"github.com/seenimoa/openseai/internal/broker"
	"github.com/seenimoa/openseai/internal/config"
	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/internal/financeql"
	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Test Helpers
// ════════════════════════════════════════════════════════════════════

func testServer(t *testing.T) *Server {
	t.Helper()
	// Build a minimal server without real LLM/broker setup — wire
	// only what we can construct without external dependencies.
	srv := &Server{
		cfg:   &config.Config{},
		wsHub: NewWSHub(),
	}
	go srv.wsHub.Run()

	return srv
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder) APIResponse {
	t.Helper()
	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

// ════════════════════════════════════════════════════════════════════
// APIResponse type tests
// ════════════════════════════════════════════════════════════════════

func TestAPIResponseJSON(t *testing.T) {
	tests := []struct {
		name string
		resp APIResponse
	}{
		{
			name: "success with data",
			resp: APIResponse{Success: true, Data: map[string]string{"key": "value"}},
		},
		{
			name: "error",
			resp: APIResponse{Success: false, Error: "something went wrong"},
		},
		{
			name: "success with nil data",
			resp: APIResponse{Success: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var got APIResponse
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if got.Success != tt.resp.Success {
				t.Errorf("Success: got %v, want %v", got.Success, tt.resp.Success)
			}
			if got.Error != tt.resp.Error {
				t.Errorf("Error: got %q, want %q", got.Error, tt.resp.Error)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════════
// Request type tests
// ════════════════════════════════════════════════════════════════════

func TestAnalyzeRequestJSON(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		ticker string
		deep   bool
	}{
		{"basic", `{"ticker":"RELIANCE"}`, "RELIANCE", false},
		{"deep", `{"ticker":"TCS","deep":true}`, "TCS", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req AnalyzeRequest
			if err := json.Unmarshal([]byte(tt.json), &req); err != nil {
				t.Fatal(err)
			}
			if req.Ticker != tt.ticker {
				t.Errorf("Ticker: got %q, want %q", req.Ticker, tt.ticker)
			}
			if req.Deep != tt.deep {
				t.Errorf("Deep: got %v, want %v", req.Deep, tt.deep)
			}
		})
	}
}

func TestBacktestRequestJSON(t *testing.T) {
	body := `{"strategy":"sma_crossover","ticker":"INFY","from":"2023-01-01","to":"2024-01-01","capital":500000}`
	var req BacktestRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatal(err)
	}
	if req.Strategy != "sma_crossover" {
		t.Errorf("Strategy: got %q", req.Strategy)
	}
	if req.Ticker != "INFY" {
		t.Errorf("Ticker: got %q", req.Ticker)
	}
	if req.From != "2023-01-01" {
		t.Errorf("From: got %q", req.From)
	}
	if req.To != "2024-01-01" {
		t.Errorf("To: got %q", req.To)
	}
	if req.Capital != 500000 {
		t.Errorf("Capital: got %f", req.Capital)
	}
}

func TestChatRequestJSON(t *testing.T) {
	body := `{"message":"analyze TCS","deep":false,"history":[{"role":"user","content":"hello"},{"role":"assistant","content":"hi"}]}`
	var req ChatRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatal(err)
	}
	if req.Message != "analyze TCS" {
		t.Errorf("Message: got %q", req.Message)
	}
	if req.Deep {
		t.Error("Deep should be false")
	}
	if len(req.History) != 2 {
		t.Fatalf("History: got %d items", len(req.History))
	}
	if req.History[0].Role != "user" || req.History[0].Content != "hello" {
		t.Errorf("History[0]: %+v", req.History[0])
	}
	if req.History[1].Role != "assistant" || req.History[1].Content != "hi" {
		t.Errorf("History[1]: %+v", req.History[1])
	}
}

func TestQueryRequestJSON(t *testing.T) {
	body := `{"expression":"sma(RELIANCE, 20)"}`
	var req QueryRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatal(err)
	}
	if req.Expression != "sma(RELIANCE, 20)" {
		t.Errorf("Expression: got %q", req.Expression)
	}
}

func TestQueryNLRequestJSON(t *testing.T) {
	body := `{"query":"What is the 20-day moving average of RELIANCE?"}`
	var req QueryNLRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatal(err)
	}
	if req.Query != "What is the 20-day moving average of RELIANCE?" {
		t.Errorf("Query: got %q", req.Query)
	}
}

// ════════════════════════════════════════════════════════════════════
// Health handler tests
// ════════════════════════════════════════════════════════════════════

func TestHandleHealth(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	srv.handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("data should be a map")
	}
	if data["status"] != "ok" {
		t.Errorf("status: got %q", data["status"])
	}
	if _, ok := data["market_status"]; !ok {
		t.Error("missing market_status")
	}
	if _, ok := data["time_ist"]; !ok {
		t.Error("missing time_ist")
	}
	if _, ok := data["version"]; !ok {
		t.Error("missing version")
	}
}

// ════════════════════════════════════════════════════════════════════
// Analyze handler tests (validation only — no real LLM)
// ════════════════════════════════════════════════════════════════════

func TestHandleAnalyze_InvalidJSON(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/analyze", strings.NewReader("{invalid"))
	srv.handleAnalyze(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rec)
	if resp.Success {
		t.Error("expected success=false for invalid JSON")
	}
	if resp.Error == "" {
		t.Error("expected non-empty error")
	}
}

func TestHandleAnalyze_MissingTicker(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	body := `{"deep":true}`
	req := httptest.NewRequest("POST", "/api/v1/analyze", strings.NewReader(body))
	srv.handleAnalyze(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rec)
	if resp.Success {
		t.Error("expected success=false")
	}
	if !strings.Contains(resp.Error, "ticker") {
		t.Errorf("error should mention 'ticker': %q", resp.Error)
	}
}

// ════════════════════════════════════════════════════════════════════
// Backtest handler tests (validation only — no data fetch)
// ════════════════════════════════════════════════════════════════════

func TestHandleBacktest_InvalidJSON(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/backtest", strings.NewReader("not json"))
	srv.handleBacktest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleBacktest_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"missing both", `{}`, "strategy and ticker"},
		{"missing ticker", `{"strategy":"sma_crossover"}`, "strategy and ticker"},
		{"missing strategy", `{"ticker":"RELIANCE"}`, "strategy and ticker"},
	}

	srv := testServer(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/backtest", strings.NewReader(tt.body))
			srv.handleBacktest(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
			}

			resp := decodeResponse(t, rec)
			if resp.Success {
				t.Error("expected success=false")
			}
			if !strings.Contains(strings.ToLower(resp.Error), tt.want) {
				t.Errorf("error %q should contain %q", resp.Error, tt.want)
			}
		})
	}
}

func TestHandleBacktest_InvalidFromDate(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	body := `{"strategy":"sma_crossover","ticker":"RELIANCE","from":"invalid-date"}`
	req := httptest.NewRequest("POST", "/api/v1/backtest", strings.NewReader(body))
	srv.handleBacktest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rec)
	if !strings.Contains(resp.Error, "from date") {
		t.Errorf("error should mention from date: %q", resp.Error)
	}
}

func TestHandleBacktest_InvalidToDate(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	body := `{"strategy":"sma_crossover","ticker":"RELIANCE","from":"2023-01-01","to":"bad"}`
	req := httptest.NewRequest("POST", "/api/v1/backtest", strings.NewReader(body))
	srv.handleBacktest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rec)
	if !strings.Contains(resp.Error, "to date") {
		t.Errorf("error should mention to date: %q", resp.Error)
	}
}

func TestHandleBacktest_UnknownStrategy(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	body := `{"strategy":"nonexistent_strategy","ticker":"RELIANCE","from":"2023-01-01"}`
	req := httptest.NewRequest("POST", "/api/v1/backtest", strings.NewReader(body))
	srv.handleBacktest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rec)
	if !strings.Contains(resp.Error, "unknown strategy") {
		t.Errorf("error should mention unknown strategy: %q", resp.Error)
	}
}

// ════════════════════════════════════════════════════════════════════
// Chat handler tests (validation only)
// ════════════════════════════════════════════════════════════════════

func TestHandleChat_InvalidJSON(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat", strings.NewReader("{bad"))
	srv.handleChat(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleChat_MissingMessage(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	body := `{"deep":true}`
	req := httptest.NewRequest("POST", "/api/v1/chat", strings.NewReader(body))
	srv.handleChat(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rec)
	if !strings.Contains(resp.Error, "message") {
		t.Errorf("error should mention 'message': %q", resp.Error)
	}
}

// ════════════════════════════════════════════════════════════════════
// Query handler tests (validation + expression eval)
// ════════════════════════════════════════════════════════════════════

func TestHandleQuery_InvalidJSON(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/query", strings.NewReader("bad"))
	srv.handleQuery(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleQuery_EmptyExpression(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	body := `{"expression":""}`
	req := httptest.NewRequest("POST", "/api/v1/query", strings.NewReader(body))
	srv.handleQuery(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rec)
	if !strings.Contains(resp.Error, "expression") {
		t.Errorf("error should mention 'expression': %q", resp.Error)
	}
}

func TestHandleQuery_ValidArithmeticExpression(t *testing.T) {
	srv := testServer(t)
	// Wire an aggregator so the handler has what it needs
	// For pure arithmetic, no datasource is needed
	srv.agg = nil // EvalContext with nil aggregator is fine for pure arithmetic

	rec := httptest.NewRecorder()
	body := `{"expression":"2 + 3 * 4"}`
	req := httptest.NewRequest("POST", "/api/v1/query", strings.NewReader(body))
	srv.handleQuery(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("data should be a map")
	}
	if data["type"] != "scalar" {
		t.Errorf("type: got %q, want %q", data["type"], "scalar")
	}
	if val, ok := data["value"].(float64); !ok || val != 14 {
		t.Errorf("value: got %v, want 14", data["value"])
	}
}

// ════════════════════════════════════════════════════════════════════
// Query explain handler tests
// ════════════════════════════════════════════════════════════════════

func TestHandleQueryExplain_InvalidJSON(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/query/explain", strings.NewReader("bad"))
	srv.handleQueryExplain(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleQueryExplain_EmptyExpression(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	body := `{"expression":""}`
	req := httptest.NewRequest("POST", "/api/v1/query/explain", strings.NewReader(body))
	srv.handleQueryExplain(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleQueryExplain_ValidExpression(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	body := `{"expression":"1 + 2"}`
	req := httptest.NewRequest("POST", "/api/v1/query/explain", strings.NewReader(body))
	srv.handleQueryExplain(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("data should be a map")
	}
	if data["valid"] != true {
		t.Error("expected valid=true")
	}
	if data["expression"] != "1 + 2" {
		t.Errorf("expression: got %q", data["expression"])
	}
}

func TestHandleQueryExplain_InvalidExpression(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	body := `{"expression":"((("}`
	req := httptest.NewRequest("POST", "/api/v1/query/explain", strings.NewReader(body))
	srv.handleQueryExplain(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d — explain returns 200 even for invalid exprs", rec.Code, http.StatusOK)
	}

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true (envelope is always success)")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("data should be a map")
	}
	if data["valid"] != false {
		t.Error("expected valid=false for malformed expression")
	}
	if data["error"] == nil || data["error"] == "" {
		t.Error("expected non-empty error for invalid expression")
	}
}

// ════════════════════════════════════════════════════════════════════
// Query NL handler tests (validation only)
// ════════════════════════════════════════════════════════════════════

func TestHandleQueryNL_InvalidJSON(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/query/nl", strings.NewReader("bad"))
	srv.handleQueryNL(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleQueryNL_EmptyQuery(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	body := `{"query":""}`
	req := httptest.NewRequest("POST", "/api/v1/query/nl", strings.NewReader(body))
	srv.handleQueryNL(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rec)
	if !strings.Contains(resp.Error, "query") {
		t.Errorf("error should mention 'query': %q", resp.Error)
	}
}

// ════════════════════════════════════════════════════════════════════
// Alerts handler tests
// ════════════════════════════════════════════════════════════════════

func TestHandleAlerts(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/alerts", nil)
	srv.handleAlerts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}

	// Data should be empty array
	arr, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("data should be an array, got %T", resp.Data)
	}
	if len(arr) != 0 {
		t.Errorf("expected empty alerts, got %d", len(arr))
	}
}

// ════════════════════════════════════════════════════════════════════
// Quote handler tests (validation only)
// ════════════════════════════════════════════════════════════════════

func TestHandleQuote_EmptyTicker(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	// Without URL params set via chi context, ticker will be empty
	req := httptest.NewRequest("GET", "/api/v1/quote/", nil)
	srv.handleQuote(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// ════════════════════════════════════════════════════════════════════
// Portfolio handler tests (with paper broker)
// ════════════════════════════════════════════════════════════════════

func TestHandlePortfolio(t *testing.T) {
	srv := testServer(t)
	// Wire in a paper broker
	srv.broker = newTestBroker()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/portfolio", nil)
	srv.handlePortfolio(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Errorf("expected success=true, error: %s", resp.Error)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("data should be a map")
	}

	// Paper broker should return empty arrays for positions/holdings/orders and valid margins
	for _, key := range []string{"margins", "positions", "holdings", "orders"} {
		if _, ok := data[key]; !ok {
			t.Errorf("missing key %q in portfolio data", key)
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// writeJSON / writeError tests
// ════════════════════════════════════════════════════════════════════

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusCreated, APIResponse{
		Success: true,
		Data:    "hello",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q", ct)
	}

	resp := decodeResponse(t, rec)
	if !resp.Success || resp.Data != "hello" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusNotFound, "not found")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}

	resp := decodeResponse(t, rec)
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error != "not found" {
		t.Errorf("error: got %q, want %q", resp.Error, "not found")
	}
}

// ════════════════════════════════════════════════════════════════════
// valueToQueryResult tests
// ════════════════════════════════════════════════════════════════════

func TestValueToQueryResult(t *testing.T) {
	tests := []struct {
		name     string
		val      financeql.Value
		wantType string
	}{
		{"scalar", financeql.ScalarValue(3.14), "scalar"},
		{"string", financeql.StringValue("hello"), "string"},
		{"bool", financeql.Value{Type: financeql.TypeBool, Bool: true}, "bool"},
		{"vector", financeql.Value{Type: financeql.TypeVector, Vector: []financeql.TimePoint{{Time: time.Now(), Value: 1}}}, "vector"},
		{"matrix", financeql.Value{Type: financeql.TypeMatrix, Matrix: map[string][]financeql.TimePoint{"a": {}}}, "matrix"},
		{"table", financeql.Value{Type: financeql.TypeTable, Table: []map[string]interface{}{{"k": "v"}}}, "table"},
		{"nil", financeql.Value{Type: financeql.TypeNil}, "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueToQueryResult(tt.val)
			if got.Type != tt.wantType {
				t.Errorf("Type: got %q, want %q", got.Type, tt.wantType)
			}
		})
	}
}

func TestValueToQueryResult_ScalarValue(t *testing.T) {
	result := valueToQueryResult(financeql.ScalarValue(42.5))
	if result.Type != "scalar" {
		t.Fatalf("Type: got %q", result.Type)
	}
	if v, ok := result.Value.(float64); !ok || v != 42.5 {
		t.Errorf("Value: got %v", result.Value)
	}
}

func TestValueToQueryResult_StringValue(t *testing.T) {
	result := valueToQueryResult(financeql.StringValue("RELIANCE"))
	if result.Type != "string" {
		t.Fatalf("Type: got %q", result.Type)
	}
	if v, ok := result.Value.(string); !ok || v != "RELIANCE" {
		t.Errorf("Value: got %v", result.Value)
	}
}

func TestValueToQueryResult_BoolValue(t *testing.T) {
	result := valueToQueryResult(financeql.Value{Type: financeql.TypeBool, Bool: true})
	if result.Type != "bool" {
		t.Fatalf("Type: got %q", result.Type)
	}
	if v, ok := result.Value.(bool); !ok || !v {
		t.Errorf("Value: got %v", result.Value)
	}
}

// ════════════════════════════════════════════════════════════════════
// findStrategy tests
// ════════════════════════════════════════════════════════════════════

func TestFindStrategy(t *testing.T) {
	strategies := backtest.BuiltinStrategies()
	if len(strategies) == 0 {
		t.Skip("no builtin strategies registered")
	}

	// findStrategy expects underscore-separated lowercase names (CLI convention)
	first := strategies[0]
	searchName := strings.ToLower(strings.ReplaceAll(first.Name(), " ", "_"))
	found := findStrategy(searchName)
	if found == nil {
		t.Fatalf("findStrategy(%q): got nil", searchName)
	}
	if found.Name() != first.Name() {
		t.Errorf("found.Name(): got %q, want %q", found.Name(), first.Name())
	}
}

func TestFindStrategy_Unknown(t *testing.T) {
	found := findStrategy("definitely_does_not_exist_xyz")
	if found != nil {
		t.Errorf("expected nil for unknown strategy, got %q", found.Name())
	}
}

func TestFindStrategy_CaseInsensitive(t *testing.T) {
	strategies := backtest.BuiltinStrategies()
	if len(strategies) == 0 {
		t.Skip("no builtin strategies registered")
	}

	first := strategies[0]
	// Search with upper case underscore variant
	searchName := strings.ToUpper(strings.ReplaceAll(first.Name(), " ", "_"))
	found := findStrategy(searchName)
	if found == nil {
		t.Fatalf("findStrategy(upper): got nil for %q", searchName)
	}
}

// ════════════════════════════════════════════════════════════════════
// WebSocket Hub tests
// ════════════════════════════════════════════════════════════════════

func TestWSHub_NewWSHub(t *testing.T) {
	hub := NewWSHub()
	if hub == nil {
		t.Fatal("NewWSHub returned nil")
	}
	if hub.ClientCount() != 0 {
		t.Errorf("ClientCount: got %d, want 0", hub.ClientCount())
	}
}

func TestWSHub_RegisterAndUnregister(t *testing.T) {
	hub := NewWSHub()
	go hub.Run()

	// Give hub time to start
	time.Sleep(10 * time.Millisecond)

	client := &WSClient{
		hub:  hub,
		send: make(chan WSMessage, 256),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)
	if hub.ClientCount() != 1 {
		t.Errorf("after register: ClientCount=%d, want 1", hub.ClientCount())
	}

	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)
	if hub.ClientCount() != 0 {
		t.Errorf("after unregister: ClientCount=%d, want 0", hub.ClientCount())
	}
}

func TestWSHub_Broadcast(t *testing.T) {
	hub := NewWSHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client1 := &WSClient{hub: hub, send: make(chan WSMessage, 256)}
	client2 := &WSClient{hub: hub, send: make(chan WSMessage, 256)}

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(10 * time.Millisecond)

	msg := WSMessage{Type: "test", Data: "hello"}
	hub.Broadcast(msg)
	time.Sleep(10 * time.Millisecond)

	// Both clients should receive the message
	select {
	case got := <-client1.send:
		if got.Type != "test" {
			t.Errorf("client1 got type=%q, want 'test'", got.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client1 did not receive message")
	}

	select {
	case got := <-client2.send:
		if got.Type != "test" {
			t.Errorf("client2 got type=%q, want 'test'", got.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client2 did not receive message")
	}

	// Cleanup
	hub.Unregister(client1)
	hub.Unregister(client2)
}

func TestWSHub_BroadcastDropsWhenBufferFull(t *testing.T) {
	hub := NewWSHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	// Calling Broadcast with no clients and a full broadcast channel
	// should not block (message is dropped).
	done := make(chan bool)
	go func() {
		for i := 0; i < 300; i++ {
			hub.Broadcast(WSMessage{Type: "test"})
		}
		done <- true
	}()

	select {
	case <-done:
		// Good — didn't block
	case <-time.After(2 * time.Second):
		t.Fatal("Broadcast blocked when buffer was full")
	}
}

func TestWSHub_ConcurrentRegisterUnregister(t *testing.T) {
	hub := NewWSHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	numClients := 50

	clients := make([]*WSClient, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = &WSClient{hub: hub, send: make(chan WSMessage, 256)}
	}

	// Register all concurrently
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(c *WSClient) {
			defer wg.Done()
			hub.Register(c)
		}(clients[i])
	}
	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	count := hub.ClientCount()
	if count != numClients {
		t.Errorf("after all registered: ClientCount=%d, want %d", count, numClients)
	}

	// Unregister all concurrently
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(c *WSClient) {
			defer wg.Done()
			hub.Unregister(c)
		}(clients[i])
	}
	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	count = hub.ClientCount()
	if count != 0 {
		t.Errorf("after all unregistered: ClientCount=%d, want 0", count)
	}
}

func TestWSHub_MultipleMessages(t *testing.T) {
	hub := NewWSHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &WSClient{hub: hub, send: make(chan WSMessage, 256)}
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	msgs := []WSMessage{
		{Type: "type1", Data: "d1"},
		{Type: "type2", Data: "d2"},
		{Type: "type3", Data: "d3"},
	}

	for _, m := range msgs {
		hub.Broadcast(m)
	}
	time.Sleep(50 * time.Millisecond)

	received := make([]WSMessage, 0)
	for {
		select {
		case m := <-client.send:
			received = append(received, m)
		default:
			goto done
		}
	}
done:

	if len(received) != 3 {
		t.Fatalf("received %d messages, want 3", len(received))
	}
	for i, m := range received {
		expected := fmt.Sprintf("type%d", i+1)
		if m.Type != expected {
			t.Errorf("msg[%d].Type: got %q, want %q", i, m.Type, expected)
		}
	}

	hub.Unregister(client)
}

// ════════════════════════════════════════════════════════════════════
// WSMessage JSON tests
// ════════════════════════════════════════════════════════════════════

func TestWSMessageJSON(t *testing.T) {
	msg := WSMessage{
		Type: "analysis_complete",
		Data: map[string]interface{}{
			"ticker": "RELIANCE",
			"status": "done",
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	var got WSMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	if got.Type != "analysis_complete" {
		t.Errorf("Type: got %q", got.Type)
	}
}

func TestWSMessageJSON_NoData(t *testing.T) {
	msg := WSMessage{Type: "pong"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	var got WSMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "pong" {
		t.Errorf("Type: got %q", got.Type)
	}
	if got.Data != nil {
		t.Errorf("Data should be nil: %v", got.Data)
	}
}

// ════════════════════════════════════════════════════════════════════
// AlertInfo JSON tests
// ════════════════════════════════════════════════════════════════════

func TestAlertInfoJSON(t *testing.T) {
	alert := AlertInfo{
		ID:         "alert-1",
		Expression: "price(RELIANCE) > 3000",
		Status:     "active",
	}

	data, err := json.Marshal(alert)
	if err != nil {
		t.Fatal(err)
	}

	var got AlertInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	if got.ID != "alert-1" || got.Expression != "price(RELIANCE) > 3000" || got.Status != "active" {
		t.Errorf("unexpected: %+v", got)
	}
}

// ════════════════════════════════════════════════════════════════════
// QueryExplainResponse JSON tests
// ════════════════════════════════════════════════════════════════════

func TestQueryExplainResponseJSON(t *testing.T) {
	resp := QueryExplainResponse{
		Expression: "1 + 2",
		AST:        "(+ 1 2)",
		Valid:      true,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var got QueryExplainResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Expression != "1 + 2" || got.AST != "(+ 1 2)" || !got.Valid {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestQueryExplainResponseJSON_WithError(t *testing.T) {
	resp := QueryExplainResponse{
		Expression: "(((",
		Valid:      false,
		Error:      "parse error",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var got QueryExplainResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Valid || got.Error != "parse error" {
		t.Errorf("unexpected: %+v", got)
	}
}

// ════════════════════════════════════════════════════════════════════
// QueryResult JSON tests
// ════════════════════════════════════════════════════════════════════

func TestQueryResultJSON(t *testing.T) {
	tests := []struct {
		name string
		qr   QueryResult
	}{
		{"scalar", QueryResult{Type: "scalar", Value: 42.5}},
		{"string", QueryResult{Type: "string", Value: "RELIANCE"}},
		{"bool", QueryResult{Type: "bool", Value: true}},
		{"nil", QueryResult{Type: "nil", Value: nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.qr)
			if err != nil {
				t.Fatal(err)
			}

			var got QueryResult
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if got.Type != tt.qr.Type {
				t.Errorf("Type: got %q, want %q", got.Type, tt.qr.Type)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════════
// ChatMessage JSON tests
// ════════════════════════════════════════════════════════════════════

func TestChatMessageJSON(t *testing.T) {
	msg := ChatMessage{Role: "user", Content: "analyze RELIANCE"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	var got ChatMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Role != "user" || got.Content != "analyze RELIANCE" {
		t.Errorf("unexpected: %+v", got)
	}
}

// ════════════════════════════════════════════════════════════════════
// Integration-style: handler with valid FinanceQL eval
// ════════════════════════════════════════════════════════════════════

func TestHandleQuery_BooleanExpression(t *testing.T) {
	srv := testServer(t)
	srv.agg = nil

	rec := httptest.NewRecorder()
	body := `{"expression":"3 > 2"}`
	req := httptest.NewRequest("POST", "/api/v1/query", strings.NewReader(body))
	srv.handleQuery(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("data should be a map")
	}
	if data["type"] != "bool" {
		t.Errorf("type: got %q, want 'bool'", data["type"])
	}
	if data["value"] != true {
		t.Errorf("value: got %v, want true", data["value"])
	}
}

func TestHandleQuery_StringExpression(t *testing.T) {
	srv := testServer(t)
	srv.agg = nil

	rec := httptest.NewRecorder()
	body := `{"expression":"\"hello\""}`
	req := httptest.NewRequest("POST", "/api/v1/query", strings.NewReader(body))
	srv.handleQuery(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestHandleQuery_InvalidExpression(t *testing.T) {
	srv := testServer(t)
	srv.agg = nil

	rec := httptest.NewRecorder()
	body := `{"expression":"((("}`
	req := httptest.NewRequest("POST", "/api/v1/query", strings.NewReader(body))
	srv.handleQuery(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rec)
	if resp.Success {
		t.Error("expected success=false for invalid expression")
	}
}

// ════════════════════════════════════════════════════════════════════
// Content-Type header tests
// ════════════════════════════════════════════════════════════════════

func TestHealthResponse_ContentType(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	srv.handleHealth(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}

// ════════════════════════════════════════════════════════════════════
// Server struct tests
// ════════════════════════════════════════════════════════════════════

func TestWSClient_SendChannel(t *testing.T) {
	client := &WSClient{
		send: make(chan WSMessage, 10),
	}

	// Should be able to send without blocking
	client.send <- WSMessage{Type: "test"}

	msg := <-client.send
	if msg.Type != "test" {
		t.Errorf("Type: got %q", msg.Type)
	}
}

// ════════════════════════════════════════════════════════════════════
// Backtest with real strategy integration
// ════════════════════════════════════════════════════════════════════

func TestHandleBacktest_ValidStrategyWithAggregator(t *testing.T) {
	// This tests real strategy lookup with a valid strategy name.
	// With a real aggregator (no API key), the data fetch may fail but won't panic.
	strategies := backtest.BuiltinStrategies()
	if len(strategies) == 0 {
		t.Skip("no builtin strategies")
	}

	srv := testServer(t)
	srv.agg = datasource.NewAggregator()

	rec := httptest.NewRecorder()
	body := fmt.Sprintf(`{"strategy":"%s","ticker":"RELIANCE","from":"2023-01-01"}`,
		strings.ToLower(strings.ReplaceAll(strategies[0].Name(), " ", "_")))
	req := httptest.NewRequest("POST", "/api/v1/backtest", strings.NewReader(body))
	srv.handleBacktest(rec, req)

	// The request should not panic — it may fail on data fetch or succeed
	if rec.Code == 0 {
		t.Error("expected a non-zero status code")
	}
}

// ════════════════════════════════════════════════════════════════════
// Batch test: verifying all error responses are valid JSON
// ════════════════════════════════════════════════════════════════════

func TestErrorResponsesAreValidJSON(t *testing.T) {
	srv := testServer(t)

	scenarios := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"analyze_invalid", "POST", "/api/v1/analyze", "{bad"},
		{"backtest_invalid", "POST", "/api/v1/backtest", "{bad"},
		{"chat_invalid", "POST", "/api/v1/chat", "{bad"},
		{"query_invalid", "POST", "/api/v1/query", "{bad"},
		{"explain_invalid", "POST", "/api/v1/query/explain", "{bad"},
		{"nl_invalid", "POST", "/api/v1/query/nl", "{bad"},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			var bodyReader *strings.Reader
			if sc.body != "" {
				bodyReader = strings.NewReader(sc.body)
			}

			var handler func(http.ResponseWriter, *http.Request)
			switch {
			case strings.Contains(sc.path, "analyze"):
				handler = srv.handleAnalyze
			case strings.Contains(sc.path, "backtest"):
				handler = srv.handleBacktest
			case strings.Contains(sc.path, "chat"):
				handler = srv.handleChat
			case strings.Contains(sc.path, "query/explain"):
				handler = srv.handleQueryExplain
			case strings.Contains(sc.path, "query/nl"):
				handler = srv.handleQueryNL
			case strings.Contains(sc.path, "query"):
				handler = srv.handleQuery
			}

			req := httptest.NewRequest(sc.method, sc.path, bodyReader)
			handler(rec, req)

			// Verify response is valid JSON
			var resp APIResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("response for %s is not valid JSON: %v\nbody: %s", sc.path, err, rec.Body.String())
			}

			if resp.Success {
				t.Errorf("expected success=false for invalid JSON input at %s", sc.path)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════════
// testBroker — a minimal broker mock for portfolio tests
// ════════════════════════════════════════════════════════════════════

// mockBroker implements broker.Broker for testing.
type mockBroker struct {
	name string
}

var _ broker.Broker = (*mockBroker)(nil)

func newTestBroker() *mockBroker { return &mockBroker{name: "test"} }

func (b *mockBroker) Name() string { return b.name }

func (b *mockBroker) GetMargins(ctx context.Context) (*models.Margins, error) {
	return &models.Margins{
		AvailableCash:   1000000,
		UsedMargin:      0,
		AvailableMargin: 1000000,
	}, nil
}

func (b *mockBroker) GetPositions(ctx context.Context) ([]models.Position, error) {
	return []models.Position{}, nil
}

func (b *mockBroker) GetHoldings(ctx context.Context) ([]models.Holding, error) {
	return []models.Holding{}, nil
}

func (b *mockBroker) GetOrders(ctx context.Context) ([]models.Order, error) {
	return []models.Order{}, nil
}

func (b *mockBroker) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	return nil, fmt.Errorf("order not found: %s", orderID)
}

func (b *mockBroker) PlaceOrder(ctx context.Context, req models.OrderRequest) (*models.OrderResponse, error) {
	return &models.OrderResponse{OrderID: "test-001", Status: "placed"}, nil
}

func (b *mockBroker) ModifyOrder(ctx context.Context, orderID string, req models.OrderRequest) (*models.OrderResponse, error) {
	return &models.OrderResponse{OrderID: orderID, Status: "modified"}, nil
}

func (b *mockBroker) CancelOrder(ctx context.Context, orderID string) error { return nil }

func (b *mockBroker) SubscribeQuotes(ctx context.Context, tickers []string) (<-chan models.Quote, error) {
	ch := make(chan models.Quote)
	close(ch)
	return ch, nil
}

// ════════════════════════════════════════════════════════════════════
// Portfolio handler with mock broker
// ════════════════════════════════════════════════════════════════════

func TestHandlePortfolio_WithMockBroker(t *testing.T) {
	srv := testServer(t)
	srv.broker = newTestBroker()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/portfolio", nil)
	srv.handlePortfolio(rec, req)

	if rec.Code != http.StatusOK {
		body := rec.Body.String()
		t.Fatalf("status: got %d, want %d\nbody: %s", rec.Code, http.StatusOK, body)
	}

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Errorf("expected success=true, error: %s", resp.Error)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("data should be a map")
	}

	for _, key := range []string{"margins", "positions", "holdings", "orders"} {
		if _, ok := data[key]; !ok {
			t.Errorf("missing key %q in portfolio data", key)
		}
	}
}

func TestHandlePortfolio_MarginsData(t *testing.T) {
	srv := testServer(t)
	srv.broker = newTestBroker()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/portfolio", nil)
	srv.handlePortfolio(rec, req)

	resp := decodeResponse(t, rec)
	data := resp.Data.(map[string]interface{})
	margins, ok := data["margins"].(map[string]interface{})
	if !ok {
		t.Fatal("margins should be a map")
	}

	if v, ok := margins["available_cash"].(float64); !ok || v != 1000000 {
		t.Errorf("available_cash: got %v", margins["available_cash"])
	}
}

func TestHealthReturn_FieldsPresent(t *testing.T) {
	srv := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	srv.handleHealth(rec, req)

	var raw map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}

	data, ok := raw["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data not present or not an object")
	}

	required := []string{"status", "version", "market_status", "time_ist"}
	for _, key := range required {
		if _, ok := data[key]; !ok {
			t.Errorf("missing required field in health data: %q", key)
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// Compile-time interface checks
// ════════════════════════════════════════════════════════════════════

// Ensure WSHub methods exist
var _ = (*WSHub)(nil).ClientCount
var _ = (*WSHub)(nil).Broadcast
var _ = (*WSHub)(nil).Register
var _ = (*WSHub)(nil).Unregister
var _ = (*WSHub)(nil).Run

// ════════════════════════════════════════════════════════════════════
// Batch: edge cases for writeJSON
// ════════════════════════════════════════════════════════════════════

func TestWriteJSON_NilData(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusOK, APIResponse{Success: true})

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestWriteJSON_NestedData(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"nested": map[string]interface{}{
				"deep": "value",
			},
		},
	})

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}

	if _, ok := raw["data"]; !ok {
		t.Error("missing data field")
	}
}

func TestWriteError_EmptyMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusInternalServerError, "")

	resp := decodeResponse(t, rec)
	if resp.Success {
		t.Error("expected success=false")
	}
}

func TestWriteError_VariousStatusCodes(t *testing.T) {
	codes := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
		http.StatusServiceUnavailable,
	}

	for _, code := range codes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeError(rec, code, "test error")

			if rec.Code != code {
				t.Errorf("status: got %d, want %d", rec.Code, code)
			}

			resp := decodeResponse(t, rec)
			if resp.Success {
				t.Error("expected success=false")
			}
		})
	}
}
