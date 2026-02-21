package fred

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---- FredSearch fetcher ----

type fredSearchFetcher struct {
	provider.BaseFetcher
}

func newFredSearchFetcher() *fredSearchFetcher {
	return &fredSearchFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelFredSearch,
			"Search FRED for economic data series",
			[]string{provider.ParamQuery},
			[]string{provider.ParamLimit},
			10*time.Minute, 10, time.Second,
		),
	}
}

func (f *fredSearchFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	query := params[provider.ParamQuery]
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("series/search?search_text=%s", url.QueryEscape(query))
	if lim := params[provider.ParamLimit]; lim != "" {
		endpoint += "&limit=" + lim
	} else {
		endpoint += "&limit=25"
	}

	var resp fredSearchResponse
	if err := fetchFredJSON(ctx, endpoint, apiKey, &resp); err != nil {
		return nil, fmt.Errorf("fred search: %w", err)
	}

	var results []models.FREDSearchResult
	for _, s := range resp.Seriess {
		results = append(results, models.FREDSearchResult{
			SeriesID:           s.ID,
			Title:              s.Title,
			ObservationStart:   parseFredDate(s.ObservationStart),
			ObservationEnd:     parseFredDate(s.ObservationEnd),
			Frequency:          s.Frequency,
			Units:              s.Units,
			SeasonalAdjustment: s.SeasonalAdjustment,
			Popularity:         s.Popularity,
		})
	}

	f.CacheSet(cacheKey, results)
	return newResult(results), nil
}

// ---- FredSeries fetcher ----

type fredSeriesFetcher struct {
	provider.BaseFetcher
}

func newFredSeriesFetcher() *fredSeriesFetcher {
	return &fredSeriesFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelFredSeries,
			"Get FRED time series observations by series ID",
			[]string{provider.ParamSymbol}, // series_id passed as symbol
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			10*time.Minute, 10, time.Second,
		),
	}
}

func (f *fredSeriesFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	seriesID := params[provider.ParamSymbol]
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	obs, err := fetchFredSeries(ctx, seriesID, apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred series %s: %w", seriesID, err)
	}

	var data []models.FREDSeriesData
	for _, o := range obs {
		if o.Value == "." {
			continue // Skip missing values
		}
		data = append(data, models.FREDSeriesData{
			Date:  parseFredDate(o.Date),
			Value: parseFloat(o.Value),
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// ---- FredReleaseTable fetcher ----

type fredReleaseTableFetcher struct {
	provider.BaseFetcher
}

func newFredReleaseTableFetcher() *fredReleaseTableFetcher {
	return &fredReleaseTableFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelFredReleaseTable,
			"Get FRED release table data",
			[]string{provider.ParamSymbol}, // release_id passed as symbol
			nil,
			30*time.Minute, 10, time.Second,
		),
	}
}

func (f *fredReleaseTableFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	releaseID := params[provider.ParamSymbol]
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("release/tables?release_id=%s&include_observation_values=true&observation_date=9999-12-31", releaseID)
	var resp fredReleaseTableResponse
	if err := fetchFredJSON(ctx, endpoint, apiKey, &resp); err != nil {
		return nil, fmt.Errorf("fred release table %s: %w", releaseID, err)
	}

	f.CacheSet(cacheKey, resp)
	return newResult(resp), nil
}

// ---- FredRegional fetcher ----

type fredRegionalFetcher struct {
	provider.BaseFetcher
}

func newFredRegionalFetcher() *fredRegionalFetcher {
	return &fredRegionalFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelFredRegional,
			"Get FRED regional (GeoFRED) data by series ID",
			[]string{provider.ParamSymbol}, // series_id
			[]string{provider.ParamStartDate},
			30*time.Minute, 10, time.Second,
		),
	}
}

func (f *fredRegionalFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	seriesID := params[provider.ParamSymbol]
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("geofred/series/data?series_id=%s", seriesID)
	if sd := params[provider.ParamStartDate]; sd != "" {
		endpoint += "&start_date=" + sd
	}

	var resp fredRegionalResponse
	if err := fetchFredJSON(ctx, endpoint, apiKey, &resp); err != nil {
		return nil, fmt.Errorf("fred regional %s: %w", seriesID, err)
	}

	f.CacheSet(cacheKey, resp)
	return newResult(resp), nil
}
