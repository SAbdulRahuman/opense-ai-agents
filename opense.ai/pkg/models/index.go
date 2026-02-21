package models

import "time"

// IndexInfo represents detailed information about a market index.
type IndexInfo struct {
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Exchange    string    `json:"exchange"`
	Country     string    `json:"country,omitempty"`
	Currency    string    `json:"currency,omitempty"`
	Description string    `json:"description,omitempty"`
	Methodology string    `json:"methodology,omitempty"`
	LaunchDate  time.Time `json:"launch_date,omitempty"`
}

// IndexConstituent represents a single constituent of an index.
type IndexConstituent struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Sector    string  `json:"sector,omitempty"`
	Industry  string  `json:"industry,omitempty"`
	Weight    float64 `json:"weight,omitempty"` // percentage weight in index
	MarketCap float64 `json:"market_cap,omitempty"`
}

// IndexSnapshot represents current snapshot of an index.
type IndexSnapshot struct {
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Change    float64   `json:"change"`
	ChangePct float64   `json:"change_pct"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	PrevClose float64   `json:"prev_close"`
	Volume    int64     `json:"volume,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// IndexSectorPerformance represents sector-level performance within an index.
type IndexSectorPerformance struct {
	Sector    string  `json:"sector"`
	ChangePct float64 `json:"change_pct"`
	Weight    float64 `json:"weight,omitempty"`
}

// SP500Multiple represents S&P 500 valuation multiples over time.
type SP500Multiple struct {
	Date        time.Time `json:"date"`
	PE          float64   `json:"pe,omitempty"`
	ShillerPE   float64   `json:"shiller_pe,omitempty"`
	PB          float64   `json:"pb,omitempty"`
	DividendYield float64 `json:"dividend_yield,omitempty"`
	EarningsYield float64 `json:"earnings_yield,omitempty"`
}
