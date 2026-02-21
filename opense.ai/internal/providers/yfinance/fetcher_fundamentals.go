package yfinance

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// --- BalanceSheet fetcher ---

type balanceSheetFetcher struct {
	provider.BaseFetcher
}

func newBalanceSheetFetcher() *balanceSheetFetcher {
	return &balanceSheetFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelBalanceSheet,
			"Balance sheet data from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamPeriod},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *balanceSheetFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	period := params[provider.ParamPeriod]
	modules := "balanceSheetHistory"
	if period == "quarterly" {
		modules = "balanceSheetHistoryQuarterly"
	}

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=%s",
		yfTicker, modules,
	)

	var resp yfQuoteSummaryResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance balance sheet %s: %w", yfTicker, err)
	}
	if resp.QuoteSummary.Error != nil {
		return nil, fmt.Errorf("yfinance API error: %s", resp.QuoteSummary.Error.Description)
	}
	if len(resp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no balance sheet data for %s", symbol)
	}

	r := resp.QuoteSummary.Result[0]
	var stmts *yfStatementContainer
	periodType := "annual"
	if period == "quarterly" {
		stmts = r.BalanceSheetHistoryQuarterly
		periodType = "quarterly"
	} else {
		stmts = r.BalanceSheetHistory
	}

	sheets := parseBalanceSheets(stmts, periodType)
	f.CacheSetTTL(cacheKey, sheets, 1*time.Hour)
	return newResult(sheets), nil
}

func parseBalanceSheets(container *yfStatementContainer, periodType string) []models.BalanceSheet {
	if container == nil || len(container.Statements) == 0 {
		return nil
	}
	sheets := make([]models.BalanceSheet, 0, len(container.Statements))
	for _, stmt := range container.Statements {
		bs := models.BalanceSheet{
			Period:     extractDate(stmt),
			PeriodType: periodType,
		}
		bs.TotalAssets = valRaw(stmt, "totalAssets")
		bs.CurrentAssets = valRaw(stmt, "totalCurrentAssets")
		bs.CashEquivalents = valRaw(stmt, "cash")
		bs.Inventory = valRaw(stmt, "inventory")
		bs.TradeReceivables = valRaw(stmt, "netReceivables")
		bs.FixedAssets = valRaw(stmt, "propertyPlantEquipment")
		bs.TotalLiabilities = valRaw(stmt, "totalLiab")
		bs.CurrentLiabilities = valRaw(stmt, "totalCurrentLiabilities")
		bs.LongTermBorrowings = valRaw(stmt, "longTermDebt")
		bs.ShortTermBorrowings = valRaw(stmt, "shortLongTermDebt")
		bs.TotalDebt = bs.LongTermBorrowings + bs.ShortTermBorrowings
		bs.TotalEquity = valRaw(stmt, "totalStockholderEquity")
		bs.ShareCapital = valRaw(stmt, "commonStock")
		bs.Reserves = valRaw(stmt, "retainedEarnings")
		sheets = append(sheets, bs)
	}
	return sheets
}

// --- IncomeStatement fetcher ---

type incomeStatementFetcher struct {
	provider.BaseFetcher
}

func newIncomeStatementFetcher() *incomeStatementFetcher {
	return &incomeStatementFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelIncomeStatement,
			"Income statement data from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamPeriod},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *incomeStatementFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	period := params[provider.ParamPeriod]
	modules := "incomeStatementHistory"
	if period == "quarterly" {
		modules = "incomeStatementHistoryQuarterly"
	}

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=%s",
		yfTicker, modules,
	)

	var resp yfQuoteSummaryResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance income statement %s: %w", yfTicker, err)
	}
	if resp.QuoteSummary.Error != nil {
		return nil, fmt.Errorf("yfinance API error: %s", resp.QuoteSummary.Error.Description)
	}
	if len(resp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no income statement data for %s", symbol)
	}

	r := resp.QuoteSummary.Result[0]
	var stmts *yfStatementContainer
	periodType := "annual"
	if period == "quarterly" {
		stmts = r.IncomeStatementHistoryQuarterly
		periodType = "quarterly"
	} else {
		stmts = r.IncomeStatementHistory
	}

	income := parseIncomeStatements(stmts, periodType)
	f.CacheSetTTL(cacheKey, income, 1*time.Hour)
	return newResult(income), nil
}

