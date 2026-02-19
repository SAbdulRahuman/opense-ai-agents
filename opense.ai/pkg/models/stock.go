// Package models defines the core data structures used throughout OpeNSE.ai.
package models

import "time"

// Stock represents basic stock information.
type Stock struct {
	Ticker       string  `json:"ticker"`        // e.g., "RELIANCE"
	NSETicker    string  `json:"nse_ticker"`     // e.g., "RELIANCE.NS"
	Name         string  `json:"name"`           // e.g., "Reliance Industries Limited"
	Exchange     string  `json:"exchange"`       // "NSE" or "BSE"
	Sector       string  `json:"sector"`         // e.g., "Oil & Gas"
	Industry     string  `json:"industry"`       // e.g., "Refineries"
	ISIN         string  `json:"isin"`           // e.g., "INE002A01018"
	MarketCap    float64 `json:"market_cap"`     // in INR (raw value, not formatted)
	FaceValue    float64 `json:"face_value"`     // e.g., 10.0
	ListingDate  string  `json:"listing_date"`   // e.g., "1995-11-29"
	IsIndex      bool    `json:"is_index"`       // true for NIFTY50, BANKNIFTY etc.
	LotSize      int     `json:"lot_size"`       // F&O lot size
	TickSize     float64 `json:"tick_size"`      // minimum price movement
}

// OHLCV represents a single candlestick bar of price data.
type OHLCV struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	AdjClose  float64   `json:"adj_close,omitempty"`
}

// Quote represents a real-time stock quote.
type Quote struct {
	Ticker         string    `json:"ticker"`
	Name           string    `json:"name"`
	LastPrice      float64   `json:"last_price"`
	Change         float64   `json:"change"`
	ChangePct      float64   `json:"change_pct"`
	Open           float64   `json:"open"`
	High           float64   `json:"high"`
	Low            float64   `json:"low"`
	PrevClose      float64   `json:"prev_close"`
	Volume         int64     `json:"volume"`
	Value          float64   `json:"value"` // traded value in INR
	UpperCircuit   float64   `json:"upper_circuit"`
	LowerCircuit   float64   `json:"lower_circuit"`
	WeekHigh52     float64   `json:"week_high_52"`
	WeekLow52      float64   `json:"week_low_52"`
	MarketCap      float64   `json:"market_cap"`
	PE             float64   `json:"pe,omitempty"`
	PB             float64   `json:"pb,omitempty"`
	DividendYield  float64   `json:"dividend_yield,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// Timeframe represents chart timeframe for OHLCV data.
type Timeframe string

const (
	Timeframe1Min  Timeframe = "1m"
	Timeframe5Min  Timeframe = "5m"
	Timeframe15Min Timeframe = "15m"
	Timeframe1Hour Timeframe = "1h"
	Timeframe1Day  Timeframe = "1d"
	Timeframe1Week Timeframe = "1w"
	Timeframe1Mon  Timeframe = "1M"
)

// StockProfile aggregates data from multiple sources for a single stock.
type StockProfile struct {
	Stock       Stock           `json:"stock"`
	Quote       *Quote          `json:"quote,omitempty"`
	Historical  []OHLCV         `json:"historical,omitempty"`
	Financials  *FinancialData  `json:"financials,omitempty"`
	Ratios      *FinancialRatios `json:"ratios,omitempty"`
	Promoter    *PromoterData   `json:"promoter,omitempty"`
	FetchedAt   time.Time       `json:"fetched_at"`
}

// PromoterData represents promoter holding information.
type PromoterData struct {
	PromoterHolding    float64   `json:"promoter_holding"`     // percentage
	PromoterPledge     float64   `json:"promoter_pledge"`      // percentage of promoter holding pledged
	FIIHolding         float64   `json:"fii_holding"`
	DIIHolding         float64   `json:"dii_holding"`
	PublicHolding      float64   `json:"public_holding"`
	MFHolding          float64   `json:"mf_holding"`
	Quarter            string    `json:"quarter"`              // e.g., "Dec 2025"
	PromoterTrend      []HoldingPoint `json:"promoter_trend,omitempty"`
}

// HoldingPoint represents a single data point in holding trend.
type HoldingPoint struct {
	Quarter string  `json:"quarter"`
	Pct     float64 `json:"pct"`
}

// FinancialRatios contains key financial ratios.
type FinancialRatios struct {
	PE               float64 `json:"pe"`
	PB               float64 `json:"pb"`
	EVBITDA          float64 `json:"ev_ebitda"`
	ROE              float64 `json:"roe"`
	ROCE             float64 `json:"roce"`
	DebtEquity       float64 `json:"debt_equity"`
	CurrentRatio     float64 `json:"current_ratio"`
	InterestCoverage float64 `json:"interest_coverage"`
	DividendYield    float64 `json:"dividend_yield"`
	EPS              float64 `json:"eps"`
	BookValue        float64 `json:"book_value"`
	PEGRatio         float64 `json:"peg_ratio"`
	GrahamNumber     float64 `json:"graham_number"`
}
