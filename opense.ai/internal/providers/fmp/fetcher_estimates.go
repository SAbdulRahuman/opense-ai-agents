package fmp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// === Estimates Fetchers ===

// --- PriceTarget fetcher ---

type priceTargetFetcher struct {
	provider.BaseFetcher
}

func newPriceTargetFetcher() *priceTargetFetcher {
	return &priceTargetFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelPriceTarget,
			"Analyst price targets from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamLimit},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *priceTargetFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	limit := params[provider.ParamLimit]
	if limit == "" {
		limit = "20"
	}
	path := fmt.Sprintf("/price-target/%s?limit=%s", symbol, limit)

	var results []fmpPriceTarget
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp price target %s: %w", symbol, err)
	}

	targets := make([]models.PriceTargetData, 0, len(results))
	for _, r := range results {
		pubDate, _ := time.Parse("2006-01-02T15:04:05.000Z", r.PublishedDate)
		if pubDate.IsZero() {
			pubDate, _ = time.Parse("2006-01-02", r.PublishedDate)
		}
		targets = append(targets, models.PriceTargetData{
			Symbol:         symbol,
			PublishedDate:  pubDate,
			AnalystName:    r.AnalystName,
			AnalystCompany: r.AnalystCompany,
			Rating:         r.NewBaseFormula,
			PriceTarget:    r.PriceTarget,
			AdjPriceTarget: r.AdjPriceTarget,
		})
	}

	f.CacheSetTTL(cacheKey, targets, 1*time.Hour)
	return newResult(targets), nil
}

// --- PriceTargetConsensus fetcher ---

type priceTargetConsensusFetcher struct {
	provider.BaseFetcher
}

func newPriceTargetConsensusFetcher() *priceTargetConsensusFetcher {
	return &priceTargetConsensusFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelPriceTargetConsensus,
			"Price target consensus from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *priceTargetConsensusFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/price-target-consensus/%s", symbol)
	var results []fmpPriceTargetConsensus
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp price target consensus %s: %w", symbol, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no price target consensus for %s", symbol)
	}

	r := results[0]
	consensus := &models.PriceTargetConsensusData{
		Symbol:  symbol,
		High:    r.TargetHigh,
		Low:     r.TargetLow,
		Median:  r.TargetMedian,
		Average: r.TargetConsensus,
	}

	f.CacheSetTTL(cacheKey, consensus, 1*time.Hour)
	return newResult(consensus), nil
}

// --- AnalystEstimates fetcher ---

type analystEstimatesFetcher struct {
	provider.BaseFetcher
}

func newAnalystEstimatesFetcher() *analystEstimatesFetcher {
	return &analystEstimatesFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelAnalystEstimates,
			"Analyst financial estimates from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamPeriod, provider.ParamLimit},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *analystEstimatesFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/analyst-estimates/%s?", symbol)
	if params[provider.ParamPeriod] == "quarterly" {
		path += "period=quarter&"
	}
	limit := params[provider.ParamLimit]
	if limit == "" {
		limit = "10"
	}
	path += "limit=" + limit

	var results []fmpAnalystEstimate
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp analyst estimates %s: %w", symbol, err)
	}

	estimates := make([]models.AnalystEstimate, 0, len(results))
	for _, r := range results {
		estimates = append(estimates, models.AnalystEstimate{
			Symbol:                symbol,
			Date:                  r.Date,
			EstimatedRevenueAvg:   r.EstimatedRevenueAvg,
			EstimatedRevenueHigh:  r.EstimatedRevenueHigh,
			EstimatedRevenueLow:   r.EstimatedRevenueLow,
			EstimatedEBITDAAvg:    r.EstimatedEbitdaAvg,
			EstimatedEBITDAHigh:   r.EstimatedEbitdaHigh,
			EstimatedEBITDALow:    r.EstimatedEbitdaLow,
			EstimatedEPSAvg:       r.EstimatedEpsAvg,
			EstimatedEPSHigh:      r.EstimatedEpsHigh,
			EstimatedEPSLow:       r.EstimatedEpsLow,
			EstimatedNetIncomeAvg: r.EstimatedNetIncomeAvg,
			NumberOfAnalysts:      r.NumberAnalystsEstimated,
		})
	}

	f.CacheSetTTL(cacheKey, estimates, 1*time.Hour)
	return newResult(estimates), nil
}

