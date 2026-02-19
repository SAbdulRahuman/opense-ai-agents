package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// ════════════════════════════════════════════════════════════════════
// provider.go — Types & Helpers
// ════════════════════════════════════════════════════════════════════

func TestMessageConstructors(t *testing.T) {
	sys := SystemMessage("You are helpful.")
	if sys.Role != RoleSystem || sys.Content != "You are helpful." {
		t.Fatalf("SystemMessage: got %+v", sys)
	}

	user := UserMessage("hello")
	if user.Role != RoleUser || user.Content != "hello" {
		t.Fatalf("UserMessage: got %+v", user)
	}

	asst := AssistantMessage("hi there")
	if asst.Role != RoleAssistant || asst.Content != "hi there" {
		t.Fatalf("AssistantMessage: got %+v", asst)
	}

	tool := ToolResultMessage("call_1", "get_price", "₹2847.50")
	if tool.Role != RoleTool || tool.ToolCallID != "call_1" || tool.Name != "get_price" || tool.Content != "₹2847.50" {
		t.Fatalf("ToolResultMessage: got %+v", tool)
	}

	tc := AssistantToolCallMessage([]ToolCall{{ID: "c1", Name: "fn"}})
	if tc.Role != RoleAssistant || len(tc.ToolCalls) != 1 {
		t.Fatalf("AssistantToolCallMessage: got %+v", tc)
	}
}

func TestResponseHasToolCalls(t *testing.T) {
	r := &Response{Content: "hello"}
	if r.HasToolCalls() {
		t.Fatal("should not have tool calls")
	}
	r.ToolCalls = []ToolCall{{ID: "1", Name: "fn"}}
	if !r.HasToolCalls() {
		t.Fatal("should have tool calls")
	}
}

func TestResponseString(t *testing.T) {
	r := &Response{
		Provider: "openai", Model: "gpt-4o",
		Content: "short answer",
		Usage:   Usage{TotalTokens: 50},
		Latency: 100 * time.Millisecond,
	}
	s := r.String()
	if !strings.Contains(s, "openai/gpt-4o") || !strings.Contains(s, "50 tokens") {
		t.Fatalf("unexpected String(): %s", s)
	}

	// With tool calls
	r.ToolCalls = []ToolCall{{ID: "1", Name: "fn"}}
	s = r.String()
	if !strings.Contains(s, "1 tool call") {
		t.Fatalf("unexpected String() with tools: %s", s)
	}

	// Long content (truncation)
	r.ToolCalls = nil
	r.Content = strings.Repeat("x", 200)
	s = r.String()
	if !strings.Contains(s, "...") {
		t.Fatal("expected truncation for long content")
	}
}

func TestDefaultProviderConfig(t *testing.T) {
	cfg := DefaultProviderConfig()
	if cfg.Model != "gpt-4o" || cfg.Temperature != 0.1 || cfg.MaxTokens != 4096 || cfg.Timeout != 120*time.Second {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

// ════════════════════════════════════════════════════════════════════
// tools.go — ToolRegistry & RunToolLoop
// ════════════════════════════════════════════════════════════════════

func TestToolRegistryBasic(t *testing.T) {
	reg := NewToolRegistry()
	if reg.Count() != 0 {
		t.Fatal("new registry should be empty")
	}

	reg.Register(Tool{
		Name:        "get_price",
		Description: "Get stock price",
		Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			return "₹2847.50", nil
		},
	})

	if reg.Count() != 1 {
		t.Fatalf("count: got %d", reg.Count())
	}
	tool, ok := reg.Get("get_price")
	if !ok || tool.Name != "get_price" {
		t.Fatal("Get failed")
	}
	_, ok = reg.Get("nonexistent")
	if ok {
		t.Fatal("should not find nonexistent")
	}

	names := reg.Names()
	if len(names) != 1 || names[0] != "get_price" {
		t.Fatalf("Names: got %v", names)
	}

	list := reg.List()
	if len(list) != 1 {
		t.Fatalf("List: got %d", len(list))
	}
}

func TestToolRegistryRegisterFunc(t *testing.T) {
	reg := NewToolRegistry()
	reg.RegisterFunc("add", "Add numbers", nil, func(ctx context.Context, args json.RawMessage) (string, error) {
		return "42", nil
	})
	if reg.Count() != 1 {
		t.Fatal("RegisterFunc should add tool")
	}
}

func TestToolRegistryExecute(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(Tool{
		Name: "echo",
		Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			return string(args), nil
		},
	})

	result, err := reg.Execute(context.Background(), ToolCall{ID: "1", Name: "echo", Arguments: json.RawMessage(`"hello"`)})
	if err != nil || result != `"hello"` {
		t.Fatalf("Execute: got %q, err=%v", result, err)
	}

	// Not found
	_, err = reg.Execute(context.Background(), ToolCall{ID: "2", Name: "missing"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got: %v", err)
	}

	// Nil handler
	reg.Register(Tool{Name: "nohandler"})
	_, err = reg.Execute(context.Background(), ToolCall{ID: "3", Name: "nohandler"})
	if err == nil || !strings.Contains(err.Error(), "no handler") {
		t.Fatalf("expected no handler error, got: %v", err)
	}
}

func TestToolRegistryExecuteAll(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(Tool{
		Name: "slow",
		Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			time.Sleep(10 * time.Millisecond)
			return "done", nil
		},
	})
	reg.Register(Tool{
		Name: "fast",
		Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			return "fast_done", nil
		},
	})

	calls := []ToolCall{
		{ID: "1", Name: "slow"},
		{ID: "2", Name: "fast"},
	}
	results := reg.ExecuteAll(context.Background(), calls)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Content != "done" || results[1].Content != "fast_done" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestToolResultToMessage(t *testing.T) {
	// Success case
	tr := ToolResult{ToolCallID: "c1", Name: "fn", Content: "result"}
	msg := tr.ToMessage()
	if msg.Role != RoleTool || msg.Content != "result" || msg.ToolCallID != "c1" {
		t.Fatalf("success ToMessage: %+v", msg)
	}

	// Error case
	tr = ToolResult{ToolCallID: "c2", Name: "fn", Err: fmt.Errorf("boom")}
	msg = tr.ToMessage()
	if !strings.Contains(msg.Content, "Error") || !strings.Contains(msg.Content, "boom") {
		t.Fatalf("error ToMessage: %+v", msg)
	}
}

