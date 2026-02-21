package yfinance

// --- Yahoo Finance API response types ---

// yfQuoteResponse wraps the v7 quote API response.
type yfQuoteResponse struct {
	QuoteResponse struct {
		Result []yfQuoteResult `json:"result"`
		Error  *yfError        `json:"error"`
	} `json:"quoteResponse"`
}

type yfQuoteResult struct {
	Symbol                     string  `json:"symbol"`
	ShortName                  string  `json:"shortName"`
	LongName                   string  `json:"longName"`
	QuoteType                  string  `json:"quoteType"`
	Exchange                   string  `json:"exchange"`
	FullExchangeName           string  `json:"fullExchangeName"`
	Market                     string  `json:"market"`
	Currency                   string  `json:"currency"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	RegularMarketChange        float64 `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketOpen          float64 `json:"regularMarketOpen"`
	RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
	RegularMarketVolume        int64   `json:"regularMarketVolume"`
	MarketCap                  float64 `json:"marketCap"`
	FiftyTwoWeekHigh           float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow            float64 `json:"fiftyTwoWeekLow"`
	TrailingPE                 float64 `json:"trailingPE"`
	ForwardPE                  float64 `json:"forwardPE"`
	PriceToBook                float64 `json:"priceToBook"`
	DividendYield              float64 `json:"dividendYield"`
	DividendRate               float64 `json:"dividendRate"`
	TrailingAnnualDividendRate float64 `json:"trailingAnnualDividendRate"`
	TrailingAnnualDividendYield float64 `json:"trailingAnnualDividendYield"`
	EpsTrailingTwelveMonths    float64 `json:"epsTrailingTwelveMonths"`
	EpsForward                 float64 `json:"epsForward"`
	BookValue                  float64 `json:"bookValue"`
	SharesOutstanding          float64 `json:"sharesOutstanding"`
	FiftyDayAverage            float64 `json:"fiftyDayAverage"`
	TwoHundredDayAverage       float64 `json:"twoHundredDayAverage"`
	AverageDailyVolume3Month   int64   `json:"averageDailyVolume3Month"`
	AverageDailyVolume10Day    int64   `json:"averageDailyVolume10Day"`
	Beta                       float64 `json:"beta"`
	RegularMarketTime          int64   `json:"regularMarketTime"`
	// ETF-specific fields
	TotalAssets       float64 `json:"totalAssets"`
	YtdReturn         float64 `json:"ytdReturn"`
	TrailingThreeMonthReturns float64 `json:"trailingThreeMonthReturns"`
}

// yfChartResponse wraps the v8 chart API response.
type yfChartResponse struct {
	Chart struct {
		Result []yfChartResult `json:"result"`
		Error  *yfError        `json:"error"`
	} `json:"chart"`
}

type yfChartResult struct {
	Meta       yfChartMeta  `json:"meta"`
	Timestamp  []int64      `json:"timestamp"`
	Indicators yfIndicators `json:"indicators"`
}

type yfChartMeta struct {
	Symbol             string  `json:"symbol"`
	Currency           string  `json:"currency"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
	InstrumentType     string  `json:"instrumentType"`
	ExchangeName       string  `json:"exchangeName"`
}

type yfIndicators struct {
	Quote    []yfOHLCV    `json:"quote"`
	AdjClose []yfAdjClose `json:"adjclose"`
}

type yfOHLCV struct {
	Open   []*float64 `json:"open"`
	High   []*float64 `json:"high"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	Volume []*int64   `json:"volume"`
}

type yfAdjClose struct {
	AdjClose []*float64 `json:"adjclose"`
}

// yfQuoteSummaryResponse wraps the v10 quoteSummary API response.
type yfQuoteSummaryResponse struct {
	QuoteSummary struct {
		Result []yfQuoteSummaryResult `json:"result"`
		Error  *yfError               `json:"error"`
	} `json:"quoteSummary"`
}

type yfQuoteSummaryResult struct {
	// Financials modules
	IncomeStatementHistory            *yfStatementContainer `json:"incomeStatementHistory"`
	IncomeStatementHistoryQuarterly   *yfStatementContainer `json:"incomeStatementHistoryQuarterly"`
	BalanceSheetHistory               *yfStatementContainer `json:"balanceSheetHistory"`
	BalanceSheetHistoryQuarterly      *yfStatementContainer `json:"balanceSheetHistoryQuarterly"`
	CashflowStatementHistory          *yfStatementContainer `json:"cashflowStatementHistory"`
	CashflowStatementHistoryQuarterly *yfStatementContainer `json:"cashflowStatementHistoryQuarterly"`

	// Profile module
	AssetProfile *yfAssetProfile `json:"assetProfile"`

	// Key stats module
	DefaultKeyStatistics *yfDefaultKeyStatistics `json:"defaultKeyStatistics"`

	// Summary detail
	SummaryDetail *yfSummaryDetail `json:"summaryDetail"`

	// Financial data module
	FinancialData *yfFinancialData `json:"financialData"`
}

