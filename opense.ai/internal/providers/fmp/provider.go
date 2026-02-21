// Package fmp implements the Financial Modeling Prep (FMP) data provider.
// FMP offers comprehensive financial data via a REST API with API key authentication.
// It covers equities, fundamentals, estimates, ETFs, crypto, forex, calendar,
// index, news, ESG, government trades, treasury, and yield curves.
//
// Free tier: 250 requests/day.
// Docs: https://financialmodelingprep.com/developer/docs
package fmp

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
	providerName = "fmp"
	baseURL      = "https://financialmodelingprep.com/api/v3"
	credAPIKey   = "api_key"
)

// Provider implements provider.Provider for FMP.
type Provider struct {
	provider.BaseProvider
	apiKey string
}

// New creates a new FMP provider and registers all fetchers.
func New() *Provider {
	p := &Provider{
		BaseProvider: provider.NewBaseProvider(
			providerName,
			"Financial Modeling Prep - comprehensive financial data",
			"https://financialmodelingprep.com",
			[]provider.ProviderCredential{
				{
					Name:        credAPIKey,
					Description: "FMP API key from financialmodelingprep.com",
					Required:    true,
					EnvVar:      "FMP_API_KEY",
				},
			},
		),
	}

	// --- Equity / Price ---
	p.RegisterFetcher(newEquityHistoricalFetcher())
	p.RegisterFetcher(newEquityQuoteFetcher())
	p.RegisterFetcher(newEquityInfoFetcher())
	p.RegisterFetcher(newEquitySearchFetcher())
	p.RegisterFetcher(newEquityScreenerFetcher())
	p.RegisterFetcher(newEquityPeersFetcher())
	p.RegisterFetcher(newPricePerformanceFetcher())
	p.RegisterFetcher(newMarketSnapshotsFetcher())

	// --- Equity / Fundamentals ---
	p.RegisterFetcher(newBalanceSheetFetcher())
	p.RegisterFetcher(newIncomeStatementFetcher())
	p.RegisterFetcher(newCashFlowStatementFetcher())
	p.RegisterFetcher(newKeyMetricsFetcher())
	p.RegisterFetcher(newFinancialRatiosFetcher())
	p.RegisterFetcher(newKeyExecutivesFetcher())
	p.RegisterFetcher(newHistoricalDividendsFetcher())
	p.RegisterFetcher(newShareStatisticsFetcher())

	// --- Equity / Estimates ---
	p.RegisterFetcher(newPriceTargetFetcher())
	p.RegisterFetcher(newPriceTargetConsensusFetcher())
	p.RegisterFetcher(newAnalystEstimatesFetcher())

	// --- Equity / Discovery ---
	p.RegisterFetcher(newEquityGainersFetcher())
	p.RegisterFetcher(newEquityLosersFetcher())
	p.RegisterFetcher(newEquityActiveFetcher())

	// --- Equity / Calendar ---
	p.RegisterFetcher(newCalendarEarningsFetcher())
	p.RegisterFetcher(newCalendarDividendFetcher())
	p.RegisterFetcher(newCalendarIpoFetcher())

	// --- ETF ---
	p.RegisterFetcher(newEtfHistoricalFetcher())
	p.RegisterFetcher(newEtfInfoFetcher())

	// --- Index ---
	p.RegisterFetcher(newIndexHistoricalFetcher())

	// --- Crypto ---
	p.RegisterFetcher(newCryptoHistoricalFetcher())

	// --- Currency ---
	p.RegisterFetcher(newCurrencyHistoricalFetcher())

	// --- News ---
	p.RegisterFetcher(newCompanyNewsFetcher())
	p.RegisterFetcher(newWorldNewsFetcher())

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

// Ping checks connectivity to FMP.
func (p *Provider) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/quote/AAPL?apikey=%s", baseURL, p.apiKey)
	body, _, err := infra.DoGet(ctx, url, jsonHeaders())
	if err != nil {
		return fmt.Errorf("fmp ping: %w", err)
	}
	body.Close()
	return nil
}

// APIKey returns the stored API key (used by fetchers).
func (p *Provider) APIKey() string {
	return p.apiKey
}

// Fetcher overrides BaseProvider.Fetcher to return a wrapper that
// auto-injects the FMP API key into query params before delegating.
func (p *Provider) Fetcher(model provider.ModelType) provider.Fetcher {
	inner := p.BaseProvider.Fetcher(model)
	if inner == nil {
		return nil
	}
	return &apiKeyInjector{inner: inner, apiKey: &p.apiKey}
}

// apiKeyInjector wraps a Fetcher and injects the FMP API key.
type apiKeyInjector struct {
	inner  provider.Fetcher
	apiKey *string
}

func (w *apiKeyInjector) ModelType() provider.ModelType     { return w.inner.ModelType() }
func (w *apiKeyInjector) Description() string               { return w.inner.Description() }
func (w *apiKeyInjector) RequiredParams() []string           { return w.inner.RequiredParams() }
func (w *apiKeyInjector) OptionalParams() []string           { return w.inner.OptionalParams() }

func (w *apiKeyInjector) Fetch(ctx context.Context, params provider.QueryParams) (*provider.FetchResult, error) {
	// Inject API key so fetchers don't need to know about credential management.
	enriched := make(provider.QueryParams, len(params)+1)
	for k, v := range params {
		enriched[k] = v
	}
	enriched["_fmp_api_key"] = *w.apiKey
	return w.inner.Fetch(ctx, enriched)
}

// --- Shared helpers ---

func jsonHeaders() map[string]string {
	return map[string]string{"Accept": "application/json"}
}

// fmpURL builds a full FMP API URL with the API key appended.
func fmpURL(path, apiKey string) string {
	sep := "?"
	if len(path) > 0 && path[len(path)-1] == '?' || containsQuery(path) {
		sep = "&"
	}
	return baseURL + path + sep + "apikey=" + apiKey
}

func containsQuery(s string) bool {
	for _, c := range s {
		if c == '?' {
			return true
		}
	}
	return false
}

// fetchFMPJSON performs a GET request to FMP and decodes the response.
func fetchFMPJSON(ctx context.Context, path, apiKey string, dest any) error {
	url := fmpURL(path, apiKey)
	body, _, err := infra.DoGet(ctx, url, jsonHeaders())
	if err != nil {
		return err
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("parse FMP JSON: %w", err)
	}
	return nil
}

// newResult creates a FetchResult.
func newResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
	}
}

// newCachedResult creates a cached FetchResult.
func newCachedResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
		Cached:    true,
	}
}

// defaultDateRange parses start_date/end_date from params or uses defaults.
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