func TestToolRegistryConcurrency(t *testing.T) {
	reg := NewToolRegistry()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := fmt.Sprintf("tool_%d", n)
			reg.Register(Tool{Name: name})
			reg.Get(name)
			reg.Names()
			reg.List()
			reg.Count()
		}(i)
	}
	wg.Wait()
	if reg.Count() != 100 {
		t.Fatalf("expected 100 tools, got %d", reg.Count())
	}
}

func TestJSONSchemaHelpers(t *testing.T) {
	schema := ObjectSchema("Test params",
		map[string]*JSONSchema{
			"ticker": StringProp("Stock ticker symbol"),
			"period": IntProp("Number of periods"),
			"price":  NumberProp("Price value"),
			"active": BoolProp("Is active"),
			"type":   EnumProp("Analysis type", "technical", "fundamental"),
			"items":  ArrayProp("List of items", StringProp("item name")),
		},
		"ticker",
	)

	if schema.Type != "object" || len(schema.Properties) != 6 || len(schema.Required) != 1 {
		t.Fatalf("ObjectSchema: %+v", schema)
	}
	if schema.Properties["ticker"].Type != "string" {
		t.Fatal("StringProp type mismatch")
	}
	if schema.Properties["period"].Type != "integer" {
		t.Fatal("IntProp type mismatch")
	}
	if schema.Properties["price"].Type != "number" {
		t.Fatal("NumberProp type mismatch")
	}
	if schema.Properties["active"].Type != "boolean" {
		t.Fatal("BoolProp type mismatch")
	}
	if len(schema.Properties["type"].Enum) != 2 {
		t.Fatal("EnumProp enum mismatch")
	}
	if schema.Properties["items"].Items == nil || schema.Properties["items"].Items.Type != "string" {
		t.Fatal("ArrayProp items mismatch")
	}
}

// ════════════════════════════════════════════════════════════════════
// openai.go — OpenAI Provider with mock server
// ════════════════════════════════════════════════════════════════════