// === Calendar Fetchers ===

// --- CalendarEarnings fetcher ---

type calendarEarningsFetcher struct {
	provider.BaseFetcher
}

func newCalendarEarningsFetcher() *calendarEarningsFetcher {
	return &calendarEarningsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCalendarEarnings,
			"Earnings calendar from Financial Modeling Prep",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
			30*time.Minute, 5, time.Second,
		),
	}
}

func (f *calendarEarningsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	start, end := defaultDateRange(params)
	path := fmt.Sprintf("/earning_calendar?from=%s&to=%s", start, end)

	var results []fmpEarningsCalendar
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp earnings calendar: %w", err)
	}

	entries := make([]models.EarningsCalendarEntry, 0, len(results))
	for _, r := range results {
		rd, _ := time.Parse("2006-01-02", r.Date)
		entries = append(entries, models.EarningsCalendarEntry{
			Symbol:          r.Symbol,
			ReportDate:      rd,
			EPSEstimate:     r.EPSEstimated,
			EPSActual:       r.EPS,
			RevenueEstimate: r.RevenueEstimated,
			RevenueActual:   r.Revenue,
		})
	}

	f.CacheSetTTL(cacheKey, entries, 30*time.Minute)
	return newResult(entries), nil
}

// --- CalendarDividend fetcher ---

type calendarDividendFetcher struct {
	provider.BaseFetcher
}

func newCalendarDividendFetcher() *calendarDividendFetcher {
	return &calendarDividendFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCalendarDividend,
			"Dividend calendar from Financial Modeling Prep",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
			30*time.Minute, 5, time.Second,
		),
	}
}

func (f *calendarDividendFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	start, end := defaultDateRange(params)
	path := fmt.Sprintf("/stock_dividend_calendar?from=%s&to=%s", start, end)

	var results []fmpDividendCalendar
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp dividend calendar: %w", err)
	}

	entries := make([]models.DividendCalendarEntry, 0, len(results))
	for _, r := range results {
		exDate, _ := time.Parse("2006-01-02", r.Date)
		payDate, _ := time.Parse("2006-01-02", r.PaymentDate)
		recDate, _ := time.Parse("2006-01-02", r.RecordDate)
		entries = append(entries, models.DividendCalendarEntry{
			Symbol:         r.Symbol,
			ExDividendDate: exDate,
			PaymentDate:    payDate,
			RecordDate:     recDate,
			Amount:         r.Dividend,
		})
	}

	f.CacheSetTTL(cacheKey, entries, 30*time.Minute)
	return newResult(entries), nil
}

// --- CalendarIpo fetcher ---

type calendarIpoFetcher struct {
	provider.BaseFetcher
}

func newCalendarIpoFetcher() *calendarIpoFetcher {
	return &calendarIpoFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCalendarIpo,
			"IPO calendar from Financial Modeling Prep",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
			30*time.Minute, 5, time.Second,
		),
	}
}

func (f *calendarIpoFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	start, end := defaultDateRange(params)
	path := fmt.Sprintf("/ipo_calendar?from=%s&to=%s", start, end)

	var results []fmpIPOCalendar
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp ipo calendar: %w", err)
	}

	entries := make([]models.IPOCalendarEntry, 0, len(results))
	for _, r := range results {
		ipoDate, _ := time.Parse("2006-01-02", r.Date)
		var priceLow, priceHigh float64
		if r.PriceRange != "" {
			fmt.Sscanf(r.PriceRange, "%f-%f", &priceLow, &priceHigh)
		}
		entries = append(entries, models.IPOCalendarEntry{
			Symbol:         r.Symbol,
			Name:           r.Company,
			IPODate:        ipoDate,
			PriceRangeLow:  priceLow,
			PriceRangeHigh: priceHigh,
			Shares:         r.Shares,
			Exchange:       r.Exchange,
		})
	}

	f.CacheSetTTL(cacheKey, entries, 30*time.Minute)
	return newResult(entries), nil
}

