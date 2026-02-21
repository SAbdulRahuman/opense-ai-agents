package models

import "time"

// --- Extended Derivatives Models ---
// Complements option.go with global options/futures data structures.

// OptionsChainData represents a provider-agnostic option chain (global format).
type OptionsChainData struct {
	Symbol     string             `json:"symbol"`
	Underlying string             `json:"underlying,omitempty"`
	Exchange   string             `json:"exchange,omitempty"`
	Expiries   []time.Time        `json:"expiries"`
	Contracts  []OptionContractEx `json:"contracts"`
	FetchedAt  time.Time          `json:"fetched_at"`
}

// OptionContractEx represents an extended option contract with additional fields
// beyond the existing OptionContract (which is India/NSE-focused).
type OptionContractEx struct {
	Symbol         string    `json:"symbol"`
	Underlying     string    `json:"underlying"`
	StrikePrice    float64   `json:"strike_price"`
	OptionType     string    `json:"option_type"` // "call" or "put"
	ExpirationDate time.Time `json:"expiration_date"`
	ContractSymbol string    `json:"contract_symbol,omitempty"`
	Exchange       string    `json:"exchange,omitempty"`

	// Price data
	LastPrice float64 `json:"last_price"`
	BidPrice  float64 `json:"bid_price"`
	AskPrice  float64 `json:"ask_price"`
	BidSize   int64   `json:"bid_size,omitempty"`
	AskSize   int64   `json:"ask_size,omitempty"`
	Volume    int64   `json:"volume"`
	OpenInterest int64 `json:"open_interest"`
	Change    float64 `json:"change,omitempty"`
	ChangePct float64 `json:"change_pct,omitempty"`

	// Greeks
	IV    float64 `json:"iv,omitempty"`
	Delta float64 `json:"delta,omitempty"`
	Gamma float64 `json:"gamma,omitempty"`
	Theta float64 `json:"theta,omitempty"`
	Vega  float64 `json:"vega,omitempty"`
	Rho   float64 `json:"rho,omitempty"`

	// Misc
	InTheMoney bool `json:"in_the_money,omitempty"`
}

// UnusualOption represents unusual options activity.
type UnusualOption struct {
	Symbol         string    `json:"symbol"`
	Underlying     string    `json:"underlying"`
	StrikePrice    float64   `json:"strike_price"`
	OptionType     string    `json:"option_type"` // "call" or "put"
	ExpirationDate time.Time `json:"expiration_date"`
	Volume         int64     `json:"volume"`
	OpenInterest   int64     `json:"open_interest"`
	VolumeOIRatio  float64   `json:"volume_oi_ratio,omitempty"`
	IV             float64   `json:"iv,omitempty"`
	Sentiment      string    `json:"sentiment,omitempty"` // "bullish", "bearish", "neutral"
	TradeDate      time.Time `json:"trade_date"`
}

// --- Extended Futures Models ---

// FuturesInstrument represents a tradeable futures instrument.
type FuturesInstrument struct {
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Exchange     string `json:"exchange"`
	Category     string `json:"category,omitempty"` // "Energy", "Metals", "Agricultural", "Index", etc.
	ContractSize float64 `json:"contract_size,omitempty"`
	Unit         string `json:"unit,omitempty"`
	Currency     string `json:"currency,omitempty"`
}

// FuturesCurvePoint represents a single point on a futures forward curve.
type FuturesCurvePoint struct {
	Symbol     string    `json:"symbol"`
	Expiration time.Time `json:"expiration"`
	Price      float64   `json:"price"`
	Volume     int64     `json:"volume,omitempty"`
	OI         int64     `json:"oi,omitempty"`
}

// FuturesHistoricalData represents historical futures data.
type FuturesHistoricalData struct {
	Symbol    string    `json:"symbol"`
	Date      time.Time `json:"date"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	OI        int64     `json:"oi,omitempty"`
	Expiration time.Time `json:"expiration,omitempty"`
}