func TestOpenAIProviderNew(t *testing.T) {
	_, err := NewOpenAIProvider("")
	if err != ErrNoAPIKey {
		t.Fatalf("expected ErrNoAPIKey, got: %v", err)
	}

	p, err := NewOpenAIProvider("sk-test", WithOpenAIModel("gpt-4"), WithOpenAIBaseURL("http://custom"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "openai" || p.model != "gpt-4" || p.baseURL != "http://custom" {
		t.Fatalf("unexpected config: %+v", p)
	}
	if len(p.Models()) == 0 {
		t.Fatal("Models() should not be empty")
	}
}

func newMockOpenAIServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestOpenAIChat(t *testing.T) {
	server := newMockOpenAIServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Fatal("missing auth header")
		}

		// Decode the request to verify structure
		var req openAIChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "gpt-4o" {
			t.Fatalf("unexpected model: %s", req.Model)
		}
		if len(req.Messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(req.Messages))
		}

		resp := openAIChatResponse{
			ID: "chatcmpl-123",
			Choices: []openAIChoice{{
				Index:        0,
				Message:      openAIMessage{Role: "assistant", Content: "RSI of RELIANCE is 62.4"},
				FinishReason: "stop",
			}},
			Usage: openAIUsage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
			Model: "gpt-4o",
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	p, _ := NewOpenAIProvider("sk-test", WithOpenAIBaseURL(server.URL))
	resp, err := p.Chat(context.Background(),
		[]Message{SystemMessage("You are helpful."), UserMessage("What is the RSI of RELIANCE?")},
		nil, nil)

	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "RSI of RELIANCE is 62.4" {
		t.Fatalf("unexpected content: %s", resp.Content)
	}
	if resp.Provider != "openai" || resp.Usage.TotalTokens != 30 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.FinishReason != FinishStop {
		t.Fatalf("expected stop, got %s", resp.FinishReason)
	}
}

func TestOpenAIChatWithToolCalls(t *testing.T) {
	server := newMockOpenAIServer(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIChatResponse{
			ID: "chatcmpl-456",
			Choices: []openAIChoice{{
				Index: 0,
				Message: openAIMessage{
					Role: "assistant",
					ToolCalls: []openAIToolCall{{
						ID:   "call_abc",
						Type: "function",
						Function: openAIFunctionCall{
							Name:      "get_price",
							Arguments: `{"ticker":"RELIANCE"}`,
						},
					}},
				},
				FinishReason: "tool_calls",
			}},
			Usage: openAIUsage{TotalTokens: 25},
			Model: "gpt-4o",
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	p, _ := NewOpenAIProvider("sk-test", WithOpenAIBaseURL(server.URL))
	resp, err := p.Chat(context.Background(),
		[]Message{UserMessage("price of RELIANCE")},
		[]Tool{{Name: "get_price", Description: "Get price"}}, nil)

	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasToolCalls() {
		t.Fatal("expected tool calls")
	}
	if resp.ToolCalls[0].Name != "get_price" || resp.ToolCalls[0].ID != "call_abc" {
		t.Fatalf("unexpected tool call: %+v", resp.ToolCalls[0])
	}
	if resp.FinishReason != FinishToolCalls {
		t.Fatalf("expected tool_calls finish, got %s", resp.FinishReason)
	}
}

func TestOpenAIChatWithOptions(t *testing.T) {
	server := newMockOpenAIServer(func(w http.ResponseWriter, r *http.Request) {
		var req openAIChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "gpt-4-turbo" {
			t.Fatalf("expected model override, got %s", req.Model)
		}
		if req.Temperature == nil || *req.Temperature != 0.5 {
			t.Fatal("expected temperature 0.5")
		}
		resp := openAIChatResponse{
			Choices: []openAIChoice{{Message: openAIMessage{Content: "ok"}, FinishReason: "stop"}},
			Model:   "gpt-4-turbo",
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	p, _ := NewOpenAIProvider("sk-test", WithOpenAIBaseURL(server.URL))
	_, err := p.Chat(context.Background(),
		[]Message{UserMessage("test")}, nil,
		&ChatOptions{Model: "gpt-4-turbo", Temperature: 0.5, MaxTokens: 100})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOpenAIErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		expectErr  string
	}{
		{
			name:       "unauthorized",
			statusCode: 401,
			body:       `{"error":{"message":"Invalid key","type":"auth","code":"invalid_api_key"}}`,
			expectErr:  "api key",
		},
		{
			name:       "rate_limit",
			statusCode: 429,
			body:       `{"error":{"message":"Rate limit exceeded","type":"rate_limit"}}`,
			expectErr:  "rate limit",
		},
		{
			name:       "context_length",
			statusCode: 400,
			body:       `{"error":{"message":"Too many tokens","code":"context_length_exceeded"}}`,
			expectErr:  "context length",
		},
		{
			name:       "model_not_found",
			statusCode: 400,
			body:       `{"error":{"message":"Model not found","code":"model_not_found"}}`,
			expectErr:  "invalid model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newMockOpenAIServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			})
			defer server.Close()

			p, _ := NewOpenAIProvider("sk-test", WithOpenAIBaseURL(server.URL))
			_, err := p.Chat(context.Background(), []Message{UserMessage("test")}, nil, nil)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(strings.ToLower(err.Error()), tt.expectErr) {
				t.Fatalf("expected error containing %q, got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestOpenAIPing(t *testing.T) {
	server := newMockOpenAIServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":[]}`))
			return
		}
	})
	defer server.Close()

	p, _ := NewOpenAIProvider("sk-test", WithOpenAIBaseURL(server.URL))
	if err := p.Ping(context.Background()); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestOpenAIChatStream(t *testing.T) {
	server := newMockOpenAIServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("server does not support flushing")
		}

		chunks := []string{
			`data: {"choices":[{"delta":{"content":"Hello"},"index":0}]}`,
			`data: {"choices":[{"delta":{"content":" world"},"index":0}]}`,
			`data: {"choices":[{"delta":{},"finish_reason":"stop","index":0}]}`,
			`data: [DONE]`,
		}
		for _, chunk := range chunks {
			fmt.Fprintln(w, chunk)
			flusher.Flush()
		}
	})
	defer server.Close()

	p, _ := NewOpenAIProvider("sk-test", WithOpenAIBaseURL(server.URL))
	ch, err := p.ChatStream(context.Background(), []Message{UserMessage("hi")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var content strings.Builder
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatal(chunk.Err)
		}
		content.WriteString(chunk.Content)
	}
	if content.String() != "Hello world" {
		t.Fatalf("unexpected stream content: %q", content.String())
	}
}

// ════════════════════════════════════════════════════════════════════
// ollama.go — Ollama Provider with mock server
// ════════════════════════════════════════════════════════════════════

func TestOllamaProviderNew(t *testing.T) {
	p, err := NewOllamaProvider("", WithOllamaModel("llama3.1:8b"))
	if err != nil {
		t.Fatal(err)
	}
	if p.baseURL != "http://localhost:11434" || p.model != "llama3.1:8b" {
		t.Fatalf("unexpected config: %+v", p)
	}
	if p.Name() != "ollama" || len(p.Models()) == 0 {
		t.Fatal("basic methods failed")
	}
}

func TestOllamaChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var req ollamaChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "qwen2.5:7b" {
			t.Fatalf("unexpected model: %s", req.Model)
		}
		if req.Stream {
			t.Fatal("stream should be false for Chat")
		}

		resp := ollamaChatResponse{
			Model:   "qwen2.5:7b",
			Message: ollamaMessage{Role: "assistant", Content: "RELIANCE RSI = 58.2"},
			Done:    true,
			PromptEvalCount: 15,
			EvalCount:       8,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p, _ := NewOllamaProvider(server.URL)
	resp, err := p.Chat(context.Background(),
		[]Message{UserMessage("RSI of RELIANCE?")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "RELIANCE RSI = 58.2" {
		t.Fatalf("unexpected content: %s", resp.Content)
	}
	if resp.Provider != "ollama" || resp.Usage.TotalTokens != 23 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestOllamaChatWithTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Tools) != 1 {
			t.Fatalf("expected 1 tool, got %d", len(req.Tools))
		}

		resp := ollamaChatResponse{
			Model: "qwen2.5:7b",
			Message: ollamaMessage{
				Role: "assistant",
				ToolCalls: []ollamaToolCall{{
					Function: ollamaFunctionCall{
						Name:      "get_rsi",
						Arguments: json.RawMessage(`{"ticker":"TCS","period":14}`),
					},
				}},
			},
			Done: true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p, _ := NewOllamaProvider(server.URL)
	resp, err := p.Chat(context.Background(),
		[]Message{UserMessage("RSI of TCS")},
		[]Tool{{Name: "get_rsi", Description: "Get RSI"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasToolCalls() || resp.ToolCalls[0].Name != "get_rsi" {
		t.Fatalf("unexpected tool calls: %+v", resp.ToolCalls)
	}
	if resp.FinishReason != FinishToolCalls {
		t.Fatalf("expected tool_calls, got %s", resp.FinishReason)
	}
}

func TestOllamaPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Write([]byte(`{"models":[]}`))
			return
		}
	}))
	defer server.Close()

	p, _ := NewOllamaProvider(server.URL)
	if err := p.Ping(context.Background()); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestOllamaChatStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if !req.Stream {
			t.Fatal("stream should be true")
		}

		flusher, _ := w.(http.Flusher)
		chunks := []ollamaChatResponse{
			{Message: ollamaMessage{Content: "Nifty "}, Done: false},
			{Message: ollamaMessage{Content: "is bullish"}, Done: false},
			{Message: ollamaMessage{Content: ""}, Done: true, EvalCount: 5},
		}
		for _, c := range chunks {
			data, _ := json.Marshal(c)
			fmt.Fprintln(w, string(data))
			flusher.Flush()
		}
	}))
	defer server.Close()

	p, _ := NewOllamaProvider(server.URL)
	ch, err := p.ChatStream(context.Background(), []Message{UserMessage("market view")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var content strings.Builder
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatal(chunk.Err)
		}
		content.WriteString(chunk.Content)
	}
	if content.String() != "Nifty is bullish" {
		t.Fatalf("unexpected stream: %q", content.String())
	}
}

// ════════════════════════════════════════════════════════════════════
// gemini.go — Gemini Provider with mock server
// ════════════════════════════════════════════════════════════════════

func TestGeminiProviderNew(t *testing.T) {
	_, err := NewGeminiProvider("")
	if err != ErrNoAPIKey {
		t.Fatalf("expected ErrNoAPIKey, got: %v", err)
	}

	p, err := NewGeminiProvider("test-key", WithGeminiModel("gemini-1.5-pro"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "gemini" || p.model != "gemini-1.5-pro" {
		t.Fatalf("unexpected config: %+v", p)
	}
}

func TestGeminiChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "generateContent") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "key=gem-key") {
			t.Fatal("missing API key in query")
		}

		resp := geminiResponse{
			Candidates: []geminiCandidate{{
				Content: geminiContent{
					Role:  "model",
					Parts: []geminiPart{{Text: "PE of TCS is 32.5"}},
				},
				FinishReason: "STOP",
			}},
			UsageMetadata: geminiUsageMetadata{
				PromptTokenCount:     10,
				CandidatesTokenCount: 8,
				TotalTokenCount:      18,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p, _ := NewGeminiProvider("gem-key", WithGeminiModel("gemini-2.0-flash"))
	p.baseURL = server.URL

	resp, err := p.Chat(context.Background(),
		[]Message{SystemMessage("Financial analyst"), UserMessage("PE of TCS?")},
		nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "PE of TCS is 32.5" {
		t.Fatalf("unexpected content: %s", resp.Content)
	}
	if resp.Provider != "gemini" || resp.Usage.TotalTokens != 18 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGeminiChatWithToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiResponse{
			Candidates: []geminiCandidate{{
				Content: geminiContent{
					Role: "model",
					Parts: []geminiPart{{
						FunctionCall: &geminiFunctionCall{
							Name: "get_financials",
							Args: json.RawMessage(`{"ticker":"INFY"}`),
						},
					}},
				},
				FinishReason: "STOP",
			}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p, _ := NewGeminiProvider("gem-key")
	p.baseURL = server.URL

	resp, err := p.Chat(context.Background(),
		[]Message{UserMessage("financials of INFY")},
		[]Tool{{Name: "get_financials"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasToolCalls() || resp.ToolCalls[0].Name != "get_financials" {
		t.Fatalf("unexpected tool calls: %+v", resp.ToolCalls)
	}
}

func TestGeminiPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/models") {
			w.Write([]byte(`{"models":[]}`))
			return
		}
	}))
	defer server.Close()

	p, _ := NewGeminiProvider("gem-key")
	p.baseURL = server.URL
	if err := p.Ping(context.Background()); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestGeminiStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "streamGenerateContent") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		chunks := []geminiResponse{
			{Candidates: []geminiCandidate{{Content: geminiContent{Parts: []geminiPart{{Text: "Market "}}}}}},
			{Candidates: []geminiCandidate{{Content: geminiContent{Parts: []geminiPart{{Text: "is up"}}}, FinishReason: "STOP"}}},
		}
		for _, c := range chunks {
			data, _ := json.Marshal(c)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}))
	defer server.Close()

	p, _ := NewGeminiProvider("gem-key")
	p.baseURL = server.URL

	ch, err := p.ChatStream(context.Background(), []Message{UserMessage("market status")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var content strings.Builder
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatal(chunk.Err)
		}
		content.WriteString(chunk.Content)
	}
	if content.String() != "Market is up" {
		t.Fatalf("unexpected stream: %q", content.String())
	}
}