type yfStatementContainer struct {
	Statements []map[string]yfFinVal `json:"incomeStatementHistory,omitempty"`
}

type yfFinVal struct {
	Raw float64 `json:"raw"`
	Fmt string  `json:"fmt"`
}

type yfAssetProfile struct {
	Industry           string              `json:"industry"`
	Sector             string              `json:"sector"`
	FullTimeEmployees  int64               `json:"fullTimeEmployees"`
	LongBusinessSummary string             `json:"longBusinessSummary"`
	City               string              `json:"city"`
	State              string              `json:"state"`
	Country            string              `json:"country"`
	Website            string              `json:"website"`
	CompanyOfficers    []yfCompanyOfficer  `json:"companyOfficers"`
}

type yfCompanyOfficer struct {
	Name      string  `json:"name"`
	Title     string  `json:"title"`
	Age       int     `json:"age"`
	TotalPay  yfFinVal `json:"totalPay"`
	YearBorn  int     `json:"yearBorn"`
}

type yfDefaultKeyStatistics struct {
	EnterpriseValue        yfFinVal `json:"enterpriseValue"`
	ForwardPE              yfFinVal `json:"forwardPE"`
	ProfitMargins          yfFinVal `json:"profitMargins"`
	FloatShares            yfFinVal `json:"floatShares"`
	SharesOutstanding      yfFinVal `json:"sharesOutstanding"`
	SharesShort            yfFinVal `json:"sharesShort"`
	SharesShortPriorMonth  yfFinVal `json:"sharesShortPriorMonth"`
	ShortRatio             yfFinVal `json:"shortRatio"`
	ShortPercentOfFloat    yfFinVal `json:"shortPercentOfFloat"`
	Beta                   yfFinVal `json:"beta"`
	BookValue              yfFinVal `json:"bookValue"`
	PriceToBook            yfFinVal `json:"priceToBook"`
	EnterpriseToRevenue    yfFinVal `json:"enterpriseToRevenue"`
	EnterpriseToEbitda     yfFinVal `json:"enterpriseToEbitda"`
	PegRatio               yfFinVal `json:"pegRatio"`
	TrailingEps            yfFinVal `json:"trailingEps"`
	ForwardEps             yfFinVal `json:"forwardEps"`
	FiftyTwoWeekChange     yfFinVal `json:"52WeekChange"`
	DividendYield          yfFinVal `json:"dividendYield"`
	LastDividendValue      yfFinVal `json:"lastDividendValue"`
	LastDividendDate       yfFinVal `json:"lastDividendDate"`
}

type yfSummaryDetail struct {
	PreviousClose      yfFinVal `json:"previousClose"`
	Open               yfFinVal `json:"open"`
	DayLow             yfFinVal `json:"dayLow"`
	DayHigh            yfFinVal `json:"dayHigh"`
	Volume             yfFinVal `json:"volume"`
	AverageVolume      yfFinVal `json:"averageVolume"`
	MarketCap          yfFinVal `json:"marketCap"`
	FiftyTwoWeekLow    yfFinVal `json:"fiftyTwoWeekLow"`
	FiftyTwoWeekHigh   yfFinVal `json:"fiftyTwoWeekHigh"`
	DividendRate       yfFinVal `json:"dividendRate"`
	DividendYield      yfFinVal `json:"dividendYield"`
	ExDividendDate     yfFinVal `json:"exDividendDate"`
	FiveYearAvgDividendYield yfFinVal `json:"fiveYearAvgDividendYield"`
	PayoutRatio        yfFinVal `json:"payoutRatio"`
	Beta               yfFinVal `json:"beta"`
	TrailingPE         yfFinVal `json:"trailingPE"`
	ForwardPE          yfFinVal `json:"forwardPE"`
	PriceToSalesTrailing12Months yfFinVal `json:"priceToSalesTrailing12Months"`
	TotalAssets        yfFinVal `json:"totalAssets"`
	NavPrice           yfFinVal `json:"navPrice"`
	YTDReturn          yfFinVal `json:"ytdReturn"`
}

