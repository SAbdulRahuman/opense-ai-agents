package cboe

import (
	"context"
	"fmt"
	"strings"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// AvailableIndices — List all CBOE indices.
// URL: https://cdn.cboe.com/api/global/us_indices/definitions/all_indices.json
// ---------------------------------------------------------------------------

type availableIndicesFetcher struct {
	provider.BaseFetcher
	prov *Provider
}

func newAvailableIndicesFetcher(p *Provider) *availableIndicesFetcher {
	return &availableIndicesFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelAvailableIndices,
			"List all CBOE available indices with metadata",
			nil,
			nil,
		),
		prov: p,
	}
}

func (f *availableIndicesFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelAvailableIndices, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	indices, err := f.prov.getIndexDirectory(ctx)
	if err != nil {
		return nil, fmt.Errorf("cboe available indices: %w", err)
	}

	var results []models.IndexInfo
	for _, idx := range indices {
		results = append(results, models.IndexInfo{
			Symbol:      idx.IndexSymbol,
			Name:        idx.Name,
			Exchange:    "CBOE",
			Currency:    idx.Currency,
			Description: idx.Description,
		})
	}

	result := newResult(results)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// IndexHistorical — Historical OHLCV for CBOE indices.
// URL: https://cdn.cboe.com/api/global/delayed_quotes/charts/historical/_{SYMBOL}.json
// ---------------------------------------------------------------------------

type indexHistoricalFetcher struct {
	provider.BaseFetcher
	prov *Provider
}

func newIndexHistoricalFetcher(p *Provider) *indexHistoricalFetcher {
	return &indexHistoricalFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelIndexHistorical,
			"CBOE index historical OHLCV data (daily or intraday)",
			[]string{provider.ParamSymbol},
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamInterval},
		),
		prov: p,
	}
}

func (f *indexHistoricalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	symbol := strings.ToUpper(params[provider.ParamSymbol])
	if symbol == "" {
		return nil, fmt.Errorf("cboe: %s is required", provider.ParamSymbol)
	}
	symbol = strings.ReplaceAll(symbol, "^", "")

	cacheKey := provider.CacheKey(provider.ModelIndexHistorical, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	_, _ = f.prov.getIndexDirectory(ctx)

	interval := params[provider.ParamInterval]
	if interval == "" {
		interval = "1d"
	}

	// Index symbols always get underscore prefix in CBOE.
	symPath := "_" + symbol
	url := chartURL(symPath, interval)

	raw, err := fetchCBOERaw(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("cboe index historical: %w", err)
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

// ---------------------------------------------------------------------------
// IndexSearch — Search CBOE index directory by name/symbol.
// ---------------------------------------------------------------------------

type indexSearchFetcher struct {
	provider.BaseFetcher
	prov *Provider
}

func newIndexSearchFetcher(p *Provider) *indexSearchFetcher {
	return &indexSearchFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelIndexSearch,
			"Search CBOE indices by name or symbol",
			[]string{provider.ParamQuery},
			nil,
		),
		prov: p,
	}
}

func (f *indexSearchFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	query := params[provider.ParamQuery]
	if query == "" {
		return nil, fmt.Errorf("cboe: %s is required", provider.ParamQuery)
	}

	cacheKey := provider.CacheKey(provider.ModelIndexSearch, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	indices, err := f.prov.getIndexDirectory(ctx)
	if err != nil {
		return nil, fmt.Errorf("cboe index search: %w", err)
	}

	var results []models.IndexInfo
	for _, idx := range indices {
		if containsCI(idx.IndexSymbol, query) || containsCI(idx.Name, query) || containsCI(idx.Description, query) {
			results = append(results, models.IndexInfo{
				Symbol:      idx.IndexSymbol,
				Name:        idx.Name,
				Exchange:    "CBOE",
				Currency:    idx.Currency,
				Description: idx.Description,
			})
		}
	}

	result := newResult(results)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// IndexSnapshots — Current snapshot of all US or EU CBOE indices.
// US: https://cdn.cboe.com/api/global/delayed_quotes/quotes/all_us_indices.json
// EU: https://cdn.cboe.com/api/global/european_indices/index_quotes/all-indices.json
// ---------------------------------------------------------------------------

type indexSnapshotsFetcher struct {
	provider.BaseFetcher
}

func newIndexSnapshotsFetcher() *indexSnapshotsFetcher {
	return &indexSnapshotsFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelIndexSnapshots,
			"CBOE real-time index snapshots (US and EU)",
			nil,
			[]string{"region"}, // "us" (default) or "eu"
		),
	}
}

func (f *indexSnapshotsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelIndexSnapshots, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	region := strings.ToLower(params["region"])
	if region == "" {
		region = "us"
	}

	url := urlAllUSSnaps
	if region == "eu" {
		url = urlAllEUSnaps
	}

	var resp cboeSnapshotResponse
	if err := fetchCBOEJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("cboe index snapshots: %w", err)
	}

	var snapshots []models.IndexSnapshot
	for _, s := range resp.Data {
		snapshots = append(snapshots, models.IndexSnapshot{
			Symbol:    strings.ReplaceAll(s.Symbol, "^", ""),
			Name:      s.Name,
			Value:     s.CurrentPrice,
			Change:    s.PriceChange,
			ChangePct: s.PriceChangePct / 100, // normalize
			Open:      s.Open,
			High:      s.High,
			Low:       s.Low,
			PrevClose: s.PrevDayClose,
			Volume:    s.Volume,
			Timestamp: parseCBOETime(s.LastTradeTime),
		})
	}

	result := newResult(snapshots)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// IndexConstituents — Constituents of CBOE European indices.
// URL: https://cdn.cboe.com/api/global/european_indices/constituent_quotes/{SYMBOL}.json
// ---------------------------------------------------------------------------

type indexConstituentsFetcher struct {
	provider.BaseFetcher
}

func newIndexConstituentsFetcher() *indexConstituentsFetcher {
	return &indexConstituentsFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelIndexConstituents,
			"CBOE European index constituents with current prices",
			[]string{provider.ParamSymbol},
			nil,
		),
	}
}

func (f *indexConstituentsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	symbol := strings.ToUpper(params[provider.ParamSymbol])
	if symbol == "" {
		return nil, fmt.Errorf("cboe: %s is required", provider.ParamSymbol)
	}

	cacheKey := provider.CacheKey(provider.ModelIndexConstituents, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	url := baseEUIndices + "/constituent_quotes/" + symbol + ".json"

	var resp cboeConstituentResponse
	if err := fetchCBOEJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("cboe index constituents: %w", err)
	}

	var constituents []models.IndexConstituent
	for _, c := range resp.Data {
		constituents = append(constituents, models.IndexConstituent{
			Symbol: strings.ReplaceAll(c.Symbol, "^", ""),
			Name:   c.Name,
		})
	}

	result := newResult(constituents)
	f.CacheSet(cacheKey, result)
	return result, nil
}