func parseIncomeStatements(container *yfStatementContainer, periodType string) []models.IncomeStatement {
	if container == nil || len(container.Statements) == 0 {
		return nil
	}
	stmts := make([]models.IncomeStatement, 0, len(container.Statements))
	for _, stmt := range container.Statements {
		is := models.IncomeStatement{
			Period:     extractDate(stmt),
			PeriodType: periodType,
		}
		is.Revenue = valRaw(stmt, "totalRevenue")
		is.TotalIncome = valRaw(stmt, "totalRevenue")
		is.TotalExpenses = valRaw(stmt, "totalOperatingExpenses")
		is.EBITDA = valRaw(stmt, "ebitda")
		is.EBIT = valRaw(stmt, "ebit")
		is.InterestExpense = valRaw(stmt, "interestExpense")
		is.PBT = valRaw(stmt, "incomeBeforeTax")
		is.Tax = valRaw(stmt, "incomeTaxExpense")
		is.PAT = valRaw(stmt, "netIncome")
		is.Depreciation = valRaw(stmt, "depreciation")
		is.EPS = valRaw(stmt, "dilutedEPS")
		if is.Revenue > 0 {
			is.OPMPct = (is.EBIT / is.Revenue) * 100
			is.NPMPct = (is.PAT / is.Revenue) * 100
		}
		stmts = append(stmts, is)
	}
	return stmts
}

// --- CashFlowStatement fetcher ---

type cashFlowStatementFetcher struct {
	provider.BaseFetcher
}

func newCashFlowStatementFetcher() *cashFlowStatementFetcher {
	return &cashFlowStatementFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCashFlowStatement,
			"Cash flow statement data from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamPeriod},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *cashFlowStatementFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	period := params[provider.ParamPeriod]
	modules := "cashflowStatementHistory"
	if period == "quarterly" {
		modules = "cashflowStatementHistoryQuarterly"
	}

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=%s",
		yfTicker, modules,
	)

	var resp yfQuoteSummaryResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance cash flow %s: %w", yfTicker, err)
	}
	if resp.QuoteSummary.Error != nil {
		return nil, fmt.Errorf("yfinance API error: %s", resp.QuoteSummary.Error.Description)
	}
	if len(resp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no cash flow data for %s", symbol)
	}

	r := resp.QuoteSummary.Result[0]
	var stmts *yfStatementContainer
	periodType := "annual"
	if period == "quarterly" {
		stmts = r.CashflowStatementHistoryQuarterly
		periodType = "quarterly"
	} else {
		stmts = r.CashflowStatementHistory
	}

	cfs := parseCashFlowStatements(stmts, periodType)
	f.CacheSetTTL(cacheKey, cfs, 1*time.Hour)
	return newResult(cfs), nil
}

func parseCashFlowStatements(container *yfStatementContainer, periodType string) []models.CashFlow {
	if container == nil || len(container.Statements) == 0 {
		return nil
	}
	cfs := make([]models.CashFlow, 0, len(container.Statements))
	for _, stmt := range container.Statements {
		cf := models.CashFlow{
			Period:     extractDate(stmt),
			PeriodType: periodType,
		}
		cf.OperatingCashFlow = valRaw(stmt, "totalCashFromOperatingActivities")
		cf.InvestingCashFlow = valRaw(stmt, "totalCashflowsFromInvestingActivities")
		cf.FinancingCashFlow = valRaw(stmt, "totalCashFromFinancingActivities")
		cf.NetCashFlow = valRaw(stmt, "changeInCash")
		cf.CapEx = valRaw(stmt, "capitalExpenditures")
		cf.DividendsPaid = valRaw(stmt, "dividendsPaid")
		cf.FreeCashFlow = cf.OperatingCashFlow + cf.CapEx // capex is negative
		cfs = append(cfs, cf)
	}
	return cfs
}

// --- KeyMetrics fetcher ---

type keyMetricsFetcher struct {
	provider.BaseFetcher
}

func newKeyMetricsFetcher() *keyMetricsFetcher {
	return &keyMetricsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelKeyMetrics,
			"Key financial metrics from Yahoo Finance",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *keyMetricsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	modules := "defaultKeyStatistics,financialData,summaryDetail"
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=%s",
		yfTicker, modules,
	)

	var resp yfQuoteSummaryResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance key metrics %s: %w", yfTicker, err)
	}
	if resp.QuoteSummary.Error != nil {
		return nil, fmt.Errorf("yfinance API error: %s", resp.QuoteSummary.Error.Description)
	}
	if len(resp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no metrics for %s", symbol)
	}

	r := resp.QuoteSummary.Result[0]
	ratios := &models.FinancialRatios{}
	if ks := r.DefaultKeyStatistics; ks != nil {
		ratios.PB = ks.PriceToBook.Raw
		ratios.BookValue = ks.BookValue.Raw
		ratios.PEGRatio = ks.PegRatio.Raw
		ratios.EPS = ks.TrailingEps.Raw
		ratios.EVBITDA = ks.EnterpriseToEbitda.Raw
	}
	if fd := r.FinancialData; fd != nil {
		ratios.ROE = fd.ReturnOnEquity.Raw * 100
		ratios.ROCE = fd.ReturnOnAssets.Raw * 100
		ratios.DebtEquity = fd.DebtToEquity.Raw
		ratios.CurrentRatio = fd.CurrentRatio.Raw
	}
	if sd := r.SummaryDetail; sd != nil {
		ratios.PE = sd.TrailingPE.Raw
		ratios.DividendYield = sd.DividendYield.Raw * 100
	}

	f.CacheSetTTL(cacheKey, ratios, 1*time.Hour)
	return newResult(ratios), nil
}

