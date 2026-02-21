package cboe

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// EquityHistorical — OHLCV data from CBOE delayed charts.
// URL: https://cdn.cboe.com/api/global/delayed_quotes/charts/historical/{SYMBOL}.json
// ---------------------------------------------------------------------------

type equityHistoricalFetcher struct {
	provider.BaseFetcher
	model provider.ModelType
	prov  *Provider
}

func newEquityHistoricalFetcher(p *Provider) *equityHistoricalFetcher {
	return &equityHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelEquityHistorical,
			"CBOE equity historical OHLCV data (daily or intraday)",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamInterval},
		),
		model: provider.ModelEquityHistorical,
		prov:  p,
	}
}

func newEtfHistoricalFetcher(p *Provider) *equityHistoricalFetcher {
	return &equityHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelEtfHistorical,
			"CBOE ETF historical OHLCV data (same source as equity)",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamInterval},
		),
		model: provider.ModelEtfHistorical,
		prov:  p,
	}
}

func (f *equityHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	symbol := strings.ToUpper(params[provider.ParamSymbol])
	if symbol == "" {
		return nil, fmt.Errorf("cboe: %s is required", provider.ParamSymbol)
	}

	cacheKey := provider.CacheKey(f.model, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	// Ensure index directory is loaded for symbol path resolution.
	_, _ = f.prov.getIndexDirectory(ctx)

	interval := params[provider.ParamInterval]
	if interval == "" {
		interval = "1d"
	}

	url := chartURL(f.prov.symbolPath(symbol), interval)

	raw, err := fetchCBOERaw(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("cboe equity historical: %w", err)
	}

	var bars []models.OHLCV
	if interval == "1m" {
		bars, err = parseIntradayChart(raw)
	} else {
		bars, err = parseDailyChart(raw, params)
	}
	if err != nil {
		return nil, err
	}

	result := newResult(bars)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// parseDailyChart parses daily chart JSON into OHLCV slices.
func parseDailyChart(raw []byte, params provider.QueryParams) ([]models.OHLCV, error) {
	var resp struct {
		Symbol string `json:"symbol"`
		Data   []struct {
			Date        string  `json:"date"`
			Open        float64 `json:"open"`
			High        float64 `json:"high"`
			Low         float64 `json:"low"`
			Close       float64 `json:"close"`
			StockVolume int64   `json:"stock_volume"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("cboe: parse daily chart: %w", err)
	}

	startDate := params[provider.ParamStartDate]
	endDate := params[provider.ParamEndDate]

	var bars []models.OHLCV
	for _, d := range resp.Data {
		if startDate != "" && d.Date < startDate {
			continue
		}
		if endDate != "" && d.Date > endDate {
			continue
		}
		bars = append(bars, models.OHLCV{
			Timestamp: parseCBOEDate(d.Date),
			Open:      d.Open,
			High:      d.High,
			Low:       d.Low,
			Close:     d.Close,
			Volume:    d.StockVolume,
		})
	}
	return bars, nil
}

// parseIntradayChart parses intraday chart JSON into OHLCV slices.
func parseIntradayChart(raw []byte) ([]models.OHLCV, error) {
	var resp struct {
		Symbol string `json:"symbol"`
		Data   []struct {
			Datetime string `json:"datetime"`
			Price    struct {
				Open  float64 `json:"open"`
				High  float64 `json:"high"`
				Low   float64 `json:"low"`
				Close float64 `json:"close"`
			} `json:"price"`
			Volume struct {
				StockVolume int64 `json:"stock_volume"`
			} `json:"volume"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("cboe: parse intraday chart: %w", err)
	}

	var bars []models.OHLCV
	for _, d := range resp.Data {
		bars = append(bars, models.OHLCV{
			Timestamp: parseCBOETime(d.Datetime),
			Open:      d.Price.Open,
			High:      d.Price.High,
			Low:       d.Price.Low,
			Close:     d.Price.Close,
			Volume:    d.Volume.StockVolume,
		})
	}
	return bars, nil
}

// ---------------------------------------------------------------------------
// EquityQuote — Delayed quote from CBOE.
// URL: https://cdn.cboe.com/api/global/delayed_quotes/quotes/{SYMBOL}.json
// ---------------------------------------------------------------------------

type equityQuoteFetcher struct {
	provider.BaseFetcher
	prov *Provider
}

func newEquityQuoteFetcher(p *Provider) *equityQuoteFetcher {
	return &equityQuoteFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelEquityQuote,
			"CBOE delayed equity quote with IV metrics",
			[]string{provider.ParamSymbol},
			nil,
		),
		prov: p,
	}
}

func (f *equityQuoteFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	symbol := strings.ToUpper(params[provider.ParamSymbol])
	if symbol == "" {
		return nil, fmt.Errorf("cboe: %s is required", provider.ParamSymbol)
	}

	cacheKey := provider.CacheKey(provider.ModelEquityQuote, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	_, _ = f.prov.getIndexDirectory(ctx)
	url := quotesURL(f.prov.symbolPath(symbol))

	var resp cboeQuoteResponse
	if err := fetchCBOEJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("cboe equity quote: %w", err)
	}

	q := resp.Data
	quote := models.Quote{
		Ticker:     strings.ReplaceAll(q.Symbol, "^", ""),
		LastPrice:  q.CurrentPrice,
		Change:     q.PriceChange,
		ChangePct:  q.PriceChangePct / 100, // normalize from percentage
		Open:       q.Open,
		High:       q.High,
		Low:        q.Low,
		PrevClose:  q.PrevDayClose,
		Volume:     q.Volume,
		WeekHigh52: q.AnnualHigh,
		WeekLow52:  q.AnnualLow,
		Timestamp:  parseCBOETime(q.LastTradeTime),
	}

	result := newResult(quote)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// EquitySearch — Search CBOE index directory (symbol/name matching).
// ---------------------------------------------------------------------------

type equitySearchFetcher struct {
	provider.BaseFetcher
	prov *Provider
}

func newEquitySearchFetcher(p *Provider) *equitySearchFetcher {
	return &equitySearchFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelEquitySearch,
			"Search CBOE listed equities and indices",
			[]string{provider.ParamQuery},
			nil,
		),
		prov: p,
	}
}

func (f *equitySearchFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	query := strings.ToUpper(params[provider.ParamQuery])
	if query == "" {
		return nil, fmt.Errorf("cboe: %s is required", provider.ParamQuery)
	}

	cacheKey := provider.CacheKey(provider.ModelEquitySearch, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	indices, err := f.prov.getIndexDirectory(ctx)
	if err != nil {
		return nil, fmt.Errorf("cboe equity search: %w", err)
	}

	var results []models.EquitySearchResult
	for _, idx := range indices {
		if containsCI(idx.IndexSymbol, query) || containsCI(idx.Name, query) {
			results = append(results, models.EquitySearchResult{
				Symbol:   idx.IndexSymbol,
				Name:     idx.Name,
				Exchange: "CBOE",
			})
		}
	}

	result := newResult(results)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// containsCI checks if s contains substr (case-insensitive).
func containsCI(s, substr string) bool {
	return strings.Contains(strings.ToUpper(s), strings.ToUpper(substr))
}
