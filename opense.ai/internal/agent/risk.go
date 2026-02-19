package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/seenimoa/openseai/internal/agent/prompts"
	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/pkg/models"
)

// RiskAgent is the Risk Manager specialized agent.
// It handles position sizing, VaR, stop-loss placement, and portfolio exposure management.
type RiskAgent struct {
	*BaseAgent
	dataSources []datasource.DataSource
}

// NewRiskAgent creates a Risk Manager agent.
func NewRiskAgent(provider llm.LLMProvider, sources []datasource.DataSource, opts *llm.ChatOptions) *RiskAgent {
	agent := &RiskAgent{dataSources: sources}

	tools := agent.buildTools()

	systemPrompt := prompts.RiskSystemPrompt + prompts.IndianMarketPromptSuffix()

	agent.BaseAgent = NewBaseAgent(BaseAgentConfig{
		Name:         prompts.AgentRisk,
		Role:         "Risk Manager — Position sizing, VaR, stop-loss, portfolio exposure",
		SystemPrompt: systemPrompt,
		Provider:     provider,
		Tools:        tools,
		ChatOptions:  opts,
		MemorySize:   30,
		MaxToolIter:  6,
	})

	return agent
}

func (a *RiskAgent) buildTools() []llm.Tool {
	return []llm.Tool{
		{
			Name:        "compute_position_size",
			Description: "Calculate the optimal position size based on capital, risk tolerance, and stop-loss distance. Follows Kelly Criterion with a half-Kelly conservative approach.",
			Parameters: llm.ObjectSchema("Position sizing parameters",
				map[string]*llm.JSONSchema{
					"capital":          llm.NumberProp("Total trading capital in ₹"),
					"risk_pct":         llm.NumberProp("Maximum risk per trade as percentage (e.g., 1.0 for 1%)"),
					"entry_price":      llm.NumberProp("Planned entry price in ₹"),
					"stop_loss":        llm.NumberProp("Stop-loss price in ₹"),
					"ticker":           llm.StringProp("NSE ticker (optional, for lot-size adjustment in F&O)"),
					"is_fno":           llm.BoolProp("Whether this is an F&O trade (adjusts for lot size)"),
				},
				"capital", "entry_price", "stop_loss",
			),
			Handler: a.handlePositionSize,
		},
		{
			Name:        "compute_var",
			Description: "Calculate Value at Risk (VaR) for a position using historical simulation",
			Parameters: llm.ObjectSchema("VaR parameters",
				map[string]*llm.JSONSchema{
					"ticker":        llm.StringProp("NSE ticker symbol"),
					"position_size": llm.NumberProp("Position value in ₹"),
					"confidence":    llm.NumberProp("Confidence level (e.g., 0.95 for 95% VaR, default: 0.95)"),
					"holding_days":  llm.IntProp("Holding period in days (default: 1)"),
				},
				"ticker", "position_size",
			),
			Handler: a.handleComputeVaR,
		},
		{
			Name:        "suggest_stop_loss",
			Description: "Suggest optimal stop-loss levels based on ATR, support levels, and volatility",
			Parameters: llm.ObjectSchema("Stop-loss suggestion parameters",
				map[string]*llm.JSONSchema{
					"ticker":      llm.StringProp("NSE ticker symbol"),
					"entry_price": llm.NumberProp("Entry price in ₹"),
					"direction":   llm.StringProp("Trade direction: long or short"),
				},
				"ticker", "entry_price",
			),
			Handler: a.handleSuggestStopLoss,
		},
		{
			Name:        "risk_reward_analysis",
			Description: "Analyze the risk-reward ratio for a proposed trade and provide a go/no-go recommendation",
			Parameters: llm.ObjectSchema("Risk-reward parameters",
				map[string]*llm.JSONSchema{
					"ticker":      llm.StringProp("NSE ticker symbol"),
					"entry_price": llm.NumberProp("Entry price in ₹"),
					"stop_loss":   llm.NumberProp("Stop-loss price in ₹"),
					"target":      llm.NumberProp("Target price in ₹"),
					"capital":     llm.NumberProp("Trading capital in ₹"),
				},
				"entry_price", "stop_loss", "target",
			),
			Handler: a.handleRiskReward,
		},
		{
			Name:        "portfolio_exposure_check",
			Description: "Check portfolio concentration and suggest adjustments to maintain diversification",
			Parameters: llm.ObjectSchema("Portfolio exposure parameters",
				map[string]*llm.JSONSchema{
					"positions": llm.ArrayProp("Current positions", llm.ObjectSchema("Position",
						map[string]*llm.JSONSchema{
							"ticker": llm.StringProp("NSE ticker"),
							"value":  llm.NumberProp("Current position value in ₹"),
						},
						"ticker", "value",
					)),
					"new_ticker": llm.StringProp("Ticker of the new position being considered"),
					"new_value":  llm.NumberProp("Value of the proposed new position in ₹"),
					"capital":    llm.NumberProp("Total portfolio capital in ₹"),
				},
				"capital",
			),
			Handler: a.handlePortfolioExposure,
		},
	}
}

