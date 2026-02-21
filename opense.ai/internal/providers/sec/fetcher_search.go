package sec

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---- EquitySearch fetcher ----
// Searches SEC EDGAR for companies by name, ticker, CIK, or keyword.

type equitySearchFetcher struct {
	provider.BaseFetcher
}

func newEquitySearchFetcher() *equitySearchFetcher {
	return &equitySearchFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelEquitySearch,
			"Search SEC EDGAR for companies by name, ticker, or CIK",
			[]string{provider.ParamQuery},
			[]string{provider.ParamLimit},
			10*time.Minute, 8, time.Second,
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

	// Use SEC company tickers JSON for matching.
	u := edgarDataURL + "/files/company_tickers.json"
	var tickers map[string]edgarTickerEntry
	if err := fetchSECJSON(ctx, u, &tickers); err != nil {
		return nil, fmt.Errorf("sec equity search: %w", err)
	}

	queryUpper := strings.ToUpper(query)
	var results []models.EquitySearchResult
	for _, entry := range tickers {
		if strings.Contains(strings.ToUpper(entry.Ticker), queryUpper) ||
			strings.Contains(strings.ToUpper(entry.Title), queryUpper) ||
			strings.Contains(entry.CIKStr, query) {
			results = append(results, models.EquitySearchResult{
				Symbol:   entry.Ticker,
				Name:     entry.Title,
				Exchange: "US", // SEC data — US-listed
			})
		}
	}

	limit := 25
	if lim := params[provider.ParamLimit]; lim != "" {
		if n := parseInt(lim); n > 0 {
			limit = n
		}
	}
	if len(results) > limit {
		results = results[:limit]
	}

	f.CacheSet(cacheKey, results)
	return newResult(results), nil
}

// ---- CikMap fetcher ----
// Returns CIK-to-ticker mappings from SEC.

type cikMapFetcher struct {
	provider.BaseFetcher
}

func newCikMapFetcher() *cikMapFetcher {
	return &cikMapFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCikMap,
			"SEC CIK to ticker symbol mapping",
			nil,
			[]string{provider.ParamQuery, provider.ParamLimit},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *cikMapFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	u := edgarDataURL + "/files/company_tickers.json"
	var tickers map[string]edgarTickerEntry
	if err := fetchSECJSON(ctx, u, &tickers); err != nil {
		return nil, fmt.Errorf("sec cik map: %w", err)
	}

	filter := strings.ToUpper(params[provider.ParamQuery])
	var mappings []models.CIKMapping
	for _, entry := range tickers {
		if filter != "" {
			if !strings.Contains(strings.ToUpper(entry.Ticker), filter) &&
				!strings.Contains(strings.ToUpper(entry.Title), filter) &&
				!strings.Contains(entry.CIKStr, filter) {
				continue
			}
		}
		mappings = append(mappings, models.CIKMapping{
			CIK:    entry.CIKStr,
			Symbol: entry.Ticker,
			Name:   entry.Title,
		})
	}

	limit := 100
	if lim := params[provider.ParamLimit]; lim != "" {
		if n := parseInt(lim); n > 0 {
			limit = n
		}
	}
	if len(mappings) > limit {
		mappings = mappings[:limit]
	}

	f.CacheSet(cacheKey, mappings)
	return newResult(mappings), nil
}

// ---- SymbolMap fetcher ----
// Returns ticker-to-CIK mappings (reverse of CikMap).

type symbolMapFetcher struct {
	provider.BaseFetcher
}

func newSymbolMapFetcher() *symbolMapFetcher {
	return &symbolMapFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelSymbolMap,
			"SEC ticker symbol to CIK mapping",
			[]string{provider.ParamSymbol},
			nil,
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *symbolMapFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	symbol := params[provider.ParamSymbol]
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	u := edgarDataURL + "/files/company_tickers.json"
	var tickers map[string]edgarTickerEntry
	if err := fetchSECJSON(ctx, u, &tickers); err != nil {
		return nil, fmt.Errorf("sec symbol map: %w", err)
	}

	sym := strings.ToUpper(strings.TrimSpace(symbol))
	for _, entry := range tickers {
		if strings.EqualFold(entry.Ticker, sym) {
			mapping := models.CIKMapping{
				CIK:    entry.CIKStr,
				Symbol: entry.Ticker,
				Name:   entry.Title,
			}
			f.CacheSet(cacheKey, mapping)
			return newResult(mapping), nil
		}
	}

	return nil, fmt.Errorf("symbol %s not found in SEC data", symbol)
}

// ---- SicSearch fetcher ----
// Searches SEC SIC (Standard Industrial Classification) codes.

type sicSearchFetcher struct {
	provider.BaseFetcher
}

func newSicSearchFetcher() *sicSearchFetcher {
	return &sicSearchFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelSicSearch,
			"Search SEC Standard Industrial Classification (SIC) codes",
			[]string{provider.ParamQuery},
			[]string{provider.ParamLimit},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *sicSearchFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	query := params[provider.ParamQuery]
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// SEC provides SIC codes via EDGAR company search.
	u := fmt.Sprintf("%s/search-index?q=%s&dateRange=custom", edgarBaseURL, url.QueryEscape(query))
	var resp edgarSearchResponse
	if err := fetchSECJSON(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("sec sic search: %w", err)
	}

	// Deduplicate by entity type.
	seen := make(map[string]bool)
	var results []models.SICEntry
	for _, hit := range resp.Hits.Hits {
		doc := hit.Source
		key := doc.EntityType + doc.EntityName
		if seen[key] {
			continue
		}
		seen[key] = true
		results = append(results, models.SICEntry{
			Description: doc.EntityName,
			Industry:    doc.EntityType,
		})
	}

	limit := 25
	if lim := params[provider.ParamLimit]; lim != "" {
		if n := parseInt(lim); n > 0 {
			limit = n
		}
	}
	if len(results) > limit {
		results = results[:limit]
	}

	f.CacheSet(cacheKey, results)
	return newResult(results), nil
}

// ---- InstitutionsSearch fetcher ----
// Searches for financial institutions registered with the SEC.

type institutionsSearchFetcher struct {
	provider.BaseFetcher
}

func newInstitutionsSearchFetcher() *institutionsSearchFetcher {
	return &institutionsSearchFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelInstitutionsSearch,
			"Search SEC-registered financial institutions",
			[]string{provider.ParamQuery},
			[]string{provider.ParamLimit},
			1*time.Hour, 5, time.Second,
		),
	}
}

func (f *institutionsSearchFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	query := params[provider.ParamQuery]
	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Search for institutions via company tickers where names match.
	u := edgarDataURL + "/files/company_tickers.json"
	var tickers map[string]edgarTickerEntry
	if err := fetchSECJSON(ctx, u, &tickers); err != nil {
		return nil, fmt.Errorf("sec institutions search: %w", err)
	}

	queryUpper := strings.ToUpper(query)
	var results []models.InstitutionEntry
	for _, entry := range tickers {
		if strings.Contains(strings.ToUpper(entry.Title), queryUpper) {
			results = append(results, models.InstitutionEntry{
				CIK:  entry.CIKStr,
				Name: entry.Title,
			})
		}
	}

	limit := 25
	if lim := params[provider.ParamLimit]; lim != "" {
		if n := parseInt(lim); n > 0 {
			limit = n
		}
	}
	if len(results) > limit {
		results = results[:limit]
	}

	f.CacheSet(cacheKey, results)
	return newResult(results), nil
}
