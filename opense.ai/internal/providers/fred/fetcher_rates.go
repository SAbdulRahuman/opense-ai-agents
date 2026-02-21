package fred

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// seriesRateFetcher is a generic fetcher for FRED rate series.
// It fetches observations for a given seriesID and maps them to InterestRateData.
type seriesRateFetcher struct {
	provider.BaseFetcher
	seriesID string
	rateType string
}

func newSeriesRateFetcher(model provider.ModelType, desc, seriesID, rateType string) *seriesRateFetcher {
	return &seriesRateFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			model, desc,
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
		seriesID: seriesID,
		rateType: rateType,
	}
}

func (f *seriesRateFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
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
		return nil, fmt.Errorf("fred %s (%s): %w", f.rateType, f.seriesID, err)
	}

	var data []models.InterestRateData
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.InterestRateData{
			Date:     parseFredDate(o.Date),
			Rate:     parseFloat(o.Value),
			RateType: f.rateType,
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// --- Rate fetcher constructors ---

func newSOFRFetcher() *seriesRateFetcher {
	return newSeriesRateFetcher(
		provider.ModelSOFR,
		"Secured Overnight Financing Rate (SOFR) from FRED",
		"SOFR", "SOFR",
	)
}

func newSONIAFetcher() *seriesRateFetcher {
	return newSeriesRateFetcher(
		provider.ModelSONIA,
		"Sterling Overnight Index Average (SONIA) from FRED",
		"IUDSOIA", "SONIA",
	)
}

func newAmeriborFetcher() *seriesRateFetcher {
	return newSeriesRateFetcher(
		provider.ModelAmeribor,
		"AMERIBOR overnight unsecured rate from FRED",
		"AMERIBOR", "AMERIBOR",
	)
}

func newFederalFundsRateFetcher() *seriesRateFetcher {
	return newSeriesRateFetcher(
		provider.ModelFederalFundsRate,
		"Effective Federal Funds Rate from FRED",
		"DFF", "Federal Funds",
	)
}

func newIORBFetcher() *seriesRateFetcher {
	return newSeriesRateFetcher(
		provider.ModelIORB,
		"Interest on Reserve Balances (IORB) from FRED",
		"IORB", "IORB",
	)
}

func newDiscountWindowFetcher() *seriesRateFetcher {
	return newSeriesRateFetcher(
		provider.ModelDiscountWindowPrimaryCreditRate,
		"Discount Window Primary Credit Rate from FRED",
		"DPCREDIT", "Discount Window",
	)
}

func newOvernightBankFundingFetcher() *seriesRateFetcher {
	return newSeriesRateFetcher(
		provider.ModelOvernightBankFundingRate,
		"Overnight Bank Funding Rate (OBFR) from FRED",
		"OBFR", "OBFR",
	)
}

func newEuroShortTermRateFetcher() *seriesRateFetcher {
	return newSeriesRateFetcher(
		provider.ModelEuroShortTermRate,
		"Euro Short-Term Rate (€STR) from FRED",
		"ECBESTRVOLWGTTRMDMNRT", "€STR",
	)
}

func newECBInterestRatesFetcher() *seriesRateFetcher {
	return newSeriesRateFetcher(
		provider.ModelEuropeanCentralBankInterestRates,
		"ECB Main Refinancing Rate from FRED",
		"ECBMRRFR", "ECB Main Refinancing",
	)
}

// --- Projections fetcher (FOMC projections via multiple series) ---

type projectionsFetcher struct {
	provider.BaseFetcher
}

func newProjectionsFetcher() *projectionsFetcher {
	return &projectionsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelProjections,
			"FOMC projections for federal funds rate from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			30*time.Minute, 10, time.Second,
		),
	}
}

func (f *projectionsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Use the FOMC Summary of Economic Projections: Fed Funds Rate Median
	obs, err := fetchFredSeries(ctx, "FEDTARMD", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred projections: %w", err)
	}

	var data []models.RateProjection
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		val := parseFloat(o.Value)
		data = append(data, models.RateProjection{
			Date:       parseFredDate(o.Date),
			RateMedian: val,
			RateLow:    val,
			RateHigh:   val,
			Source:      "FOMC",
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}
