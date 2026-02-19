package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/seenimoa/openseai/internal/agent/prompts"
	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/pkg/models"
)

// OrchestratorMode determines how the orchestrator coordinates agents.
type OrchestratorMode string

const (
	// ModeSingle uses a single agent with all available tools for quick queries.
	ModeSingle OrchestratorMode = "single"
	// ModeMulti uses multiple specialized agents coordinated by a CIO agent.
	ModeMulti OrchestratorMode = "multi"
)

// Orchestrator coordinates agents in single-agent or multi-agent mode.
type Orchestrator struct {
	mu sync.RWMutex

	// Agents
	fundamental *FundamentalAgent
	technical   *TechnicalAgent
	sentiment   *SentimentAgent
	fno         *FnOAgent
	risk        *RiskAgent
	executor    *ExecutorAgent
	reporter    *ReporterAgent

	// CIO agent for multi-agent synthesis
	cio *BaseAgent

	// Single-agent for quick queries
	singleAgent *BaseAgent

	// LLM provider
	provider llm.LLMProvider

	// Config
	defaultMode   OrchestratorMode
	defaultCapital float64 // default trading capital in ₹
}

// OrchestratorConfig holds configuration for creating an Orchestrator.
type OrchestratorConfig struct {
	Provider    llm.LLMProvider
	Aggregator  *datasource.Aggregator
	ChatOptions *llm.ChatOptions
	DefaultMode OrchestratorMode
	Capital     float64 // default trading capital in ₹
}

// NewOrchestrator creates a fully configured Orchestrator with all specialized agents.
func NewOrchestrator(cfg OrchestratorConfig) *Orchestrator {
	sources := cfg.Aggregator.Sources()

	o := &Orchestrator{
		provider:       cfg.Provider,
		defaultMode:    cfg.DefaultMode,
		defaultCapital: cfg.Capital,
	}

	if o.defaultMode == "" {
		o.defaultMode = ModeSingle
	}
	if o.defaultCapital <= 0 {
		o.defaultCapital = 1_000_000 // ₹10 Lakh default
	}

	opts := cfg.ChatOptions

	// Create specialized agents
	o.fundamental = NewFundamentalAgent(cfg.Provider, sources, opts)
	o.technical = NewTechnicalAgent(cfg.Provider, sources, opts)
	o.sentiment = NewSentimentAgent(cfg.Provider, cfg.Aggregator.NewsSource(), opts)
	o.fno = NewFnOAgent(cfg.Provider, cfg.Aggregator.Derivatives(), sources, opts)
	o.risk = NewRiskAgent(cfg.Provider, sources, opts)
	o.executor = NewExecutorAgent(cfg.Provider, opts)
	o.reporter = NewReporterAgent(cfg.Provider, opts)

	// Create CIO agent for multi-agent coordination
	o.cio = NewBaseAgent(BaseAgentConfig{
		Name:         prompts.AgentCIO,
		Role:         "Chief Investment Officer — Team coordination, conflict resolution, final recommendation",
		SystemPrompt: prompts.CIOSystemPrompt + prompts.IndianMarketPromptSuffix(),
		Provider:     cfg.Provider,
		ChatOptions:  opts,
		MemorySize:   60,
		MaxToolIter:  4,
	})

	// Create single-agent with all tools combined
	o.buildSingleAgent(cfg.Provider, opts)

	return o
}

// buildSingleAgent creates a single agent that has tools from all specialized agents.
func (o *Orchestrator) buildSingleAgent(provider llm.LLMProvider, opts *llm.ChatOptions) {
	// Merge tools from all agents, prefixing names to avoid collisions
	var allTools []llm.Tool

	for _, agent := range []Agent{o.fundamental, o.technical, o.sentiment, o.fno, o.risk, o.executor, o.reporter} {
		for _, t := range agent.Tools() {
			allTools = append(allTools, t)
		}
	}

	// Deduplicate tools by name (some agents share tools like get_quote)
	seen := make(map[string]bool)
	var uniqueTools []llm.Tool
	for _, t := range allTools {
		if !seen[t.Name] {
			seen[t.Name] = true
			uniqueTools = append(uniqueTools, t)
		}
	}

	o.singleAgent = NewBaseAgent(BaseAgentConfig{
		Name:         "single-agent",
		Role:         "Universal Stock Analyst — handles all types of queries",
		SystemPrompt: buildSingleAgentPrompt(),
		Provider:     provider,
		Tools:        uniqueTools,
		ChatOptions:  opts,
		MemorySize:   50,
		MaxToolIter:  12,
	})
}

// buildSingleAgentPrompt creates a combined system prompt for the single agent.
func buildSingleAgentPrompt() string {
	return `You are OpeNSE.ai — an expert AI stock analyst for the Indian market (NSE/BSE).
You have access to tools for technical analysis, fundamental analysis, derivatives (F&O) analysis,
sentiment analysis, risk management, and trade execution.

For simple queries (e.g., "What's the RSI of RELIANCE?"), use the appropriate tool directly.
For complex queries, combine multiple tools to provide comprehensive analysis.

Always:
- Use Indian number formatting (₹, Lakhs, Crores)
- Include NSE-specific context (circuit limits, F&O lot sizes, expiry schedules)
- Provide confidence levels with your analysis
- Highlight key risks and caveats

` + prompts.IndianMarketPromptSuffix()
}

