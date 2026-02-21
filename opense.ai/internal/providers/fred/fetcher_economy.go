package fred

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// seriesEconomyFetcher is a generic fetcher for FRED economic indicator series.
// It fetches observations and maps them to EconomicIndicatorData.
type seriesEconomyFetcher struct {
	provider.BaseFetcher
	seriesID  string
	indicator string
	country   string
}

func newSeriesEconomyFetcher(model provider.ModelType, desc, seriesID, indicator, country string) *seriesEconomyFetcher {
	return &seriesEconomyFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			model, desc,
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
		seriesID:  seriesID,
		indicator: indicator,
		country:   country,
	}
}

func (f *seriesEconomyFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	obs, err := fetchFredSeries(ctx, f.seriesID, apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred %s (%s): %w", f.indicator, f.seriesID, err)
	}

	var data []models.EconomicIndicatorData
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.EconomicIndicatorData{
			Date:    parseFredDate(o.Date),
			Value:   parseFloat(o.Value),
			Country: f.country,
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// --- CPI fetcher ---

type cpiFetcher struct {
	provider.BaseFetcher
}

func newConsumerPriceIndexFetcher() *cpiFetcher {
	return &cpiFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelConsumerPriceIndex,
			"Consumer Price Index (CPI) from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *cpiFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Fetch headline CPI (all urban consumers).
	headlineObs, err := fetchFredSeries(ctx, "CPIAUCSL", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred cpi: %w", err)
	}

	var data []models.CPIData
	for _, o := range headlineObs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.CPIData{
			Date:      parseFredDate(o.Date),
			Value:     parseFloat(o.Value),
			Country:   "US",
			Frequency: "monthly",
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// --- Non-Farm Payrolls fetcher ---

func newNonFarmPayrollsFetcher() *seriesEconomyFetcher {
	return newSeriesEconomyFetcher(
		provider.ModelNonFarmPayrolls,
		"Total Nonfarm Payrolls from FRED (BLS)",
		"PAYEMS", "Non-Farm Payrolls", "US",
	)
}

// --- PCE fetcher ---

func newPersonalConsumptionExpendituresFetcher() *seriesEconomyFetcher {
	return newSeriesEconomyFetcher(
		provider.ModelPersonalConsumptionExpenditures,
		"Personal Consumption Expenditures (PCE) from FRED",
		"PCE", "PCE", "US",
	)
}

// --- University of Michigan Consumer Sentiment ---

type umichFetcher struct {
	provider.BaseFetcher
}

func newUniversityOfMichiganFetcher() *umichFetcher {
	return &umichFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelUniversityOfMichigan,
			"University of Michigan Consumer Sentiment from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *umichFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	obs, err := fetchFredSeries(ctx, "UMCSENT", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred umich sentiment: %w", err)
	}

	var data []models.ConsumerSentimentData
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.ConsumerSentimentData{
			Date:           parseFredDate(o.Date),
			SentimentIndex: parseFloat(o.Value),
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// --- Manufacturing NY (Empire State) ---

func newManufacturingNYFetcher() *seriesEconomyFetcher {
	return newSeriesEconomyFetcher(
		provider.ModelManufacturingOutlookNY,
		"NY Empire State Manufacturing Index from FRED",
		"GACDISA066MSFRBNY", "Manufacturing Outlook NY", "US",
	)
}

// --- Manufacturing Texas (Dallas Fed) ---

func newManufacturingTexasFetcher() *seriesEconomyFetcher {
	return newSeriesEconomyFetcher(
		provider.ModelManufacturingOutlookTexas,
		"Texas Manufacturing Outlook from FRED (Dallas Fed)",
		"TXMKSURVEYPRODUCTION", "Manufacturing Outlook Texas", "US",
	)
}

// --- Retail Prices ---

func newRetailPricesFetcher() *seriesEconomyFetcher {
	return newSeriesEconomyFetcher(
		provider.ModelRetailPrices,
		"Retail price data from FRED",
		"CUSR0000SA0", "Retail Prices", "US",
	)
}

// --- Commodity Spot Prices ---

func newCommoditySpotPricesFetcher() *seriesEconomyFetcher {
	return newSeriesEconomyFetcher(
		provider.ModelCommoditySpotPrices,
		"WTI Crude Oil and commodity spot prices from FRED",
		"DCOILWTICO", "Commodity Spot Prices", "US",
	)
}

// --- Unemployment ---

type unemploymentFetcher struct {
	provider.BaseFetcher
}

func newUnemploymentFetcher() *unemploymentFetcher {
	return &unemploymentFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelUnemployment,
			"Unemployment rate from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit, provider.ParamCountry},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *unemploymentFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	obs, err := fetchFredSeries(ctx, "UNRATE", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred unemployment: %w", err)
	}

	var data []models.UnemploymentData
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.UnemploymentData{
			Date:    parseFredDate(o.Date),
			Value:   parseFloat(o.Value),
			Country: "US",
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// --- GDP Real ---

type gdpRealFetcher struct {
	provider.BaseFetcher
}

func newGDPRealFetcher() *gdpRealFetcher {
	return &gdpRealFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelGdpReal,
			"Real GDP from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			30*time.Minute, 10, time.Second,
		),
	}
}

func (f *gdpRealFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	obs, err := fetchFredSeries(ctx, "GDPC1", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred gdp real: %w", err)
	}

	var data []models.GDPData
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.GDPData{
			Date:    parseFredDate(o.Date),
			Value:   parseFloat(o.Value),
			Country: "US",
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}
