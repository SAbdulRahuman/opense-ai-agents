package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/agent/prompts"
	"github.com/seenimoa/openseai/internal/analysis/fundamental"
	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/pkg/models"
)

// FundamentalAgent is the Fundamental Analyst specialized agent.
// It analyzes company financials, ratios, valuation, and peer comparison.
type FundamentalAgent struct {
	*BaseAgent
	dataSources []datasource.DataSource
}

// NewFundamentalAgent creates a Fundamental Analyst agent with the given data sources.
func NewFundamentalAgent(provider llm.LLMProvider, sources []datasource.DataSource, opts *llm.ChatOptions) *FundamentalAgent {
	agent := &FundamentalAgent{dataSources: sources}

	tools := agent.buildTools()

	systemPrompt := prompts.FundamentalSystemPrompt + prompts.IndianMarketPromptSuffix()

	agent.BaseAgent = NewBaseAgent(BaseAgentConfig{
		Name:         prompts.AgentFundamental,
		Role:         "Fundamental Analyst — Company financials, valuation, peer comparison",
		SystemPrompt: systemPrompt,
		Provider:     provider,
		Tools:        tools,
		ChatOptions:  opts,
		MemorySize:   40,
		MaxToolIter:  8,
	})

	return agent
}

// buildTools creates the Fundamental Analyst's tool set.
func (a *FundamentalAgent) buildTools() []llm.Tool {
	return []llm.Tool{
		{
			Name:        "get_financials",
			Description: "Fetch financial statements (income statement, balance sheet, cash flow) for an NSE-listed company",
			Parameters: llm.ObjectSchema("Financial data parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker symbol (e.g., RELIANCE, TCS, INFY)"),
				},
				"ticker",
			),
			Handler: a.handleGetFinancials,
		},
		{
			Name:        "get_stock_quote",
			Description: "Get the latest stock quote with current price, volume, day range, 52-week range",
			Parameters: llm.ObjectSchema("Quote parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker symbol"),
				},
				"ticker",
			),
			Handler: a.handleGetQuote,
		},
		{
			Name:        "compute_ratios",
			Description: "Compute financial ratios (PE, PB, ROE, ROCE, D/E, Current Ratio) from financial data",
			Parameters: llm.ObjectSchema("Ratio computation parameters",
				map[string]*llm.JSONSchema{
					"ticker":             llm.StringProp("NSE ticker symbol"),
					"price":              llm.NumberProp("Current stock price in ₹"),
					"shares_outstanding": llm.NumberProp("Total shares outstanding"),
				},
				"ticker", "price",
			),
			Handler: a.handleComputeRatios,
		},
		{
			Name:        "compute_valuation",
			Description: "Compute valuation metrics (intrinsic value, DCF estimate, Graham Number, PEG ratio)",
			Parameters: llm.ObjectSchema("Valuation parameters",
				map[string]*llm.JSONSchema{
					"ticker":             llm.StringProp("NSE ticker symbol"),
					"price":              llm.NumberProp("Current stock price in ₹"),
					"shares_outstanding": llm.NumberProp("Total shares outstanding"),
					"sector_pe":          llm.NumberProp("Sector average PE ratio"),
				},
				"ticker", "price",
			),
			Handler: a.handleComputeValuation,
		},
		{
			Name:        "get_peer_comparison",
			Description: "Compare the stock against its sector peers on key financial metrics",
			Parameters: llm.ObjectSchema("Peer comparison parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker symbol"),
				},
				"ticker",
			),
			Handler: a.handlePeerComparison,
		},
		{
			Name:        "get_stock_profile",
			Description: "Get a comprehensive stock profile including company info, sector, and key metrics",
			Parameters: llm.ObjectSchema("Profile parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker symbol"),
				},
				"ticker",
			),
			Handler: a.handleGetProfile,
		},
	}
}

// ── Tool Handlers ──

func (a *FundamentalAgent) handleGetFinancials(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	for _, src := range a.dataSources {
		fin, err := src.GetFinancials(ctx, params.Ticker)
		if err != nil {
			continue
		}
		data, _ := json.MarshalIndent(fin, "", "  ")
		return string(data), nil
	}
	return fmt.Sprintf("Could not fetch financials for %s from any data source", params.Ticker), nil
}