// ── Public API ──

// Process handles a user query, automatically selecting single or multi-agent mode.
func (o *Orchestrator) Process(ctx context.Context, query string) (*AgentResult, error) {
	return o.ProcessWithMode(ctx, query, o.defaultMode)
}

// ProcessWithMode handles a query with an explicit mode selection.
func (o *Orchestrator) ProcessWithMode(ctx context.Context, query string, mode OrchestratorMode) (*AgentResult, error) {
	switch mode {
	case ModeSingle:
		return o.processSingle(ctx, query)
	case ModeMulti:
		return o.processMulti(ctx, query)
	default:
		return o.processSingle(ctx, query)
	}
}

// QuickQuery runs a single-agent query (convenience method).
func (o *Orchestrator) QuickQuery(ctx context.Context, query string) (*AgentResult, error) {
	return o.processSingle(ctx, query)
}

// FullAnalysis runs a multi-agent analysis for a ticker (convenience method).
func (o *Orchestrator) FullAnalysis(ctx context.Context, ticker string) (*AgentResult, error) {
	query := fmt.Sprintf("Perform a comprehensive investment analysis of %s for the Indian market.", ticker)
	return o.processMulti(ctx, query)
}

// Chat handles an interactive chat message with conversation history.
func (o *Orchestrator) Chat(ctx context.Context, message string, history []llm.Message) (*AgentResult, error) {
	return o.singleAgent.ProcessWithMessages(ctx, message, history)
}

// ── Internal modes ──

// processSingle routes the query to the single all-tools agent.
func (o *Orchestrator) processSingle(ctx context.Context, query string) (*AgentResult, error) {
	return o.singleAgent.Process(ctx, query)
}

// processMulti runs the CIO-led multi-agent workflow.
func (o *Orchestrator) processMulti(ctx context.Context, query string) (*AgentResult, error) {
	ticker := extractTicker(query)
	if ticker == "" {
		// Fall back to single agent if we can't extract a ticker
		return o.processSingle(ctx, query)
	}

	start := time.Now()

	// Phase 1: Run specialized agents concurrently
	type agentResult struct {
		name   string
		result *AgentResult
		err    error
	}

	ch := make(chan agentResult, 5)
	var wg sync.WaitGroup

	// Launch agents concurrently
	agents := []struct {
		name string
		fn   func(context.Context, string) (*AgentResult, error)
	}{
		{"fundamental", func(ctx context.Context, t string) (*AgentResult, error) {
			return o.fundamental.AnalyzeWithTimestamp(ctx, t)
		}},
		{"technical", func(ctx context.Context, t string) (*AgentResult, error) {
			return o.technical.AnalyzeWithTimestamp(ctx, t)
		}},
		{"sentiment", func(ctx context.Context, t string) (*AgentResult, error) {
			return o.sentiment.AnalyzeWithTimestamp(ctx, t)
		}},
		{"fno", func(ctx context.Context, t string) (*AgentResult, error) {
			return o.fno.AnalyzeWithTimestamp(ctx, t)
		}},
		{"risk", func(ctx context.Context, t string) (*AgentResult, error) {
			return o.risk.AnalyzeWithTimestamp(ctx, t, o.defaultCapital)
		}},
	}

	for _, a := range agents {
		wg.Add(1)
		go func(name string, fn func(context.Context, string) (*AgentResult, error)) {
			defer wg.Done()
			result, err := fn(ctx, ticker)
			ch <- agentResult{name: name, result: result, err: err}
		}(a.name, a.fn)
	}

	// Close channel once all agents complete
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Collect results
	results := make(map[string]*AgentResult)
	var errors []string
	for ar := range ch {
		if ar.err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", ar.name, ar.err))
			continue
		}
		results[ar.name] = ar.result
	}

	// Phase 2: CIO synthesis
	synthesisTask := buildSynthesisPrompt(ticker, query, results, errors)
	cioResult, err := o.cio.Process(ctx, synthesisTask)
	if err != nil {
		// If CIO fails, try to compile results manually
		return compileFallbackResult(ticker, results, errors, start), nil
	}

	// Phase 3: Generate report
	allResults := []*AgentResult{}
	for _, r := range results {
		allResults = append(allResults, r)
	}
	if cioResult != nil {
		allResults = append(allResults, cioResult)
	}

	reportResult, reportErr := o.reporter.GenerateReport(ctx, ticker, allResults)

	// Build final orchestrator result
	final := &AgentResult{
		AgentName: "orchestrator",
		Role:      "Multi-Agent Orchestrator",
		Duration:  time.Since(start),
	}

	if reportErr == nil && reportResult != nil {
		final.Content = reportResult.Content
		final.Tokens = reportResult.Tokens
	} else {
		final.Content = cioResult.Content
		final.Tokens = cioResult.Tokens
	}

	// Count total tool calls across all agents
	for _, r := range results {
		final.ToolCalls += r.ToolCalls
	}
	final.ToolCalls += cioResult.ToolCalls

	// Attach composite analysis
	final.Analysis = &models.AnalysisResult{
		Ticker:    ticker,
		Type:      models.AnalysisComposite,
		AgentName: "orchestrator",
		Timestamp: time.Now(),
	}

	return final, nil
}

