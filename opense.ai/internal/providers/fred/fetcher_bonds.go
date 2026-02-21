package fred

import (
	"context"
	"fmt"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---- YieldCurve fetcher ----
// Returns US Treasury yield curve points from FRED.

type yieldCurveFetcher struct {
	provider.BaseFetcher
}

func newYieldCurveFetcher() *yieldCurveFetcher {
	return &yieldCurveFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelYieldCurve,
			"US Treasury yield curve from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate},
			15*time.Minute, 10, time.Second,
		),
	}
}

// yieldCurveSeries maps maturity labels to FRED series IDs.
var yieldCurveSeries = []struct {
	maturity string
	seriesID string
}{
	{"1M", "DGS1MO"},
	{"3M", "DGS3MO"},
	{"6M", "DGS6MO"},
	{"1Y", "DGS1"},
	{"2Y", "DGS2"},
	{"3Y", "DGS3"},
	{"5Y", "DGS5"},
	{"7Y", "DGS7"},
	{"10Y", "DGS10"},
	{"20Y", "DGS20"},
	{"30Y", "DGS30"},
}

func (f *yieldCurveFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Limit to last observation for each maturity to form the curve.
	curveParams := make(provider.QueryParams, len(params))
	for k, v := range params {
		curveParams[k] = v
	}
	curveParams[provider.ParamLimit] = "1"

	var points []models.YieldCurvePoint
	for _, s := range yieldCurveSeries {
		obs, err := fetchFredSeries(ctx, s.seriesID, apiKey, curveParams)
		if err != nil {
			continue // Skip unavailable maturities
		}
		for _, o := range obs {
			if o.Value == "." {
				continue
			}
			points = append(points, models.YieldCurvePoint{
				Date:     parseFredDate(o.Date),
				Maturity: s.maturity,
				Rate:     parseFloat(o.Value),
			})
		}
	}

	f.CacheSet(cacheKey, points)
	return newResult(points), nil
}

// ---- TreasuryConstantMaturity fetcher ----
// Returns Treasury Constant Maturity rates via FRED.

type treasuryConstantMaturityFetcher struct {
	provider.BaseFetcher
}