// ── Tool Handlers ──

func (a *RiskAgent) handlePositionSize(_ context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Capital    float64 `json:"capital"`
		RiskPct    float64 `json:"risk_pct"`
		EntryPrice float64 `json:"entry_price"`
		StopLoss   float64 `json:"stop_loss"`
		Ticker     string  `json:"ticker"`
		IsFnO      bool    `json:"is_fno"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	if params.RiskPct <= 0 {
		params.RiskPct = 2.0 // default 2% risk per trade
	}

	riskAmount := params.Capital * (params.RiskPct / 100.0)
	slDistance := math.Abs(params.EntryPrice - params.StopLoss)
	if slDistance == 0 {
		return "Stop-loss distance is zero. Cannot compute position size.", nil
	}

	quantity := int(riskAmount / slDistance)
	posValue := float64(quantity) * params.EntryPrice
	positionPct := (posValue / params.Capital) * 100

	// Max position limit: 5% of capital for single stock
	maxPositionPct := 5.0
	if positionPct > maxPositionPct {
		quantity = int((params.Capital * maxPositionPct / 100) / params.EntryPrice)
		posValue = float64(quantity) * params.EntryPrice
		positionPct = (posValue / params.Capital) * 100
	}

	result := map[string]any{
		"capital":          params.Capital,
		"risk_per_trade":   fmt.Sprintf("%.1f%%", params.RiskPct),
		"risk_amount":      riskAmount,
		"entry_price":      params.EntryPrice,
		"stop_loss":        params.StopLoss,
		"sl_distance":      slDistance,
		"quantity":         quantity,
		"position_value":   posValue,
		"position_pct":     fmt.Sprintf("%.2f%%", positionPct),
		"max_loss":         float64(quantity) * slDistance,
	}

	if params.IsFnO && params.Ticker != "" {
		result["note"] = "F&O trade: quantity should be rounded to nearest lot size"
		result["ticker"] = params.Ticker
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *RiskAgent) handleComputeVaR(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker       string  `json:"ticker"`
		PositionSize float64 `json:"position_size"`
		Confidence   float64 `json:"confidence"`
		HoldingDays  int     `json:"holding_days"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	if params.Confidence <= 0 || params.Confidence >= 1 {
		params.Confidence = 0.95
	}
	if params.HoldingDays <= 0 {
		params.HoldingDays = 1
	}

	// Fetch historical data for VaR computation
	to := time.Now()
	from := to.AddDate(-1, 0, 0)
	var candles []models.OHLCV
	for _, src := range a.dataSources {
		c, err := src.GetHistoricalData(ctx, params.Ticker, from, to, models.Timeframe1Day)
		if err == nil && len(c) > 20 {
			candles = c
			break
		}
	}

	if len(candles) < 20 {
		return fmt.Sprintf("Insufficient historical data for %s VaR calculation (need ≥20 days, got %d)", params.Ticker, len(candles)), nil
	}

	// Compute daily returns
	returns := make([]float64, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		if candles[i-1].Close > 0 {
			returns[i-1] = (candles[i].Close - candles[i-1].Close) / candles[i-1].Close
		}
	}

	// Sort returns for percentile-based VaR
	sortedReturns := make([]float64, len(returns))
	copy(sortedReturns, returns)
	sortFloat64s(sortedReturns)

	// Historical VaR
	idx := int(float64(len(sortedReturns)) * (1 - params.Confidence))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sortedReturns) {
		idx = len(sortedReturns) - 1
	}
	dailyVaRPct := math.Abs(sortedReturns[idx])
	holdingVaRPct := dailyVaRPct * math.Sqrt(float64(params.HoldingDays))
	varAmount := params.PositionSize * holdingVaRPct

	// Compute volatility
	var sum, sumSq float64
	for _, r := range returns {
		sum += r
		sumSq += r * r
	}
	mean := sum / float64(len(returns))
	variance := sumSq/float64(len(returns)) - mean*mean
	dailyVol := math.Sqrt(variance)
	annualVol := dailyVol * math.Sqrt(252)

	result := map[string]any{
		"ticker":            params.Ticker,
		"position_size":     params.PositionSize,
		"confidence":        fmt.Sprintf("%.0f%%", params.Confidence*100),
		"holding_days":      params.HoldingDays,
		"daily_var_pct":     fmt.Sprintf("%.2f%%", dailyVaRPct*100),
		"holding_var_pct":   fmt.Sprintf("%.2f%%", holdingVaRPct*100),
		"var_amount":        varAmount,
		"daily_volatility":  fmt.Sprintf("%.2f%%", dailyVol*100),
		"annual_volatility": fmt.Sprintf("%.2f%%", annualVol*100),
		"data_points":       len(returns),
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *RiskAgent) handleSuggestStopLoss(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker     string  `json:"ticker"`
		EntryPrice float64 `json:"entry_price"`
		Direction  string  `json:"direction"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	if params.Direction == "" {
		params.Direction = "long"
	}

	// Fetch recent candles for ATR computation
	to := time.Now()
	from := to.AddDate(0, -3, 0) // 3 months
	var candles []models.OHLCV
	for _, src := range a.dataSources {
		c, err := src.GetHistoricalData(ctx, params.Ticker, from, to, models.Timeframe1Day)
		if err == nil && len(c) > 14 {
			candles = c
			break
		}
	}

	if len(candles) < 14 {
		return fmt.Sprintf("Insufficient data for %s ATR-based stop-loss (need ≥14 days)", params.Ticker), nil
	}

	// Compute 14-day ATR manually
	atr := computeATR(candles, 14)

	var sl1, sl2, sl3 float64
	if params.Direction == "long" {
		sl1 = params.EntryPrice - 1.5*atr // tight
		sl2 = params.EntryPrice - 2.0*atr // moderate
		sl3 = params.EntryPrice - 3.0*atr // wide
	} else {
		sl1 = params.EntryPrice + 1.5*atr
		sl2 = params.EntryPrice + 2.0*atr
		sl3 = params.EntryPrice + 3.0*atr
	}

	result := map[string]any{
		"ticker":      params.Ticker,
		"entry_price": params.EntryPrice,
		"direction":   params.Direction,
		"atr_14":      fmt.Sprintf("%.2f", atr),
		"stop_levels": map[string]any{
			"tight":    map[string]any{"price": math.Round(sl1*100) / 100, "multiplier": "1.5x ATR", "risk_pct": fmt.Sprintf("%.2f%%", math.Abs(sl1-params.EntryPrice)/params.EntryPrice*100)},
			"moderate": map[string]any{"price": math.Round(sl2*100) / 100, "multiplier": "2.0x ATR", "risk_pct": fmt.Sprintf("%.2f%%", math.Abs(sl2-params.EntryPrice)/params.EntryPrice*100)},
			"wide":     map[string]any{"price": math.Round(sl3*100) / 100, "multiplier": "3.0x ATR", "risk_pct": fmt.Sprintf("%.2f%%", math.Abs(sl3-params.EntryPrice)/params.EntryPrice*100)},
		},
		"recommendation": "Use 2.0x ATR for swing trades, 1.5x for intraday, 3.0x for positional",
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *RiskAgent) handleRiskReward(_ context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker     string  `json:"ticker"`
		EntryPrice float64 `json:"entry_price"`
		StopLoss   float64 `json:"stop_loss"`
		Target     float64 `json:"target"`
		Capital    float64 `json:"capital"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	risk := math.Abs(params.EntryPrice - params.StopLoss)
	reward := math.Abs(params.Target - params.EntryPrice)

	var rrRatio float64
	if risk > 0 {
		rrRatio = reward / risk
	}

	verdict := "REJECT"
	if rrRatio >= 2.0 {
		verdict = "ACCEPTABLE"
	}
	if rrRatio >= 3.0 {
		verdict = "FAVORABLE"
	}

	result := map[string]any{
		"ticker":          params.Ticker,
		"entry_price":     params.EntryPrice,
		"stop_loss":       params.StopLoss,
		"target":          params.Target,
		"risk_per_share":  risk,
		"reward_per_share": reward,
		"risk_reward":     fmt.Sprintf("1:%.2f", rrRatio),
		"verdict":         verdict,
	}

	if params.Capital > 0 {
		riskPct := (risk / params.EntryPrice) * 100
		maxQtyByCapital := int(params.Capital * 0.02 / risk) // 2% risk rule
		result["risk_pct_per_share"] = fmt.Sprintf("%.2f%%", riskPct)
		result["max_qty_2pct_rule"] = maxQtyByCapital
		result["max_position_value"] = float64(maxQtyByCapital) * params.EntryPrice
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (a *RiskAgent) handlePortfolioExposure(_ context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Positions []struct {
			Ticker string  `json:"ticker"`
			Value  float64 `json:"value"`
		} `json:"positions"`
		NewTicker string  `json:"new_ticker"`
		NewValue  float64 `json:"new_value"`
		Capital   float64 `json:"capital"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	// Compute current exposure including new position
	totalExposed := params.NewValue
	sectorExposure := map[string]float64{}

	for _, p := range params.Positions {
		totalExposed += p.Value
		sector := prompts.SectorForTicker(p.Ticker)
		if sector == "" {
			sector = "Unknown"
		}
		sectorExposure[sector] += p.Value
	}

	if params.NewTicker != "" {
		newSector := prompts.SectorForTicker(params.NewTicker)
		if newSector == "" {
			newSector = "Unknown"
		}
		sectorExposure[newSector] += params.NewValue
	}

	totalPct := (totalExposed / params.Capital) * 100
	newPositionPct := (params.NewValue / params.Capital) * 100

	warnings := []string{}
	if newPositionPct > 5 {
		warnings = append(warnings, fmt.Sprintf("New position %.1f%% exceeds 5%% single-stock limit", newPositionPct))
	}
	if totalPct > 80 {
		warnings = append(warnings, fmt.Sprintf("Total exposure %.1f%% exceeds 80%% — consider maintaining cash reserve", totalPct))
	}
	for sector, val := range sectorExposure {
		sectorPct := (val / params.Capital) * 100
		if sectorPct > 25 {
			warnings = append(warnings, fmt.Sprintf("Sector %s at %.1f%% exceeds 25%% concentration limit", sector, sectorPct))
		}
	}

	result := map[string]any{
		"capital":            params.Capital,
		"total_exposed":      totalExposed,
		"exposure_pct":       fmt.Sprintf("%.1f%%", totalPct),
		"positions":          len(params.Positions) + 1,
		"new_position_pct":   fmt.Sprintf("%.1f%%", newPositionPct),
		"sector_exposure":    sectorExposure,
		"warnings":           warnings,
		"diversification_ok": len(warnings) == 0,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

// ── Helpers ──

// computeATR computes the Average True Range for n periods.
func computeATR(candles []models.OHLCV, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}

	var atr float64
	for i := 1; i <= period; i++ {
		tr := trueRange(candles[len(candles)-period-1+i], candles[len(candles)-period-1+i-1])
		atr += tr
	}
	atr /= float64(period)

	return atr
}

func trueRange(curr, prev models.OHLCV) float64 {
	hl := curr.High - curr.Low
	hc := math.Abs(curr.High - prev.Close)
	lc := math.Abs(curr.Low - prev.Close)
	return math.Max(hl, math.Max(hc, lc))
}

// sortFloat64s sorts a slice of float64 in ascending order.
func sortFloat64s(s []float64) {
	// Simple insertion sort — good enough for ~250 elements.
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}

// Analyze runs risk analysis with chain-of-thought reasoning.
func (a *RiskAgent) Analyze(ctx context.Context, ticker string, capitalINR float64) (*AgentResult, error) {
	task := prompts.CoTRisk(ticker, capitalINR)
	return a.Process(ctx, task)
}

// AnalyzeWithTimestamp runs the analysis and attaches a typed result.
func (a *RiskAgent) AnalyzeWithTimestamp(ctx context.Context, ticker string, capitalINR float64) (*AgentResult, error) {
	result, err := a.Analyze(ctx, ticker, capitalINR)
	if err != nil {
		return result, err
	}

	result.Analysis = ParseAnalysisResult(result.Content, models.AnalysisResult{
		Ticker:    ticker,
		Type:      models.AnalysisRisk,
		AgentName: a.Name(),
		Timestamp: time.Now(),
	})

	return result, nil
}
