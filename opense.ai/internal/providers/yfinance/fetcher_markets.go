package yfinance

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// --- EtfHistorical fetcher ---

type etfHistoricalFetcher struct {
	provider.BaseFetcher
}

func newEtfHistoricalFetcher() *etfHistoricalFetcher {
	return &etfHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEtfHistorical,
			"Historical ETF OHLCV data from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamInterval},
			15*time.Minute, 5, time.Second,
		),
	}
}

func (f *etfHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchHistorical(ctx, f, params)
}

// --- EtfInfo fetcher ---

type etfInfoFetcher struct {
	provider.BaseFetcher
}

func newEtfInfoFetcher() *etfInfoFetcher {
	return &etfInfoFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEtfInfo,
			"ETF profile and holdings info from Yahoo Finance",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *etfInfoFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Get quote and summary data
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s", yfTicker)
	var qResp yfQuoteResponse
	if err := fetchJSON(ctx, url, &qResp); err != nil {
		return nil, fmt.Errorf("yfinance etf info %s: %w", yfTicker, err)
	}
	if qResp.QuoteResponse.Error != nil {
		return nil, fmt.Errorf("yfinance API error: %s", qResp.QuoteResponse.Error.Description)
	}
	if len(qResp.QuoteResponse.Result) == 0 {
		return nil, fmt.Errorf("no data for %s", symbol)
	}

	q := qResp.QuoteResponse.Result[0]
	info := &models.ETFInfo{
		Symbol:        q.Symbol,
		Name:          coalesce(q.LongName, q.ShortName),
		Exchange:      q.FullExchangeName,
		AUM:           q.TotalAssets,
		DividendYield: q.TrailingAnnualDividendYield * 100,
		PE:            q.TrailingPE,
		Beta:          q.Beta,
		YTDReturn:     q.YtdReturn * 100,
	}

	f.CacheSetTTL(cacheKey, info, 1*time.Hour)
	return newResult(info), nil
}

// --- IndexHistorical fetcher ---

type indexHistoricalFetcher struct {
	provider.BaseFetcher
}

func newIndexHistoricalFetcher() *indexHistoricalFetcher {
	return &indexHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelIndexHistorical,
			"Historical index data from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamInterval},
			15*time.Minute, 5, time.Second,
		),
	}
}

func (f *indexHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	return fetchHistorical(ctx, f, params)
}

// --- OptionsChains fetcher ---

type optionsChainsFetcher struct {
	provider.BaseFetcher
}

func newOptionsChainsFetcher() *optionsChainsFetcher {
	return &optionsChainsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelOptionsChains,
			"Options chain data from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamExpiry},
			5*time.Minute, 5, time.Second,
		),
	}
}

func (f *optionsChainsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	yfTicker := toYFTicker(symbol)

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/options/%s", yfTicker)
	if expiry := params[provider.ParamExpiry]; expiry != "" {
		if t, err := time.Parse("2006-01-02", expiry); err == nil {
			url += fmt.Sprintf("?date=%d", t.Unix())
		}
	}

	var resp yfOptionsResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("yfinance options %s: %w", yfTicker, err)
	}
	if resp.OptionChain.Error != nil {
		return nil, fmt.Errorf("yfinance options error: %s", resp.OptionChain.Error.Description)
	}
	if len(resp.OptionChain.Result) == 0 {
		return nil, fmt.Errorf("no options data for %s", symbol)
	}

	r := resp.OptionChain.Result[0]
	chain := &models.OptionChain{
		Ticker:    fromYFTicker(r.UnderlyingSymbol),
		SpotPrice: r.Quote.RegularMarketPrice,
		FetchedAt: time.Now(),
	}

	// Map expiration dates
	expiries := make([]string, 0, len(r.ExpirationDates))
	for _, ts := range r.ExpirationDates {
		expiries = append(expiries, time.Unix(ts, 0).Format("2006-01-02"))
	}
	chain.Expiries = expiries

	// Parse contracts
	var totalCEOI, totalPEOI int64
	for _, opt := range r.Options {
		chain.ExpiryDate = time.Unix(opt.ExpirationDate, 0).Format("2006-01-02")
		for _, c := range opt.Calls {
			contract := yfContractToModel(c, "CE")
			chain.Contracts = append(chain.Contracts, contract)
			totalCEOI += c.OpenInterest
		}
		for _, c := range opt.Puts {
			contract := yfContractToModel(c, "PE")
			chain.Contracts = append(chain.Contracts, contract)
			totalPEOI += c.OpenInterest
		}
	}
	chain.TotalCEOI = totalCEOI
	chain.TotalPEOI = totalPEOI
	if totalCEOI > 0 {
		chain.PCR = float64(totalPEOI) / float64(totalCEOI)
	}

	f.CacheSetTTL(cacheKey, chain, 5*time.Minute)
	return newResult(chain), nil
}

func yfContractToModel(c yfContract, optType string) models.OptionContract {
	return models.OptionContract{
		StrikePrice: c.Strike,
		OptionType:  optType,
		ExpiryDate:  time.Unix(c.Expiration, 0).Format("2006-01-02"),
		LTP:         c.LastPrice,
		Change:      c.Change,
		ChangePct:   c.PercentChange,
		Volume:      c.Volume,
		OI:          c.OpenInterest,
		BidPrice:    c.Bid,
		AskPrice:    c.Ask,
		IV:          c.ImpliedVolatility * 100,
	}
}

// --- CryptoHistorical fetcher ---

type cryptoHistoricalFetcher struct {
	provider.BaseFetcher
}

func newCryptoHistoricalFetcher() *cryptoHistoricalFetcher {
	return &cryptoHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCryptoHistorical,
			"Historical crypto OHLCV data from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamInterval},
			15*time.Minute, 5, time.Second,
		),
	}
}

func (f *cryptoHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	// Crypto tickers on YF use format like BTC-USD
	return fetchHistorical(ctx, f, params)
}

// --- CurrencyHistorical fetcher ---

type currencyHistoricalFetcher struct {
	provider.BaseFetcher
}

func newCurrencyHistoricalFetcher() *currencyHistoricalFetcher {
	return &currencyHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCurrencyHistorical,
			"Historical currency pair OHLCV data from Yahoo Finance",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamInterval},
			15*time.Minute, 5, time.Second,
		),
	}
}

func (f *currencyHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	// Currency pairs on YF use format like USDINR=X
	return fetchHistorical(ctx, f, params)
}

// --- Shared historical fetcher ---

type historicalFetcherBase interface {
	CacheGet(key string) (any, bool)
	CacheSetTTL(key string, value any, ttl time.Duration)
	RateLimit(ctx context.Context) error
	ModelType() provider.ModelType
}

func fetchHistorical(ctx context.Context, f historicalFetcherBase, params provider.QueryParams) (*provider.FetchResult, error) {
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
	interval := params[provider.ParamInterval]
	if interval == "" {
		interval = "1d"
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
