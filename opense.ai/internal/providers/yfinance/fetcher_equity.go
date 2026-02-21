package yfinance

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// --- EquityHistorical fetcher ---

type equityHistoricalFetcher struct {
	provider.BaseFetcher
}

func newEquityHistoricalFetcher() *equityHistoricalFetcher {
	return &equityHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityHistorical,
			"Historical OHLCV price data from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamInterval},
			15*time.Minute, 5, time.Second,
		),
	}
}

func (f *equityHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	// Parse date range.
	startDate, endDate := defaultDateRange(params)

	interval := params[provider.ParamInterval]
	if interval == "" {
		interval = "1d"
	}

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}

	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=%s",
		yfTicker, startDate.Unix(), endDate.Unix(), interval,
	)

	var resp yfChartResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance chart %s: %w", yfTicker, err)
	}

	if resp.Chart.Error != nil {
		return nil, fmt.Errorf("yfinance chart error: %s", resp.Chart.Error.Description)
	}
	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data for %s", symbol)
	}

	candles := parseCandles(resp.Chart.Result[0])
	f.CacheSetTTL(cacheKey, candles, 15*time.Minute)
	return newResult(candles), nil
}

// --- EquityQuote fetcher ---

type equityQuoteFetcher struct {
	provider.BaseFetcher
}

func newEquityQuoteFetcher() *equityQuoteFetcher {
	return &equityQuoteFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityQuote,
			"Real-time stock quote from Yahoo Finance",
			[]string{provider.ParamSymbol},
			nil,
			5*time.Minute, 5, time.Second,
		),
	}
}

func (f *equityQuoteFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}

	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s", yfTicker)

	var resp yfQuoteResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance quote %s: %w", yfTicker, err)
	}

	if resp.QuoteResponse.Error != nil {
		return nil, fmt.Errorf("yfinance API error: %s", resp.QuoteResponse.Error.Description)
	}
	if len(resp.QuoteResponse.Result) == 0 {
		return nil, fmt.Errorf("no quote for %s", symbol)
	}

	r := resp.QuoteResponse.Result[0]
	quote := &models.Quote{
		Ticker:        fromYFTicker(r.Symbol),
		Name:          coalesce(r.LongName, r.ShortName),
		LastPrice:     r.RegularMarketPrice,
		Change:        r.RegularMarketChange,
		ChangePct:     r.RegularMarketChangePercent,
		Open:          r.RegularMarketOpen,
		High:          r.RegularMarketDayHigh,
		Low:           r.RegularMarketDayLow,
		PrevClose:     r.RegularMarketPreviousClose,
		Volume:        r.RegularMarketVolume,
		WeekHigh52:    r.FiftyTwoWeekHigh,
		WeekLow52:     r.FiftyTwoWeekLow,
		MarketCap:     r.MarketCap,
		PE:            r.TrailingPE,
		PB:            r.PriceToBook,
		DividendYield: r.DividendYield * 100,
		Timestamp:     time.Unix(r.RegularMarketTime, 0),
	}

	f.CacheSet(cacheKey, quote)
	return newResult(quote), nil
}

// --- EquityInfo fetcher ---

type equityInfoFetcher struct {
	provider.BaseFetcher
}

func newEquityInfoFetcher() *equityInfoFetcher {
	return &equityInfoFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityInfo,
			"Company profile and summary info from Yahoo Finance",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *equityInfoFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}

	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	modules := "assetProfile,defaultKeyStatistics,summaryDetail,financialData"
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=%s",
		yfTicker, modules,
	)

	var resp yfQuoteSummaryResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance info %s: %w", yfTicker, err)
	}

	if resp.QuoteSummary.Error != nil {
		return nil, fmt.Errorf("yfinance API error: %s", resp.QuoteSummary.Error.Description)
	}
	if len(resp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no info for %s", symbol)
	}

	r := resp.QuoteSummary.Result[0]
	info := buildEquityInfo(symbol, yfTicker, r)

	f.CacheSetTTL(cacheKey, info, 1*time.Hour)
	return newResult(info), nil
}

