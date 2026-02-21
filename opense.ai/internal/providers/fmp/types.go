package fmp

// --- FMP API response types ---

// fmpQuote represents a real-time quote from FMP.
type fmpQuote struct {
	Symbol                string  `json:"symbol"`
	Name                  string  `json:"name"`
	Price                 float64 `json:"price"`
	ChangesPercentage     float64 `json:"changesPercentage"`
	Change                float64 `json:"change"`
	DayLow                float64 `json:"dayLow"`
	DayHigh               float64 `json:"dayHigh"`
	YearHigh              float64 `json:"yearHigh"`
	YearLow               float64 `json:"yearLow"`
	MarketCap             float64 `json:"marketCap"`
	PriceAvg50            float64 `json:"priceAvg50"`
	PriceAvg200           float64 `json:"priceAvg200"`
	Volume                int64   `json:"volume"`
	AvgVolume             int64   `json:"avgVolume"`
	Exchange              string  `json:"exchange"`
	Open                  float64 `json:"open"`
	PreviousClose         float64 `json:"previousClose"`
	EPS                   float64 `json:"eps"`
	PE                    float64 `json:"pe"`
	SharesOutstanding     float64 `json:"sharesOutstanding"`
	Timestamp             int64   `json:"timestamp"`
}

// fmpHistorical represents historical OHLCV data from FMP.
type fmpHistoricalPrice struct {
	Historical []fmpHistoricalEntry `json:"historical"`
	Symbol     string               `json:"symbol"`
}

type fmpHistoricalEntry struct {
	Date             string  `json:"date"`
	Open             float64 `json:"open"`
	High             float64 `json:"high"`
	Low              float64 `json:"low"`
	Close            float64 `json:"close"`
	AdjClose         float64 `json:"adjClose"`
	Volume           int64   `json:"volume"`
	UnadjustedVolume int64   `json:"unadjustedVolume"`
	Change           float64 `json:"change"`
	ChangePercent    float64 `json:"changePercent"`
	VWAP             float64 `json:"vwap"`
}

// fmpProfile represents company profile from FMP.
type fmpProfile struct {
	Symbol            string  `json:"symbol"`
	Price             float64 `json:"price"`
	Beta              float64 `json:"beta"`
	VolAvg            int64   `json:"volAvg"`
	MktCap            float64 `json:"mktCap"`
	LastDiv           float64 `json:"lastDiv"`
	Range             string  `json:"range"`
	Changes           float64 `json:"changes"`
	CompanyName       string  `json:"companyName"`
	Currency          string  `json:"currency"`
	CIK               string  `json:"cik"`
	ISIN              string  `json:"isin"`
	Exchange          string  `json:"exchange"`
	ExchangeShortName string  `json:"exchangeShortName"`
	Industry          string  `json:"industry"`
	Website           string  `json:"website"`
	Description       string  `json:"description"`
	CEO               string  `json:"ceo"`
	Sector            string  `json:"sector"`
	Country           string  `json:"country"`
	FullTimeEmployees string  `json:"fullTimeEmployees"`
	Phone             string  `json:"phone"`
	Address           string  `json:"address"`
	City              string  `json:"city"`
	State             string  `json:"state"`
	Zip               string  `json:"zip"`
	DCFDiff           float64 `json:"dcfDiff"`
	DCF               float64 `json:"dcf"`
	IPODate           string  `json:"ipoDate"`
	DefaultImage      bool    `json:"defaultImage"`
	IsETF             bool    `json:"isEtf"`
	IsActivelyTrading bool    `json:"isActivelyTrading"`
	IsFund            bool    `json:"isFund"`
}

// fmpSearchResult represents a search result from FMP.
type fmpSearchResult struct {
	Symbol            string `json:"symbol"`
	Name              string `json:"name"`
	Currency          string `json:"currency"`
	StockExchange     string `json:"stockExchange"`
	ExchangeShortName string `json:"exchangeShortName"`
}

// fmpScreenerResult represents a screener result from FMP.
type fmpScreenerResult struct {
	Symbol            string  `json:"symbol"`
	CompanyName       string  `json:"companyName"`
	MarketCap         float64 `json:"marketCap"`
	Sector            string  `json:"sector"`
	Industry          string  `json:"industry"`
	Beta              float64 `json:"beta"`
	Price             float64 `json:"price"`
	LastAnnualDividend float64 `json:"lastAnnualDividend"`
	Volume            int64   `json:"volume"`
	Exchange          string  `json:"exchange"`
	ExchangeShortName string  `json:"exchangeShortName"`
	Country           string  `json:"country"`
	IsETF             bool    `json:"isEtf"`
	IsActivelyTrading bool    `json:"isActivelyTrading"`
}

