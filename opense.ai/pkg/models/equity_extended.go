package models

import "time"

// --- Equity / Search & Screening ---

// EquitySearchResult represents a single result from equity search.
type EquitySearchResult struct {
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Exchange  string `json:"exchange"`
	Sector    string `json:"sector,omitempty"`
	Industry  string `json:"industry,omitempty"`
	MarketCap float64 `json:"market_cap,omitempty"`
	Country   string `json:"country,omitempty"`
	IsETF     bool   `json:"is_etf,omitempty"`
}

// ScreenerCriteria defines filter criteria for stock screening.
type ScreenerCriteria struct {
	MarketCapMin    *float64 `json:"market_cap_min,omitempty"`
	MarketCapMax    *float64 `json:"market_cap_max,omitempty"`
	PEMin           *float64 `json:"pe_min,omitempty"`
	PEMax           *float64 `json:"pe_max,omitempty"`
	DividendYieldMin *float64 `json:"dividend_yield_min,omitempty"`
	Sector          string   `json:"sector,omitempty"`
	Industry        string   `json:"industry,omitempty"`
	Country         string   `json:"country,omitempty"`
	Exchange        string   `json:"exchange,omitempty"`
	BetaMin         *float64 `json:"beta_min,omitempty"`
	BetaMax         *float64 `json:"beta_max,omitempty"`
	VolumeMin       *int64   `json:"volume_min,omitempty"`
	PriceMin        *float64 `json:"price_min,omitempty"`
	PriceMax        *float64 `json:"price_max,omitempty"`
	Limit           int      `json:"limit,omitempty"`
}

// ScreenerResult represents a stock that matched screening criteria.
type ScreenerResult struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Exchange      string  `json:"exchange"`
	Sector        string  `json:"sector"`
	Industry      string  `json:"industry"`
	MarketCap     float64 `json:"market_cap"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePct     float64 `json:"change_pct"`
	Volume        int64   `json:"volume"`
	PE            float64 `json:"pe"`
	PB            float64 `json:"pb"`
	DividendYield float64 `json:"dividend_yield"`
	Beta          float64 `json:"beta"`
	EPS           float64 `json:"eps"`
	ROE           float64 `json:"roe"`
	DebtEquity    float64 `json:"debt_equity"`
}

// --- Equity / Peers ---

// EquityPeer represents a peer company.
type EquityPeer struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Sector    string  `json:"sector"`
	MarketCap float64 `json:"market_cap"`
	PE        float64 `json:"pe"`
	PB        float64 `json:"pb"`
	ROE       float64 `json:"roe"`
}

// --- Equity / Estimates ---

// PriceTargetData represents analyst price targets.
type PriceTargetData struct {
	Symbol          string    `json:"symbol"`
	PublishedDate   time.Time `json:"published_date"`
	AnalystName     string    `json:"analyst_name"`
	AnalystCompany  string    `json:"analyst_company"`
	Rating          string    `json:"rating"`          // "Buy", "Hold", "Sell" etc.
	PriceTarget     float64   `json:"price_target"`
	PriceTargetPrev float64   `json:"price_target_prev,omitempty"`
	AdjPriceTarget  float64   `json:"adj_price_target,omitempty"`
}

// PriceTargetConsensusData represents consensus price target data.
type PriceTargetConsensusData struct {
	Symbol    string  `json:"symbol"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Median    float64 `json:"median"`
	Average   float64 `json:"average"`
	Current   float64 `json:"current"`
	NumBuy    int     `json:"num_buy"`
	NumHold   int     `json:"num_hold"`
	NumSell   int     `json:"num_sell"`
	NumTotal  int     `json:"num_total"`
}

// AnalystEstimate represents analyst financial estimates.
type AnalystEstimate struct {
	Symbol                string  `json:"symbol"`
	Date                  string  `json:"date"`
	EstimatedRevenueAvg   float64 `json:"estimated_revenue_avg"`
	EstimatedRevenueHigh  float64 `json:"estimated_revenue_high"`
	EstimatedRevenueLow   float64 `json:"estimated_revenue_low"`
	EstimatedEBITDAAvg    float64 `json:"estimated_ebitda_avg"`
	EstimatedEBITDAHigh   float64 `json:"estimated_ebitda_high"`
	EstimatedEBITDALow    float64 `json:"estimated_ebitda_low"`
	EstimatedEPSAvg       float64 `json:"estimated_eps_avg"`
	EstimatedEPSHigh      float64 `json:"estimated_eps_high"`
	EstimatedEPSLow       float64 `json:"estimated_eps_low"`
	EstimatedNetIncomeAvg float64 `json:"estimated_net_income_avg"`
	NumberOfAnalysts      int     `json:"number_of_analysts"`
}

