package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/seenimoa/openseai/internal/agent/prompts"
	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Mock LLM Provider
// ════════════════════════════════════════════════════════════════════

// mockProvider implements llm.LLMProvider for testing.
type mockProvider struct {
	name     string
	chatFunc func(ctx context.Context, messages []llm.Message, tools []llm.Tool, opts *llm.ChatOptions) (*llm.Response, error)
	calls    int
	mu       sync.Mutex
}

func newMockProvider(fn func(ctx context.Context, messages []llm.Message, tools []llm.Tool, opts *llm.ChatOptions) (*llm.Response, error)) *mockProvider {
	return &mockProvider{name: "mock", chatFunc: fn}
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Chat(ctx context.Context, messages []llm.Message, tools []llm.Tool, opts *llm.ChatOptions) (*llm.Response, error) {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
	if m.chatFunc != nil {
		return m.chatFunc(ctx, messages, tools, opts)
	}
	return &llm.Response{
		Content:      "Mock analysis complete.",
		FinishReason: llm.FinishStop,
		Usage:        llm.Usage{TotalTokens: 100},
		Model:        "mock-model",
		Provider:     "mock",
	}, nil
}

func (m *mockProvider) ChatStream(_ context.Context, _ []llm.Message, _ []llm.Tool, _ *llm.ChatOptions) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: "streamed", Done: true}
	close(ch)
	return ch, nil
}

func (m *mockProvider) Models() []string { return []string{"mock-model"} }

func (m *mockProvider) HealthCheck(_ context.Context) error { return nil }

func (m *mockProvider) Ping(_ context.Context) error { return nil }

func (m *mockProvider) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// simpleProvider returns a mock that always responds with the given content.
func simpleProvider(content string) *mockProvider {
	return newMockProvider(func(ctx context.Context, msgs []llm.Message, tools []llm.Tool, opts *llm.ChatOptions) (*llm.Response, error) {
		return &llm.Response{
			Content:      content,
			FinishReason: llm.FinishStop,
			Usage:        llm.Usage{TotalTokens: 50},
			Model:        "mock",
			Provider:     "mock",
		}, nil
	})
}

// toolCallingProvider returns a mock that processes one tool call then returns content.
func toolCallingProvider(toolName string, toolResult string, finalContent string) *mockProvider {
	callNum := 0
	var mu sync.Mutex
	return newMockProvider(func(ctx context.Context, msgs []llm.Message, tools []llm.Tool, opts *llm.ChatOptions) (*llm.Response, error) {
		mu.Lock()
		callNum++
		n := callNum
		mu.Unlock()

		if n == 1 {
			// First call: respond with a tool call
			return &llm.Response{
				ToolCalls: []llm.ToolCall{{
					ID:        "tc_1",
					Name:      toolName,
					Arguments: json.RawMessage(`{"ticker": "TCS"}`),
				}},
				FinishReason: llm.FinishToolCalls,
				Usage:        llm.Usage{TotalTokens: 30},
			}, nil
		}
		// Second call: return final answer
		return &llm.Response{
			Content:      finalContent,
			FinishReason: llm.FinishStop,
			Usage:        llm.Usage{TotalTokens: 80},
		}, nil
	})
}

// ════════════════════════════════════════════════════════════════════
// Memory Tests
// ════════════════════════════════════════════════════════════════════

func TestMemoryNew(t *testing.T) {
	m := NewMemory(10)
	if m.Size() != 0 {
		t.Fatalf("new memory should be empty, got %d", m.Size())
	}
	if m.NeedsSummarization() {
		t.Fatal("empty memory should not need summarization")
	}
}

func TestMemoryNewDefaultSize(t *testing.T) {
	m := NewMemory(0)
	if m.maxSize != 50 {
		t.Fatalf("expected default max size 50, got %d", m.maxSize)
	}
}

func TestMemoryAddAndMessages(t *testing.T) {
	m := NewMemory(100)
	m.Add(llm.UserMessage("hello"))
	m.Add(llm.AssistantMessage("hi"))

	msgs := m.Messages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != llm.RoleUser || msgs[0].Content != "hello" {
		t.Fatalf("first message: %+v", msgs[0])
	}
	if msgs[1].Role != llm.RoleAssistant || msgs[1].Content != "hi" {
		t.Fatalf("second message: %+v", msgs[1])
	}
}

func TestMemoryAddAll(t *testing.T) {
	m := NewMemory(100)
	msgs := []llm.Message{
		llm.UserMessage("a"),
		llm.AssistantMessage("b"),
		llm.UserMessage("c"),
	}
	m.AddAll(msgs)
	if m.Size() != 3 {
		t.Fatalf("expected 3, got %d", m.Size())
	}
}

func TestMemoryNeedsSummarization(t *testing.T) {
	m := NewMemory(5)
	for i := 0; i < 6; i++ {
		m.Add(llm.UserMessage(fmt.Sprintf("msg %d", i)))
	}
	if !m.NeedsSummarization() {
		t.Fatal("should need summarization when over max size")
	}
}

func TestMemorySummarize(t *testing.T) {
	m := NewMemory(5)
	for i := 0; i < 10; i++ {
		m.Add(llm.UserMessage(fmt.Sprintf("msg %d", i)))
	}

	err := m.Summarize(context.Background(), 3, func(ctx context.Context, msgs []llm.Message) (string, error) {
		return fmt.Sprintf("Summary of %d messages", len(msgs)), nil
	})
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}

	if m.Size() != 3 {
		t.Fatalf("after summarize, expected 3 messages, got %d", m.Size())
	}

	msgs := m.Messages()
	// Should have summary as first message + 3 recent
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages (summary + 3), got %d", len(msgs))
	}
	if !strings.Contains(msgs[0].Content, "Summary of") {
		t.Fatalf("first message should be summary: %s", msgs[0].Content)
	}
}

func TestMemorySummarizeError(t *testing.T) {
	m := NewMemory(5)
	for i := 0; i < 10; i++ {
		m.Add(llm.UserMessage(fmt.Sprintf("msg %d", i)))
	}

	err := m.Summarize(context.Background(), 3, func(ctx context.Context, msgs []llm.Message) (string, error) {
		return "", fmt.Errorf("summarizer failed")
	})
	if err == nil || !strings.Contains(err.Error(), "summarizer failed") {
		t.Fatalf("expected summarizer error, got: %v", err)
	}
	// Messages should be unchanged on error
	if m.Size() != 10 {
		t.Fatalf("messages should be unchanged on error, got %d", m.Size())
	}
}