// buildEquityInfo assembles a StockProfile from quoteSummary response.
func buildEquityInfo(symbol, yfTicker string, r yfQuoteSummaryResult) *models.StockProfile {
	profile := &models.StockProfile{
		Stock: models.Stock{
			Ticker:    fromYFTicker(yfTicker),
			NSETicker: yfTicker,
			Name:      symbol,
		},
		FetchedAt: time.Now(),
	}

	if r.AssetProfile != nil {
		ap := r.AssetProfile
		profile.Stock.Sector = ap.Sector
		profile.Stock.Industry = ap.Industry
	}

	if r.SummaryDetail != nil {
		sd := r.SummaryDetail
		profile.Stock.MarketCap = sd.MarketCap.Raw
	}

	if r.DefaultKeyStatistics != nil || r.FinancialData != nil {
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
			ratios.DebtEquity = fd.DebtToEquity.Raw
			ratios.CurrentRatio = fd.CurrentRatio.Raw
			ratios.DividendYield = fd.ProfitMargins.Raw * 100 // placeholder
		}
		if sd := r.SummaryDetail; sd != nil {
			ratios.PE = sd.TrailingPE.Raw
			ratios.DividendYield = sd.DividendYield.Raw * 100
		}
		profile.Ratios = ratios
	}

	return profile
}

// --- EquitySearch fetcher ---

type equitySearchFetcher struct {
	provider.BaseFetcher
}

func newEquitySearchFetcher() *equitySearchFetcher {
	return &equitySearchFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquitySearch,
			"Search for equities by name or symbol on Yahoo Finance",
			[]string{provider.ParamQuery},
			[]string{provider.ParamLimit},
			5*time.Minute, 5, time.Second,
		),
	}
}

func (f *equitySearchFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	query := params[provider.ParamQuery]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}

	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v1/finance/search?q=%s&quotesCount=20&newsCount=0", query)

	var resp yfSearchResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance search %q: %w", query, err)
	}

	results := make([]models.EquitySearchResult, 0, len(resp.Quotes))
	for _, q := range resp.Quotes {
		if !q.IsYahooFinance {
			continue
		}
		results = append(results, models.EquitySearchResult{
			Symbol:   q.Symbol,
			Name:     coalesce(q.LongName, q.ShortName),
			Exchange: q.Exchange,
			Sector:   q.Sector,
			Industry: q.Industry,
			IsETF:    q.QuoteType == "ETF",
		})
	}

	f.CacheSet(cacheKey, results)
	return newResult(results), nil
}

// --- Helpers ---

// parseCandles converts YF chart data to OHLCV slices.
func parseCandles(result yfChartResult) []models.OHLCV {
	if len(result.Indicators.Quote) == 0 {
		return nil
	}

	q := result.Indicators.Quote[0]
	var adjCloses []*float64
	if len(result.Indicators.AdjClose) > 0 {
		adjCloses = result.Indicators.AdjClose[0].AdjClose
	}

	candles := make([]models.OHLCV, 0, len(result.Timestamp))
	for i, ts := range result.Timestamp {
		c := models.OHLCV{
			Timestamp: time.Unix(ts, 0),
		}
		if i < len(q.Open) && q.Open[i] != nil {
			c.Open = *q.Open[i]
		}
		if i < len(q.High) && q.High[i] != nil {
			c.High = *q.High[i]
		}
		if i < len(q.Low) && q.Low[i] != nil {
			c.Low = *q.Low[i]
		}
		if i < len(q.Close) && q.Close[i] != nil {
			c.Close = *q.Close[i]
		}
		if i < len(q.Volume) && q.Volume[i] != nil {
			c.Volume = *q.Volume[i]
		}
		if i < len(adjCloses) && adjCloses[i] != nil {
			c.AdjClose = *adjCloses[i]
		}
		candles = append(candles, c)
	}
	return candles
}

// defaultDateRange parses start_date/end_date from params or uses defaults.
func defaultDateRange(params provider.QueryParams) (time.Time, time.Time) {
	now := time.Now()
	endDate := now
	startDate := now.AddDate(-1, 0, 0) // default: 1 year

	if s := params[provider.ParamStartDate]; s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = t
		}
	}
	if s := params[provider.ParamEndDate]; s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			endDate = t
		}
	}
	return startDate, endDate
}