type yfFinancialData struct {
	CurrentPrice        yfFinVal `json:"currentPrice"`
	TargetHighPrice     yfFinVal `json:"targetHighPrice"`
	TargetLowPrice      yfFinVal `json:"targetLowPrice"`
	TargetMeanPrice     yfFinVal `json:"targetMeanPrice"`
	TargetMedianPrice   yfFinVal `json:"targetMedianPrice"`
	RecommendationMean  yfFinVal `json:"recommendationMean"`
	RecommendationKey   string   `json:"recommendationKey"`
	NumberOfAnalystOpinions yfFinVal `json:"numberOfAnalystOpinions"`
	TotalRevenue        yfFinVal `json:"totalRevenue"`
	RevenuePerShare     yfFinVal `json:"revenuePerShare"`
	RevenueGrowth       yfFinVal `json:"revenueGrowth"`
	GrossProfits        yfFinVal `json:"grossProfits"`
	GrossMargins        yfFinVal `json:"grossMargins"`
	EbitdaMargins       yfFinVal `json:"ebitdaMargins"`
	OperatingMargins    yfFinVal `json:"operatingMargins"`
	ProfitMargins       yfFinVal `json:"profitMargins"`
	ReturnOnAssets      yfFinVal `json:"returnOnAssets"`
	ReturnOnEquity      yfFinVal `json:"returnOnEquity"`
	TotalCash           yfFinVal `json:"totalCash"`
	TotalDebt           yfFinVal `json:"totalDebt"`
	DebtToEquity        yfFinVal `json:"debtToEquity"`
	CurrentRatio        yfFinVal `json:"currentRatio"`
	FreeCashflow        yfFinVal `json:"freeCashflow"`
	OperatingCashflow   yfFinVal `json:"operatingCashflow"`
	EarningsGrowth      yfFinVal `json:"earningsGrowth"`
}

// yfSearchResponse wraps the v1 search API response.
type yfSearchResponse struct {
	Quotes []yfSearchQuote `json:"quotes"`
	News   []yfSearchNews  `json:"news"`
}

type yfSearchQuote struct {
	Exchange  string `json:"exchange"`
	ShortName string `json:"shortname"`
	LongName  string `json:"longname"`
	QuoteType string `json:"quoteType"`
	Symbol    string `json:"symbol"`
	Sector    string `json:"sector"`
	Industry  string `json:"industry"`
	IsYahooFinance bool `json:"isYahooFinance"`
}

type yfSearchNews struct {
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
	Link      string `json:"link"`
	UUID      string `json:"uuid"`
}

// yfOptionsResponse wraps the options API response.
type yfOptionsResponse struct {
	OptionChain struct {
		Result []yfOptionsResult `json:"result"`
		Error  *yfError          `json:"error"`
	} `json:"optionChain"`
}

type yfOptionsResult struct {
	UnderlyingSymbol string          `json:"underlyingSymbol"`
	ExpirationDates  []int64         `json:"expirationDates"`
	Strikes          []float64       `json:"strikes"`
	Quote            yfQuoteResult   `json:"quote"`
	Options          []yfOptionChain `json:"options"`
}

type yfOptionChain struct {
	ExpirationDate int64        `json:"expirationDate"`
	Calls          []yfContract `json:"calls"`
	Puts           []yfContract `json:"puts"`
}

type yfContract struct {
	ContractSymbol    string  `json:"contractSymbol"`
	Strike            float64 `json:"strike"`
	Currency          string  `json:"currency"`
	LastPrice         float64 `json:"lastPrice"`
	Change            float64 `json:"change"`
	PercentChange     float64 `json:"percentChange"`
	Volume            int64   `json:"volume"`
	OpenInterest      int64   `json:"openInterest"`
	Bid               float64 `json:"bid"`
	Ask               float64 `json:"ask"`
	ImpliedVolatility float64 `json:"impliedVolatility"`
	InTheMoney        bool    `json:"inTheMoney"`
	Expiration        int64   `json:"expiration"`
}

// yfScreenerResponse wraps the screener API response.
type yfScreenerResponse struct {
	Finance struct {
		Result []yfScreenerResult `json:"result"`
		Error  *yfError           `json:"error"`
	} `json:"finance"`
}

type yfScreenerResult struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Quotes []yfQuoteResult `json:"quotes"`
	Total  int    `json:"total"`
}

type yfError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}
