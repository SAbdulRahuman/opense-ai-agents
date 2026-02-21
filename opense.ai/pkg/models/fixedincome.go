package models

import "time"

// --- Fixed Income / Rates ---

// InterestRateData represents an interest rate data point (SOFR, SONIA, Fed Funds, etc.).
type InterestRateData struct {
	Date     time.Time `json:"date"`
	Rate     float64   `json:"rate"`
	RateType string    `json:"rate_type"` // "SOFR", "SONIA", "Ameribor", "FedFunds", etc.
	Maturity string    `json:"maturity,omitempty"` // e.g., "overnight", "30d", "90d"
}

// RateProjection represents a central bank rate projection.
type RateProjection struct {
	Date       time.Time `json:"date"`
	MeetingDate time.Time `json:"meeting_date,omitempty"`
	RateLow    float64   `json:"rate_low"`
	RateHigh   float64   `json:"rate_high"`
	RateMedian float64   `json:"rate_median,omitempty"`
	Source     string    `json:"source,omitempty"` // "FOMC", "ECB", "RBI"
}

// ECBRate represents European Central Bank interest rate data.
type ECBRate struct {
	Date          time.Time `json:"date"`
	MainRate      float64   `json:"main_rate"`       // Main Refinancing Operations
	DepositRate   float64   `json:"deposit_rate"`    // Deposit Facility Rate
	LendingRate   float64   `json:"lending_rate"`    // Marginal Lending Facility
}

// --- Fixed Income / Government ---

// YieldCurvePoint represents a single point on a yield curve.
type YieldCurvePoint struct {
	Date     time.Time `json:"date"`
	Maturity string    `json:"maturity"` // "1M", "3M", "6M", "1Y", "2Y", "5Y", "10Y", "30Y"
	Rate     float64   `json:"rate"`
}

// TreasuryRate represents a treasury rate data point.
type TreasuryRate struct {
	Date     time.Time          `json:"date"`
	Rates    map[string]float64 `json:"rates"` // maturity → rate (e.g., "10Y" → 4.25)
}

// TreasuryAuction represents a treasury auction result.
type TreasuryAuction struct {
	Date            time.Time `json:"date"`
	SecurityType    string    `json:"security_type"` // "Bill", "Note", "Bond", "TIPS"
	Maturity        string    `json:"maturity"`
	CUSIP           string    `json:"cusip,omitempty"`
	HighYield       float64   `json:"high_yield,omitempty"`
	AllotmentRatio  float64   `json:"allotment_ratio,omitempty"`
	BidToCover      float64   `json:"bid_to_cover,omitempty"`
	TotalAccepted   float64   `json:"total_accepted,omitempty"`
	TotalTendered   float64   `json:"total_tendered,omitempty"`
	Currency        string    `json:"currency,omitempty"`
}

// TreasuryPrice represents a treasury security price.
type TreasuryPrice struct {
	Date     time.Time `json:"date"`
	CUSIP    string    `json:"cusip"`
	Type     string    `json:"type"` // "Bill", "Note", "Bond", "TIPS"
	Maturity string    `json:"maturity"`
	Price    float64   `json:"price"`
	Yield    float64   `json:"yield"`
	Coupon   float64   `json:"coupon,omitempty"`
}

// --- Fixed Income / Corporate ---

// BondPrice represents a corporate bond price.
type BondPrice struct {
	Date      time.Time `json:"date"`
	ISIN      string    `json:"isin,omitempty"`
	CUSIP     string    `json:"cusip,omitempty"`
	Issuer    string    `json:"issuer"`
	CouponRate float64  `json:"coupon_rate,omitempty"`
	Maturity  time.Time `json:"maturity"`
	Price     float64   `json:"price"`
	Yield     float64   `json:"yield"`
	Rating    string    `json:"rating,omitempty"` // credit rating
	Sector    string    `json:"sector,omitempty"`
}

// BondIndex represents a bond index data point.
type BondIndex struct {
	Date      time.Time `json:"date"`
	IndexName string    `json:"index_name"`
	Value     float64   `json:"value"`
	Change    float64   `json:"change,omitempty"`
	Yield     float64   `json:"yield,omitempty"`
	Duration  float64   `json:"duration,omitempty"`
}

// MortgageIndex represents a mortgage rate data point.
type MortgageIndex struct {
	Date           time.Time `json:"date"`
	Rate30YrFixed  float64   `json:"rate_30yr_fixed,omitempty"`
	Rate15YrFixed  float64   `json:"rate_15yr_fixed,omitempty"`
	Rate5YrARM     float64   `json:"rate_5yr_arm,omitempty"`
	Points30Yr     float64   `json:"points_30yr,omitempty"`
	Source         string    `json:"source,omitempty"`
}

// SpotRateData represents a spot rate from the yield curve.
type SpotRateData struct {
	Date     time.Time `json:"date"`
	Maturity string    `json:"maturity"`
	Rate     float64   `json:"rate"`
}

// CommercialPaperRate represents commercial paper rate data.
type CommercialPaperRate struct {
	Date     time.Time `json:"date"`
	Maturity string    `json:"maturity"`
	Rate     float64   `json:"rate"`
	Type     string    `json:"type,omitempty"` // "financial", "nonfinancial", "AA"
}