// ════════════════════════════════════════════════════════════════════
// anthropic.go — Anthropic Provider with mock server
// ════════════════════════════════════════════════════════════════════

func TestAnthropicProviderNew(t *testing.T) {
	_, err := NewAnthropicProvider("")
	if err != ErrNoAPIKey {
		t.Fatalf("expected ErrNoAPIKey, got: %v", err)
	}

	p, err := NewAnthropicProvider("sk-ant-test",
		WithAnthropicModel("claude-3-5-sonnet-20241022"),
		WithAnthropicBaseURL("http://custom"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "anthropic" || p.model != "claude-3-5-sonnet-20241022" || p.baseURL != "http://custom" {
		t.Fatalf("unexpected config: %+v", p)
	}
}

func TestAnthropicChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "sk-ant-test" {
			t.Fatal("missing x-api-key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Fatal("missing anthropic-version header")
		}

		var req anthropicRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.System == "" {
			t.Fatal("expected system prompt")
		}

		resp := anthropicResponse{
			ID:   "msg_123",
			Type: "message",
			Role: "assistant",
			Content: []anthropicContentBlock{{
				Type: "text",
				Text: "HDFC Bank PE is 19.8",
			}},
			Model:      "claude-sonnet-4-20250514",
			StopReason: "end_turn",
			Usage:      anthropicUsage{InputTokens: 15, OutputTokens: 10},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p, _ := NewAnthropicProvider("sk-ant-test", WithAnthropicBaseURL(server.URL))
	resp, err := p.Chat(context.Background(),
		[]Message{SystemMessage("Financial analyst"), UserMessage("PE of HDFC Bank?")},
		nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "HDFC Bank PE is 19.8" {
		t.Fatalf("unexpected content: %s", resp.Content)
	}
	if resp.Provider != "anthropic" || resp.Usage.TotalTokens != 25 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.FinishReason != FinishStop {
		t.Fatalf("expected stop, got %s", resp.FinishReason)
	}
}

func TestAnthropicChatWithToolUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req anthropicRequest
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Tools) != 1 {
			t.Fatalf("expected 1 tool, got %d", len(req.Tools))
		}

		resp := anthropicResponse{
			Content: []anthropicContentBlock{
				{Type: "text", Text: "Let me check."},
				{Type: "tool_use", ID: "toolu_01", Name: "get_oi", Input: json.RawMessage(`{"ticker":"NIFTY"}`)},
			},
			StopReason: "tool_use",
			Usage:      anthropicUsage{InputTokens: 20, OutputTokens: 15},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p, _ := NewAnthropicProvider("sk-ant-test", WithAnthropicBaseURL(server.URL))
	resp, err := p.Chat(context.Background(),
		[]Message{UserMessage("OI analysis for NIFTY")},
		[]Tool{{Name: "get_oi", Description: "Get OI data"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasToolCalls() {
		t.Fatal("expected tool calls")
	}
	if resp.Content != "Let me check." {
		t.Fatalf("unexpected content: %s", resp.Content)
	}
	if resp.ToolCalls[0].Name != "get_oi" || resp.ToolCalls[0].ID != "toolu_01" {
		t.Fatalf("unexpected tool call: %+v", resp.ToolCalls[0])
	}
	if resp.FinishReason != FinishToolCalls {
		t.Fatalf("expected tool_calls, got %s", resp.FinishReason)
	}
}

func TestAnthropicStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req anthropicRequest
		json.NewDecoder(r.Body).Decode(&req)
		if !req.Stream {
			t.Fatal("expected stream=true")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		events := []string{
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Bullish "}}`,
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"on TCS"}}`,
			`data: {"type":"content_block_stop","index":0}`,
			`data: {"type":"message_stop"}`,
		}
		for _, e := range events {
			fmt.Fprintln(w, e)
			flusher.Flush()
		}
	}))
	defer server.Close()

	p, _ := NewAnthropicProvider("sk-ant-test", WithAnthropicBaseURL(server.URL))
	ch, err := p.ChatStream(context.Background(), []Message{UserMessage("view on TCS")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var content strings.Builder
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatal(chunk.Err)
		}
		content.WriteString(chunk.Content)
	}
	if content.String() != "Bullish on TCS" {
		t.Fatalf("unexpected stream: %q", content.String())
	}
}

func TestAnthropicPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content:    []anthropicContentBlock{{Type: "text", Text: "hi"}},
			StopReason: "end_turn",
			Usage:      anthropicUsage{InputTokens: 1, OutputTokens: 1},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p, _ := NewAnthropicProvider("sk-ant-test", WithAnthropicBaseURL(server.URL))
	if err := p.Ping(context.Background()); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

// ════════════════════════════════════════════════════════════════════
// router.go — Router tests
// ════════════════════════════════════════════════════════════════════

// mockProvider implements LLMProvider for testing the router.
type mockProvider struct {
	name      string
	chatFunc  func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error)
	pingErr   error
}

func (m *mockProvider) Name() string    { return m.name }
func (m *mockProvider) Models() []string { return []string{"mock-model"} }
func (m *mockProvider) Ping(ctx context.Context) error { return m.pingErr }
func (m *mockProvider) Chat(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
	if m.chatFunc != nil {
		return m.chatFunc(ctx, messages, tools, opts)
	}
	return &Response{Content: "mock response", Provider: m.name}, nil
}
func (m *mockProvider) ChatStream(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	ch <- StreamChunk{Content: "streamed", Done: true}
	close(ch)
	return ch, nil
}

func TestRouterBasic(t *testing.T) {
	r := NewRouter("primary")
	r.RegisterProvider(&mockProvider{name: "primary"})

	p, err := r.Primary()
	if err != nil || p.Name() != "primary" {
		t.Fatalf("Primary: %v, %v", p, err)
	}

	names := r.ProviderNames()
	if len(names) != 1 || names[0] != "primary" {
		t.Fatalf("ProviderNames: %v", names)
	}
}

func TestRouterChat(t *testing.T) {
	r := NewRouter("main")
	r.RegisterProvider(&mockProvider{
		name: "main",
		chatFunc: func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
			return &Response{Content: "from main", Provider: "main"}, nil
		},
	})

	resp, err := r.Chat(context.Background(), []Message{UserMessage("test")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "from main" {
		t.Fatalf("unexpected: %s", resp.Content)
	}
}

func TestRouterFallback(t *testing.T) {
	callCount := 0
	r := NewRouter("primary",
		WithFallbacks("backup"),
		WithMaxRetries(0), // no retries to speed up test
	)
	r.RegisterProvider(&mockProvider{
		name: "primary",
		chatFunc: func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
			callCount++
			return nil, fmt.Errorf("%w: primary down", ErrProviderDown)
		},
	})
	r.RegisterProvider(&mockProvider{
		name: "backup",
		chatFunc: func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
			callCount++
			return &Response{Content: "from backup", Provider: "backup"}, nil
		},
	})

	resp, err := r.Chat(context.Background(), []Message{UserMessage("test")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "from backup" || resp.Provider != "backup" {
		t.Fatalf("expected fallback response, got: %+v", resp)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 calls (primary + backup), got %d", callCount)
	}
}

func TestRouterAllFail(t *testing.T) {
	r := NewRouter("a",
		WithFallbacks("b"),
		WithMaxRetries(0),
	)
	r.RegisterProvider(&mockProvider{
		name: "a",
		chatFunc: func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
			return nil, ErrProviderDown
		},
	})
	r.RegisterProvider(&mockProvider{
		name: "b",
		chatFunc: func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
			return nil, ErrProviderDown
		},
	})

	_, err := r.Chat(context.Background(), []Message{UserMessage("test")}, nil, nil)
	if err == nil {
		t.Fatal("expected error when all fail")
	}
	if !strings.Contains(err.Error(), "all providers failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRouterNoProviders(t *testing.T) {
	r := NewRouter("nonexistent")
	_, err := r.Chat(context.Background(), []Message{UserMessage("test")}, nil, nil)
	if err == nil {
		t.Fatal("expected error when no providers registered")
	}
	if !strings.Contains(err.Error(), "all providers failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRouterChatWithComplexity(t *testing.T) {
	r := NewRouter("main",
		WithModelMap(map[TaskComplexity]string{
			TaskSimple:  "fast-model",
			TaskComplex: "powerful-model",
		}),
	)
	var capturedModel string
	r.RegisterProvider(&mockProvider{
		name: "main",
		chatFunc: func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
			if opts != nil {
				capturedModel = opts.Model
			}
			return &Response{Content: "ok"}, nil
		},
	})

	_, err := r.ChatWithComplexity(context.Background(), TaskSimple, []Message{UserMessage("hi")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if capturedModel != "fast-model" {
		t.Fatalf("expected fast-model, got %s", capturedModel)
	}

	_, err = r.ChatWithComplexity(context.Background(), TaskComplex, []Message{UserMessage("full analysis")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if capturedModel != "powerful-model" {
		t.Fatalf("expected powerful-model, got %s", capturedModel)
	}
}

func TestRouterHealthCheck(t *testing.T) {
	r := NewRouter("a")
	r.RegisterProvider(&mockProvider{name: "a", pingErr: nil})
	r.RegisterProvider(&mockProvider{name: "b", pingErr: fmt.Errorf("down")})

	results := r.HealthCheck(context.Background())
	if results["a"] != nil {
		t.Fatalf("expected a=nil, got %v", results["a"])
	}
	if results["b"] == nil {
		t.Fatal("expected b=error")
	}
}

func TestRouterNonRetryableError(t *testing.T) {
	r := NewRouter("main", WithFallbacks("backup"), WithMaxRetries(3))
	r.RegisterProvider(&mockProvider{
		name: "main",
		chatFunc: func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
			return nil, ErrNoAPIKey // non-retryable
		},
	})
	r.RegisterProvider(&mockProvider{name: "backup"})

	_, err := r.Chat(context.Background(), []Message{UserMessage("test")}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("expected non-retryable error, got: %v", err)
	}
}

func TestRouterStream(t *testing.T) {
	r := NewRouter("main")
	r.RegisterProvider(&mockProvider{name: "main"})

	ch, err := r.ChatStream(context.Background(), []Message{UserMessage("test")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	chunk := <-ch
	if chunk.Content != "streamed" {
		t.Fatalf("unexpected stream: %s", chunk.Content)
	}
}

// ════════════════════════════════════════════════════════════════════
// RunToolLoop — Integration test
// ════════════════════════════════════════════════════════════════════

func TestRunToolLoop(t *testing.T) {
	callNum := 0
	provider := &mockProvider{
		name: "test",
		chatFunc: func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
			callNum++
			if callNum == 1 {
				// First call: request a tool call
				return &Response{
					ToolCalls: []ToolCall{{
						ID:        "call_1",
						Name:      "get_price",
						Arguments: json.RawMessage(`{"ticker":"TCS"}`),
					}},
					FinishReason: FinishToolCalls,
				}, nil
			}
			// Second call: return final answer
			return &Response{
				Content:      "TCS price is ₹4,200",
				FinishReason: FinishStop,
			}, nil
		},
	}

	registry := NewToolRegistry()
	registry.Register(Tool{
		Name: "get_price",
		Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			return "₹4,200.00", nil
		},
	})

	msgs := []Message{UserMessage("Price of TCS?")}
	tools := []Tool{{Name: "get_price", Description: "Get stock price"}}

	resp, finalMsgs, err := RunToolLoop(context.Background(), provider, registry, msgs, tools, nil, 5)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "TCS price is ₹4,200" {
		t.Fatalf("unexpected content: %s", resp.Content)
	}
	if callNum != 2 {
		t.Fatalf("expected 2 LLM calls, got %d", callNum)
	}
	// Original message + assistant tool call + tool result = 3
	if len(finalMsgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(finalMsgs))
	}
}

func TestRunToolLoopMaxIterations(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		chatFunc: func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
			// Always request tool calls (infinite loop)
			return &Response{
				ToolCalls:    []ToolCall{{ID: "c1", Name: "fn", Arguments: json.RawMessage(`{}`)}},
				FinishReason: FinishToolCalls,
			}, nil
		},
	}

	registry := NewToolRegistry()
	registry.Register(Tool{
		Name:    "fn",
		Handler: func(ctx context.Context, args json.RawMessage) (string, error) { return "ok", nil },
	})

	_, _, err := RunToolLoop(context.Background(), provider, registry,
		[]Message{UserMessage("test")}, []Tool{{Name: "fn"}}, nil, 3)
	if err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("expected max iterations error, got: %v", err)
	}
}

func TestRunToolLoopNoToolCalls(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		chatFunc: func(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
			return &Response{Content: "direct answer", FinishReason: FinishStop}, nil
		},
	}

	resp, msgs, err := RunToolLoop(context.Background(), provider, NewToolRegistry(),
		[]Message{UserMessage("hello")}, nil, nil, 5)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "direct answer" {
		t.Fatalf("unexpected: %s", resp.Content)
	}
	if len(msgs) != 1 { // only original message, no tool call messages added
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

// ════════════════════════════════════════════════════════════════════
// gemini.go — quoteIfNeeded helper
// ════════════════════════════════════════════════════════════════════

func TestQuoteIfNeeded(t *testing.T) {
	// Already valid JSON
	if got := quoteIfNeeded(`{"key":"value"}`); got != `{"key":"value"}` {
		t.Fatalf("should pass through valid JSON: %s", got)
	}
	// Plain string that needs quoting
	if got := quoteIfNeeded(`hello world`); got != `"hello world"` {
		t.Fatalf("should quote plain string: %s", got)
	}
	// Number
	if got := quoteIfNeeded(`42`); got != `42` {
		t.Fatalf("should pass through number: %s", got)
	}
}

// ════════════════════════════════════════════════════════════════════
// Conversion helpers — OpenAI
// ════════════════════════════════════════════════════════════════════

func TestConvertToOpenAIMessages(t *testing.T) {
	msgs := []Message{
		SystemMessage("system"),
		UserMessage("user msg"),
		AssistantMessage("assistant msg"),
		ToolResultMessage("c1", "fn", "result"),
	}
	oai := convertToOpenAIMessages(msgs)
	if len(oai) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(oai))
	}
	if oai[0].Role != "system" || oai[1].Role != "user" || oai[2].Role != "assistant" || oai[3].Role != "tool" {
		t.Fatal("role mismatch")
	}
	if oai[3].ToolCallID != "c1" || oai[3].Name != "fn" {
		t.Fatal("tool result fields mismatch")
	}
}

func TestConvertToOpenAITools(t *testing.T) {
	tools := []Tool{{
		Name:        "get_price",
		Description: "Get stock price",
		Parameters:  ObjectSchema("", map[string]*JSONSchema{"ticker": StringProp("symbol")}, "ticker"),
	}}
	oaiTools := convertToOpenAITools(tools)
	if len(oaiTools) != 1 || oaiTools[0].Type != "function" || oaiTools[0].Function.Name != "get_price" {
		t.Fatalf("tool conversion failed: %+v", oaiTools)
	}
}

func TestMapFinishReason(t *testing.T) {
	tests := map[string]FinishReason{
		"stop":       FinishStop,
		"tool_calls": FinishToolCalls,
		"length":     FinishLength,
		"unknown":    FinishReason("unknown"),
	}
	for input, expected := range tests {
		if got := mapFinishReason(input); got != expected {
			t.Fatalf("mapFinishReason(%q): got %s, want %s", input, got, expected)
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// OpenAI streaming error
// ════════════════════════════════════════════════════════════════════

func TestOpenAIStreamError(t *testing.T) {
	server := newMockOpenAIServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"Rate limit","type":"rate_limit"}}`))
	})
	defer server.Close()

	p, _ := NewOpenAIProvider("sk-test", WithOpenAIBaseURL(server.URL))
	_, err := p.ChatStream(context.Background(), []Message{UserMessage("test")}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "rate limit") {
		t.Fatalf("expected rate limit error, got: %v", err)
	}
}

// ════════════════════════════════════════════════════════════════════
// Anthropic error handling
// ════════════════════════════════════════════════════════════════════

func TestAnthropicErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"type":"error","error":{"type":"authentication_error","message":"Invalid key"}}`))
	}))
	defer server.Close()

	p, _ := NewAnthropicProvider("bad-key", WithAnthropicBaseURL(server.URL))
	_, err := p.Chat(context.Background(), []Message{UserMessage("test")}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("expected auth error, got: %v", err)
	}
}

// ════════════════════════════════════════════════════════════════════
// Anthropic stream with tool use
// ════════════════════════════════════════════════════════════════════

func TestAnthropicStreamWithToolUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		events := []string{
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_01","name":"get_rsi"}}`,
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"ticker\":"}}`,
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"\"RELIANCE\"}"}}`,
			`data: {"type":"content_block_stop","index":0}`,
			`data: {"type":"message_delta","delta":{"stop_reason":"tool_use"}}`,
		}
		for _, e := range events {
			fmt.Fprintln(w, e)
			flusher.Flush()
		}
	}))
	defer server.Close()

	p, _ := NewAnthropicProvider("sk-test", WithAnthropicBaseURL(server.URL))
	ch, err := p.ChatStream(context.Background(), []Message{UserMessage("RSI")}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var toolCalls []ToolCall
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatal(chunk.Err)
		}
		toolCalls = append(toolCalls, chunk.ToolCalls...)
	}
	if len(toolCalls) != 1 || toolCalls[0].Name != "get_rsi" {
		t.Fatalf("expected tool call, got: %+v", toolCalls)
	}
	// Verify the arguments were assembled
	var args map[string]string
	json.Unmarshal(toolCalls[0].Arguments, &args)
	if args["ticker"] != "RELIANCE" {
		t.Fatalf("unexpected args: %s", string(toolCalls[0].Arguments))
	}
}