// ForwardEstimate represents a forward-looking estimate (EPS, EBITDA, PE, Sales).
type ForwardEstimate struct {
	Symbol           string  `json:"symbol"`
	FiscalYear       string  `json:"fiscal_year,omitempty"`
	FiscalQuarter    string  `json:"fiscal_quarter,omitempty"`
	Date             string  `json:"date"`
	EstimateLow      float64 `json:"estimate_low"`
	EstimateHigh     float64 `json:"estimate_high"`
	EstimateAvg      float64 `json:"estimate_avg"`
	EstimateMedian   float64 `json:"estimate_median,omitempty"`
	NumberOfAnalysts int     `json:"number_of_analysts"`
}

// --- Equity / Calendar ---

// EarningsCalendarEntry represents an entry in the earnings calendar.
type EarningsCalendarEntry struct {
	Symbol       string    `json:"symbol"`
	Name         string    `json:"name"`
	ReportDate   time.Time `json:"report_date"`
	FiscalQuarter string   `json:"fiscal_quarter,omitempty"`
	EPSEstimate  float64   `json:"eps_estimate,omitempty"`
	EPSActual    float64   `json:"eps_actual,omitempty"`
	RevenueEstimate float64 `json:"revenue_estimate,omitempty"`
	RevenueActual   float64 `json:"revenue_actual,omitempty"`
	Surprise     float64   `json:"surprise,omitempty"`
	SurprisePct  float64   `json:"surprise_pct,omitempty"`
	Timing       string    `json:"timing,omitempty"` // "BMO" (before market open), "AMC" (after market close)
}

// DividendCalendarEntry represents an upcoming dividend event.
type DividendCalendarEntry struct {
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name"`
	ExDividendDate time.Time `json:"ex_dividend_date"`
	PaymentDate   time.Time `json:"payment_date,omitempty"`
	RecordDate    time.Time `json:"record_date,omitempty"`
	Amount        float64   `json:"amount"`
	Yield         float64   `json:"yield,omitempty"`
	Frequency     string    `json:"frequency,omitempty"` // "quarterly", "annual", etc.
}

// IPOCalendarEntry represents an upcoming IPO.
type IPOCalendarEntry struct {
	Symbol        string    `json:"symbol,omitempty"`
	Name          string    `json:"name"`
	FilingDate    time.Time `json:"filing_date,omitempty"`
	IPODate       time.Time `json:"ipo_date,omitempty"`
	PriceRangeLow  float64  `json:"price_range_low,omitempty"`
	PriceRangeHigh float64  `json:"price_range_high,omitempty"`
	OfferPrice    float64   `json:"offer_price,omitempty"`
	Shares        int64     `json:"shares,omitempty"`
	Exchange      string    `json:"exchange,omitempty"`
	Status        string    `json:"status,omitempty"` // "upcoming", "priced", "withdrawn"
}

// SplitCalendarEntry represents an upcoming stock split.
type SplitCalendarEntry struct {
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name,omitempty"`
	Date      time.Time `json:"date"`
	Ratio     string    `json:"ratio"` // e.g., "2:1", "5:1"
	Numerator   float64 `json:"numerator"`
	Denominator float64 `json:"denominator"`
}

// --- Equity / Discovery ---

