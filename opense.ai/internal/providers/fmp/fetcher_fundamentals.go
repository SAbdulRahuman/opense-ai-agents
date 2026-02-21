package fmp

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// --- IncomeStatement fetcher ---

type incomeStatementFetcher struct {
	provider.BaseFetcher
}

func newIncomeStatementFetcher() *incomeStatementFetcher {
	return &incomeStatementFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelIncomeStatement,
			"Income statement from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamPeriod, provider.ParamLimit},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *incomeStatementFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	period := params[provider.ParamPeriod]
	path := fmt.Sprintf("/income-statement/%s?", symbol)
	if period == "quarterly" {
		path += "period=quarter&"
	}
	if limit := params[provider.ParamLimit]; limit != "" {
		path += "limit=" + limit
	} else {
		path += "limit=10"
	}

	var results []fmpIncomeStatement
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp income statement %s: %w", symbol, err)
	}

	periodType := "annual"
	if period == "quarterly" {
		periodType = "quarterly"
	}

	stmts := make([]models.IncomeStatement, 0, len(results))
	for _, r := range results {
		is := models.IncomeStatement{
			Period:          r.Date,
			PeriodType:      periodType,
			Revenue:         r.Revenue,
			TotalIncome:     r.Revenue,
			TotalExpenses:   r.OperatingExpenses,
			EBITDA:          r.EBITDA,
			Depreciation:    r.DepreciationAndAmortization,
			EBIT:            r.OperatingIncome,
			InterestExpense: r.InterestExpense,
			PBT:             r.IncomeBeforeTax,
			Tax:             r.IncomeTaxExpense,
			PAT:             r.NetIncome,
			EPS:             r.EPSDiluted,
			OPMPct:          r.OperatingIncomeRatio * 100,
			NPMPct:          r.NetIncomeRatio * 100,
		}
		stmts = append(stmts, is)
	}

	f.CacheSetTTL(cacheKey, stmts, 1*time.Hour)
	return newResult(stmts), nil
}

// --- BalanceSheet fetcher ---

type balanceSheetFetcher struct {
	provider.BaseFetcher
}

func newBalanceSheetFetcher() *balanceSheetFetcher {
	return &balanceSheetFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelBalanceSheet,
			"Balance sheet from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamPeriod, provider.ParamLimit},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *balanceSheetFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	period := params[provider.ParamPeriod]
	path := fmt.Sprintf("/balance-sheet-statement/%s?", symbol)
	if period == "quarterly" {
		path += "period=quarter&"
	}
	if limit := params[provider.ParamLimit]; limit != "" {
		path += "limit=" + limit
	} else {
		path += "limit=10"
	}

	var results []fmpBalanceSheet
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp balance sheet %s: %w", symbol, err)
	}

	periodType := "annual"
	if period == "quarterly" {
		periodType = "quarterly"
	}

	sheets := make([]models.BalanceSheet, 0, len(results))
	for _, r := range results {
		bs := models.BalanceSheet{
			Period:              r.Date,
			PeriodType:          periodType,
			TotalAssets:         r.TotalAssets,
			CurrentAssets:       r.TotalCurrentAssets,
			CashEquivalents:     r.CashAndCashEquivalents,
			Inventory:           r.Inventory,
			TradeReceivables:    r.NetReceivables,
			FixedAssets:         r.PropertyPlantEquipmentNet,
			Investments:         r.LongTermInvestments,
			TotalLiabilities:    r.TotalLiabilities,
			CurrentLiabilities:  r.TotalCurrentLiabilities,
			LongTermBorrowings:  r.LongTermDebt,
			ShortTermBorrowings: r.ShortTermDebt,
			TotalDebt:           r.LongTermDebt + r.ShortTermDebt,
			TotalEquity:         r.TotalStockholdersEquity,
			ShareCapital:        r.CommonStock,
			Reserves:            r.RetainedEarnings,
		}
		sheets = append(sheets, bs)
	}

	f.CacheSetTTL(cacheKey, sheets, 1*time.Hour)
	return newResult(sheets), nil
}

// --- CashFlowStatement fetcher ---

type cashFlowStatementFetcher struct {
	provider.BaseFetcher
}

func newCashFlowStatementFetcher() *cashFlowStatementFetcher {
	return &cashFlowStatementFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCashFlowStatement,
			"Cash flow statement from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamPeriod, provider.ParamLimit},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *cashFlowStatementFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	period := params[provider.ParamPeriod]
	path := fmt.Sprintf("/cash-flow-statement/%s?", symbol)
	if period == "quarterly" {
		path += "period=quarter&"
	}
	if limit := params[provider.ParamLimit]; limit != "" {
		path += "limit=" + limit
	} else {
		path += "limit=10"
	}

	var results []fmpCashFlow
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp cash flow %s: %w", symbol, err)
	}

	periodType := "annual"
	if period == "quarterly" {
		periodType = "quarterly"
	}

	cfs := make([]models.CashFlow, 0, len(results))
	for _, r := range results {
		cf := models.CashFlow{
			Period:            r.Date,
			PeriodType:        periodType,
			OperatingCashFlow: r.OperatingCashFlow,
			InvestingCashFlow: r.InvestingActivitiesCF,
			FinancingCashFlow: r.FinancingActivitiesCF,
			NetCashFlow:       r.NetChangeInCash,
			FreeCashFlow:      r.FreeCashFlow,
			CapEx:             r.CapitalExpenditure,
			DividendsPaid:     r.DividendsPaid,
		}
		cfs = append(cfs, cf)
	}

	f.CacheSetTTL(cacheKey, cfs, 1*time.Hour)
	return newResult(cfs), nil
}

