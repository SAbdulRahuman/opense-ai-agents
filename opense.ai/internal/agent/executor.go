package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/agent/prompts"
	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/pkg/models"
)

// ExecutorAgent is the Trade Executor specialized agent.
// It translates analysis recommendations into actionable trade plans with
// human-in-the-loop confirmation. NEVER executes trades without explicit approval.
type ExecutorAgent struct {
	*BaseAgent
}

// TradeProposal represents a structured trade proposal requiring human approval.
type TradeProposal struct {
	ID         string    `json:"id"`
	Ticker     string    `json:"ticker"`
	Action     string    `json:"action"`      // BUY, SELL, HOLD
	OrderType  string    `json:"order_type"`   // LIMIT, MARKET, SL, SL-M
	Price      float64   `json:"price"`        // target entry price
	StopLoss   float64   `json:"stop_loss"`
	Target     float64   `json:"target"`
	Quantity   int       `json:"quantity"`
	Rationale  string    `json:"rationale"`
	RiskReward string    `json:"risk_reward"`
	Approved   bool      `json:"approved"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewExecutorAgent creates a Trade Executor agent.
func NewExecutorAgent(provider llm.LLMProvider, opts *llm.ChatOptions) *ExecutorAgent {
	agent := &ExecutorAgent{}

	tools := agent.buildTools()

	systemPrompt := prompts.ExecutorSystemPrompt + prompts.IndianMarketPromptSuffix()

	agent.BaseAgent = NewBaseAgent(BaseAgentConfig{
		Name:         prompts.AgentExecutor,
		Role:         "Trade Executor — Order planning with human-in-the-loop confirmation",
		SystemPrompt: systemPrompt,
		Provider:     provider,
		Tools:        tools,
		ChatOptions:  opts,
		MemorySize:   20,
		MaxToolIter:  4,
	})

	return agent
}

func (a *ExecutorAgent) buildTools() []llm.Tool {
	return []llm.Tool{
		{
			Name:        "create_trade_proposal",
			Description: "Create a structured trade proposal from analysis recommendations. The proposal requires human approval before any execution.",
			Parameters: llm.ObjectSchema("Trade proposal parameters",
				map[string]*llm.JSONSchema{
					"ticker":      llm.StringProp("NSE ticker symbol"),
					"action":      llm.StringProp("Trade action: BUY, SELL, or HOLD"),
					"order_type":  llm.StringProp("Order type: LIMIT (default), MARKET, SL, SL-M"),
					"price":       llm.NumberProp("Entry price in ₹"),
					"stop_loss":   llm.NumberProp("Stop-loss price in ₹"),
					"target":      llm.NumberProp("Target price in ₹"),
					"quantity":    llm.IntProp("Number of shares"),
					"rationale":   llm.StringProp("Brief rationale for the trade"),
				},
				"ticker", "action", "price",
			),
			Handler: a.handleCreateProposal,
		},
		{
			Name:        "estimate_brokerage",
			Description: "Estimate brokerage and charges (STT, exchange fees, GST, stamp duty) for a proposed trade",
			Parameters: llm.ObjectSchema("Brokerage estimation parameters",
				map[string]*llm.JSONSchema{
					"buy_price":   llm.NumberProp("Buy price per share in ₹"),
					"sell_price":  llm.NumberProp("Sell price per share in ₹"),
					"quantity":    llm.IntProp("Number of shares"),
					"is_delivery": llm.BoolProp("True for delivery trade, false for intraday"),
				},
				"buy_price", "sell_price", "quantity",
			),
			Handler: a.handleEstimateBrokerage,
		},
		{
			Name:        "validate_trade",
			Description: "Validate a trade proposal against risk rules: max position size 5%, max daily loss 2%, circuit limits, market hours",
			Parameters: llm.ObjectSchema("Trade validation parameters",
				map[string]*llm.JSONSchema{
					"ticker":    llm.StringProp("NSE ticker symbol"),
					"action":    llm.StringProp("BUY or SELL"),
					"price":     llm.NumberProp("Trade price in ₹"),
					"quantity":  llm.IntProp("Number of shares"),
					"capital":   llm.NumberProp("Total trading capital in ₹"),
					"stop_loss": llm.NumberProp("Stop-loss price in ₹"),
				},
				"ticker", "action", "price", "quantity",
			),
			Handler: a.handleValidateTrade,
		},
	}
}

// ── Tool Handlers ──

func (a *ExecutorAgent) handleCreateProposal(_ context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker    string  `json:"ticker"`
		Action    string  `json:"action"`
		OrderType string  `json:"order_type"`
		Price     float64 `json:"price"`
		StopLoss  float64 `json:"stop_loss"`
		Target    float64 `json:"target"`
		Quantity  int     `json:"quantity"`
		Rationale string  `json:"rationale"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	if params.OrderType == "" {
		params.OrderType = "LIMIT"
	}

	var rr string
	if params.StopLoss > 0 && params.Target > 0 {
		risk := abs(params.Price - params.StopLoss)
		reward := abs(params.Target - params.Price)
		if risk > 0 {
			rr = fmt.Sprintf("1:%.2f", reward/risk)
		}
	}

	proposal := TradeProposal{
		ID:         fmt.Sprintf("%s-%s-%d", params.Ticker, params.Action, time.Now().UnixMilli()),
		Ticker:     params.Ticker,
		Action:     params.Action,
		OrderType:  params.OrderType,
		Price:      params.Price,
		StopLoss:   params.StopLoss,
		Target:     params.Target,
		Quantity:   params.Quantity,
		Rationale:  params.Rationale,
		RiskReward: rr,
		Approved:   false,
		CreatedAt:  time.Now(),
	}

	data, _ := json.MarshalIndent(proposal, "", "  ")

	return fmt.Sprintf(
		"⚠️ TRADE PROPOSAL CREATED — REQUIRES HUMAN APPROVAL ⚠️\n\n%s\n\n"+
			"This trade will NOT be executed until explicitly approved by the user.",
		string(data),
	), nil
}