// fmpPeerResult represents stock peers.
type fmpPeerResult struct {
	Symbol    string   `json:"symbol"`
	PeersList []string `json:"peersList"`
}

// fmpIncomeStatement represents an income statement from FMP.
type fmpIncomeStatement struct {
	Date                   string  `json:"date"`
	Symbol                 string  `json:"symbol"`
	Period                 string  `json:"period"` // "FY" or "Q1", "Q2", etc.
	Revenue                float64 `json:"revenue"`
	CostOfRevenue          float64 `json:"costOfRevenue"`
	GrossProfit            float64 `json:"grossProfit"`
	OperatingExpenses      float64 `json:"operatingExpenses"`
	OperatingIncome        float64 `json:"operatingIncome"`
	InterestExpense        float64 `json:"interestExpense"`
	IncomeBeforeTax        float64 `json:"incomeBeforeTax"`
	IncomeTaxExpense       float64 `json:"incomeTaxExpense"`
	NetIncome              float64 `json:"netIncome"`
	EPS                    float64 `json:"eps"`
	EPSDiluted             float64 `json:"epsdiluted"`
	EBITDA                 float64 `json:"ebitda"`
	DepreciationAndAmortization float64 `json:"depreciationAndAmortization"`
	GrossProfitRatio       float64 `json:"grossProfitRatio"`
	OperatingIncomeRatio   float64 `json:"operatingIncomeRatio"`
	NetIncomeRatio         float64 `json:"netIncomeRatio"`
}

// fmpBalanceSheet represents a balance sheet from FMP.
type fmpBalanceSheet struct {
	Date                      string  `json:"date"`
	Symbol                    string  `json:"symbol"`
	Period                    string  `json:"period"`
	CashAndCashEquivalents    float64 `json:"cashAndCashEquivalents"`
	ShortTermInvestments      float64 `json:"shortTermInvestments"`
	NetReceivables            float64 `json:"netReceivables"`
	Inventory                 float64 `json:"inventory"`
	TotalCurrentAssets        float64 `json:"totalCurrentAssets"`
	PropertyPlantEquipmentNet float64 `json:"propertyPlantEquipmentNet"`
	LongTermInvestments       float64 `json:"longTermInvestments"`
	TotalNonCurrentAssets     float64 `json:"totalNonCurrentAssets"`
	TotalAssets               float64 `json:"totalAssets"`
	AccountPayables           float64 `json:"accountPayables"`
	ShortTermDebt             float64 `json:"shortTermDebt"`
	TotalCurrentLiabilities   float64 `json:"totalCurrentLiabilities"`
	LongTermDebt              float64 `json:"longTermDebt"`
	TotalNonCurrentLiabilities float64 `json:"totalNonCurrentLiabilities"`
	TotalLiabilities          float64 `json:"totalLiabilities"`
	CommonStock               float64 `json:"commonStock"`
	RetainedEarnings          float64 `json:"retainedEarnings"`
	TotalStockholdersEquity   float64 `json:"totalStockholdersEquity"`
	TotalEquity               float64 `json:"totalEquity"`
}

// fmpCashFlow represents a cash flow statement from FMP.
type fmpCashFlow struct {
	Date                       string  `json:"date"`
	Symbol                     string  `json:"symbol"`
	Period                     string  `json:"period"`
	NetIncome                  float64 `json:"netIncome"`
	DepreciationAndAmortization float64 `json:"depreciationAndAmortization"`
	OperatingCashFlow          float64 `json:"operatingCashFlow"`
	CapitalExpenditure         float64 `json:"capitalExpenditure"`
	InvestmentsInPPE           float64 `json:"investmentsInPropertyPlantAndEquipment"`
	InvestingActivitiesCF      float64 `json:"netCashUsedForInvestingActivites"`
	DebtRepayment              float64 `json:"debtRepayment"`
	DividendsPaid              float64 `json:"dividendsPaid"`
	FinancingActivitiesCF      float64 `json:"netCashUsedProvidedByFinancingActivities"`
	NetChangeInCash            float64 `json:"netChangeInCash"`
	FreeCashFlow               float64 `json:"freeCashFlow"`
}