// ════════════════════════════════════════════════════════════════════
// Client custom HTTP
// ════════════════════════════════════════════════════════════════════

func TestOpenAICustomHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	p, _ := NewOpenAIProvider("sk-test", WithOpenAIHTTPClient(custom))
	if p.client != custom {
		t.Fatal("custom client not set")
	}
}

func TestOllamaCustomHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	p, _ := NewOllamaProvider("http://localhost:11434", WithOllamaHTTPClient(custom))
	if p.client != custom {
		t.Fatal("custom client not set")
	}
}

func TestGeminiCustomHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	p, _ := NewGeminiProvider("key", WithGeminiHTTPClient(custom))
	if p.client != custom {
		t.Fatal("custom client not set")
	}
}

func TestAnthropicCustomHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	p, _ := NewAnthropicProvider("key", WithAnthropicHTTPClient(custom))
	if p.client != custom {
		t.Fatal("custom client not set")
	}
}

// ════════════════════════════════════════════════════════════════════
// Ollama HTTP error
// ════════════════════════════════════════════════════════════════════

func TestOllamaHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`model not found`))
	}))
	defer server.Close()

	p, _ := NewOllamaProvider(server.URL)
	_, err := p.Chat(context.Background(), []Message{UserMessage("test")}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected 404 error, got: %v", err)
	}
}