// === Discovery Fetchers ===

// --- EquityGainers fetcher ---

type equityGainersFetcher struct {
	provider.BaseFetcher
}

func newEquityGainersFetcher() *equityGainersFetcher {
	return &equityGainersFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityGainers,
			"Top gainers from Financial Modeling Prep",
			nil, nil,
			5*time.Minute, 5, time.Second,
		),
	}
}

func (f *equityGainersFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchFMPMovers(ctx, &f.BaseFetcher, "/gainers", params)
}

// --- EquityLosers fetcher ---

type equityLosersFetcher struct {
	provider.BaseFetcher
}

func newEquityLosersFetcher() *equityLosersFetcher {
	return &equityLosersFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityLosers,
			"Top losers from Financial Modeling Prep",
			nil, nil,
			5*time.Minute, 5, time.Second,
		),
	}
}

func (f *equityLosersFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchFMPMovers(ctx, &f.BaseFetcher, "/losers", params)
}

// --- EquityActive fetcher ---

type equityActiveFetcher struct {
	provider.BaseFetcher
}

func newEquityActiveFetcher() *equityActiveFetcher {
	return &equityActiveFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityActive,
			"Most active stocks from Financial Modeling Prep",
			nil, nil,
			5*time.Minute, 5, time.Second,
		),
	}
}

func (f *equityActiveFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchFMPMovers(ctx, &f.BaseFetcher, "/actives", params)
}

// fetchFMPMovers is a shared helper for gainers/losers/active fetchers.
func fetchFMPMovers(ctx context.Context, f *provider.BaseFetcher, endpoint string, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	var results []fmpGainer
	if err := fetchFMPJSON(ctx, endpoint, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp movers %s: %w", endpoint, err)
	}

	movers := make([]models.MarketMover, 0, len(results))
	for _, r := range results {
		movers = append(movers, models.MarketMover{
			Symbol:    r.Symbol,
			Name:      r.Name,
			Price:     r.Price,
			Change:    r.Change,
			ChangePct: r.ChangesPercentage,
		})
	}

	f.CacheSetTTL(cacheKey, movers, 5*time.Minute)
	return newResult(movers), nil
}

// === ETF / Index / Crypto / Currency Fetchers ===

// --- ETF Historical ---

type etfHistoricalFetcher struct {
	provider.BaseFetcher
}

func newEtfHistoricalFetcher() *etfHistoricalFetcher {
	return &etfHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEtfHistorical,
			"ETF historical prices from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *etfHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchFMPHistorical(ctx, &f.BaseFetcher, params)
}

// --- ETF Info ---

type etfInfoFetcher struct {
	provider.BaseFetcher
}

func newEtfInfoFetcher() *etfInfoFetcher {
	return &etfInfoFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEtfInfo,
			"ETF profile/info from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *etfInfoFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/profile/%s", symbol)
	var profiles []fmpProfile
	if err := fetchFMPJSON(ctx, path, apiKey, &profiles); err != nil {
		return nil, fmt.Errorf("fmp etf info %s: %w", symbol, err)
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profile for %s", symbol)
	}

	p := profiles[0]
	info := &models.ETFInfo{
		Symbol:      p.Symbol,
		Name:        p.CompanyName,
		Exchange:    p.Exchange,
		Description: p.Description,
	}

	f.CacheSetTTL(cacheKey, info, 1*time.Hour)
	return newResult(info), nil
}

// --- Index Historical ---

type indexHistoricalFetcher struct {
	provider.BaseFetcher
}

func newIndexHistoricalFetcher() *indexHistoricalFetcher {
	return &indexHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelIndexHistorical,
			"Index historical prices from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *indexHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchFMPHistorical(ctx, &f.BaseFetcher, params)
}

// --- Crypto Historical ---

type cryptoHistoricalFetcher struct {
	provider.BaseFetcher
}

func newCryptoHistoricalFetcher() *cryptoHistoricalFetcher {
	return &cryptoHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCryptoHistorical,
			"Crypto historical prices from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *cryptoHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchFMPHistorical(ctx, &f.BaseFetcher, params)
}

