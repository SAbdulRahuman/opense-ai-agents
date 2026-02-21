package fmp

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
			"Historical OHLCV from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate},
			15*time.Minute, 5, time.Second,
		),
	}
}

func (f *equityHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	startDate, endDate := defaultDateRange(params)
	path := fmt.Sprintf("/historical-price-full/%s?from=%s&to=%s", symbol, startDate, endDate)

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
			"Real-time quote from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Minute, 5, time.Second,
		),
	}
}

func (f *equityQuoteFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/quote/%s", symbol)
	var results []fmpQuote
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp quote %s: %w", symbol, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no quote for %s", symbol)
	}

	q := results[0]
	quote := &models.Quote{
		Ticker:    q.Symbol,
		Name:      q.Name,
		LastPrice: q.Price,
		Change:    q.Change,
		ChangePct: q.ChangesPercentage,
		Open:      q.Open,
		High:      q.DayHigh,
		Low:       q.DayLow,
		PrevClose: q.PreviousClose,
		Volume:    q.Volume,
		WeekHigh52: q.YearHigh,
		WeekLow52:  q.YearLow,
		MarketCap:  q.MarketCap,
		PE:         q.PE,
		Timestamp:  time.Unix(q.Timestamp, 0),
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
			"Company profile from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *equityInfoFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
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
		return nil, fmt.Errorf("fmp profile %s: %w", symbol, err)
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profile for %s", symbol)
	}

	p := profiles[0]
	profile := &models.StockProfile{
		Stock: models.Stock{
			Ticker:    p.Symbol,
			Name:      p.CompanyName,
			Exchange:  p.ExchangeShortName,
			Sector:    p.Sector,
			Industry:  p.Industry,
			ISIN:      p.ISIN,
			MarketCap: p.MktCap,
		},
		FetchedAt: time.Now(),
	}

	f.CacheSetTTL(cacheKey, profile, 1*time.Hour)
	return newResult(profile), nil
}

// --- EquitySearch fetcher ---

type equitySearchFetcher struct {
	provider.BaseFetcher
}

func newEquitySearchFetcher() *equitySearchFetcher {
	return &equitySearchFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquitySearch,
			"Search for equities on Financial Modeling Prep",
			[]string{provider.ParamQuery},
			[]string{provider.ParamLimit, provider.ParamExchange},
			5*time.Minute, 5, time.Second,
		),
	}
}

func (f *equitySearchFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	query := params[provider.ParamQuery]
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
	path := fmt.Sprintf("/search?query=%s&limit=%s", query, limit)
	if exchange := params[provider.ParamExchange]; exchange != "" {
		path += "&exchange=" + exchange
	}

	var results []fmpSearchResult
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp search %q: %w", query, err)
	}

	out := make([]models.EquitySearchResult, 0, len(results))
	for _, r := range results {
		out = append(out, models.EquitySearchResult{
			Symbol:   r.Symbol,
			Name:     r.Name,
			Exchange: r.ExchangeShortName,
		})
	}

	f.CacheSet(cacheKey, out)
	return newResult(out), nil
}

// --- EquityScreener fetcher ---

type equityScreenerFetcher struct {
	provider.BaseFetcher
}

func newEquityScreenerFetcher() *equityScreenerFetcher {
	return &equityScreenerFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityScreener,
			"Stock screener from Financial Modeling Prep",
			nil,
			[]string{provider.ParamExchange, provider.ParamLimit, "sector", "industry", "market_cap_gt", "market_cap_lt"},
			10*time.Minute, 3, time.Second,
		),
	}
}