func TestMemoryClear(t *testing.T) {
	m := NewMemory(100)
	m.Add(llm.UserMessage("test"))
	m.Clear()
	if m.Size() != 0 {
		t.Fatal("clear should empty memory")
	}
}

func TestMemoryConcurrency(t *testing.T) {
	m := NewMemory(1000)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.Add(llm.UserMessage(fmt.Sprintf("msg %d", n)))
			m.Size()
			m.Messages()
			m.NeedsSummarization()
		}(i)
	}
	wg.Wait()
	if m.Size() != 100 {
		t.Fatalf("expected 100 messages, got %d", m.Size())
	}
}

// ════════════════════════════════════════════════════════════════════
// BaseAgent Tests
// ════════════════════════════════════════════════════════════════════

func TestBaseAgentDefaults(t *testing.T) {
	provider := simpleProvider("test response")
	agent := NewBaseAgent(BaseAgentConfig{
		Name:     "test-agent",
		Role:     "Test Role",
		Provider: provider,
	})

	if agent.Name() != "test-agent" {
		t.Fatalf("Name: got %q", agent.Name())
	}
	if agent.Role() != "Test Role" {
		t.Fatalf("Role: got %q", agent.Role())
	}
	if len(agent.Tools()) != 0 {
		t.Fatalf("Tools: expected 0, got %d", len(agent.Tools()))
	}
	if agent.Provider() == nil {
		t.Fatal("Provider should not be nil")
	}
	if agent.Memory() == nil {
		t.Fatal("Memory should not be nil")
	}
}

func TestBaseAgentProcess(t *testing.T) {
	provider := simpleProvider("Analysis: TCS is bullish. Recommendation: BUY.")
	agent := NewBaseAgent(BaseAgentConfig{
		Name:         "test-agent",
		Role:         "Test",
		SystemPrompt: "You are a test agent.",
		Provider:     provider,
	})

	result, err := agent.Process(context.Background(), "Analyze TCS")
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	if result.AgentName != "test-agent" {
		t.Fatalf("AgentName: got %q", result.AgentName)
	}
	if result.Content != "Analysis: TCS is bullish. Recommendation: BUY." {
		t.Fatalf("Content: got %q", result.Content)
	}
	if result.ToolCalls != 0 {
		t.Fatalf("ToolCalls: expected 0, got %d", result.ToolCalls)
	}
	if result.Duration <= 0 {
		t.Fatal("Duration should be positive")
	}
	if len(result.Messages) == 0 {
		t.Fatal("Messages should not be empty")
	}
}

func TestBaseAgentProcessWithHistory(t *testing.T) {
	provider := simpleProvider("Follow-up response.")
	agent := NewBaseAgent(BaseAgentConfig{
		Name:         "test-agent",
		Role:         "Test",
		SystemPrompt: "System.",
		Provider:     provider,
	})

	history := []llm.Message{
		llm.UserMessage("previous question"),
		llm.AssistantMessage("previous answer"),
	}

	result, err := agent.ProcessWithMessages(context.Background(), "follow up", history)
	if err != nil {
		t.Fatalf("ProcessWithMessages: %v", err)
	}
	if result.Content != "Follow-up response." {
		t.Fatalf("Content: %q", result.Content)
	}
}

func TestBaseAgentProcessWithTools(t *testing.T) {
	provider := toolCallingProvider("get_price", "₹3500", "TCS is at ₹3500.")

	tools := []llm.Tool{{
		Name:        "get_price",
		Description: "Get stock price",
		Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			return "₹3500", nil
		},
	}}

	agent := NewBaseAgent(BaseAgentConfig{
		Name:         "test-agent",
		Role:         "Test",
		SystemPrompt: "You are a test agent.",
		Provider:     provider,
		Tools:        tools,
	})

	result, err := agent.Process(context.Background(), "What is the price of TCS?")
	if err != nil {
		t.Fatalf("Process with tools: %v", err)
	}

	if result.Content != "TCS is at ₹3500." {
		t.Fatalf("Content: %q", result.Content)
	}
	if result.ToolCalls == 0 {
		t.Fatal("should have at least one tool call")
	}
	if provider.callCount() != 2 {
		t.Fatalf("expected 2 LLM calls, got %d", provider.callCount())
	}
}

func TestBaseAgentProcessError(t *testing.T) {
	provider := newMockProvider(func(ctx context.Context, msgs []llm.Message, tools []llm.Tool, opts *llm.ChatOptions) (*llm.Response, error) {
		return nil, fmt.Errorf("provider error")
	})

	agent := NewBaseAgent(BaseAgentConfig{
		Name:     "err-agent",
		Role:     "Test",
		Provider: provider,
	})

	result, err := agent.Process(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error")
	}
	if result.Error != "provider error" {
		t.Fatalf("Error: got %q", result.Error)
	}
	if result.AgentName != "err-agent" {
		t.Fatalf("AgentName: got %q", result.AgentName)
	}
}

func TestBaseAgentMemoryGrowth(t *testing.T) {
	provider := simpleProvider("response")
	agent := NewBaseAgent(BaseAgentConfig{
		Name:     "mem-agent",
		Role:     "Test",
		Provider: provider,
	})

	// Process multiple queries
	for i := 0; i < 5; i++ {
		_, err := agent.Process(context.Background(), fmt.Sprintf("query %d", i))
		if err != nil {
			t.Fatalf("Process %d: %v", i, err)
		}
	}

	// Memory should have grown (user + assistant messages per call, minus system)
	if agent.Memory().Size() == 0 {
		t.Fatal("memory should have messages after processing")
	}
}

// ════════════════════════════════════════════════════════════════════
// ParseAnalysisResult Tests
// ════════════════════════════════════════════════════════════════════