// --- HistoricalDividends fetcher ---

type historicalDividendsFetcher struct {
	provider.BaseFetcher
}

func newHistoricalDividendsFetcher() *historicalDividendsFetcher {
	return &historicalDividendsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelHistoricalDividends,
			"Historical dividend data from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *historicalDividendsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	startDate, endDate := defaultDateRange(params)
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d&events=div",
		yfTicker, startDate.Unix(), endDate.Unix(),
	)

	var resp yfChartDividendResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance dividends %s: %w", yfTicker, err)
	}
	if resp.Chart.Error != nil {
		return nil, fmt.Errorf("yfinance error: %s", resp.Chart.Error.Description)
	}
	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no dividend data for %s", symbol)
	}

	divs := make([]models.DividendRecord, 0)
	if events := resp.Chart.Result[0].Events; events != nil {
		for _, d := range events.Dividends {
			divs = append(divs, models.DividendRecord{
				Symbol: fromYFTicker(yfTicker),
				ExDate: time.Unix(d.Date, 0),
				Amount: d.Amount,
			})
		}
	}

	f.CacheSetTTL(cacheKey, divs, 1*time.Hour)
	return newResult(divs), nil
}

// --- ShareStatistics fetcher ---

type shareStatisticsFetcher struct {
	provider.BaseFetcher
}

func newShareStatisticsFetcher() *shareStatisticsFetcher {
	return &shareStatisticsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelShareStatistics,
			"Share statistics (float, short interest) from Yahoo Finance",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *shareStatisticsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	modules := "defaultKeyStatistics"
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=%s",
		yfTicker, modules,
	)

	var resp yfQuoteSummaryResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance share stats %s: %w", yfTicker, err)
	}
	if resp.QuoteSummary.Error != nil {
		return nil, fmt.Errorf("yfinance API error: %s", resp.QuoteSummary.Error.Description)
	}
	if len(resp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no share stats for %s", symbol)
	}

	ks := resp.QuoteSummary.Result[0].DefaultKeyStatistics
	stats := &models.ShareStatisticsData{
		Symbol: fromYFTicker(yfTicker),
	}
	if ks != nil {
		stats.SharesOutstanding = int64(ks.SharesOutstanding.Raw)
		stats.FloatShares = int64(ks.FloatShares.Raw)
		stats.SharesShort = int64(ks.SharesShort.Raw)
		stats.ShortRatio = ks.ShortRatio.Raw
		stats.ShortPercentFloat = ks.ShortPercentOfFloat.Raw * 100
	}

	f.CacheSetTTL(cacheKey, stats, 1*time.Hour)
	return newResult(stats), nil
}

// --- Shared financial statement helpers ---

// extractDate tries to extract a date string from a YF statement map.
func extractDate(stmt map[string]yfFinVal) string {
	if v, ok := stmt["endDate"]; ok {
		if v.Fmt != "" {
			return v.Fmt
		}
		if v.Raw > 0 {
			return time.Unix(int64(v.Raw), 0).Format("2006-01-02")
		}
	}
	return ""
}

// valRaw extracts the raw numeric value for a key from a YF statement map.
func valRaw(stmt map[string]yfFinVal, key string) float64 {
	if v, ok := stmt[key]; ok {
		return v.Raw
	}
	return 0
}

// yfChartDividendResponse extends chart response with events (dividends).
type yfChartDividendResponse struct {
	Chart struct {
		Result []yfChartDividendResult `json:"result"`
		Error  *yfError                `json:"error"`
	} `json:"chart"`
}

type yfChartDividendResult struct {
	Meta       yfChartMeta       `json:"meta"`
	Timestamp  []int64           `json:"timestamp"`
	Events     *yfChartEvents    `json:"events"`
	Indicators yfIndicators      `json:"indicators"`
}

type yfChartEvents struct {
	Dividends map[string]yfDividendEvent `json:"dividends"`
	Splits    map[string]yfSplitEvent    `json:"splits"`
}

type yfDividendEvent struct {
	Amount float64 `json:"amount"`
	Date   int64   `json:"date"`
}

type yfSplitEvent struct {
	Date        int64   `json:"date"`
	Numerator   float64 `json:"numerator"`
	Denominator float64 `json:"denominator"`
	Ratio       string  `json:"splitRatio"`
}