func (f *equityScreenerFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := "/stock-screener?"
	if v := params[provider.ParamExchange]; v != "" {
		path += "exchange=" + v + "&"
	}
	if v := params["sector"]; v != "" {
		path += "sector=" + v + "&"
	}
	if v := params["industry"]; v != "" {
		path += "industry=" + v + "&"
	}
	if v := params["market_cap_gt"]; v != "" {
		path += "marketCapMoreThan=" + v + "&"
	}
	if v := params["market_cap_lt"]; v != "" {
		path += "marketCapLowerThan=" + v + "&"
	}
	limit := params[provider.ParamLimit]
	if limit == "" {
		limit = "50"
	}
	path += "limit=" + limit

	var results []fmpScreenerResult
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp screener: %w", err)
	}

	out := make([]models.ScreenerResult, 0, len(results))
	for _, r := range results {
		out = append(out, models.ScreenerResult{
			Symbol:    r.Symbol,
			Name:      r.CompanyName,
			Exchange:  r.ExchangeShortName,
			Sector:    r.Sector,
			Industry:  r.Industry,
			MarketCap: r.MarketCap,
			Price:     r.Price,
			Volume:    r.Volume,
			Beta:      r.Beta,
		})
	}

	f.CacheSetTTL(cacheKey, out, 10*time.Minute)
	return newResult(out), nil
}

// --- EquityPeers fetcher ---

type equityPeersFetcher struct {
	provider.BaseFetcher
}

func newEquityPeersFetcher() *equityPeersFetcher {
	return &equityPeersFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquityPeers,
			"Stock peers from Financial Modeling Prep",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *equityPeersFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/stock_peers?symbol=%s", symbol)
	var results []fmpPeerResult
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp peers %s: %w", symbol, err)
	}

	peers := make([]models.EquityPeer, 0)
	if len(results) > 0 {
		for _, p := range results[0].PeersList {
			peers = append(peers, models.EquityPeer{Symbol: p})
		}
	}

	f.CacheSetTTL(cacheKey, peers, 1*time.Hour)
	return newResult(peers), nil
}

// --- PricePerformance fetcher ---

type pricePerformanceFetcher struct {
	provider.BaseFetcher
}

func newPricePerformanceFetcher() *pricePerformanceFetcher {
	return &pricePerformanceFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelPricePerformance,
			"Price performance over various periods from FMP",
			[]string{provider.ParamSymbol},
			nil,
			15*time.Minute, 5, time.Second,
		),
	}
}

func (f *pricePerformanceFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/stock-price-change/%s", symbol)
	var results []fmpPricePerformance
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp price perf %s: %w", symbol, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no price performance for %s", symbol)
	}

	r := results[0]
	perf := &models.PricePerformanceData{
		Symbol:     r.Symbol,
		OneDay:     r.OneDay,
		OneMonth:   r.OneMonth,
		ThreeMonth: r.ThreeMonth,
		SixMonth:   r.SixMonth,
		YTD:        r.YTD,
		OneYear:    r.OneYear,
		ThreeYear:  r.ThreeYear,
		FiveYear:   r.FiveYear,
		TenYear:    r.TenYear,
		MaxReturn:  r.Max,
	}

	f.CacheSetTTL(cacheKey, perf, 15*time.Minute)
	return newResult(perf), nil
}

// --- MarketSnapshots fetcher ---

type marketSnapshotsFetcher struct {
	provider.BaseFetcher
}

func newMarketSnapshotsFetcher() *marketSnapshotsFetcher {
	return &marketSnapshotsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelMarketSnapshots,
			"Market snapshots (all quotes) from FMP",
			nil,
			[]string{provider.ParamExchange},
			5*time.Minute, 2, time.Second,
		),
	}
}

func (f *marketSnapshotsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fmp_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	exchange := params[provider.ParamExchange]
	path := "/quotes/nyse"
	if exchange != "" {
		path = "/quotes/" + exchange
	}

	var results []fmpQuote
	if err := fetchFMPJSON(ctx, path, apiKey, &results); err != nil {
		return nil, fmt.Errorf("fmp snapshots: %w", err)
	}

	quotes := make([]models.Quote, 0, len(results))
	for _, q := range results {
		quotes = append(quotes, models.Quote{
			Ticker:    q.Symbol,
			Name:      q.Name,
			LastPrice: q.Price,
			Change:    q.Change,
			ChangePct: q.ChangesPercentage,
			Open:      q.Open,
			High:      q.DayHigh,
			Low:       q.DayLow,
			PrevClose: q.PreviousClose,
			Volume:    q.Volume,
			MarketCap: q.MarketCap,
		})
	}

	f.CacheSetTTL(cacheKey, quotes, 5*time.Minute)
	return newResult(quotes), nil
}
