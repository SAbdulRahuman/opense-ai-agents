package models

// IncomeStatement represents a single period income statement.
type IncomeStatement struct {
	Period           string  `json:"period"`            // e.g., "Mar 2025", "Q3 FY26"
	PeriodType       string  `json:"period_type"`       // "annual" or "quarterly"
	Revenue          float64 `json:"revenue"`           // Total revenue / Net sales
	OtherIncome      float64 `json:"other_income"`
	TotalIncome      float64 `json:"total_income"`
	RawMaterials     float64 `json:"raw_materials"`
	EmployeeCost     float64 `json:"employee_cost"`
	OtherExpenses    float64 `json:"other_expenses"`
	TotalExpenses    float64 `json:"total_expenses"`
	EBITDA           float64 `json:"ebitda"`
	Depreciation     float64 `json:"depreciation"`
	EBIT             float64 `json:"ebit"`
	InterestExpense  float64 `json:"interest_expense"`
	PBT              float64 `json:"pbt"`              // Profit Before Tax
	Tax              float64 `json:"tax"`
	PAT              float64 `json:"pat"`              // Profit After Tax
	EPS              float64 `json:"eps"`
	OPMPct           float64 `json:"opm_pct"`          // Operating Profit Margin %
	NPMPct           float64 `json:"npm_pct"`          // Net Profit Margin %
}

// BalanceSheet represents a single period balance sheet.
type BalanceSheet struct {
	Period              string  `json:"period"`
	PeriodType          string  `json:"period_type"`
	// Assets
	TotalAssets         float64 `json:"total_assets"`
	FixedAssets         float64 `json:"fixed_assets"`
	CWIP                float64 `json:"cwip"`                  // Capital Work in Progress
	Investments         float64 `json:"investments"`
	OtherAssets         float64 `json:"other_assets"`
	CurrentAssets       float64 `json:"current_assets"`
	Inventory           float64 `json:"inventory"`
	TradeReceivables    float64 `json:"trade_receivables"`
	CashEquivalents     float64 `json:"cash_equivalents"`
	// Liabilities
	TotalLiabilities    float64 `json:"total_liabilities"`
	ShareCapital        float64 `json:"share_capital"`
	Reserves            float64 `json:"reserves"`
	TotalEquity         float64 `json:"total_equity"`
	LongTermBorrowings  float64 `json:"long_term_borrowings"`
	ShortTermBorrowings float64 `json:"short_term_borrowings"`
	TotalDebt           float64 `json:"total_debt"`
	CurrentLiabilities  float64 `json:"current_liabilities"`
	TradePayables       float64 `json:"trade_payables"`
	OtherLiabilities    float64 `json:"other_liabilities"`
}

// CashFlow represents a single period cash flow statement.
type CashFlow struct {
	Period               string  `json:"period"`
	PeriodType           string  `json:"period_type"`
	OperatingCashFlow    float64 `json:"operating_cash_flow"`
	InvestingCashFlow    float64 `json:"investing_cash_flow"`
	FinancingCashFlow    float64 `json:"financing_cash_flow"`
	NetCashFlow          float64 `json:"net_cash_flow"`
	FreeCashFlow         float64 `json:"free_cash_flow"`
	CapEx                float64 `json:"capex"`
	DividendsPaid        float64 `json:"dividends_paid"`
}

// FinancialData aggregates all financial statements for a stock.
type FinancialData struct {
	Ticker               string            `json:"ticker"`
	AnnualIncome         []IncomeStatement `json:"annual_income"`
	QuarterlyIncome      []IncomeStatement `json:"quarterly_income"`
	AnnualBalanceSheet   []BalanceSheet    `json:"annual_balance_sheet"`
	QuarterlyBalanceSheet []BalanceSheet   `json:"quarterly_balance_sheet"`
	AnnualCashFlow       []CashFlow        `json:"annual_cash_flow"`
	QuarterlyCashFlow    []CashFlow        `json:"quarterly_cash_flow"`
}

// GrowthRates holds computed growth metrics.
type GrowthRates struct {
	RevenueGrowthQoQ  float64 `json:"revenue_growth_qoq"`
	RevenueGrowthYoY  float64 `json:"revenue_growth_yoy"`
	RevenueCAGR3Y     float64 `json:"revenue_cagr_3y"`
	RevenueCAGR5Y     float64 `json:"revenue_cagr_5y"`
	ProfitGrowthQoQ   float64 `json:"profit_growth_qoq"`
	ProfitGrowthYoY   float64 `json:"profit_growth_yoy"`
	ProfitCAGR3Y      float64 `json:"profit_cagr_3y"`
	ProfitCAGR5Y      float64 `json:"profit_cagr_5y"`
	EPSGrowthYoY      float64 `json:"eps_growth_yoy"`
	EPSCAGR3Y         float64 `json:"eps_cagr_3y"`
}
