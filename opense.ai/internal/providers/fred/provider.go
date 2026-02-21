// Package fred implements the FRED (Federal Reserve Economic Data) provider.
// FRED provides free access to over 800,000 economic time series from dozens
// of sources via the FRED API.
//
// Requires a free API key from https://fred.stlouisfed.org/docs/api/api_key.html
// Rate limit: 120 requests/minute.
// Docs: https://fred.stlouisfed.org/docs/api/fred/
package fred

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/seenimoa/openseai/internal/infra"
	"github.com/seenimoa/openseai/internal/provider"
)

const (
	providerName = "fred"
	baseURL      = "https://api.stlouisfed.org/fred"
	credAPIKey   = "api_key"
)

// Provider implements provider.Provider for FRED.
type Provider struct {
	provider.BaseProvider
	apiKey string
}

// New creates a new FRED provider and registers all fetchers.
func New() *Provider {
	p := &Provider{
		BaseProvider: provider.NewBaseProvider(
			providerName,
			"Federal Reserve Economic Data - 800K+ economic time series",
			"https://fred.stlouisfed.org",
			[]provider.ProviderCredential{
				{
					Name:        credAPIKey,
					Description: "FRED API key from fred.stlouisfed.org",
					Required:    true,
					EnvVar:      "FRED_API_KEY",
				},
			},
		),
	}

	// --- Core FRED ---
	p.RegisterFetcher(newFredSearchFetcher())
	p.RegisterFetcher(newFredSeriesFetcher())
	p.RegisterFetcher(newFredReleaseTableFetcher())
	p.RegisterFetcher(newFredRegionalFetcher())

	// --- Fixed Income / Rates ---
	p.RegisterFetcher(newSOFRFetcher())
	p.RegisterFetcher(newSONIAFetcher())
	p.RegisterFetcher(newAmeriborFetcher())
	p.RegisterFetcher(newFederalFundsRateFetcher())
	p.RegisterFetcher(newProjectionsFetcher())
	p.RegisterFetcher(newIORBFetcher())
	p.RegisterFetcher(newDiscountWindowFetcher())
	p.RegisterFetcher(newOvernightBankFundingFetcher())
	p.RegisterFetcher(newEuroShortTermRateFetcher())
	p.RegisterFetcher(newECBInterestRatesFetcher())

	// --- Economy ---
	p.RegisterFetcher(newConsumerPriceIndexFetcher())
	p.RegisterFetcher(newNonFarmPayrollsFetcher())
	p.RegisterFetcher(newPersonalConsumptionExpendituresFetcher())
	p.RegisterFetcher(newUniversityOfMichiganFetcher())
	p.RegisterFetcher(newManufacturingNYFetcher())
	p.RegisterFetcher(newManufacturingTexasFetcher())
	p.RegisterFetcher(newRetailPricesFetcher())
	p.RegisterFetcher(newCommoditySpotPricesFetcher())
	p.RegisterFetcher(newUnemploymentFetcher())
	p.RegisterFetcher(newGDPRealFetcher())

	// --- Fixed Income / Government & Corporate ---
	p.RegisterFetcher(newYieldCurveFetcher())
	p.RegisterFetcher(newTreasuryConstantMaturityFetcher())
	p.RegisterFetcher(newSelectedTreasuryConstantMaturityFetcher())
	p.RegisterFetcher(newSelectedTreasuryBillFetcher())
	p.RegisterFetcher(newTipsYieldsFetcher())
	p.RegisterFetcher(newHighQualityMarketCorporateBondFetcher())
	p.RegisterFetcher(newSpotRateFetcher())
	p.RegisterFetcher(newCommercialPaperFetcher())
	p.RegisterFetcher(newBondIndicesFetcher())
	p.RegisterFetcher(newMortgageIndicesFetcher())

	return p
}

// Init stores the API key.
func (p *Provider) Init(credentials map[string]string) error {
	if err := p.BaseProvider.Init(credentials); err != nil {
		return err
	}
	p.apiKey = credentials[credAPIKey]
	return nil
}

// Ping checks connectivity to FRED API.
func (p *Provider) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/series?series_id=GDP&api_key=%s&file_type=json", baseURL, p.apiKey)
	body, _, err := infra.DoGet(ctx, url, jsonHeaders())
	if err != nil {
		return fmt.Errorf("fred ping: %w", err)
	}
	body.Close()
	return nil
}