// fmpKeyMetrics represents key financial metrics from FMP.
type fmpKeyMetrics struct {
	Symbol            string  `json:"symbol"`
	Date              string  `json:"date"`
	Period            string  `json:"period"`
	RevenuePerShare   float64 `json:"revenuePerShare"`
	NetIncomePerShare float64 `json:"netIncomePerShare"`
	BookValuePerShare float64 `json:"bookValuePerShare"`
	PERatio           float64 `json:"peRatio"`
	PBRatio           float64 `json:"pbRatio"`
	PEGRatio          float64 `json:"pegRatio"`
	EVToEBITDA        float64 `json:"enterpriseValueOverEBITDA"`
	DebtToEquity      float64 `json:"debtToEquity"`
	CurrentRatio      float64 `json:"currentRatio"`
	InterestCoverage  float64 `json:"interestCoverage"`
	DividendYield     float64 `json:"dividendYield"`
	ROE               float64 `json:"roe"`
	ROIC              float64 `json:"roic"`
	ROA               float64 `json:"returnOnTangibleAssets"`
	EarningsYield     float64 `json:"earningsYield"`
	FreeCashFlowYield float64 `json:"freeCashFlowYield"`
	GrahamNumber      float64 `json:"grahamNumber"`
}

// fmpRatios represents financial ratios from FMP.
type fmpRatios struct {
	Symbol               string  `json:"symbol"`
	Date                 string  `json:"date"`
	Period               string  `json:"period"`
	CurrentRatio         float64 `json:"currentRatio"`
	QuickRatio           float64 `json:"quickRatio"`
	DebtEquityRatio      float64 `json:"debtEquityRatio"`
	InterestCoverage     float64 `json:"interestCoverage"`
	GrossProfitMargin    float64 `json:"grossProfitMargin"`
	OperatingProfitMargin float64 `json:"operatingProfitMargin"`
	NetProfitMargin      float64 `json:"netProfitMargin"`
	ROE                  float64 `json:"returnOnEquity"`
	ROA                  float64 `json:"returnOnAssets"`
	ROIC                 float64 `json:"returnOnCapitalEmployed"`
	DividendYield        float64 `json:"dividendYield"`
	PriceEarningsRatio   float64 `json:"priceEarningsRatio"`
	PriceBookRatio       float64 `json:"priceToBookRatio"`
	PEGRatio             float64 `json:"priceEarningsToGrowthRatio"`
	EBITPerRevenue       float64 `json:"ebitPerRevenue"`
	EVToEBITDA           float64 `json:"enterpriseValueMultiple"`
}

// fmpKeyExecutive represents a company executive.
type fmpKeyExecutive struct {
	Title       string  `json:"title"`
	Name        string  `json:"name"`
	Pay         float64 `json:"pay"`
	CurrencyPay string  `json:"currencyPay"`
	Gender      string  `json:"gender"`
	YearBorn    int     `json:"yearBorn"`
}

// fmpHistoricalDividend represents historical dividend data.
type fmpHistoricalDividend struct {
	Historical []fmpDividendEntry `json:"historical"`
	Symbol     string             `json:"symbol"`
}

type fmpDividendEntry struct {
	Date            string  `json:"date"`
	Label           string  `json:"label"`
	AdjDividend     float64 `json:"adjDividend"`
	Dividend        float64 `json:"dividend"`
	RecordDate      string  `json:"recordDate"`
	PaymentDate     string  `json:"paymentDate"`
	DeclarationDate string  `json:"declarationDate"`
}

// fmpShareFloat represents share statistics.
type fmpShareFloat struct {
	Symbol               string  `json:"symbol"`
	FreeFloat            float64 `json:"freeFloat"`
	FloatShares          float64 `json:"floatShares"`
	OutstandingShares    float64 `json:"outstandingShares"`
	Date                 string  `json:"date"`
}

// fmpPriceTarget represents analyst price targets.
type fmpPriceTarget struct {
	Symbol          string `json:"symbol"`
	PublishedDate   string `json:"publishedDate"`
	AnalystName     string `json:"analystName"`
	AnalystCompany  string `json:"analystCompany"`
	PriceTarget     float64 `json:"priceTarget"`
	AdjPriceTarget  float64 `json:"adjPriceTarget"`
	PriceWhenPosted float64 `json:"priceWhenPosted"`
	NewsURL         string `json:"newsURL"`
	NewsTitle       string `json:"newsTitle"`
	NewBaseFormula  string `json:"newGrade"`
	OldGrade        string `json:"previousGrade"`
}

