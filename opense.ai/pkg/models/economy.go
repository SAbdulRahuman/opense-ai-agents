package models

import "time"

// EconomicCalendarEvent represents an event in the economic calendar.
type EconomicCalendarEvent struct {
	Date          time.Time `json:"date"`
	Country       string    `json:"country"`
	Category      string    `json:"category,omitempty"`
	Event         string    `json:"event"`
	Importance    string    `json:"importance,omitempty"` // "low", "medium", "high"
	Actual        *float64  `json:"actual,omitempty"`
	Forecast      *float64  `json:"forecast,omitempty"`
	Previous      *float64  `json:"previous,omitempty"`
	Unit          string    `json:"unit,omitempty"`
	Source        string    `json:"source,omitempty"`
}

// EconomicIndicatorData represents a time series economic indicator.
type EconomicIndicatorData struct {
	Date    time.Time `json:"date"`
	Value   float64   `json:"value"`
	Country string    `json:"country,omitempty"`
}

// CPIData represents Consumer Price Index data.
type CPIData struct {
	Date       time.Time `json:"date"`
	Value      float64   `json:"value"`
	Country    string    `json:"country"`
	Frequency  string    `json:"frequency,omitempty"`  // "monthly", "annual"
	Harmonized bool      `json:"harmonized,omitempty"` // EU harmonized CPI
}

// GDPData represents GDP data (real or nominal).
type GDPData struct {
	Date      time.Time `json:"date"`
	Value     float64   `json:"value"`
	Country   string    `json:"country"`
	Currency  string    `json:"currency,omitempty"`
	Type      string    `json:"type,omitempty"` // "real", "nominal"
	YoYGrowth float64   `json:"yoy_growth,omitempty"`
	QoQGrowth float64   `json:"qoq_growth,omitempty"`
}

// UnemploymentData represents unemployment data.
type UnemploymentData struct {
	Date    time.Time `json:"date"`
	Value   float64   `json:"value"` // unemployment rate %
	Country string    `json:"country"`
}

// CountryProfile represents economic profile of a country.
type CountryProfile struct {
	Country          string  `json:"country"`
	ISOCode          string  `json:"iso_code,omitempty"`
	GDP              float64 `json:"gdp,omitempty"`
	GDPGrowth        float64 `json:"gdp_growth,omitempty"`
	GDPPerCapita     float64 `json:"gdp_per_capita,omitempty"`
	Inflation        float64 `json:"inflation,omitempty"`
	Unemployment     float64 `json:"unemployment,omitempty"`
	Population       int64   `json:"population,omitempty"`
	Currency         string  `json:"currency,omitempty"`
	CentralBankRate  float64 `json:"central_bank_rate,omitempty"`
	DebtToGDP        float64 `json:"debt_to_gdp,omitempty"`
	CurrentAccount   float64 `json:"current_account,omitempty"`
}

// AvailableEconomicIndicator represents an available economic indicator for querying.
type AvailableEconomicIndicator struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Country     string `json:"country,omitempty"`
	Frequency   string `json:"frequency,omitempty"`
	Source      string `json:"source,omitempty"`
	Unit        string `json:"unit,omitempty"`
}

// BalanceOfPaymentsData represents balance of payments data.
type BalanceOfPaymentsData struct {
	Date            time.Time `json:"date"`
	Country         string    `json:"country"`
	CurrentAccount  float64   `json:"current_account"`
	CapitalAccount  float64   `json:"capital_account,omitempty"`
	FinancialAccount float64  `json:"financial_account,omitempty"`
	TradeBalance    float64   `json:"trade_balance,omitempty"`
	Currency        string    `json:"currency,omitempty"`
}

// MoneyMeasureData represents money supply data (M0, M1, M2, M3).
type MoneyMeasureData struct {
	Date     time.Time `json:"date"`
	Country  string    `json:"country"`
	Measure  string    `json:"measure"` // "M0", "M1", "M2", "M3"
	Value    float64   `json:"value"`
	Currency string    `json:"currency,omitempty"`
}