func (a *FundamentalAgent) handleGetQuote(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	for _, src := range a.dataSources {
		quote, err := src.GetQuote(ctx, params.Ticker)
		if err != nil {
			continue
		}
		data, _ := json.MarshalIndent(quote, "", "  ")
		return string(data), nil
	}
	return fmt.Sprintf("Could not fetch quote for %s", params.Ticker), nil
}

func (a *FundamentalAgent) handleComputeRatios(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker            string  `json:"ticker"`
		Price             float64 `json:"price"`
		SharesOutstanding float64 `json:"shares_outstanding"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	// Fetch financials
	var fin *models.FinancialData
	for _, src := range a.dataSources {
		f, err := src.GetFinancials(ctx, params.Ticker)
		if err == nil {
			fin = f
			break
		}
	}
	if fin == nil {
		return fmt.Sprintf("Could not fetch financial data for %s", params.Ticker), nil
	}

	shares := params.SharesOutstanding
	if shares <= 0 {
		shares = 1e8 // default estimate
	}

	ratios := fundamental.ComputeRatios(fin, params.Price, shares)
	growth := fundamental.ComputeGrowth(fin)

	result := map[string]any{
		"ticker": params.Ticker,
		"ratios": ratios,
		"growth": growth,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *FundamentalAgent) handleComputeValuation(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker            string  `json:"ticker"`
		Price             float64 `json:"price"`
		SharesOutstanding float64 `json:"shares_outstanding"`
		SectorPE          float64 `json:"sector_pe"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	var fin *models.FinancialData
	for _, src := range a.dataSources {
		f, err := src.GetFinancials(ctx, params.Ticker)
		if err == nil {
			fin = f
			break
		}
	}
	if fin == nil {
		return fmt.Sprintf("Could not fetch financial data for %s", params.Ticker), nil
	}

	shares := params.SharesOutstanding
	if shares <= 0 {
		shares = 1e8
	}
	sectorPE := params.SectorPE
	if sectorPE <= 0 {
		sectorPE = 20.0 // broad market PE default
	}

	ratios := fundamental.ComputeRatios(fin, params.Price, shares)
	valuation := fundamental.ComputeValuation(params.Ticker, params.Price, ratios, fin, shares, sectorPE)

	data, _ := json.MarshalIndent(valuation, "", "  ")
	return string(data), nil
}

func (a *FundamentalAgent) handlePeerComparison(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	peers := prompts.SectorPeers(params.Ticker)
	if len(peers) == 0 {
		return fmt.Sprintf("No known sector peers for %s", params.Ticker), nil
	}

	// Limit to top 5 peers
	if len(peers) > 5 {
		peers = peers[:5]
	}

	result := map[string]any{
		"ticker": params.Ticker,
		"sector": prompts.SectorForTicker(params.Ticker),
		"peers":  peers,
		"note":   "Use get_financials and compute_ratios for each peer to build a comparison table",
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *FundamentalAgent) handleGetProfile(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker string `json:"ticker"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	for _, src := range a.dataSources {
		profile, err := src.GetStockProfile(ctx, params.Ticker)
		if err != nil {
			continue
		}
		data, _ := json.MarshalIndent(profile, "", "  ")
		return string(data), nil
	}
	return fmt.Sprintf("Could not fetch profile for %s", params.Ticker), nil
}

// Analyze runs a full fundamental analysis with chain-of-thought reasoning.
func (a *FundamentalAgent) Analyze(ctx context.Context, ticker string) (*AgentResult, error) {
	task := prompts.CoTFundamental(ticker)
	return a.Process(ctx, task)
}

// AnalyzeWithTimestamp runs the analysis and attaches a timestamp to the result.
func (a *FundamentalAgent) AnalyzeWithTimestamp(ctx context.Context, ticker string) (*AgentResult, error) {
	result, err := a.Analyze(ctx, ticker)
	if err != nil {
		return result, err
	}

	result.Analysis = ParseAnalysisResult(result.Content, models.AnalysisResult{
		Ticker:    ticker,
		Type:      models.AnalysisFundamental,
		AgentName: a.Name(),
		Timestamp: time.Now(),
	})

	return result, nil
}
