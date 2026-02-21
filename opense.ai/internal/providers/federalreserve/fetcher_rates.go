package federalreserve

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// FederalFundsRate (EFFR) — Effective Federal Funds Rate.
// URL: https://markets.newyorkfed.org/api/rates/unsecured/effr/search.json
// ---------------------------------------------------------------------------

type federalFundsRateFetcher struct {
	provider.BaseFetcher
}

func newFederalFundsRateFetcher() *federalFundsRateFetcher {
	return &federalFundsRateFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelFederalFundsRate,
			"Federal Reserve effective federal funds rate (EFFR)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *federalFundsRateFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelFederalFundsRate, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	start := defaultDate(params, provider.ParamStartDate, "2016-03-01")
	end := defaultDate(params, provider.ParamEndDate, time.Now().Format("2006-01-02"))

	url := buildNYFedRatesURL("unsecured/effr", start, end)

	var resp nyfedRatesResponse
	if err := fetchFedJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("fed funds rate: %w", err)
	}

	var rates []models.InterestRateData
	for _, r := range resp.RefRates {
		rates = append(rates, models.InterestRateData{
			Date:     parseDate(r.EffectiveDate),
			Rate:     r.PercentRate / 100, // normalize from percentage
			RateType: "FedFunds",
			Maturity: "overnight",
		})
	}

	result := newResult(rates)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// SOFR — Secured Overnight Financing Rate.
// URL: https://markets.newyorkfed.org/api/rates/secured/sofr/search.json
// ---------------------------------------------------------------------------

type sofrFetcher struct {
	provider.BaseFetcher
}

func newSOFRFetcher() *sofrFetcher {
	return &sofrFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelSOFR,
			"Federal Reserve SOFR (Secured Overnight Financing Rate)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *sofrFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelSOFR, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	start := defaultDate(params, provider.ParamStartDate, "2018-04-02")
	end := defaultDate(params, provider.ParamEndDate, time.Now().Format("2006-01-02"))

	url := buildNYFedRatesURL("secured/sofr", start, end)

	var resp nyfedRatesResponse
	if err := fetchFedJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("sofr: %w", err)
	}

	var rates []models.InterestRateData
	for _, r := range resp.RefRates {
		rates = append(rates, models.InterestRateData{
			Date:     parseDate(r.EffectiveDate),
			Rate:     r.PercentRate / 100,
			RateType: "SOFR",
			Maturity: "overnight",
		})
	}

	result := newResult(rates)
	f.CacheSet(cacheKey, result)
	return result, nil
}

// ---------------------------------------------------------------------------
// OBFR — Overnight Bank Funding Rate.
// URL: https://markets.newyorkfed.org/api/rates/unsecured/obfr/search.json
// ---------------------------------------------------------------------------

type obfrFetcher struct {
	provider.BaseFetcher
}

func newOBFRFetcher() *obfrFetcher {
	return &obfrFetcher{
		BaseFetcher: provider.NewBaseFetcher(
			provider.ModelOvernightBankFundingRate,
			"Federal Reserve OBFR (Overnight Bank Funding Rate)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
		),
	}
}

func (f *obfrFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	cacheKey := provider.CacheKey(provider.ModelOvernightBankFundingRate, params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return cached.(*provider.FetchResult), nil
	}

	start := defaultDate(params, provider.ParamStartDate, "2016-03-01")
	end := defaultDate(params, provider.ParamEndDate, time.Now().Format("2006-01-02"))

	url := buildNYFedRatesURL("unsecured/obfr", start, end)

	var resp nyfedRatesResponse
	if err := fetchFedJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("obfr: %w", err)
	}

	var rates []models.InterestRateData
	for _, r := range resp.RefRates {
		rates = append(rates, models.InterestRateData{
			Date:     parseDate(r.EffectiveDate),
			Rate:     r.PercentRate / 100,
			RateType: "OBFR",
			Maturity: "overnight",
		})
	}

	result := newResult(rates)
	f.CacheSet(cacheKey, result)
	return result, nil
}