// MarketMover represents a stock in gainers/losers/active lists.
type MarketMover struct {
	Symbol     string  `json:"symbol"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Change     float64 `json:"change"`
	ChangePct  float64 `json:"change_pct"`
	Volume     int64   `json:"volume"`
	MarketCap  float64 `json:"market_cap,omitempty"`
}

// --- Equity / Ownership ---

// OwnershipData represents institutional/insider ownership data.
type OwnershipData struct {
	Symbol          string    `json:"symbol"`
	InvestorName    string    `json:"investor_name"`
	SecurityName    string    `json:"security_name,omitempty"`
	TypeOfSecurity  string    `json:"type_of_security,omitempty"`
	SharesHeld      int64     `json:"shares_held"`
	SharesChanged   int64     `json:"shares_changed,omitempty"`
	ChangePercent   float64   `json:"change_percent,omitempty"`
	PortfolioPercent float64  `json:"portfolio_percent,omitempty"`
	MarketValue     float64   `json:"market_value,omitempty"`
	ReportDate      time.Time `json:"report_date"`
	FilingDate      time.Time `json:"filing_date,omitempty"`
}

// InsiderTrade represents an insider trading transaction.
type InsiderTrade struct {
	Symbol            string    `json:"symbol"`
	FilingDate        time.Time `json:"filing_date"`
	TransactionDate   time.Time `json:"transaction_date"`
	OwnerName         string    `json:"owner_name"`
	OwnerTitle        string    `json:"owner_title,omitempty"`
	TransactionType   string    `json:"transaction_type"` // "Purchase", "Sale", etc.
	SharesTraded      int64     `json:"shares_traded"`
	PricePerShare     float64   `json:"price_per_share"`
	TotalValue        float64   `json:"total_value"`
	SharesOwned       int64     `json:"shares_owned,omitempty"`
}

// GovernmentTradeEntry represents a government official's trade.
type GovernmentTradeEntry struct {
	ReportDate     time.Time `json:"report_date"`
	TransactionDate time.Time `json:"transaction_date"`
	Representative string    `json:"representative"`
	Chamber        string    `json:"chamber,omitempty"` // "Senate" or "House"
	Party          string    `json:"party,omitempty"`
	Symbol         string    `json:"symbol"`
	AssetType      string    `json:"asset_type,omitempty"`
	TransactionType string   `json:"transaction_type"`
	AmountLow      float64   `json:"amount_low,omitempty"`
	AmountHigh     float64   `json:"amount_high,omitempty"`
}

// --- Equity / Shorts ---

// FailToDeliver represents fail-to-deliver data.
type FailToDeliver struct {
	Symbol       string    `json:"symbol"`
	Date         time.Time `json:"date"`
	Quantity     int64     `json:"quantity"`
	Price        float64   `json:"price"`
	Value        float64   `json:"value"`
}

// ShortVolumeData represents short volume trading data.
type ShortVolumeData struct {
	Symbol        string    `json:"symbol"`
	Date          time.Time `json:"date"`
	ShortVolume   int64     `json:"short_volume"`
	TotalVolume   int64     `json:"total_volume"`
	ShortPercent  float64   `json:"short_percent"`
}

// ShortInterestData represents short interest data.
type ShortInterestData struct {
	Symbol           string    `json:"symbol"`
	SettlementDate   time.Time `json:"settlement_date"`
	ShortInterest    int64     `json:"short_interest"`
	AvgDailyVolume   int64     `json:"avg_daily_volume"`
	DaysToCover      float64   `json:"days_to_cover"`
	ShortPercentFloat float64  `json:"short_percent_float,omitempty"`
}

// --- Equity / Fundamentals Extended ---

// KeyExecutive represents a company executive.
type KeyExecutive struct {
	Name         string  `json:"name"`
	Title        string  `json:"title"`
	Pay          float64 `json:"pay,omitempty"`
	CurrencyPay string  `json:"currency_pay,omitempty"`
	Gender       string  `json:"gender,omitempty"`
	YearBorn     int     `json:"year_born,omitempty"`
}

// ExecutiveCompensationData represents executive compensation details.
type ExecutiveCompensationData struct {
	Symbol         string  `json:"symbol"`
	CIK            string  `json:"cik,omitempty"`
	Year           int     `json:"year"`
	Name           string  `json:"name"`
	Title          string  `json:"title"`
	Salary         float64 `json:"salary"`
	Bonus          float64 `json:"bonus"`
	StockAward     float64 `json:"stock_award"`
	OptionAward    float64 `json:"option_award"`
	OtherCompensation float64 `json:"other_compensation"`
	Total          float64 `json:"total"`
}

// DividendRecord represents a historical dividend.
type DividendRecord struct {
	Symbol        string    `json:"symbol"`
	ExDate        time.Time `json:"ex_date"`
	PaymentDate   time.Time `json:"payment_date,omitempty"`
	DeclarationDate time.Time `json:"declaration_date,omitempty"`
	RecordDate    time.Time `json:"record_date,omitempty"`
	Amount        float64   `json:"amount"`
	Frequency     string    `json:"frequency,omitempty"`
	AdjDividend   float64   `json:"adj_dividend,omitempty"`
}

// StockSplit represents a historical stock split.
type StockSplit struct {
	Symbol      string    `json:"symbol"`
	Date        time.Time `json:"date"`
	Numerator   float64   `json:"numerator"`
	Denominator float64   `json:"denominator"`
}

// HistoricalEPSData represents historical EPS data.
type HistoricalEPSData struct {
	Symbol         string    `json:"symbol"`
	Date           time.Time `json:"date"`
	FiscalQuarter  string    `json:"fiscal_quarter,omitempty"`
	EPSActual      float64   `json:"eps_actual"`
	EPSEstimate    float64   `json:"eps_estimate,omitempty"`
	Surprise       float64   `json:"surprise,omitempty"`
	SurprisePct    float64   `json:"surprise_pct,omitempty"`
	RevenueActual  float64   `json:"revenue_actual,omitempty"`
	RevenueEstimate float64  `json:"revenue_estimate,omitempty"`
}

// EarningsTranscript represents an earnings call transcript.
type EarningsTranscript struct {
	Symbol   string `json:"symbol"`
	Quarter  string `json:"quarter"`
	Year     int    `json:"year"`
	Date     string `json:"date"`
	Content  string `json:"content"`
}

// ESGScore represents ESG (Environmental, Social, Governance) scores.
type ESGScore struct {
	Symbol              string  `json:"symbol"`
	CompanyName         string  `json:"company_name"`
	EnvironmentScore    float64 `json:"environment_score"`
	SocialScore         float64 `json:"social_score"`
	GovernanceScore     float64 `json:"governance_score"`
	TotalScore          float64 `json:"total_score"`
	ESGRiskRating       string  `json:"esg_risk_rating,omitempty"`
	IndustryRank        int     `json:"industry_rank,omitempty"`
}

// ShareStatisticsData represents share statistics.
type ShareStatisticsData struct {
	Symbol              string  `json:"symbol"`
	SharesOutstanding   int64   `json:"shares_outstanding"`
	FloatShares         int64   `json:"float_shares,omitempty"`
	SharesShort         int64   `json:"shares_short,omitempty"`
	ShortRatio          float64 `json:"short_ratio,omitempty"`
	ShortPercentFloat   float64 `json:"short_percent_float,omitempty"`
	AvgVolume10Day      int64   `json:"avg_volume_10day,omitempty"`
	AvgVolume30Day      int64   `json:"avg_volume_30day,omitempty"`
	InsiderOwnership    float64 `json:"insider_ownership,omitempty"`
	InstitutionOwnership float64 `json:"institution_ownership,omitempty"`
}

// RevenueBySegment represents revenue broken down by geographic or business segment.
type RevenueBySegment struct {
	Symbol     string  `json:"symbol"`
	Period     string  `json:"period"`
	Segment    string  `json:"segment"`
	Revenue    float64 `json:"revenue"`
	Percentage float64 `json:"percentage,omitempty"`
}

// HistoricalMarketCapData represents historical market cap data.
type HistoricalMarketCapData struct {
	Symbol    string    `json:"symbol"`
	Date      time.Time `json:"date"`
	MarketCap float64   `json:"market_cap"`
}

// PricePerformanceData represents price performance over various periods.
type PricePerformanceData struct {
	Symbol        string  `json:"symbol"`
	OneDay        float64 `json:"one_day"`
	OneWeek       float64 `json:"one_week"`
	OneMonth      float64 `json:"one_month"`
	ThreeMonth    float64 `json:"three_month"`
	SixMonth      float64 `json:"six_month"`
	YTD           float64 `json:"ytd"`
	OneYear       float64 `json:"one_year"`
	ThreeYear     float64 `json:"three_year,omitempty"`
	FiveYear      float64 `json:"five_year,omitempty"`
	TenYear       float64 `json:"ten_year,omitempty"`
	MaxReturn     float64 `json:"max_return,omitempty"`
}