// ════════════════════════════════════════════════════════════════════
// Ollama stream HTTP error
// ════════════════════════════════════════════════════════════════════

func TestOllamaStreamHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal error")
	}))
	defer server.Close()

	p, _ := NewOllamaProvider(server.URL)
	_, err := p.ChatStream(context.Background(), []Message{UserMessage("test")}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected 500 error, got: %v", err)
	}
}
// ════════════════════════════════════════════════════════════════════
// Router LLMProvider interface tests
// ════════════════════════════════════════════════════════════════════

// Compile-time check: Router must satisfy LLMProvider.
var _ LLMProvider = (*Router)(nil)

func TestRouterName(t *testing.T) {
	r := NewRouter("primary")
	r.RegisterProvider(&mockProvider{name: "primary"})

	name := r.Name()
	if name != "router/primary" {
		t.Errorf("Name(): got %q, want %q", name, "router/primary")
	}
}

func TestRouterModels(t *testing.T) {
	r := NewRouter("p1")
	r.RegisterProvider(&mockProvider{name: "p1"})

	models := r.Models()
	if len(models) != 1 || models[0] != "mock-model" {
		t.Errorf("Models(): got %v", models)
	}
}

func TestRouterModelsMultipleProviders(t *testing.T) {
	r := NewRouter("p1")
	r.RegisterProvider(&mockProvider{name: "p1"})
	r.RegisterProvider(&mockProvider{name: "p2"})

	models := r.Models()
	// Both providers return "mock-model" — should be de-duplicated
	if len(models) != 1 {
		t.Errorf("Models() should de-duplicate: got %v", models)
	}
}