// --- Currency Historical ---

type currencyHistoricalFetcher struct {
	provider.BaseFetcher
}

func newCurrencyHistoricalFetcher() *currencyHistoricalFetcher {
	return &currencyHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCurrencyHistorical,
			"Currency pair historical prices from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *currencyHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchFMPHistorical(ctx, &f.BaseFetcher, params)
}

// fetchFMPHistorical is a shared helper for historical price fetchers (equity/etf/index/crypto/currency).
func fetchFMPHistorical(ctx context.Context, f *provider.BaseFetcher, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	start, end := defaultDateRange(params)
	path := fmt.Sprintf("/historical-price-full/%s?from=%s&to=%s", symbol, start, end)

	var resp fmpHistoricalPrice
	if err := fetchFMPJSON(ctx, path, apiKey, &resp); err != nil {
		return nil, fmt.Errorf("fmp historical %s: %w", symbol, err)
	}

	candles := make([]models.OHLCV, 0, len(resp.Historical))
	for _, h := range resp.Historical {
		t, _ := time.Parse("2006-01-02", h.Date)
		candles = append(candles, models.OHLCV{
			Timestamp: t,
			Open:      h.Open,
			High:      h.High,
			Low:       h.Low,
			Close:     h.Close,
			AdjClose:  h.AdjClose,
			Volume:    h.Volume,
		})
	}

	f.CacheSetTTL(cacheKey, candles, 1*time.Hour)
	return newResult(candles), nil
}

// === News Fetchers ===

// --- CompanyNews fetcher ---

type companyNewsFetcher struct {
	provider.BaseFetcher
}

func newCompanyNewsFetcher() *companyNewsFetcher {
	return &companyNewsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCompanyNews,
			"Company-specific news from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamLimit},
			15*time.Minute, 5, time.Second,
		),
	}
}

func (f *companyNewsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	limit := params[provider.ParamLimit]
	if limit == "" {
		limit = "20"
	}
	path := fmt.Sprintf("/stock_news?tickers=%s&limit=%s", symbol, limit)

	var results []fmpNewsArticle
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp company news %s: %w", symbol, err)
	}

	articles := make([]models.CompanyNewsArticle, 0, len(results))
	for _, r := range results {
		pubDate, _ := time.Parse("2006-01-02 15:04:05", r.PublishedDate)
		articles = append(articles, models.CompanyNewsArticle{
			Symbol:      r.Symbol,
			Title:       r.Title,
			URL:         r.URL,
			Source:      r.Site,
			Summary:     r.Text,
			ImageURL:    r.Image,
			PublishedAt: pubDate,
		})
	}

	f.CacheSetTTL(cacheKey, articles, 15*time.Minute)
	return newResult(articles), nil
}

// --- WorldNews fetcher ---

type worldNewsFetcher struct {
	provider.BaseFetcher
}

func newWorldNewsFetcher() *worldNewsFetcher {
	return &worldNewsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelWorldNews,
			"World/market news from Financial Modeling Prep",
			nil,
			[]string{provider.ParamLimit},
			15*time.Minute, 5, time.Second,
		),
	}
}

func (f *worldNewsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	limit := params[provider.ParamLimit]
	if limit == "" {
		limit = "30"
	}
	path := fmt.Sprintf("/stock_news?limit=%s", limit)

	var results []fmpNewsArticle
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp world news: %w", err)
	}

	articles := make([]models.WorldNewsArticle, 0, len(results))
	for _, r := range results {
		pubDate, _ := time.Parse("2006-01-02 15:04:05", r.PublishedDate)
		var tickers []string
		if r.Symbol != "" {
			tickers = strings.Split(r.Symbol, ",")
		}
		articles = append(articles, models.WorldNewsArticle{
			Title:       r.Title,
			URL:         r.URL,
			Source:      r.Site,
			Summary:     r.Text,
			ImageURL:    r.Image,
			Tickers:     tickers,
			PublishedAt: pubDate,
		})
	}

	f.CacheSetTTL(cacheKey, articles, 15*time.Minute)
	return newResult(articles), nil
}