// RiskPremiumData represents equity risk premium data.
type RiskPremiumData struct {
	Country             string  `json:"country"`
	RiskPremium         float64 `json:"risk_premium"`
	CountryRiskPremium  float64 `json:"country_risk_premium,omitempty"`
	CorporateTaxRate    float64 `json:"corporate_tax_rate,omitempty"`
	MoodysRating        string  `json:"moodys_rating,omitempty"`
}

// --- FRED (Federal Reserve Economic Data) ---

// FREDSearchResult represents a FRED series search result.
type FREDSearchResult struct {
	SeriesID           string    `json:"series_id"`
	Title              string    `json:"title"`
	ObservationStart   time.Time `json:"observation_start,omitempty"`
	ObservationEnd     time.Time `json:"observation_end,omitempty"`
	Frequency          string    `json:"frequency,omitempty"`
	Units              string    `json:"units,omitempty"`
	SeasonalAdjustment string    `json:"seasonal_adjustment,omitempty"`
	Popularity         int       `json:"popularity,omitempty"`
}

// FREDSeriesData represents a FRED time series data point.
type FREDSeriesData struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// --- Surveys ---

// ConsumerSentimentData represents consumer sentiment (e.g., UMich) data.
type ConsumerSentimentData struct {
	Date            time.Time `json:"date"`
	SentimentIndex  float64   `json:"sentiment_index"`
	CurrentConditions float64 `json:"current_conditions,omitempty"`
	Expectations    float64   `json:"expectations,omitempty"`
	InflationExpectation1Y float64 `json:"inflation_expectation_1y,omitempty"`
	InflationExpectation5Y float64 `json:"inflation_expectation_5y,omitempty"`
}

// ManufacturingOutlookData represents manufacturing outlook survey data.
type ManufacturingOutlookData struct {
	Date               time.Time `json:"date"`
	Region             string    `json:"region"` // "NY", "Texas", "Chicago"
	GeneralBusinessIndex float64 `json:"general_business_index"`
	NewOrders            float64 `json:"new_orders,omitempty"`
	Shipments            float64 `json:"shipments,omitempty"`
	Employment           float64 `json:"employment,omitempty"`
	PricesPaid           float64 `json:"prices_paid,omitempty"`
}

// NonFarmPayrollData represents non-farm payroll data.
type NonFarmPayrollData struct {
	Date       time.Time `json:"date"`
	Value      int64     `json:"value"`   // jobs added
	Revised    *int64    `json:"revised,omitempty"` // revised figure
	Forecast   *int64    `json:"forecast,omitempty"`
	Previous   *int64    `json:"previous,omitempty"`
	Unemployment float64 `json:"unemployment,omitempty"`
}

// --- FOMC ---

// FOMCDocument represents a FOMC document/minutes/statement.
type FOMCDocument struct {
	Date     time.Time `json:"date"`
	Type     string    `json:"type"` // "minutes", "statement", "press_conference"
	Title    string    `json:"title"`
	URL      string    `json:"url,omitempty"`
	Content  string    `json:"content,omitempty"`
}

// InflationExpectationData represents inflation expectations.
type InflationExpectationData struct {
	Date     time.Time `json:"date"`
	Horizon  string    `json:"horizon"` // "1Y", "5Y", "10Y"
	Value    float64   `json:"value"`
	Source   string    `json:"source,omitempty"`
}

// --- Shipping / Port ---

// PortData represents port volume data.
type PortData struct {
	Port     string    `json:"port"`
	Country  string    `json:"country"`
	Date     time.Time `json:"date"`
	Volume   float64   `json:"volume"`
	Unit     string    `json:"unit,omitempty"` // "TEU", "tons"
	Category string    `json:"category,omitempty"`
}