// APIKey returns the stored API key.
func (p *Provider) APIKey() string {
	return p.apiKey
}

// Fetcher overrides BaseProvider.Fetcher to return a wrapper that
// auto-injects the FRED API key into query params before delegating.
func (p *Provider) Fetcher(model provider.ModelType) provider.Fetcher {
	inner := p.BaseProvider.Fetcher(model)
	if inner == nil {
		return nil
	}
	return &apiKeyInjector{inner: inner, apiKey: &p.apiKey}
}

// apiKeyInjector wraps a Fetcher and injects the FRED API key.
type apiKeyInjector struct {
	inner  provider.Fetcher
	apiKey *string
}

func (w *apiKeyInjector) ModelType() provider.ModelType   { return w.inner.ModelType() }
func (w *apiKeyInjector) Description() string             { return w.inner.Description() }
func (w *apiKeyInjector) RequiredParams() []string         { return w.inner.RequiredParams() }
func (w *apiKeyInjector) OptionalParams() []string         { return w.inner.OptionalParams() }

func (w *apiKeyInjector) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	enriched := make(provider.QueryParams, len(params)+1)
	for k, v := range params {
		enriched[k] = v
	}
	enriched["_fred_api_key"] = *w.apiKey
	return w.inner.Fetch(ctx, enriched)
}

// --- Shared helpers ---

func jsonHeaders() map[string]string {
	return map[string]string{"Accept": "application/json"}
}

// fredURL builds a full FRED API URL with api_key and file_type=json appended.
func fredURL(endpoint, apiKey string) string {
	sep := "?"
	if containsQuery(endpoint) {
		sep = "&"
	}
	return baseURL + "/" + endpoint + sep + "api_key=" + apiKey + "&file_type=json"
}

func containsQuery(s string) bool {
	for _, c := range s {
		if c == '?' {
			return true
		}
	}
	return false
}

// fetchFredJSON performs a GET request to FRED API and decodes JSON.
func fetchFredJSON(ctx context.Context, endpoint, apiKey string, dest any) error {
	url := fredURL(endpoint, apiKey)
	body, _, err := infra.DoGet(ctx, url, jsonHeaders())
	if err != nil {
		return err
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read FRED response: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("parse FRED JSON: %w", err)
	}
	return nil
}

// fetchFredSeries is a convenience function that fetches a FRED series by ID
// and returns the observations as a slice of FREDSeriesData.
func fetchFredSeries(ctx context.Context, seriesID, apiKey string, params provider.QueryParams) ([]fredObservation, error) {
	endpoint := fmt.Sprintf("series/observations?series_id=%s", seriesID)
	if sd := params[provider.ParamStartDate]; sd != "" {
		endpoint += "&observation_start=" + sd
	}
	if ed := params[provider.ParamEndDate]; ed != "" {
		endpoint += "&observation_end=" + ed
	}
	if lim := params[provider.ParamLimit]; lim != "" {
		endpoint += "&limit=" + lim
	}

	var resp fredObservationsResponse
	if err := fetchFredJSON(ctx, endpoint, apiKey, &resp); err != nil {
		return nil, err
	}
	return resp.Observations, nil
}

func newResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
	}
}

func newCachedResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
		Cached:    true,
	}
}

func defaultDateRange(params provider.QueryParams) (string, string) {
	now := time.Now()
	startStr := params[provider.ParamStartDate]
	if startStr == "" {
		startStr = now.AddDate(-1, 0, 0).Format("2006-01-02")
	}
	endStr := params[provider.ParamEndDate]
	if endStr == "" {
		endStr = now.Format("2006-01-02")
	}
	return startStr, endStr
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}

func parseFloat(s string) float64 {
	if s == "" || s == "." {
		return 0
	}
	var result float64
	var decimal float64 = 1
	pastDot := false
	neg := false
	for i, c := range s {
		if c == '-' && i == 0 {
			neg = true
			continue
		}
		if c == '.' {
			pastDot = true
			continue
		}
		if c >= '0' && c <= '9' {
			if pastDot {
				decimal *= 10
				result += float64(c-'0') / decimal
			} else {
				result = result*10 + float64(c-'0')
			}
		}
	}
	if neg {
		result = -result
	}
	return result
}
