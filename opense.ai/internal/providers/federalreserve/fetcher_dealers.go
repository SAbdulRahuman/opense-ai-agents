package federalreserve

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// PrimaryDealerFails — NY Fed primary dealer statistics: fail-to-deliver.
// URL: https://markets.newyorkfed.org/api/pd/get/PDFTD-CS.json
// ---------------------------------------------------------------------------

type primaryDealerFailsFetcher struct {
	provider.BaseFetcher
}

func newPrimaryDealerFailsFetcher() *primaryDealerFailsFetcher {
	return &primaryDealerFailsFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelPrimaryDealerFails,
			"Federal Reserve primary dealer fails to deliver",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *primaryDealerFailsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelPrimaryDealerFails, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	entries, err := fetchPDSeries(ctx, "PDFTD-CS", params)
	if err != nil {
		return nil, fmt.Errorf("primary dealer fails: %w", err)
	}

	result := newResult(entries)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// PrimaryDealerPositioning — NY Fed primary dealer net positioning.
// URL: https://markets.newyorkfed.org/api/pd/get/PDPOS-CS.json
// ---------------------------------------------------------------------------

type primaryDealerPositioningFetcher struct {
	provider.BaseFetcher
}

func newPrimaryDealerPositioningFetcher() *primaryDealerPositioningFetcher {
	return &primaryDealerPositioningFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelPrimaryDealerPositioning,
			"Federal Reserve primary dealer positioning",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *primaryDealerPositioningFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelPrimaryDealerPositioning, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	entries, err := fetchPDSeries(ctx, "PDPOS-CS", params)
	if err != nil {
		return nil, fmt.Errorf("primary dealer positioning: %w", err)
	}

	result := newResult(entries)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// fetchPDSeries fetches a primary dealer series from the NY Fed API.
func fetchPDSeries(ctx context.Context, series string, params provider.QueryParams) ([]models.EconomicIndicatorData, error) {
	url := baseNYFed + "/api/pd/get/" + series + ".json"

	var resp nyfedPDResponse
	if err := fetchFedJSON(ctx, url, &resp); err != nil {
		return nil, err
	}

	startDate := params[provider.ParamStartDate]
	endDate := params[provider.ParamEndDate]

	var entries []models.EconomicIndicatorData
	for _, e := range resp.PD.Timeseries {
		if startDate != "" && e.KeyID < startDate {
			continue
		}
		if endDate != "" && e.KeyID > endDate {
			continue
		}

		dt, _ := time.Parse("2006-01-02", e.KeyID)
		if dt.IsZero() {
			dt = parseDate(e.KeyID)
		}

		val := parseFloat64(e.Value)

		entries = append(entries, models.EconomicIndicatorData{
			Date:    dt,
			Country: "US",
			Value:   val,
		})
	}

	return entries, nil
}