func TestParseAnalysisResultJSON(t *testing.T) {
	content := `Based on the analysis, here's my recommendation:
{"ticker": "RELIANCE", "recommendation": "BUY", "confidence": 0.85, "summary": "Strong fundamentals"}
The stock shows bullish momentum.`

	defaults := models.AnalysisResult{
		Ticker: "RELIANCE",
		Type:   models.AnalysisFundamental,
	}

	result := ParseAnalysisResult(content, defaults)
	if result.Recommendation != "BUY" {
		t.Fatalf("Recommendation: got %q", result.Recommendation)
	}
	if result.Confidence != 0.85 {
		t.Fatalf("Confidence: got %f", result.Confidence)
	}
	if result.Summary != "Strong fundamentals" {
		t.Fatalf("Summary: got %q", result.Summary)
	}
	if result.Type != models.AnalysisFundamental {
		t.Fatalf("Type should come from defaults: got %q", result.Type)
	}
}

func TestParseAnalysisResultNoJSON(t *testing.T) {
	content := "This is a plain text analysis without any JSON."
	defaults := models.AnalysisResult{
		Ticker:    "TCS",
		Type:      models.AnalysisTechnical,
		AgentName: "technical",
	}

	result := ParseAnalysisResult(content, defaults)
	if result.Ticker != "TCS" {
		t.Fatalf("Ticker should come from defaults: got %q", result.Ticker)
	}
	if result.Summary != content {
		t.Fatalf("Summary should be full content for plain text: got %q", result.Summary)
	}
}

func TestParseAnalysisResultPartialJSON(t *testing.T) {
	content := `{"recommendation": "HOLD", "confidence": 0.6}`
	defaults := models.AnalysisResult{
		Ticker: "INFY",
		Type:   models.AnalysisFundamental,
	}

	result := ParseAnalysisResult(content, defaults)
	if result.Recommendation != "HOLD" {
		t.Fatalf("Recommendation: got %q", result.Recommendation)
	}
	if result.Ticker != "INFY" {
		t.Fatalf("Ticker should keep default: got %q", result.Ticker)
	}
}

// ════════════════════════════════════════════════════════════════════
// Agent Registry Tests
// ════════════════════════════════════════════════════════════════════

func TestRegistryEmpty(t *testing.T) {
	r := NewRegistry()
	if r.Count() != 0 {
		t.Fatal("new registry should be empty")
	}
	if len(r.Names()) != 0 {
		t.Fatal("Names should be empty")
	}
	if len(r.List()) != 0 {
		t.Fatal("List should be empty")
	}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	agent := NewBaseAgent(BaseAgentConfig{
		Name:     "test-agent",
		Role:     "Test",
		Provider: simpleProvider("ok"),
	})

	r.Register(agent)
	if r.Count() != 1 {
		t.Fatalf("Count: got %d", r.Count())
	}

	got, ok := r.Get("test-agent")
	if !ok || got.Name() != "test-agent" {
		t.Fatalf("Get: ok=%v, agent=%+v", ok, got)
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Fatal("should not find nonexistent agent")
	}
}

func TestRegistryOverwrite(t *testing.T) {
	r := NewRegistry()
	r.Register(NewBaseAgent(BaseAgentConfig{Name: "a", Role: "v1", Provider: simpleProvider("")}))
	r.Register(NewBaseAgent(BaseAgentConfig{Name: "a", Role: "v2", Provider: simpleProvider("")}))

	if r.Count() != 1 {
		t.Fatal("should overwrite same-name agent")
	}
	agent, _ := r.Get("a")
	if agent.Role() != "v2" {
		t.Fatalf("should have v2, got %q", agent.Role())
	}
}

func TestRegistryConcurrency(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := fmt.Sprintf("agent_%d", n)
			r.Register(NewBaseAgent(BaseAgentConfig{Name: name, Role: "Test", Provider: simpleProvider("")}))
			r.Get(name)
			r.Names()
			r.List()
			r.Count()
		}(i)
	}
	wg.Wait()
	if r.Count() != 50 {
		t.Fatalf("expected 50, got %d", r.Count())
	}
}

// ════════════════════════════════════════════════════════════════════
// Specialized Agent Construction Tests
// ════════════════════════════════════════════════════════════════════

func TestFundamentalAgentCreation(t *testing.T) {
	provider := simpleProvider("analysis")
	agent := NewFundamentalAgent(provider, newMockSources(), nil)

	if agent.Name() != prompts.AgentFundamental {
		t.Fatalf("Name: got %q", agent.Name())
	}
	if !strings.Contains(agent.Role(), "Fundamental") {
		t.Fatalf("Role: got %q", agent.Role())
	}
	if !strings.Contains(strings.ToLower(agent.SystemPrompt()), "fundamental") {
		t.Fatalf("SystemPrompt should contain 'fundamental', got: %.100s...", agent.SystemPrompt())
	}
	if len(agent.Tools()) == 0 {
		t.Fatal("should have tools")
	}

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range agent.Tools() {
		toolNames[tool.Name] = true
	}
	expectedTools := []string{"get_financials", "get_stock_quote", "compute_ratios", "compute_valuation", "get_peer_comparison", "get_stock_profile"}
	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Fatalf("missing tool: %s", name)
		}
	}
}

func TestTechnicalAgentCreation(t *testing.T) {
	agent := NewTechnicalAgent(simpleProvider(""), newMockSources(), nil)

	if agent.Name() != prompts.AgentTechnical {
		t.Fatalf("Name: got %q", agent.Name())
	}

	toolNames := toolNameSet(agent.Tools())
	for _, name := range []string{"get_historical_data", "compute_indicators", "generate_signals", "full_technical_analysis", "get_quote"} {
		if !toolNames[name] {
			t.Fatalf("missing tool: %s", name)
		}
	}
}

func TestSentimentAgentCreation(t *testing.T) {
	agent := NewSentimentAgent(simpleProvider(""), nil, nil)

	if agent.Name() != prompts.AgentSentiment {
		t.Fatalf("Name: got %q", agent.Name())
	}

	toolNames := toolNameSet(agent.Tools())
	for _, name := range []string{"get_stock_news", "get_market_news", "get_sector_news", "analyze_sentiment", "score_headline"} {
		if !toolNames[name] {
			t.Fatalf("missing tool: %s", name)
		}
	}
}

func TestFnOAgentCreation(t *testing.T) {
	agent := NewFnOAgent(simpleProvider(""), nil, nil, nil)

	if agent.Name() != prompts.AgentFnO {
		t.Fatalf("Name: got %q", agent.Name())
	}

	toolNames := toolNameSet(agent.Tools())
	for _, name := range []string{"get_option_chain", "analyze_option_chain", "compute_pcr", "analyze_oi_buildup", "get_futures_data", "get_india_vix", "full_derivatives_analysis"} {
		if !toolNames[name] {
			t.Fatalf("missing tool: %s", name)
		}
	}
}

