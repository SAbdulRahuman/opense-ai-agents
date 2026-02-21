package models

import "time"

// ETFInfo represents detailed ETF information.
type ETFInfo struct {
	Symbol          string    `json:"symbol"`
	Name            string    `json:"name"`
	Exchange        string    `json:"exchange"`
	Issuer          string    `json:"issuer,omitempty"`
	FundFamily      string    `json:"fund_family,omitempty"`
	Category        string    `json:"category,omitempty"`
	InceptionDate   time.Time `json:"inception_date,omitempty"`
	ExpenseRatio    float64   `json:"expense_ratio,omitempty"`
	AUM             float64   `json:"aum,omitempty"` // Assets Under Management
	AvgVolume       int64     `json:"avg_volume,omitempty"`
	NAV             float64   `json:"nav,omitempty"`
	PE              float64   `json:"pe,omitempty"`
	DividendYield   float64   `json:"dividend_yield,omitempty"`
	YTDReturn       float64   `json:"ytd_return,omitempty"`
	ThreeYearReturn float64   `json:"three_year_return,omitempty"`
	FiveYearReturn  float64   `json:"five_year_return,omitempty"`
	Beta            float64   `json:"beta,omitempty"`
	HoldingsCount   int       `json:"holdings_count,omitempty"`
	Description     string    `json:"description,omitempty"`
}

// ETFSearchResult represents an ETF search result.
type ETFSearchResult struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Exchange     string  `json:"exchange"`
	Category     string  `json:"category,omitempty"`
	ExpenseRatio float64 `json:"expense_ratio,omitempty"`
	AUM          float64 `json:"aum,omitempty"`
}

// ETFHolding represents a single holding in an ETF.
type ETFHolding struct {
	Symbol        string  `json:"symbol,omitempty"`
	Name          string  `json:"name"`
	Weight        float64 `json:"weight"`         // percentage of portfolio
	Shares        int64   `json:"shares,omitempty"`
	MarketValue   float64 `json:"market_value,omitempty"`
	Sector        string  `json:"sector,omitempty"`
	Country       string  `json:"country,omitempty"`
	AssetClass    string  `json:"asset_class,omitempty"` // "Equity", "Bond", "Cash", etc.
}

// ETFSectorExposure represents sector-level exposure of an ETF.
type ETFSectorExposure struct {
	Sector   string  `json:"sector"`
	Weight   float64 `json:"weight"`
}

// ETFCountryExposure represents geographic exposure of an ETF.
type ETFCountryExposure struct {
	Country string  `json:"country"`
	Weight  float64 `json:"weight"`
}

// ETFPricePerformance represents ETF price performance over various periods.
type ETFPricePerformance struct {
	Symbol     string  `json:"symbol"`
	OneDay     float64 `json:"one_day"`
	OneWeek    float64 `json:"one_week"`
	OneMonth   float64 `json:"one_month"`
	ThreeMonth float64 `json:"three_month"`
	SixMonth   float64 `json:"six_month"`
	YTD        float64 `json:"ytd"`
	OneYear    float64 `json:"one_year"`
	ThreeYear  float64 `json:"three_year,omitempty"`
	FiveYear   float64 `json:"five_year,omitempty"`
}
