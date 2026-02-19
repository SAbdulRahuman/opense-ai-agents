package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/agent/prompts"
	"github.com/seenimoa/openseai/internal/analysis/derivatives"
	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/pkg/models"
)

// FnOAgent is the F&O (Futures & Options) Analyst specialized agent.
// It analyzes option chains, PCR, OI buildup, futures data, and suggests strategies.
type FnOAgent struct {
	*BaseAgent
	derivSrc *datasource.NSEDerivatives
	sources  []datasource.DataSource
}

// NewFnOAgent creates an F&O Analyst agent.
func NewFnOAgent(provider llm.LLMProvider, derivSrc *datasource.NSEDerivatives, sources []datasource.DataSource, opts *llm.ChatOptions) *FnOAgent {
	agent := &FnOAgent{
		derivSrc: derivSrc,
		sources:  sources,
	}

	tools := agent.buildTools()

	systemPrompt := prompts.FnOSystemPrompt + prompts.IndianMarketPromptSuffix()

	agent.BaseAgent = NewBaseAgent(BaseAgentConfig{
		Name:         prompts.AgentFnO,
		Role:         "F&O Analyst — Options, futures, OI analysis, strategy suggestions",
		SystemPrompt: systemPrompt,
		Provider:     provider,
		Tools:        tools,
		ChatOptions:  opts,
		MemorySize:   40,
		MaxToolIter:  8,
	})

	return agent
}

func (a *FnOAgent) buildTools() []llm.Tool {
	return []llm.Tool{
		{
			Name:        "get_option_chain",
			Description: "Fetch the full option chain for an NSE F&O stock or index. Returns all strikes with CE/PE data, OI, volume, IV, Greeks.",
			Parameters: llm.ObjectSchema("Option chain parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker or index (e.g., NIFTY, BANKNIFTY, RELIANCE)"),
					"expiry": llm.StringProp("Expiry date in DD-Mon-YYYY format (optional, defaults to nearest expiry)"),
				},
				"ticker",
			),
			Handler: a.handleGetOptionChain,
		},
		{
			Name:        "analyze_option_chain",
			Description: "Analyze option chain: compute max pain, IV skew, ATM IV, OI-based support/resistance, PCR sentiment",
			Parameters: llm.ObjectSchema("Analysis parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker or index"),
					"expiry": llm.StringProp("Expiry date (optional)"),
				},
				"ticker",
			),
			Handler: a.handleAnalyzeOptionChain,
		},
		{
			Name:        "compute_pcr",
			Description: "Compute Put-Call Ratio (PCR) by OI and volume with interpretation (bullish/bearish/neutral signal)",
			Parameters: llm.ObjectSchema("PCR parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker or index"),
					"expiry": llm.StringProp("Expiry date (optional)"),
				},
				"ticker",
			),
			Handler: a.handleComputePCR,
		},
		{
			Name:        "analyze_oi_buildup",
			Description: "Analyze OI buildup patterns (long/short buildup/unwinding) from option chain and futures data",
			Parameters: llm.ObjectSchema("OI buildup parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker or index"),
					"expiry": llm.StringProp("Expiry date (optional)"),
				},
				"ticker",
			),
			Handler: a.handleAnalyzeOIBuildup,
		},
		{
			Name:        "get_futures_data",
			Description: "Fetch futures contracts for a ticker (current, next, far month) with price, OI, basis, premium/discount",
			Parameters: llm.ObjectSchema("Futures parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker or index"),
				},
				"ticker",
			),
			Handler: a.handleGetFutures,
		},
		{
			Name:        "get_india_vix",
			Description: "Get India VIX (volatility index) — a key fear/greed indicator for the Indian market",
			Parameters: llm.ObjectSchema("VIX parameters",
				map[string]*llm.JSONSchema{},
			),
			Handler: a.handleGetVIX,
		},
		{
			Name:        "full_derivatives_analysis",
			Description: "Run comprehensive derivatives analysis combining option chain, PCR, OI buildup, futures, and VIX into a single report",
			Parameters: llm.ObjectSchema("Full derivatives analysis parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker or index"),
					"expiry": llm.StringProp("Expiry date (optional)"),
				},
				"ticker",
			),
			Handler: a.handleFullAnalysis,
		},
	}
}

// ── Tool Handlers ──

func (a *FnOAgent) fetchOptionChain(ctx context.Context, ticker, expiry string) (*models.OptionChain, error) {
	return a.derivSrc.GetOptionChain(ctx, ticker, expiry)
}

