package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/agent/prompts"
	"github.com/seenimoa/openseai/internal/analysis/technical"
	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/pkg/models"
)

// TechnicalAgent is the Technical Analyst specialized agent.
// It analyzes price action, indicators, patterns, and generates trading signals.
type TechnicalAgent struct {
	*BaseAgent
	dataSources []datasource.DataSource
}

// NewTechnicalAgent creates a Technical Analyst agent.
func NewTechnicalAgent(provider llm.LLMProvider, sources []datasource.DataSource, opts *llm.ChatOptions) *TechnicalAgent {
	agent := &TechnicalAgent{dataSources: sources}

	tools := agent.buildTools()

	systemPrompt := prompts.TechnicalSystemPrompt + prompts.IndianMarketPromptSuffix()

	agent.BaseAgent = NewBaseAgent(BaseAgentConfig{
		Name:         prompts.AgentTechnical,
		Role:         "Technical Analyst — Price action, indicators, patterns, signals",
		SystemPrompt: systemPrompt,
		Provider:     provider,
		Tools:        tools,
		ChatOptions:  opts,
		MemorySize:   40,
		MaxToolIter:  8,
	})

	return agent
}

func (a *TechnicalAgent) buildTools() []llm.Tool {
	return []llm.Tool{
		{
			Name:        "get_historical_data",
			Description: "Fetch OHLCV candlestick data for a ticker. Returns price history used for all technical analysis.",
			Parameters: llm.ObjectSchema("Historical data parameters",
				map[string]*llm.JSONSchema{
					"ticker":    llm.StringProp("NSE ticker symbol (e.g., RELIANCE, TCS)"),
					"days":      llm.IntProp("Number of trading days of history to fetch (default: 200)"),
					"timeframe": llm.StringProp("Candle timeframe: 1d (default), 1h, 15m, 5m, 1m, 1w"),
				},
				"ticker",
			),
			Handler: a.handleGetHistorical,
		},
		{
			Name:        "compute_indicators",
			Description: "Compute all technical indicators (RSI, MACD, Bollinger Bands, SuperTrend, ATR, Pivot Points) from OHLCV data",
			Parameters: llm.ObjectSchema("Indicator parameters",
				map[string]*llm.JSONSchema{
					"ticker":    llm.StringProp("NSE ticker symbol"),
					"days":      llm.IntProp("Number of trading days (default: 200)"),
					"timeframe": llm.StringProp("Candle timeframe (default: 1d)"),
				},
				"ticker",
			),
			Handler: a.handleComputeIndicators,
		},
		{
			Name:        "generate_signals",
			Description: "Generate BUY/SELL/HOLD trading signals from technical analysis of price data",
			Parameters: llm.ObjectSchema("Signal generation parameters",
				map[string]*llm.JSONSchema{
					"ticker":    llm.StringProp("NSE ticker symbol"),
					"days":      llm.IntProp("Number of trading days (default: 200)"),
					"timeframe": llm.StringProp("Candle timeframe (default: 1d)"),
				},
				"ticker",
			),
			Handler: a.handleGenerateSignals,
		},
		{
			Name:        "full_technical_analysis",
			Description: "Run a comprehensive technical analysis combining all indicators, signals, support/resistance, and trend assessment",
			Parameters: llm.ObjectSchema("Full analysis parameters",
				map[string]*llm.JSONSchema{
					"ticker":    llm.StringProp("NSE ticker symbol"),
					"days":      llm.IntProp("Number of trading days (default: 200)"),
					"timeframe": llm.StringProp("Candle timeframe (default: 1d)"),
				},
				"ticker",
			),
			Handler: a.handleFullAnalysis,
		},
		{
			Name:        "get_quote",
			Description: "Get latest stock quote with current price, volume, day range, 52-week range",
			Parameters: llm.ObjectSchema("Quote parameters",
				map[string]*llm.JSONSchema{
					"ticker": llm.StringProp("NSE ticker symbol"),
				},
				"ticker",
			),
			Handler: a.handleGetQuote,
		},
	}
}

// ── Tool Handlers ──

