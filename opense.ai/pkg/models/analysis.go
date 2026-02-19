package models

import "time"

// SignalType represents a trading signal direction.
type SignalType string

const (
	SignalBuy     SignalType = "BUY"
	SignalSell    SignalType = "SELL"
	SignalNeutral SignalType = "NEUTRAL"
)

// Confidence represents the strength of a signal (0.0 to 1.0).
type Confidence float64

// Signal represents a single trading signal from an indicator or analysis.
type Signal struct {
	Source     string     `json:"source"`      // e.g., "RSI", "MACD", "Fundamental"
	Type       SignalType `json:"type"`
	Confidence Confidence `json:"confidence"`  // 0.0 to 1.0
	Reason     string     `json:"reason"`      // human-readable explanation
	Price      float64    `json:"price,omitempty"`
	Target     float64    `json:"target,omitempty"`
	StopLoss   float64    `json:"stop_loss,omitempty"`
}

// AnalysisType represents the kind of analysis performed.
type AnalysisType string

const (
	AnalysisTechnical    AnalysisType = "technical"
	AnalysisFundamental  AnalysisType = "fundamental"
	AnalysisDerivatives  AnalysisType = "derivatives"
	AnalysisSentiment    AnalysisType = "sentiment"
	AnalysisRisk         AnalysisType = "risk"
	AnalysisComposite    AnalysisType = "composite"
)

// Recommendation represents the final recommendation for a stock.
type Recommendation string

const (
	StrongBuy  Recommendation = "STRONG_BUY"
	ModerateBuy Recommendation = "BUY"
	Hold       Recommendation = "HOLD"
	ModerateSell Recommendation = "SELL"
	StrongSell Recommendation = "STRONG_SELL"
)

// AnalysisResult represents the output of a single analysis agent.
type AnalysisResult struct {
	Ticker         string         `json:"ticker"`
	Type           AnalysisType   `json:"type"`
	AgentName      string         `json:"agent_name"`
	Signals        []Signal       `json:"signals"`
	Recommendation Recommendation `json:"recommendation"`
	Confidence     Confidence     `json:"confidence"`
	Summary        string         `json:"summary"`       // LLM-generated summary
	Details        map[string]any `json:"details"`       // agent-specific details
	Timestamp      time.Time      `json:"timestamp"`
}

// CompositeAnalysis represents the final synthesized analysis across all agents.
type CompositeAnalysis struct {
	Ticker          string           `json:"ticker"`
	StockProfile    StockProfile     `json:"stock_profile"`
	Technical       *AnalysisResult  `json:"technical,omitempty"`
	Fundamental     *AnalysisResult  `json:"fundamental,omitempty"`
	Derivatives     *AnalysisResult  `json:"derivatives,omitempty"`
	Sentiment       *AnalysisResult  `json:"sentiment,omitempty"`
	Risk            *AnalysisResult  `json:"risk,omitempty"`
	Recommendation  Recommendation   `json:"recommendation"`
	Confidence      Confidence       `json:"confidence"`
	Summary         string           `json:"summary"`
	EntryPrice      float64          `json:"entry_price,omitempty"`
	TargetPrice     float64          `json:"target_price,omitempty"`
	StopLoss        float64          `json:"stop_loss,omitempty"`
	PositionSize    int              `json:"position_size,omitempty"`
	RiskRewardRatio float64          `json:"risk_reward_ratio,omitempty"`
	Timeframe       string           `json:"timeframe"`  // e.g., "short-term", "medium-term"
	Timestamp       time.Time        `json:"timestamp"`
}

// SentimentScore represents sentiment analysis output for a single source.
type SentimentScore struct {
	Source     string    `json:"source"`      // e.g., "Moneycontrol", "Economic Times"
	Headline   string    `json:"headline"`
	Score      float64   `json:"score"`       // -1.0 (very bearish) to +1.0 (very bullish)
	Confidence Confidence `json:"confidence"`
	URL        string    `json:"url,omitempty"`
	PublishedAt time.Time `json:"published_at"`
}

// AggregatedSentiment represents the combined sentiment across sources.
type AggregatedSentiment struct {
	Ticker     string           `json:"ticker"`
	Score      float64          `json:"score"`       // weighted average sentiment
	Confidence Confidence       `json:"confidence"`
	Label      string           `json:"label"`       // "Bullish", "Bearish", "Neutral"
	Sources    []SentimentScore `json:"sources"`
	ArticleCount int            `json:"article_count"`
	Timestamp  time.Time        `json:"timestamp"`
}