func newTreasuryConstantMaturityFetcher() *treasuryConstantMaturityFetcher {
	return &treasuryConstantMaturityFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelTreasuryConstantMaturity,
			"Treasury Constant Maturity rates from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *treasuryConstantMaturityFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Fetch 10-Year Treasury Constant Maturity Rate as the primary series.
	obs, err := fetchFredSeries(ctx, "DGS10", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred treasury constant maturity: %w", err)
	}

	var data []models.InterestRateData
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.InterestRateData{
			Date:     parseFredDate(o.Date),
			Rate:     parseFloat(o.Value),
			RateType: "Treasury Constant Maturity",
			Maturity: "10Y",
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// ---- SelectedTreasuryConstantMaturity fetcher ----

type selectedTreasuryConstantMaturityFetcher struct {
	provider.BaseFetcher
}

func newSelectedTreasuryConstantMaturityFetcher() *selectedTreasuryConstantMaturityFetcher {
	return &selectedTreasuryConstantMaturityFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelSelectedTreasuryConstantMaturity,
			"Selected Treasury Constant Maturity rates (2Y, 5Y, 10Y, 30Y) from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *selectedTreasuryConstantMaturityFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	selected := []struct {
		maturity string
		seriesID string
	}{
		{"2Y", "DGS2"},
		{"5Y", "DGS5"},
		{"10Y", "DGS10"},
		{"30Y", "DGS30"},
	}

	var data []models.InterestRateData
	for _, s := range selected {
		obs, err := fetchFredSeries(ctx, s.seriesID, apiKey, params)
		if err != nil {
			continue
		}
		for _, o := range obs {
			if o.Value == "." {
				continue
			}
			data = append(data, models.InterestRateData{
				Date:     parseFredDate(o.Date),
				Rate:     parseFloat(o.Value),
				RateType: "Treasury Constant Maturity",
				Maturity: s.maturity,
			})
		}
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// ---- SelectedTreasuryBill fetcher ----

type selectedTreasuryBillFetcher struct {
	provider.BaseFetcher
}

func newSelectedTreasuryBillFetcher() *selectedTreasuryBillFetcher {
	return &selectedTreasuryBillFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelSelectedTreasuryBill,
			"Selected Treasury Bill rates from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *selectedTreasuryBillFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	bills := []struct {
		maturity string
		seriesID string
	}{
		{"3M", "DTB3"},
		{"6M", "DTB6"},
		{"1Y", "DTB1YR"},
	}

	var data []models.InterestRateData
	for _, s := range bills {
		obs, err := fetchFredSeries(ctx, s.seriesID, apiKey, params)
		if err != nil {
			continue
		}
		for _, o := range obs {
			if o.Value == "." {
				continue
			}
			data = append(data, models.InterestRateData{
				Date:     parseFredDate(o.Date),
				Rate:     parseFloat(o.Value),
				RateType: "Treasury Bill",
				Maturity: s.maturity,
			})
		}
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// ---- TipsYields fetcher ----

type tipsYieldsFetcher struct {
	provider.BaseFetcher
}

func newTipsYieldsFetcher() *tipsYieldsFetcher {
	return &tipsYieldsFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelTipsYields,
			"Treasury Inflation-Protected Securities (TIPS) yields from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *tipsYieldsFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	tips := []struct {
		maturity string
		seriesID string
	}{
		{"5Y", "DFII5"},
		{"7Y", "DFII7"},
		{"10Y", "DFII10"},
		{"20Y", "DFII20"},
		{"30Y", "DFII30"},
	}

	var data []models.InterestRateData
	for _, s := range tips {
		obs, err := fetchFredSeries(ctx, s.seriesID, apiKey, params)
		if err != nil {
			continue
		}
		for _, o := range obs {
			if o.Value == "." {
				continue
			}
			data = append(data, models.InterestRateData{
				Date:     parseFredDate(o.Date),
				Rate:     parseFloat(o.Value),
				RateType: "TIPS",
				Maturity: s.maturity,
			})
		}
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// ---- HighQualityMarketCorporateBond fetcher ----

type hqmCorporateBondFetcher struct {
	provider.BaseFetcher
}

func newHighQualityMarketCorporateBondFetcher() *hqmCorporateBondFetcher {
	return &hqmCorporateBondFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelHighQualityMarketCorporateBond,
			"High Quality Market Corporate Bond yields from FRED (Moody's AAA/BAA)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *hqmCorporateBondFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Fetch Moody's AAA corporate bond yield.
	obs, err := fetchFredSeries(ctx, "DAAA", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred hqm corporate bond: %w", err)
	}

	var data []models.InterestRateData
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.InterestRateData{
			Date:     parseFredDate(o.Date),
			Rate:     parseFloat(o.Value),
			RateType: "Moody's AAA Corporate",
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// ---- SpotRate fetcher ----

type spotRateFetcher struct {
	provider.BaseFetcher
}

func newSpotRateFetcher() *spotRateFetcher {
	return &spotRateFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelSpotRate,
			"Treasury spot rates from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *spotRateFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Use Treasury constant maturity as spot rate proxy.
	obs, err := fetchFredSeries(ctx, "DGS10", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred spot rate: %w", err)
	}

	var data []models.InterestRateData
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.InterestRateData{
			Date:     parseFredDate(o.Date),
			Rate:     parseFloat(o.Value),
			RateType: "Spot Rate",
			Maturity: "10Y",
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// ---- CommercialPaper fetcher ----

type commercialPaperFetcher struct {
	provider.BaseFetcher
}

func newCommercialPaperFetcher() *commercialPaperFetcher {
	return &commercialPaperFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelCommercialPaper,
			"Commercial paper rates from FRED",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *commercialPaperFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Fetch 3-Month Financial Commercial Paper Rate.
	obs, err := fetchFredSeries(ctx, "DCPF3M", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred commercial paper: %w", err)
	}

	var data []models.InterestRateData
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.InterestRateData{
			Date:     parseFredDate(o.Date),
			Rate:     parseFloat(o.Value),
			RateType: "Commercial Paper",
			Maturity: "3M",
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// ---- BondIndices fetcher ----

type bondIndicesFetcher struct {
	provider.BaseFetcher
}

func newBondIndicesFetcher() *bondIndicesFetcher {
	return &bondIndicesFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelBondIndices,
			"Bond market indices from FRED (ICE BofA)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *bondIndicesFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// ICE BofA US Corporate Index Option-Adjusted Spread.
	obs, err := fetchFredSeries(ctx, "BAMLC0A0CM", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred bond indices: %w", err)
	}

	var data []models.BondIndex
	for _, o := range obs {
		if o.Value == "." {
			continue
		}
		data = append(data, models.BondIndex{
			Date:      parseFredDate(o.Date),
			IndexName: "ICE BofA US Corporate Index OAS",
			Value:     parseFloat(o.Value),
		})
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}

// ---- MortgageIndices fetcher ----

type mortgageIndicesFetcher struct {
	provider.BaseFetcher
}

func newMortgageIndicesFetcher() *mortgageIndicesFetcher {
	return &mortgageIndicesFetcher{
		BaseFetcher: provider.NewBaseFetcherWithOpts(
			provider.ModelMortgageIndices,
			"Mortgage rate indices from FRED (30Y, 15Y, 5/1 ARM)",
			nil,
			[]string{provider.ParamStartDate, provider.ParamEndDate, provider.ParamLimit},
			15*time.Minute, 10, time.Second,
		),
	}
}

func (f *mortgageIndicesFetcher) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	apiKey := params["_fred_api_key"]

	cacheKey := provider.CacheKey(f.ModelType(), params)
	if cached, ok := f.CacheGet(cacheKey); ok {
		return newCachedResult(cached), nil
	}
	if err := f.RateLimit(ctx); err != nil {
		return nil, err
	}

	// Fetch 30-Year Fixed Rate Mortgage Average.
	obs30, err := fetchFredSeries(ctx, "MORTGAGE30US", apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("fred mortgage 30yr: %w", err)
	}

	// Build 30Y rate map by date.
	rate30ByDate := make(map[string]float64)
	for _, o := range obs30 {
		if o.Value != "." {
			rate30ByDate[o.Date] = parseFloat(o.Value)
		}
	}

	// Fetch 15-Year Fixed.
	obs15, _ := fetchFredSeries(ctx, "MORTGAGE15US", apiKey, params)
	rate15ByDate := make(map[string]float64)
	for _, o := range obs15 {
		if o.Value != "." {
			rate15ByDate[o.Date] = parseFloat(o.Value)
		}
	}

	// Build results keyed on 30Y dates.
	var data []models.MortgageIndex
	for _, o := range obs30 {
		if o.Value == "." {
			continue
		}
		m := models.MortgageIndex{
			Date:          parseFredDate(o.Date),
			Rate30YrFixed: parseFloat(o.Value),
			Source:        "FRED / Freddie Mac",
		}
		if r15, ok := rate15ByDate[o.Date]; ok {
			m.Rate15YrFixed = r15
		}
		data = append(data, m)
	}

	f.CacheSet(cacheKey, data)
	return newResult(data), nil
}