func (a *TechnicalAgent) fetchCandles(ctx context.Context, ticker string, days int, timeframe string) ([]models.OHLCV, error) {
	if days <= 0 {
		days = 200
	}
	tf := models.Timeframe1Day
	switch timeframe {
	case "1m":
		tf = models.Timeframe1Min
	case "5m":
		tf = models.Timeframe5Min
	case "15m":
		tf = models.Timeframe15Min
	case "1h":
		tf = models.Timeframe1Hour
	case "1w":
		tf = models.Timeframe1Week
	}

	to := time.Now()
	from := to.AddDate(0, 0, -int(float64(days)*1.5)) // extra days to account for weekends/holidays

	for _, src := range a.dataSources {
		candles, err := src.GetHistoricalData(ctx, ticker, from, to, tf)
		if err == nil && len(candles) > 0 {
			if len(candles) > days {
				candles = candles[len(candles)-days:]
			}
			return candles, nil
		}
	}
	return nil, fmt.Errorf("no historical data available for %s", ticker)
}

func (a *TechnicalAgent) handleGetHistorical(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker    string `json:"ticker"`
		Days      int    `json:"days"`
		Timeframe string `json:"timeframe"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	candles, err := a.fetchCandles(ctx, params.Ticker, params.Days, params.Timeframe)
	if err != nil {
		return err.Error(), nil
	}

	// Return summary to avoid flooding context
	summary := map[string]any{
		"ticker":  params.Ticker,
		"candles": len(candles),
	}
	if len(candles) > 0 {
		last := candles[len(candles)-1]
		first := candles[0]
		summary["first_date"] = first.Timestamp.Format("2006-01-02")
		summary["last_date"] = last.Timestamp.Format("2006-01-02")
		summary["latest_close"] = last.Close
		summary["latest_volume"] = last.Volume
		summary["latest_high"] = last.High
		summary["latest_low"] = last.Low

		// Provide last 5 candles in detail
		n := 5
		if len(candles) < n {
			n = len(candles)
		}
		summary["recent_candles"] = candles[len(candles)-n:]
	}

	data, _ := json.MarshalIndent(summary, "", "  ")
	return string(data), nil
}

func (a *TechnicalAgent) handleComputeIndicators(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker    string `json:"ticker"`
		Days      int    `json:"days"`
		Timeframe string `json:"timeframe"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	candles, err := a.fetchCandles(ctx, params.Ticker, params.Days, params.Timeframe)
	if err != nil {
		return err.Error(), nil
	}

	indicators := technical.ComputeAll(params.Ticker, candles)
	data, _ := json.MarshalIndent(indicators, "", "  ")
	return string(data), nil
}

func (a *TechnicalAgent) handleGenerateSignals(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker    string `json:"ticker"`
		Days      int    `json:"days"`
		Timeframe string `json:"timeframe"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	candles, err := a.fetchCandles(ctx, params.Ticker, params.Days, params.Timeframe)
	if err != nil {
		return err.Error(), nil
	}

	signals := technical.GenerateSignals(candles)
	result := map[string]any{
		"ticker":       params.Ticker,
		"total_signals": len(signals),
	}
	// Return last 10 signals
	n := 10
	if len(signals) < n {
		n = len(signals)
	}
	if n > 0 {
		result["recent_signals"] = signals[len(signals)-n:]
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *TechnicalAgent) handleFullAnalysis(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker    string `json:"ticker"`
		Days      int    `json:"days"`
		Timeframe string `json:"timeframe"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	candles, err := a.fetchCandles(ctx, params.Ticker, params.Days, params.Timeframe)
	if err != nil {
		return err.Error(), nil
	}

	result := technical.FullTechnicalAnalysis(params.Ticker, candles)
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *TechnicalAgent) handleGetQuote(ctx context.Context, args json.RawMessage) (string, error) {
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

// Analyze runs a full technical analysis with chain-of-thought reasoning.
func (a *TechnicalAgent) Analyze(ctx context.Context, ticker string) (*AgentResult, error) {
	task := prompts.CoTTechnical(ticker)
	return a.Process(ctx, task)
}

// AnalyzeWithTimestamp runs the analysis and attaches a typed result.
func (a *TechnicalAgent) AnalyzeWithTimestamp(ctx context.Context, ticker string) (*AgentResult, error) {
	result, err := a.Analyze(ctx, ticker)
	if err != nil {
		return result, err
	}

	result.Analysis = ParseAnalysisResult(result.Content, models.AnalysisResult{
		Ticker:    ticker,
		Type:      models.AnalysisTechnical,
		AgentName: a.Name(),
		Timestamp: time.Now(),
	})

	return result, nil
}