// --- KeyMetrics fetcher ---

type keyMetricsFetcher struct {
	provider.BaseFetcher
}

func newKeyMetricsFetcher() *keyMetricsFetcher {
	return &keyMetricsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelKeyMetrics,
			"Key financial metrics from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamPeriod, provider.ParamLimit},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *keyMetricsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/key-metrics/%s?limit=1", symbol)
	var results []fmpKeyMetrics
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp key metrics %s: %w", symbol, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no key metrics for %s", symbol)
	}

	r := results[0]
	ratios := &models.FinancialRatios{
		PE:               r.PERatio,
		PB:               r.PBRatio,
		PEGRatio:         r.PEGRatio,
		EVBITDA:          r.EVToEBITDA,
		ROE:              r.ROE * 100,
		DebtEquity:       r.DebtToEquity,
		CurrentRatio:     r.CurrentRatio,
		InterestCoverage: r.InterestCoverage,
		DividendYield:    r.DividendYield * 100,
		EPS:              r.NetIncomePerShare,
		BookValue:        r.BookValuePerShare,
		GrahamNumber:     r.GrahamNumber,
	}

	f.CacheSetTTL(cacheKey, ratios, 1*time.Hour)
	return newResult(ratios), nil
}

// --- FinancialRatios fetcher ---

type financialRatiosFetcher struct {
	provider.BaseFetcher
}

func newFinancialRatiosFetcher() *financialRatiosFetcher {
	return &financialRatiosFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelFinancialRatios,
			"Financial ratios from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamPeriod, provider.ParamLimit},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *financialRatiosFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/ratios/%s?limit=1", symbol)
	var results []fmpRatios
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp ratios %s: %w", symbol, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no ratios for %s", symbol)
	}

	r := results[0]
	ratios := &models.FinancialRatios{
		PE:               r.PriceEarningsRatio,
		PB:               r.PriceBookRatio,
		PEGRatio:         r.PEGRatio,
		EVBITDA:          r.EVToEBITDA,
		ROE:              r.ROE * 100,
		ROCE:             r.ROIC * 100,
		DebtEquity:       r.DebtEquityRatio,
		CurrentRatio:     r.CurrentRatio,
		InterestCoverage: r.InterestCoverage,
		DividendYield:    r.DividendYield * 100,
	}

	f.CacheSetTTL(cacheKey, ratios, 1*time.Hour)
	return newResult(ratios), nil
}

// --- KeyExecutives fetcher ---

type keyExecutivesFetcher struct {
	provider.BaseFetcher
}

func newKeyExecutivesFetcher() *keyExecutivesFetcher {
	return &keyExecutivesFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelKeyExecutives,
			"Key executives from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *keyExecutivesFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/key-executives/%s", symbol)
	var results []fmpKeyExecutive
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp executives %s: %w", symbol, err)
	}

	execs := make([]models.KeyExecutive, 0, len(results))
	for _, r := range results {
		execs = append(execs, models.KeyExecutive{
			Name:        r.Name,
			Title:       r.Title,
			Pay:         r.Pay,
			CurrencyPay: r.CurrencyPay,
			Gender:      r.Gender,
			YearBorn:    r.YearBorn,
		})
	}

	f.CacheSetTTL(cacheKey, execs, 1*time.Hour)
	return newResult(execs), nil
}

// --- HistoricalDividends fetcher ---

type historicalDividendsFetcher struct {
	provider.BaseFetcher
}

func newHistoricalDividendsFetcher() *historicalDividendsFetcher {
	return &historicalDividendsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelHistoricalDividends,
			"Historical dividends from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *historicalDividendsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/historical-price-full/stock_dividend/%s", symbol)
	var resp fmpHistoricalDividend
	if err := fetchFMPJSON(ctx, path, apiKey, &resp); err != nil {
		return nil, fmt.Errorf("fmp dividends %s: %w", symbol, err)
	}

	divs := make([]models.DividendRecord, 0, len(resp.Historical))
	for _, d := range resp.Historical {
		exDate, _ := time.Parse("2006-01-02", d.Date)
		payDate, _ := time.Parse("2006-01-02", d.PaymentDate)
		declDate, _ := time.Parse("2006-01-02", d.DeclarationDate)
		recDate, _ := time.Parse("2006-01-02", d.RecordDate)
		divs = append(divs, models.DividendRecord{
			Symbol:          symbol,
			ExDate:          exDate,
			PaymentDate:     payDate,
			DeclarationDate: declDate,
			RecordDate:      recDate,
			Amount:          d.Dividend,
			AdjDividend:     d.AdjDividend,
		})
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
			"Share float statistics from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *shareStatisticsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/shares_float?symbol=%s", symbol)
	var results []fmpShareFloat
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp share float %s: %w", symbol, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no share float for %s", symbol)
	}

	r := results[0]
	stats := &models.ShareStatisticsData{
		Symbol:            symbol,
		SharesOutstanding: int64(r.OutstandingShares),
		FloatShares:       int64(r.FloatShares),
	}

	f.CacheSetTTL(cacheKey, stats, 1*time.Hour)
	return newResult(stats), nil
}