func TestRouterModelsMultipleDistinct(t *testing.T) {
	r := NewRouter("p1")
	r.RegisterProvider(&distinctModelProvider{name: "p1", models: []string{"gpt-4", "gpt-3.5"}})
	r.RegisterProvider(&distinctModelProvider{name: "p2", models: []string{"claude-3", "gpt-4"}})

	models := r.Models()
	// "gpt-4" appears in both — should be de-duplicated
	// Expected: gpt-4, gpt-3.5, claude-3 = 3 unique
	if len(models) != 3 {
		t.Errorf("Models() should have 3 unique models, got %d: %v", len(models), models)
	}
}

func TestRouterPing(t *testing.T) {
	r := NewRouter("ok")
	r.RegisterProvider(&mockProvider{name: "ok", pingErr: nil})

	err := r.Ping(context.Background())
	if err != nil {
		t.Errorf("Ping(): got %v, want nil", err)
	}
}

func TestRouterPingError(t *testing.T) {
	r := NewRouter("bad")
	r.RegisterProvider(&mockProvider{name: "bad", pingErr: fmt.Errorf("connection refused")})

	err := r.Ping(context.Background())
	if err == nil {
		t.Error("Ping(): expected error, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("Ping(): got %q, want 'connection refused'", err.Error())
	}
}

func TestRouterPingNoPrimary(t *testing.T) {
	r := NewRouter("missing")
	// No providers registered
	err := r.Ping(context.Background())
	if err == nil {
		t.Error("Ping(): expected error for missing primary, got nil")
	}
}

func TestRouterModelsEmpty(t *testing.T) {
	r := NewRouter("none")
	// No providers registered
	models := r.Models()
	if len(models) != 0 {
		t.Errorf("Models(): expected empty, got %v", models)
	}
}

// distinctModelProvider is a mock with configurable model lists.
type distinctModelProvider struct {
	name   string
	models []string
}

func (d *distinctModelProvider) Name() string    { return d.name }
func (d *distinctModelProvider) Models() []string { return d.models }
func (d *distinctModelProvider) Ping(ctx context.Context) error { return nil }
func (d *distinctModelProvider) Chat(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (*Response, error) {
	return &Response{Content: "ok"}, nil
}
func (d *distinctModelProvider) ChatStream(ctx context.Context, messages []Message, tools []Tool, opts *ChatOptions) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	ch <- StreamChunk{Content: "ok", Done: true}
	close(ch)
	return ch, nil
}