// fmpPriceTargetConsensus represents consensus price target.
type fmpPriceTargetConsensus struct {
	Symbol           string  `json:"symbol"`
	TargetHigh       float64 `json:"targetHigh"`
	TargetLow        float64 `json:"targetLow"`
	TargetConsensus  float64 `json:"targetConsensus"`
	TargetMedian     float64 `json:"targetMedian"`
}

// fmpAnalystEstimate represents analyst estimates from FMP.
type fmpAnalystEstimate struct {
	Symbol                  string  `json:"symbol"`
	Date                    string  `json:"date"`
	EstimatedRevenueAvg     float64 `json:"estimatedRevenueAvg"`
	EstimatedRevenueHigh    float64 `json:"estimatedRevenueHigh"`
	EstimatedRevenueLow     float64 `json:"estimatedRevenueLow"`
	EstimatedEbitdaAvg      float64 `json:"estimatedEbitdaAvg"`
	EstimatedEbitdaHigh     float64 `json:"estimatedEbitdaHigh"`
	EstimatedEbitdaLow      float64 `json:"estimatedEbitdaLow"`
	EstimatedNetIncomeAvg   float64 `json:"estimatedNetIncomeAvg"`
	EstimatedEpsAvg         float64 `json:"estimatedEpsAvg"`
	EstimatedEpsHigh        float64 `json:"estimatedEpsHigh"`
	EstimatedEpsLow         float64 `json:"estimatedEpsLow"`
	NumberAnalystsEstimated int     `json:"numberAnalystsEstimated"`
}

// fmpEarningsCalendar represents earnings calendar entry from FMP.
type fmpEarningsCalendar struct {
	Date             string  `json:"date"`
	Symbol           string  `json:"symbol"`
	EPS              float64 `json:"eps"`
	EPSEstimated     float64 `json:"epsEstimated"`
	Revenue          float64 `json:"revenue"`
	RevenueEstimated float64 `json:"revenueEstimated"`
	FiscalDateEnding string  `json:"fiscalDateEnding"`
}

// fmpDividendCalendar represents dividend calendar entry from FMP.
type fmpDividendCalendar struct {
	Date            string  `json:"date"`
	Label           string  `json:"label"`
	Symbol          string  `json:"symbol"`
	AdjDividend     float64 `json:"adjDividend"`
	Dividend        float64 `json:"dividend"`
	RecordDate      string  `json:"recordDate"`
	PaymentDate     string  `json:"paymentDate"`
	DeclarationDate string  `json:"declarationDate"`
}

// fmpIPOCalendar represents IPO calendar entry from FMP.
type fmpIPOCalendar struct {
	Date           string  `json:"date"`
	Company        string  `json:"company"`
	Symbol         string  `json:"symbol"`
	Exchange       string  `json:"exchange"`
	Actions        string  `json:"actions"`
	Shares         int64   `json:"shares"`
	PriceRange     string  `json:"priceRange"`
	MarketCap      float64 `json:"marketCap"`
}

// fmpGainer represents a market gainer/loser/active from FMP.
type fmpGainer struct {
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	Change            float64 `json:"change"`
	Price             float64 `json:"price"`
	ChangesPercentage float64 `json:"changesPercentage"`
}

// fmpNewsArticle represents a news article from FMP.
type fmpNewsArticle struct {
	Symbol        string `json:"symbol"`
	PublishedDate string `json:"publishedDate"`
	Title         string `json:"title"`
	Image         string `json:"image"`
	Site          string `json:"site"`
	Text          string `json:"text"`
	URL           string `json:"url"`
}

// fmpPricePerformance represents stock price change summary.
type fmpPricePerformance struct {
	Symbol string  `json:"symbol"`
	OneDay float64 `json:"1D"`
	FiveDay float64 `json:"5D"`
	OneMonth float64 `json:"1M"`
	ThreeMonth float64 `json:"3M"`
	SixMonth float64 `json:"6M"`
	YTD     float64 `json:"ytd"`
	OneYear float64 `json:"1Y"`
	ThreeYear float64 `json:"3Y"`
	FiveYear float64 `json:"5Y"`
	TenYear float64 `json:"10Y"`
	Max     float64 `json:"max"`
}
