package models

import "time"

// OptionChain represents the full option chain for a ticker on an expiry date.
type OptionChain struct {
	Ticker       string           `json:"ticker"`
	SpotPrice    float64          `json:"spot_price"`
	ExpiryDate   string           `json:"expiry_date"`
	Expiries     []string         `json:"expiries"`      // all available expiry dates
	Contracts    []OptionContract `json:"contracts"`
	TotalCEOI    int64            `json:"total_ce_oi"`
	TotalPEOI    int64            `json:"total_pe_oi"`
	PCR          float64          `json:"pcr"`            // Put-Call Ratio
	MaxPain      float64          `json:"max_pain"`
	FetchedAt    time.Time        `json:"fetched_at"`
}

// OptionContract represents a single option contract (CE or PE) at a strike.
type OptionContract struct {
	StrikePrice    float64 `json:"strike_price"`
	OptionType     string  `json:"option_type"`     // "CE" or "PE"
	ExpiryDate     string  `json:"expiry_date"`
	LTP            float64 `json:"ltp"`             // Last Traded Price
	Change         float64 `json:"change"`
	ChangePct      float64 `json:"change_pct"`
	Volume         int64   `json:"volume"`
	OI             int64   `json:"oi"`              // Open Interest
	OIChange       int64   `json:"oi_change"`
	OIChangePct    float64 `json:"oi_change_pct"`
	BidPrice       float64 `json:"bid_price"`
	AskPrice       float64 `json:"ask_price"`
	BidQty         int64   `json:"bid_qty"`
	AskQty         int64   `json:"ask_qty"`
	IV             float64 `json:"iv"`              // Implied Volatility
	// Greeks (computed)
	Delta          float64 `json:"delta,omitempty"`
	Gamma          float64 `json:"gamma,omitempty"`
	Theta          float64 `json:"theta,omitempty"`
	Vega           float64 `json:"vega,omitempty"`
}

// FuturesContract represents a single futures contract.
type FuturesContract struct {
	Ticker       string    `json:"ticker"`
	ExpiryDate   string    `json:"expiry_date"`
	LTP          float64   `json:"ltp"`
	Change       float64   `json:"change"`
	ChangePct    float64   `json:"change_pct"`
	Volume       int64     `json:"volume"`
	OI           int64     `json:"oi"`
	OIChange     int64     `json:"oi_change"`
	Basis        float64   `json:"basis"`          // Futures price - Spot price
	BasisPct     float64   `json:"basis_pct"`
	LotSize      int       `json:"lot_size"`
	FetchedAt    time.Time `json:"fetched_at"`
}

// OIBuildup represents OI change classification.
type OIBuildupType string

const (
	LongBuildup    OIBuildupType = "long_buildup"     // Price ↑, OI ↑
	ShortBuildup   OIBuildupType = "short_buildup"    // Price ↓, OI ↑
	LongUnwinding  OIBuildupType = "long_unwinding"   // Price ↓, OI ↓
	ShortCovering  OIBuildupType = "short_covering"   // Price ↑, OI ↓
)

// OIBuildupData holds OI buildup classification for a contract.
type OIBuildupData struct {
	Ticker     string        `json:"ticker"`
	Buildup    OIBuildupType `json:"buildup"`
	PriceChange float64      `json:"price_change"`
	OIChange   int64         `json:"oi_change"`
	OIChangePct float64      `json:"oi_change_pct"`
}

// OptionPayoff represents a data point in an option strategy payoff chart.
type OptionPayoff struct {
	UnderlyingPrice float64 `json:"underlying_price"`
	PnL             float64 `json:"pnl"`
}

// OptionStrategy represents a multi-leg option strategy.
type OptionStrategy struct {
	Name       string           `json:"name"`        // e.g., "Bull Call Spread"
	Legs       []OptionLeg      `json:"legs"`
	MaxProfit  float64          `json:"max_profit"`
	MaxLoss    float64          `json:"max_loss"`
	Breakevens []float64        `json:"breakevens"`
	NetPremium float64          `json:"net_premium"` // positive = credit, negative = debit
	Payoff     []OptionPayoff   `json:"payoff,omitempty"`
}

// OptionLeg represents a single leg of an option strategy.
type OptionLeg struct {
	OptionType  string  `json:"option_type"` // "CE" or "PE"
	StrikePrice float64 `json:"strike_price"`
	Action      string  `json:"action"`       // "BUY" or "SELL"
	Lots        int     `json:"lots"`
	Premium     float64 `json:"premium"`
}

// IndiaVIX represents the India VIX (volatility index) data.
type IndiaVIX struct {
	Value     float64   `json:"value"`
	Change    float64   `json:"change"`
	ChangePct float64   `json:"change_pct"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	PrevClose float64   `json:"prev_close"`
	Timestamp time.Time `json:"timestamp"`
}

// FIIDIIData represents FII/DII daily activity data.
type FIIDIIData struct {
	Date       string  `json:"date"`
	FIIBuy     float64 `json:"fii_buy"`     // in crores
	FIISell    float64 `json:"fii_sell"`
	FIINet     float64 `json:"fii_net"`
	DIIBuy     float64 `json:"dii_buy"`
	DIISell    float64 `json:"dii_sell"`
	DIINet     float64 `json:"dii_net"`
}