func (a *FnOAgent) handleGetOptionChain(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
		Expiry string `json:"expiry"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	oc, err := a.fetchOptionChain(ctx, params.Ticker, params.Expiry)
	if err != nil {
		return fmt.Sprintf("Could not fetch option chain for %s: %v", params.Ticker, err), nil
	}

	// Return summary, not full chain (too large)
	summary := map[string]any{
		"ticker":     oc.Ticker,
		"spot_price": oc.SpotPrice,
		"expiry":     oc.ExpiryDate,
		"contracts":  len(oc.Contracts),
		"pcr":        oc.PCR,
		"max_pain":   oc.MaxPain,
	}

	// Include ATM ± 5 strikes
	if len(oc.Contracts) > 0 {
		atmContracts := filterATMContracts(oc)
		summary["atm_contracts"] = atmContracts
	}

	data, _ := json.MarshalIndent(summary, "", "  ")
	return string(data), nil
}

func (a *FnOAgent) handleAnalyzeOptionChain(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
		Expiry string `json:"expiry"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	oc, err := a.fetchOptionChain(ctx, params.Ticker, params.Expiry)
	if err != nil {
		return fmt.Sprintf("Could not fetch option chain for %s: %v", params.Ticker, err), nil
	}

	analysis := derivatives.AnalyzeOptionChain(oc)
	data, _ := json.MarshalIndent(analysis, "", "  ")
	return string(data), nil
}

func (a *FnOAgent) handleComputePCR(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
		Expiry string `json:"expiry"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	oc, err := a.fetchOptionChain(ctx, params.Ticker, params.Expiry)
	if err != nil {
		return fmt.Sprintf("Could not fetch option chain for %s: %v", params.Ticker, err), nil
	}

	pcr := derivatives.ComputePCR(oc)
	data, _ := json.MarshalIndent(pcr, "", "  ")
	return string(data), nil
}

func (a *FnOAgent) handleAnalyzeOIBuildup(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
		Expiry string `json:"expiry"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	oc, err := a.fetchOptionChain(ctx, params.Ticker, params.Expiry)
	if err != nil {
		return fmt.Sprintf("Could not fetch option chain for %s: %v", params.Ticker, err), nil
	}

	// Try to fetch futures data for the same ticker
	var fut *models.FuturesContract
	futures, fErr := a.derivSrc.GetFuturesData(ctx, params.Ticker)
	if fErr == nil && len(futures) > 0 {
		fut = &futures[0] // current month
	}

	oiAnalysis := derivatives.AnalyzeOIBuildup(oc, fut)
	data, _ := json.MarshalIndent(oiAnalysis, "", "  ")
	return string(data), nil
}

func (a *FnOAgent) handleGetFutures(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	futures, err := a.derivSrc.GetFuturesData(ctx, params.Ticker)
	if err != nil {
		return fmt.Sprintf("Could not fetch futures data for %s: %v", params.Ticker, err), nil
	}

	data, _ := json.MarshalIndent(futures, "", "  ")
	return string(data), nil
}

func (a *FnOAgent) handleGetVIX(ctx context.Context, _ json.RawMessage) (string, error) {
	vix, err := a.derivSrc.GetIndiaVIX(ctx)
	if err != nil {
		return fmt.Sprintf("Could not fetch India VIX: %v", err), nil
	}
	data, _ := json.MarshalIndent(vix, "", "  ")
	return string(data), nil
}

func (a *FnOAgent) handleFullAnalysis(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
		Expiry string `json:"expiry"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	oc, err := a.fetchOptionChain(ctx, params.Ticker, params.Expiry)
	if err != nil {
		return fmt.Sprintf("Could not fetch option chain for %s: %v", params.Ticker, err), nil
	}

	var fut *models.FuturesContract
	futures, fErr := a.derivSrc.GetFuturesData(ctx, params.Ticker)
	if fErr == nil && len(futures) > 0 {
		fut = &futures[0]
	}

	result := derivatives.FullDerivativesAnalysis(params.Ticker, oc, fut)
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

// filterATMContracts returns contracts within ±5 strikes of ATM.
func filterATMContracts(oc *models.OptionChain) []models.OptionContract {
	atm := oc.SpotPrice
	var filtered []models.OptionContract

	// Collect unique strikes near ATM
	for _, c := range oc.Contracts {
		diff := c.StrikePrice - atm
		if diff < 0 {
			diff = -diff
		}
		// Include strikes within 5% of spot
		if diff < atm*0.05 {
			filtered = append(filtered, c)
		}
	}

	// Limit output
	if len(filtered) > 20 {
		filtered = filtered[:20]
	}
	return filtered
}

// Analyze runs a comprehensive F&O analysis with chain-of-thought reasoning.
func (a *FnOAgent) Analyze(ctx context.Context, ticker string) (*AgentResult, error) {
	task := prompts.CoTDerivatives(ticker)
	return a.Process(ctx, task)
}

// AnalyzeWithTimestamp runs the analysis and attaches a typed result.
func (a *FnOAgent) AnalyzeWithTimestamp(ctx context.Context, ticker string) (*AgentResult, error) {
	result, err := a.Analyze(ctx, ticker)
	if err != nil {
		return result, err
	}

	result.Analysis = ParseAnalysisResult(result.Content, models.AnalysisResult{
		Ticker:    ticker,
		Type:      models.AnalysisDerivatives,
		AgentName: a.Name(),
		Timestamp: time.Now(),
	})

	return result, nil
}