func TestRiskAgentCreation(t *testing.T) {
	agent := NewRiskAgent(simpleProvider(""), newMockSources(), nil)

	if agent.Name() != prompts.AgentRisk {
		t.Fatalf("Name: got %q", agent.Name())
	}

	toolNames := toolNameSet(agent.Tools())
	for _, name := range []string{"compute_position_size", "compute_var", "suggest_stop_loss", "risk_reward_analysis", "portfolio_exposure_check"} {
		if !toolNames[name] {
			t.Fatalf("missing tool: %s", name)
		}
	}
}

func TestExecutorAgentCreation(t *testing.T) {
	agent := NewExecutorAgent(simpleProvider(""), nil)

	if agent.Name() != prompts.AgentExecutor {
		t.Fatalf("Name: got %q", agent.Name())
	}

	toolNames := toolNameSet(agent.Tools())
	for _, name := range []string{"create_trade_proposal", "estimate_brokerage", "validate_trade"} {
		if !toolNames[name] {
			t.Fatalf("missing tool: %s", name)
		}
	}
}

func TestReporterAgentCreation(t *testing.T) {
	agent := NewReporterAgent(simpleProvider(""), nil)

	if agent.Name() != prompts.AgentReporter {
		t.Fatalf("Name: got %q", agent.Name())
	}

	toolNames := toolNameSet(agent.Tools())
	for _, name := range []string{"format_report", "create_summary_table"} {
		if !toolNames[name] {
			t.Fatalf("missing tool: %s", name)
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// Tool Handler Tests (direct invocations)
// ════════════════════════════════════════════════════════════════════

func TestFundamentalHandleGetQuote(t *testing.T) {
	agent := NewFundamentalAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS"}`)
	result, err := agent.handleGetQuote(context.Background(), args)
	if err != nil {
		t.Fatalf("handleGetQuote: %v", err)
	}
	if !strings.Contains(result, "TCS") || !strings.Contains(result, "3500") {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestFundamentalHandleGetFinancials(t *testing.T) {
	agent := NewFundamentalAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS"}`)
	result, err := agent.handleGetFinancials(context.Background(), args)
	if err != nil {
		t.Fatalf("handleGetFinancials: %v", err)
	}
	if !strings.Contains(result, "TCS") {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestFundamentalHandleGetProfile(t *testing.T) {
	agent := NewFundamentalAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS"}`)
	result, err := agent.handleGetProfile(context.Background(), args)
	if err != nil {
		t.Fatalf("handleGetProfile: %v", err)
	}
	if !strings.Contains(result, "TCS") {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestFundamentalHandlePeerComparison(t *testing.T) {
	agent := NewFundamentalAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS"}`)
	result, err := agent.handlePeerComparison(context.Background(), args)
	if err != nil {
		t.Fatalf("handlePeerComparison: %v", err)
	}
	// TCS is in IT sector, should find peers
	if !strings.Contains(result, "IT") {
		t.Fatalf("should show IT sector: %s", result)
	}
}

func TestFundamentalHandleComputeRatios(t *testing.T) {
	agent := NewFundamentalAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS", "price": 3500, "shares_outstanding": 366000000}`)
	result, err := agent.handleComputeRatios(context.Background(), args)
	if err != nil {
		t.Fatalf("handleComputeRatios: %v", err)
	}
	if !strings.Contains(result, "ratios") {
		t.Fatalf("should contain ratios: %s", result)
	}
}

func TestTechnicalHandleGetHistorical(t *testing.T) {
	agent := NewTechnicalAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS", "days": 30}`)
	result, err := agent.handleGetHistorical(context.Background(), args)
	if err != nil {
		t.Fatalf("handleGetHistorical: %v", err)
	}
	if !strings.Contains(result, "TCS") || !strings.Contains(result, "candles") {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestTechnicalHandleComputeIndicators(t *testing.T) {
	agent := NewTechnicalAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS", "days": 30}`)
	result, err := agent.handleComputeIndicators(context.Background(), args)
	if err != nil {
		t.Fatalf("handleComputeIndicators: %v", err)
	}
	if !strings.Contains(result, "TCS") {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestTechnicalHandleGenerateSignals(t *testing.T) {
	agent := NewTechnicalAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS", "days": 30}`)
	result, err := agent.handleGenerateSignals(context.Background(), args)
	if err != nil {
		t.Fatalf("handleGenerateSignals: %v", err)
	}
	if !strings.Contains(result, "TCS") {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestTechnicalHandleFullAnalysis(t *testing.T) {
	agent := NewTechnicalAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS", "days": 30}`)
	result, err := agent.handleFullAnalysis(context.Background(), args)
	if err != nil {
		t.Fatalf("handleFullAnalysis: %v", err)
	}
	if !strings.Contains(result, "TCS") {
		t.Fatalf("unexpected result: %s", result)
	}
}

// ════════════════════════════════════════════════════════════════════
// Risk Agent Tool Handler Tests
// ════════════════════════════════════════════════════════════════════

func TestRiskHandlePositionSize(t *testing.T) {
	agent := NewRiskAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"capital": 1000000, "entry_price": 3500, "stop_loss": 3400, "risk_pct": 2.0}`)
	result, err := agent.handlePositionSize(context.Background(), args)
	if err != nil {
		t.Fatalf("handlePositionSize: %v", err)
	}
	if !strings.Contains(result, "quantity") {
		t.Fatalf("should contain quantity: %s", result)
	}
	// With ₹10L capital, 2% risk = ₹20,000 risk. SL distance = ₹100. Qty = 200
	if !strings.Contains(result, "200") {
		t.Fatalf("expected qty ~200: %s", result)
	}
}

func TestRiskHandlePositionSizeZeroSL(t *testing.T) {
	agent := NewRiskAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"capital": 1000000, "entry_price": 3500, "stop_loss": 3500}`)
	result, err := agent.handlePositionSize(context.Background(), args)
	if err != nil {
		t.Fatalf("handlePositionSize: %v", err)
	}
	if !strings.Contains(result, "zero") {
		t.Fatalf("should warn about zero SL distance: %s", result)
	}
}

func TestRiskHandlePositionSizeCapLimit(t *testing.T) {
	agent := NewRiskAgent(simpleProvider(""), newMockSources(), nil)

	// Very low SL distance → would result in huge position -> capped at 5%
	args := json.RawMessage(`{"capital": 1000000, "entry_price": 100, "stop_loss": 99.99, "risk_pct": 2.0}`)
	result, err := agent.handlePositionSize(context.Background(), args)
	if err != nil {
		t.Fatalf("handlePositionSize: %v", err)
	}
	// Position should be capped at 5% = ₹50,000 / ₹100 = 500 shares
	if !strings.Contains(result, "500") {
		t.Fatalf("position should be capped: %s", result)
	}
}

func TestRiskHandleRiskReward(t *testing.T) {
	agent := NewRiskAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS", "entry_price": 3500, "stop_loss": 3400, "target": 3800, "capital": 1000000}`)
	result, err := agent.handleRiskReward(context.Background(), args)
	if err != nil {
		t.Fatalf("handleRiskReward: %v", err)
	}
	// Risk = 100, Reward = 300. R:R = 1:3.00 → FAVORABLE
	if !strings.Contains(result, "FAVORABLE") {
		t.Fatalf("should be FAVORABLE with 1:3 R:R: %s", result)
	}
	if !strings.Contains(result, "1:3.00") {
		t.Fatalf("should show 1:3.00 ratio: %s", result)
	}
}

func TestRiskHandleRiskRewardReject(t *testing.T) {
	agent := NewRiskAgent(simpleProvider(""), newMockSources(), nil)

	// Risk = 100, Reward = 50. R:R = 1:0.50 → REJECT
	args := json.RawMessage(`{"entry_price": 3500, "stop_loss": 3400, "target": 3550}`)
	result, err := agent.handleRiskReward(context.Background(), args)
	if err != nil {
		t.Fatalf("handleRiskReward: %v", err)
	}
	if !strings.Contains(result, "REJECT") {
		t.Fatalf("should REJECT low R:R: %s", result)
	}
}

func TestRiskHandleVaR(t *testing.T) {
	agent := NewRiskAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS", "position_size": 350000, "confidence": 0.95}`)
	result, err := agent.handleComputeVaR(context.Background(), args)
	if err != nil {
		t.Fatalf("handleComputeVaR: %v", err)
	}
	if !strings.Contains(result, "var_amount") {
		t.Fatalf("should contain var_amount: %s", result)
	}
}

func TestRiskHandleSuggestStopLoss(t *testing.T) {
	agent := NewRiskAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{"ticker": "TCS", "entry_price": 3500, "direction": "long"}`)
	result, err := agent.handleSuggestStopLoss(context.Background(), args)
	if err != nil {
		t.Fatalf("handleSuggestStopLoss: %v", err)
	}
	if !strings.Contains(result, "tight") || !strings.Contains(result, "moderate") || !strings.Contains(result, "wide") {
		t.Fatalf("should contain 3 SL levels: %s", result)
	}
}

func TestRiskHandlePortfolioExposure(t *testing.T) {
	agent := NewRiskAgent(simpleProvider(""), newMockSources(), nil)

	args := json.RawMessage(`{
		"positions": [{"ticker": "TCS", "value": 200000}, {"ticker": "INFY", "value": 150000}],
		"new_ticker": "HCLTECH",
		"new_value": 100000,
		"capital": 1000000
	}`)
	result, err := agent.handlePortfolioExposure(context.Background(), args)
	if err != nil {
		t.Fatalf("handlePortfolioExposure: %v", err)
	}
	if !strings.Contains(result, "exposure_pct") {
		t.Fatalf("should contain exposure_pct: %s", result)
	}
}

func TestRiskHandlePortfolioExposureWarnings(t *testing.T) {
	agent := NewRiskAgent(simpleProvider(""), newMockSources(), nil)

	// New position is 10% of capital → should trigger warning
	args := json.RawMessage(`{
		"positions": [],
		"new_ticker": "TCS",
		"new_value": 100000,
		"capital": 1000000
	}`)
	result, err := agent.handlePortfolioExposure(context.Background(), args)
	if err != nil {
		t.Fatalf("handlePortfolioExposure: %v", err)
	}
	if !strings.Contains(result, "exceeds 5%") {
		t.Fatalf("should warn about 10%% position: %s", result)
	}
}

// ════════════════════════════════════════════════════════════════════
// Executor Agent Tool Handler Tests
// ════════════════════════════════════════════════════════════════════

func TestExecutorHandleCreateProposal(t *testing.T) {
	agent := NewExecutorAgent(simpleProvider(""), nil)

	args := json.RawMessage(`{
		"ticker": "TCS",
		"action": "BUY",
		"price": 3500,
		"stop_loss": 3400,
		"target": 3800,
		"quantity": 100,
		"rationale": "Strong technicals and fundamentals"
	}`)
	result, err := agent.handleCreateProposal(context.Background(), args)
	if err != nil {
		t.Fatalf("handleCreateProposal: %v", err)
	}
	if !strings.Contains(result, "REQUIRES HUMAN APPROVAL") {
		t.Fatalf("should require approval: %s", result)
	}
	if !strings.Contains(result, "TCS") || !strings.Contains(result, "BUY") {
		t.Fatalf("should contain ticker and action: %s", result)
	}
	if !strings.Contains(result, "1:3.00") {
		t.Fatalf("should compute R:R: %s", result)
	}
}

func TestExecutorHandleCreateProposalDefaultOrderType(t *testing.T) {
	agent := NewExecutorAgent(simpleProvider(""), nil)

	args := json.RawMessage(`{"ticker": "TCS", "action": "BUY", "price": 3500}`)
	result, err := agent.handleCreateProposal(context.Background(), args)
	if err != nil {
		t.Fatalf("handleCreateProposal: %v", err)
	}
	if !strings.Contains(result, "LIMIT") {
		t.Fatalf("default order type should be LIMIT: %s", result)
	}
}

func TestExecutorHandleEstimateBrokerage(t *testing.T) {
	agent := NewExecutorAgent(simpleProvider(""), nil)

	args := json.RawMessage(`{"buy_price": 3500, "sell_price": 3800, "quantity": 100, "is_delivery": true}`)
	result, err := agent.handleEstimateBrokerage(context.Background(), args)
	if err != nil {
		t.Fatalf("handleEstimateBrokerage: %v", err)
	}
	if !strings.Contains(strings.ToLower(result), "stt") || !strings.Contains(strings.ToLower(result), "total") {
		t.Fatalf("should contain brokerage details: %s", result)
	}
}

func TestExecutorHandleValidateTrade(t *testing.T) {
	agent := NewExecutorAgent(simpleProvider(""), nil)

	// Valid trade: position = ₹3,50,000 which is 3.5% of ₹1Cr
	args := json.RawMessage(`{
		"ticker": "TCS",
		"action": "BUY",
		"price": 3500,
		"quantity": 100,
		"capital": 10000000,
		"stop_loss": 3400
	}`)
	result, err := agent.handleValidateTrade(context.Background(), args)
	if err != nil {
		t.Fatalf("handleValidateTrade: %v", err)
	}
	if !strings.Contains(result, `"valid": true`) {
		t.Fatalf("should be valid: %s", result)
	}
}

func TestExecutorHandleValidateTradeReject(t *testing.T) {
	agent := NewExecutorAgent(simpleProvider(""), nil)

	// Position = ₹3,50,000 which is 35% of ₹10L → exceeds 5%
	args := json.RawMessage(`{
		"ticker": "TCS",
		"action": "BUY",
		"price": 3500,
		"quantity": 100,
		"capital": 1000000,
		"stop_loss": 3400
	}`)
	result, err := agent.handleValidateTrade(context.Background(), args)
	if err != nil {
		t.Fatalf("handleValidateTrade: %v", err)
	}
	if !strings.Contains(result, `"valid": false`) {
		t.Fatalf("should be invalid (position too large): %s", result)
	}
	if !strings.Contains(result, "POSITION_SIZE") {
		t.Fatalf("should have position size issue: %s", result)
	}
}

// ════════════════════════════════════════════════════════════════════
// Reporter Agent Tool Handler Tests
// ════════════════════════════════════════════════════════════════════

func TestReporterHandleFormatReport(t *testing.T) {
	agent := NewReporterAgent(simpleProvider(""), nil)

	args := json.RawMessage(`{
		"ticker": "TCS",
		"title": "TCS Investment Analysis",
		"recommendation": "BUY",
		"target_price": 3800,
		"stop_loss": 3400,
		"timeframe": "medium-term",
		"sections": [
			{"heading": "Executive Summary", "content": "TCS shows strong momentum."},
			{"heading": "Technical Analysis", "content": "RSI at 62, MACD bullish crossover."}
		]
	}`)
	result, err := agent.handleFormatReport(context.Background(), args)
	if err != nil {
		t.Fatalf("handleFormatReport: %v", err)
	}
	if !strings.Contains(result, "TCS Investment Analysis") {
		t.Fatal("should contain title")
	}
	if !strings.Contains(result, "BUY") {
		t.Fatal("should contain recommendation")
	}
	if !strings.Contains(result, "₹3800") {
		t.Fatal("should contain target price")
	}
	if !strings.Contains(result, "Disclaimer") {
		t.Fatal("should contain disclaimer")
	}
}

func TestReporterHandleCreateTable(t *testing.T) {
	agent := NewReporterAgent(simpleProvider(""), nil)

	args := json.RawMessage(`{
		"title": "Peer Comparison",
		"headers": ["Ticker", "PE", "ROE"],
		"rows": [
			["TCS", "32.5", "45.2%"],
			["INFY", "28.1", "33.8%"]
		]
	}`)
	result, err := agent.handleCreateTable(context.Background(), args)
	if err != nil {
		t.Fatalf("handleCreateTable: %v", err)
	}
	if !strings.Contains(result, "Peer Comparison") {
		t.Fatal("should contain title")
	}
	if !strings.Contains(result, "| Ticker") {
		t.Fatal("should have markdown table")
	}
	if !strings.Contains(result, "TCS") || !strings.Contains(result, "INFY") {
		t.Fatal("should contain both rows")
	}
}

// ════════════════════════════════════════════════════════════════════
// Orchestrator Tests
// ════════════════════════════════════════════════════════════════════

func TestExtractTicker(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{"Analyze TCS for investment", "TCS"},
		{"What is RSI of RELIANCE?", "RSI"}, // extractTicker picks first alpha token
		{"Full analysis of INFY", "INFY"},
		{"How is the market today?", ""},
		{"Buy HDFCBANK at 1600", "HDFCBANK"},
	}

	for _, tt := range tests {
		got := extractTicker(tt.query)
		if got != tt.expected {
			t.Errorf("extractTicker(%q) = %q, want %q", tt.query, got, tt.expected)
		}
	}
}

func TestIsAlpha(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"TCS", true},
		{"RELIANCE", true},
		{"123", false},
		{"TCS123", false},
		{"tcs", true},
		{"", true},
	}

	for _, tt := range tests {
		got := isAlpha(tt.input)
		if got != tt.expected {
			t.Errorf("isAlpha(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestSortFloat64s(t *testing.T) {
	s := []float64{5.0, 2.0, 8.0, 1.0, 4.0}
	sortFloat64s(s)
	for i := 1; i < len(s); i++ {
		if s[i] < s[i-1] {
			t.Fatalf("not sorted: %v", s)
		}
	}
}

func TestComputeATR(t *testing.T) {
	candles := make([]models.OHLCV, 20)
	base := 100.0
	for i := range candles {
		candles[i] = models.OHLCV{
			Open:  base + float64(i),
			High:  base + float64(i) + 5,
			Low:   base + float64(i) - 3,
			Close: base + float64(i) + 1,
		}
	}

	atr := computeATR(candles, 14)
	if atr <= 0 {
		t.Fatalf("ATR should be positive: %f", atr)
	}
	// With consistent ranges of ~8 (high-low), ATR should be around 8
	if atr < 5 || atr > 12 {
		t.Fatalf("ATR outside expected range: %f", atr)
	}
}

func TestComputeATRInsufficientData(t *testing.T) {
	candles := make([]models.OHLCV, 5) // too few for period=14
	atr := computeATR(candles, 14)
	if atr != 0 {
		t.Fatalf("ATR should be 0 with insufficient data: %f", atr)
	}
}

func TestFilterATMContracts(t *testing.T) {
	oc := &models.OptionChain{
		SpotPrice: 3500,
		Contracts: []models.OptionContract{
			{StrikePrice: 3000, OptionType: "CE"}, // too far
			{StrikePrice: 3400, OptionType: "CE"}, // within 5%
			{StrikePrice: 3500, OptionType: "CE"}, // ATM
			{StrikePrice: 3500, OptionType: "PE"}, // ATM
			{StrikePrice: 3600, OptionType: "PE"}, // within 5%
			{StrikePrice: 4000, OptionType: "PE"}, // too far
		},
	}

	filtered := filterATMContracts(oc)
	if len(filtered) != 4 {
		t.Fatalf("expected 4 ATM contracts, got %d", len(filtered))
	}
}

func TestFormatArticles(t *testing.T) {
	articles := []models.NewsArticle{
		{
			Title:       "TCS reports strong Q1",
			Source:      "Moneycontrol",
			PublishedAt: time.Date(2024, 7, 10, 14, 30, 0, 0, time.UTC),
			Summary:     "Revenue up 8%",
			URL:         "https://example.com/1",
		},
		{
			Title:       "IT sector rally continues",
			Source:      "ET",
			PublishedAt: time.Date(2024, 7, 10, 15, 0, 0, 0, time.UTC),
		},
	}

	items := formatArticles(articles)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0]["title"] != "TCS reports strong Q1" {
		t.Fatalf("title: %q", items[0]["title"])
	}
	if items[0]["summary"] != "Revenue up 8%" {
		t.Fatalf("summary: %q", items[0]["summary"])
	}
	if items[1]["summary"] != "" {
		t.Fatalf("second item should have no summary: %q", items[1]["summary"])
	}
}

func TestAbsHelper(t *testing.T) {
	tests := []struct {
		in, out float64
	}{
		{5.0, 5.0},
		{-3.0, 3.0},
		{0.0, 0.0},
	}
	for _, tt := range tests {
		got := abs(tt.in)
		if got != tt.out {
			t.Errorf("abs(%f) = %f, want %f", tt.in, got, tt.out)
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// Prompts Tests
// ════════════════════════════════════════════════════════════════════

func TestSystemPromptsExist(t *testing.T) {
	promtps := []string{
		prompts.FundamentalSystemPrompt,
		prompts.TechnicalSystemPrompt,
		prompts.SentimentSystemPrompt,
		prompts.FnOSystemPrompt,
		prompts.RiskSystemPrompt,
		prompts.ExecutorSystemPrompt,
		prompts.ReporterSystemPrompt,
		prompts.CIOSystemPrompt,
	}

	for i, p := range promtps {
		if len(p) < 100 {
			t.Fatalf("prompt %d is too short (%d chars)", i, len(p))
		}
	}
}

func TestAgentNameConstants(t *testing.T) {
	names := []string{
		prompts.AgentFundamental,
		prompts.AgentTechnical,
		prompts.AgentSentiment,
		prompts.AgentFnO,
		prompts.AgentRisk,
		prompts.AgentExecutor,
		prompts.AgentReporter,
		prompts.AgentCIO,
	}

	seen := make(map[string]bool)
	for _, n := range names {
		if n == "" {
			t.Fatal("agent name should not be empty")
		}
		if seen[n] {
			t.Fatalf("duplicate agent name: %q", n)
		}
		seen[n] = true
	}
}

func TestCoTTemplates(t *testing.T) {
	templates := []struct {
		name   string
		output string
	}{
		{"CoTAnalysis", prompts.CoTAnalysis("TCS", "test task")},
		{"CoTFundamental", prompts.CoTFundamental("TCS")},
		{"CoTTechnical", prompts.CoTTechnical("TCS")},
		{"CoTDerivatives", prompts.CoTDerivatives("TCS")},
		{"CoTRisk", prompts.CoTRisk("TCS", 1000000)},
		{"CoTSynthesis", prompts.CoTSynthesis("TCS")},
	}

	for _, tt := range templates {
		if !strings.Contains(tt.output, "TCS") {
			t.Fatalf("%s should contain ticker TCS", tt.name)
		}
		if len(tt.output) < 50 {
			t.Fatalf("%s is too short (%d chars)", tt.name, len(tt.output))
		}
	}
}

func TestSectorForTicker(t *testing.T) {
	tests := []struct {
		ticker string
		sector string
	}{
		{"TCS", "IT"},
		{"HDFCBANK", "Banking"},
		{"RELIANCE", "Oil & Gas"},
		{"NONEXISTENT", ""},
	}

	for _, tt := range tests {
		got := prompts.SectorForTicker(tt.ticker)
		if got != tt.sector {
			t.Errorf("SectorForTicker(%q) = %q, want %q", tt.ticker, got, tt.sector)
		}
	}
}

func TestSectorPeers(t *testing.T) {
	peers := prompts.SectorPeers("TCS")
	if len(peers) == 0 {
		t.Fatal("TCS should have peers")
	}
	for _, p := range peers {
		if p == "TCS" {
			t.Fatal("peers should not include the ticker itself")
		}
	}
}

func TestSectorPeersUnknown(t *testing.T) {
	peers := prompts.SectorPeers("UNKNOWN123")
	if len(peers) != 0 {
		t.Fatalf("unknown ticker should have no peers, got %v", peers)
	}
}

func TestFormatTickerPrompt(t *testing.T) {
	result := prompts.FormatTickerPrompt("TCS")
	if !strings.Contains(result, "TCS") {
		t.Fatal("should contain ticker")
	}
	if !strings.Contains(result, "IT") {
		t.Fatal("should contain sector for TCS")
	}
}

func TestIndianMarketPromptSuffix(t *testing.T) {
	suffix := prompts.IndianMarketPromptSuffix()
	if !strings.Contains(suffix, "NSE") {
		t.Fatal("should contain NSE reference")
	}
	if !strings.Contains(suffix, "₹") {
		t.Fatal("should contain ₹ symbol")
	}
}

func TestIndianBrokerageEstimate(t *testing.T) {
	result := prompts.IndianBrokerageEstimate(3500.0, 3800.0, 100, true)
	if len(result) == 0 {
		t.Fatal("result should not be empty")
	}
	if !strings.Contains(result, "stt") && !strings.Contains(result, "STT") {
		t.Fatal("should contain STT reference")
	}
	if !strings.Contains(strings.ToLower(result), "total") {
		t.Fatal("should contain total")
	}
}

// ════════════════════════════════════════════════════════════════════
// Integration-style tests (mock provider with tool calling)
// ════════════════════════════════════════════════════════════════════

func TestFundamentalAgentAnalyze(t *testing.T) {
	provider := simpleProvider(`Based on analysis:
{"ticker": "TCS", "recommendation": "BUY", "confidence": 0.8, "summary": "Strong fundamentals with good growth."}`)

	agent := NewFundamentalAgent(provider, newMockSources(), nil)
	result, err := agent.Analyze(context.Background(), "TCS")
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if result.AgentName != prompts.AgentFundamental {
		t.Fatalf("AgentName: %q", result.AgentName)
	}
	if result.Content == "" {
		t.Fatal("Content should not be empty")
	}
}

func TestFundamentalAgentAnalyzeWithTimestamp(t *testing.T) {
	provider := simpleProvider(`{"recommendation": "BUY", "confidence": 0.9}`)
	agent := NewFundamentalAgent(provider, newMockSources(), nil)

	result, err := agent.AnalyzeWithTimestamp(context.Background(), "TCS")
	if err != nil {
		t.Fatalf("AnalyzeWithTimestamp: %v", err)
	}
	if result.Analysis == nil {
		t.Fatal("Analysis should not be nil")
	}
	if result.Analysis.Ticker != "TCS" {
		t.Fatalf("Ticker: %q", result.Analysis.Ticker)
	}
	if result.Analysis.Type != models.AnalysisFundamental {
		t.Fatalf("Type: %q", result.Analysis.Type)
	}
}

func TestTechnicalAgentAnalyze(t *testing.T) {
	provider := simpleProvider("Technical analysis complete. Trend: Bullish.")
	agent := NewTechnicalAgent(provider, newMockSources(), nil)

	result, err := agent.Analyze(context.Background(), "RELIANCE")
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if result.Content == "" {
		t.Fatal("Content should not be empty")
	}
}

func TestExecutorCreateTradeProposal(t *testing.T) {
	provider := simpleProvider("Trade proposal: BUY TCS at 3500, SL 3400, Target 3800.")
	agent := NewExecutorAgent(provider, nil)

	analyses := []*AgentResult{
		{AgentName: "fundamental", Role: "Fundamental", Content: "Strong buy signal"},
		{AgentName: "technical", Role: "Technical", Content: "Bullish trend"},
	}

	result, err := agent.CreateTradeProposal(context.Background(), "TCS", analyses)
	if err != nil {
		t.Fatalf("CreateTradeProposal: %v", err)
	}
	if result.Content == "" {
		t.Fatal("should produce content")
	}
}

func TestReporterGenerateReport(t *testing.T) {
	provider := simpleProvider("# TCS Analysis Report\n\nComprehensive report...")
	agent := NewReporterAgent(provider, nil)

	analyses := []*AgentResult{
		{AgentName: "fundamental", Role: "Fundamental Analyst", Content: "PE: 32.5, ROE: 45.2%"},
		{AgentName: "technical", Role: "Technical Analyst", Content: "RSI: 62, MACD bullish"},
		nil, // should handle nil gracefully
	}

	result, err := agent.GenerateReport(context.Background(), "TCS", analyses)
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	if result.Content == "" {
		t.Fatal("should produce report content")
	}
}

// ════════════════════════════════════════════════════════════════════
// Helpers
// ════════════════════════════════════════════════════════════════════

// mockDS implements datasource.DataSource for testing.
type mockDS struct {
	dsName string
}

var _ datasource.DataSource = (*mockDS)(nil) // compile-time check

func (m *mockDS) Name() string { return m.dsName }
func (m *mockDS) GetQuote(_ context.Context, ticker string) (*models.Quote, error) {
	return &models.Quote{
		Ticker: ticker, LastPrice: 3500, Change: 25, ChangePct: 0.72,
		Open: 3480, High: 3520, Low: 3470, PrevClose: 3475, Volume: 1500000,
	}, nil
}
func (m *mockDS) GetHistoricalData(_ context.Context, ticker string, from, _ time.Time, _ models.Timeframe) ([]models.OHLCV, error) {
	candles := make([]models.OHLCV, 30)
	for i := range candles {
		candles[i] = models.OHLCV{
			Timestamp: from.AddDate(0, 0, i),
			Open: 3000 + float64(i), High: 3050 + float64(i),
			Low: 2970 + float64(i), Close: 3010 + float64(i),
			Volume: 1000000,
		}
	}
	return candles, nil
}
func (m *mockDS) GetFinancials(_ context.Context, ticker string) (*models.FinancialData, error) {
	return &models.FinancialData{
		Ticker: ticker,
		QuarterlyIncome: []models.IncomeStatement{
			{Period: "Q1 2024", Revenue: 60000e6, PAT: 12000e6},
			{Period: "Q4 2023", Revenue: 58000e6, PAT: 11500e6},
		},
	}, nil
}
func (m *mockDS) GetOptionChain(_ context.Context, ticker string, _ string) (*models.OptionChain, error) {
	return &models.OptionChain{
		Ticker: ticker, SpotPrice: 3500, ExpiryDate: "27-Jun-2024",
		PCR: 1.05, MaxPain: 3500,
		Contracts: []models.OptionContract{
			{StrikePrice: 3500, OptionType: "CE", OI: 80000, IV: 17.5},
			{StrikePrice: 3500, OptionType: "PE", OI: 70000, IV: 18.0},
		},
	}, nil
}
func (m *mockDS) GetStockProfile(_ context.Context, ticker string) (*models.StockProfile, error) {
	return &models.StockProfile{
		Stock: models.Stock{Ticker: ticker, Name: ticker + " Ltd.", Sector: "IT"},
	}, nil
}

// newMockSources creates a slice of datasource.DataSource with one mock.
func newMockSources() []datasource.DataSource {
	return []datasource.DataSource{&mockDS{dsName: "mock"}}
}

func toolNameSet(tools []llm.Tool) map[string]bool {
	m := make(map[string]bool, len(tools))
	for _, t := range tools {
		m[t.Name] = true
	}
	return m
}