func (a *ExecutorAgent) handleEstimateBrokerage(_ context.Context, args json.RawMessage) (string, error) {
	var params struct {
		BuyPrice   float64 `json:"buy_price"`
		SellPrice  float64 `json:"sell_price"`
		Quantity   int     `json:"quantity"`
		IsDelivery bool    `json:"is_delivery"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	charges := prompts.IndianBrokerageEstimate(params.BuyPrice, params.SellPrice, params.Quantity, params.IsDelivery)
	data, _ := json.MarshalIndent(charges, "", "  ")
	return string(data), nil
}

func (a *ExecutorAgent) handleValidateTrade(_ context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Ticker   string  `json:"ticker"`
		Action   string  `json:"action"`
		Price    float64 `json:"price"`
		Quantity int     `json:"quantity"`
		Capital  float64 `json:"capital"`
		StopLoss float64 `json:"stop_loss"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	posValue := params.Price * float64(params.Quantity)
	issues := []string{}
	warnings := []string{}

	// Max position size check (5% of capital)
	if params.Capital > 0 {
		posPct := (posValue / params.Capital) * 100
		if posPct > 5.0 {
			issues = append(issues, fmt.Sprintf("POSITION_SIZE: %.1f%% exceeds 5%% max (₹%.0f of ₹%.0f)", posPct, posValue, params.Capital))
		} else if posPct > 3.0 {
			warnings = append(warnings, fmt.Sprintf("Position is %.1f%% of capital — on the higher side", posPct))
		}
	}

	// Max loss check (2% of capital)
	if params.Capital > 0 && params.StopLoss > 0 {
		maxLoss := abs(params.Price-params.StopLoss) * float64(params.Quantity)
		lossPct := (maxLoss / params.Capital) * 100
		if lossPct > 2.0 {
			issues = append(issues, fmt.Sprintf("MAX_LOSS: %.1f%% exceeds 2%% daily loss limit (₹%.0f)", lossPct, maxLoss))
		}
	}

	// Basic sanity
	if params.Price <= 0 {
		issues = append(issues, "PRICE: Invalid price (≤ 0)")
	}
	if params.Quantity <= 0 {
		issues = append(issues, "QUANTITY: Invalid quantity (≤ 0)")
	}

	valid := len(issues) == 0

	result := map[string]any{
		"ticker":         params.Ticker,
		"action":         params.Action,
		"valid":          valid,
		"position_value": posValue,
	}
	if len(issues) > 0 {
		result["issues"] = issues
	}
	if len(warnings) > 0 {
		result["warnings"] = warnings
	}
	if valid {
		result["message"] = "Trade passes all pre-trade risk checks"
	} else {
		result["message"] = "Trade REJECTED — fix issues before proceeding"
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// CreateTradeProposal creates a trade proposal from analysis results using the LLM.
func (a *ExecutorAgent) CreateTradeProposal(ctx context.Context, ticker string, analyses []*AgentResult) (*AgentResult, error) {
	// Build context from previous analyses
	task := fmt.Sprintf("Based on the following analysis results for %s, create a detailed trade proposal.\n\n", ticker)

	for _, ar := range analyses {
		if ar != nil && ar.Content != "" {
			task += fmt.Sprintf("--- %s (%s) ---\n%s\n\n", ar.AgentName, ar.Role, ar.Content)
		}
	}

	task += "Create a trade proposal with: action (BUY/SELL/HOLD), order type, entry price, stop-loss, target, quantity rationale, and risk-reward ratio. " +
		"If the analysis is conflicting or unclear, recommend HOLD with explanation."

	return a.Process(ctx, task)
}

// Analyze processes a generic executor task.
func (a *ExecutorAgent) Analyze(ctx context.Context, task string) (*AgentResult, error) {
	return a.Process(ctx, task)
}

// AnalyzeWithTimestamp attaches a typed result.
func (a *ExecutorAgent) AnalyzeWithTimestamp(ctx context.Context, task string, ticker string) (*AgentResult, error) {
	result, err := a.Process(ctx, task)
	if err != nil {
		return result, err
	}

	result.Analysis = ParseAnalysisResult(result.Content, models.AnalysisResult{
		Ticker:    ticker,
		Type:      "execution",
		AgentName: a.Name(),
		Timestamp: time.Now(),
	})

	return result, nil
}