// TechnicalIndicators holds computed indicator values for a stock.
type TechnicalIndicators struct {
	Ticker    string    `json:"ticker"`
	RSI       float64   `json:"rsi"`
	MACD      MACDData  `json:"macd"`
	SMA       map[int]float64 `json:"sma"`       // period → value (e.g., 20 → 2845.5)
	EMA       map[int]float64 `json:"ema"`
	Bollinger BollingerData   `json:"bollinger"`
	SuperTrend SuperTrendData  `json:"supertrend"`
	ATR       float64   `json:"atr"`
	VWAP      float64   `json:"vwap"`
	Timestamp time.Time `json:"timestamp"`
}

// MACDData contains MACD indicator values.
type MACDData struct {
	MACDLine   float64 `json:"macd_line"`
	SignalLine float64 `json:"signal_line"`
	Histogram  float64 `json:"histogram"`
}

// BollingerData contains Bollinger Bands values.
type BollingerData struct {
	Upper  float64 `json:"upper"`
	Middle float64 `json:"middle"`
	Lower  float64 `json:"lower"`
}

// SuperTrendData contains SuperTrend indicator values.
type SuperTrendData struct {
	Value    float64 `json:"value"`
	Trend    string  `json:"trend"` // "UP" or "DOWN"
}

// SupportResistance represents support and resistance levels.
type SupportResistance struct {
	Ticker      string    `json:"ticker"`
	Supports    []float64 `json:"supports"`
	Resistances []float64 `json:"resistances"`
	PivotPoint  float64   `json:"pivot_point"`
	S1          float64   `json:"s1"`
	S2          float64   `json:"s2"`
	S3          float64   `json:"s3"`
	R1          float64   `json:"r1"`
	R2          float64   `json:"r2"`
	R3          float64   `json:"r3"`
	Method      string    `json:"method"` // "classic", "fibonacci", "camarilla"
}

// BacktestResult represents the outcome of a backtest run.
type BacktestResult struct {
	StrategyName    string    `json:"strategy_name"`
	Ticker          string    `json:"ticker"`
	From            time.Time `json:"from"`
	To              time.Time `json:"to"`
	InitialCapital  float64   `json:"initial_capital"`
	FinalCapital    float64   `json:"final_capital"`
	TotalReturn     float64   `json:"total_return"`
	TotalReturnPct  float64   `json:"total_return_pct"`
	CAGR            float64   `json:"cagr"`
	SharpeRatio     float64   `json:"sharpe_ratio"`
	SortinoRatio    float64   `json:"sortino_ratio"`
	MaxDrawdown     float64   `json:"max_drawdown"`
	MaxDrawdownPct  float64   `json:"max_drawdown_pct"`
	WinRate         float64   `json:"win_rate"`
	ProfitFactor    float64   `json:"profit_factor"`
	TotalTrades     int       `json:"total_trades"`
	WinningTrades   int       `json:"winning_trades"`
	LosingTrades    int       `json:"losing_trades"`
	AvgWin          float64   `json:"avg_win"`
	AvgLoss         float64   `json:"avg_loss"`
	EquityCurve     []EquityPoint `json:"equity_curve"`
	Trades          []BacktestTrade `json:"trades"`
	BenchmarkReturn float64   `json:"benchmark_return,omitempty"`
}

// EquityPoint represents a point on the equity curve.
type EquityPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// BacktestTrade represents a single trade from a backtest.
type BacktestTrade struct {
	EntryDate  time.Time `json:"entry_date"`
	ExitDate   time.Time `json:"exit_date"`
	Side       OrderSide `json:"side"`
	EntryPrice float64   `json:"entry_price"`
	ExitPrice  float64   `json:"exit_price"`
	Quantity   int       `json:"quantity"`
	PnL        float64   `json:"pnl"`
	PnLPct     float64   `json:"pnl_pct"`
	Reason     string    `json:"reason"` // why the trade was taken/exited
}

// NewsArticle represents a single news article.
type NewsArticle struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	Summary     string    `json:"summary,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	Tickers     []string  `json:"tickers,omitempty"` // related tickers
}