// buildSynthesisPrompt creates the CIO synthesis task from agent results.
func buildSynthesisPrompt(ticker, originalQuery string, results map[string]*AgentResult, errors []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are synthesizing a comprehensive analysis of %s.\n\n", ticker))
	sb.WriteString(fmt.Sprintf("Original query: %s\n\n", originalQuery))
	sb.WriteString(prompts.CoTSynthesis(ticker))
	sb.WriteString("\n\nHere are the analysis results from your team:\n\n")

	for name, r := range results {
		sb.WriteString(fmt.Sprintf("### %s Agent (%s)\n", strings.Title(name), r.Role))
		sb.WriteString(r.Content)
		sb.WriteString("\n\n---\n\n")
	}

	if len(errors) > 0 {
		sb.WriteString("### Agent Errors\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
		sb.WriteString("\nNote: Some agents encountered errors. Factor this into your confidence level.\n\n")
	}

	sb.WriteString("Provide your final synthesis with:\n" +
		"1. Weighted assessment (fundamental 30%, technical 25%, sentiment 15%, derivatives 15%, risk 15%)\n" +
		"2. Key conflicts and how you resolve them\n" +
		"3. Overall recommendation: STRONG BUY / BUY / HOLD / SELL / STRONG SELL\n" +
		"4. Conviction level: HIGH / MEDIUM / LOW\n" +
		"5. Key risks and catalysts\n")

	return sb.String()
}

// compileFallbackResult creates a summary when the CIO agent fails.
func compileFallbackResult(ticker string, results map[string]*AgentResult, errors []string, start time.Time) *AgentResult {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Multi-Agent Analysis: %s\n\n", ticker))
	sb.WriteString("*Note: CIO synthesis unavailable. Presenting raw agent outputs.*\n\n")

	for name, r := range results {
		sb.WriteString(fmt.Sprintf("## %s Agent\n%s\n\n", strings.Title(name), r.Content))
	}

	if len(errors) > 0 {
		sb.WriteString("## Errors\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	totalTools := 0
	for _, r := range results {
		totalTools += r.ToolCalls
	}

	return &AgentResult{
		AgentName: "orchestrator",
		Role:      "Multi-Agent Orchestrator (fallback)",
		Content:   sb.String(),
		ToolCalls: totalTools,
		Duration:  time.Since(start),
		Analysis: &models.AnalysisResult{
			Ticker:    ticker,
			Type:      models.AnalysisComposite,
			AgentName: "orchestrator",
			Timestamp: time.Now(),
		},
	}
}

// extractTicker attempts to extract an NSE ticker from a query string.
// It looks for known patterns and uppercase words that look like tickers.
func extractTicker(query string) string {
	words := strings.Fields(query)
	for _, w := range words {
		// Clean punctuation
		w = strings.TrimRight(w, ".,;:!?")

		// Skip short words and common words
		if len(w) < 2 || len(w) > 20 {
			continue
		}

		// Check if it's a known sector ticker
		upper := strings.ToUpper(w)
		if prompts.SectorForTicker(upper) != "" {
			return upper
		}

		// Check if it looks like a ticker (all uppercase, 2-20 chars)
		if w == upper && isAlpha(w) && len(w) >= 2 {
			return upper
		}
	}
	return ""
}

// isAlpha returns true if a string contains only ASCII letters.
func isAlpha(s string) bool {
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			return false
		}
	}
	return true
}

// ── Accessors ──

// FundamentalAgent returns the fundamental analysis agent.
func (o *Orchestrator) FundamentalAgent() *FundamentalAgent { return o.fundamental }

// TechnicalAgent returns the technical analysis agent.
func (o *Orchestrator) TechnicalAgent() *TechnicalAgent { return o.technical }

// SentimentAgent returns the sentiment analysis agent.
func (o *Orchestrator) SentimentAgent() *SentimentAgent { return o.sentiment }

// FnOAgent returns the F&O analysis agent.
func (o *Orchestrator) FnOAgent() *FnOAgent { return o.fno }

// RiskAgent returns the risk management agent.
func (o *Orchestrator) RiskAgent() *RiskAgent { return o.risk }

// ExecutorAgent returns the trade executor agent.
func (o *Orchestrator) ExecutorAgent() *ExecutorAgent { return o.executor }

// ReporterAgent returns the report generator agent.
func (o *Orchestrator) ReporterAgent() *ReporterAgent { return o.reporter }

// SetMode sets the default orchestration mode.
func (o *Orchestrator) SetMode(mode OrchestratorMode) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.defaultMode = mode
}

// Mode returns the current default mode.
func (o *Orchestrator) Mode() OrchestratorMode {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.defaultMode
}
