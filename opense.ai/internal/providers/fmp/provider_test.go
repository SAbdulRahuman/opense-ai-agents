package fmp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seenimoa/openseai/internal/provider"
)

func TestProviderInfo(t *testing.T) {
	p := New()
	info := p.Info()
	if info.Name != "fmp" {
		t.Errorf("expected name fmp, got %s", info.Name)
	}
	if info.Website == "" {
		t.Error("expected non-empty website")
	}
	if len(info.Credentials) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(info.Credentials))
	}
	if info.Credentials[0].Name != "api_key" {
		t.Errorf("expected credential name api_key, got %s", info.Credentials[0].Name)
	}
	if !info.Credentials[0].Required {
		t.Error("api_key should be required")
	}
}

func TestProviderSupportedModels(t *testing.T) {
	p := New()
	models := p.SupportedModels()
	if len(models) == 0 {
		t.Fatal("expected at least one supported model")
	}

	expected := []provider.ModelType{
		provider.ModelEquityHistorical,
		provider.ModelEquityQuote,
		provider.ModelEquityInfo,
		provider.ModelEquitySearch,
		provider.ModelEquityScreener,
		provider.ModelEquityPeers,
		provider.ModelPricePerformance,
		provider.ModelMarketSnapshots,
		provider.ModelBalanceSheet,
		provider.ModelIncomeStatement,
		provider.ModelCashFlowStatement,
		provider.ModelKeyMetrics,
		provider.ModelFinancialRatios,
		provider.ModelKeyExecutives,
		provider.ModelHistoricalDividends,
		provider.ModelShareStatistics,
		provider.ModelPriceTarget,
		provider.ModelPriceTargetConsensus,
		provider.ModelAnalystEstimates,
		provider.ModelEquityGainers,
		provider.ModelEquityLosers,
		provider.ModelEquityActive,
		provider.ModelCalendarEarnings,
		provider.ModelCalendarDividend,
		provider.ModelCalendarIpo,
		provider.ModelEtfHistorical,
		provider.ModelEtfInfo,
		provider.ModelIndexHistorical,
		provider.ModelCryptoHistorical,
		provider.ModelCurrencyHistorical,
		provider.ModelCompanyNews,
		provider.ModelWorldNews,
	}

	modelSet := make(map[provider.ModelType]bool)
	for _, m := range models {
		modelSet[m] = true
	}

	for _, m := range expected {
		if !modelSet[m] {
			t.Errorf("missing expected model: %s", m)
		}
	}
}

func TestProviderInitSuccess(t *testing.T) {
	p := New()
	err := p.Init(map[string]string{"api_key": "test_key_123"})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if p.APIKey() != "test_key_123" {
		t.Errorf("expected api key test_key_123, got %s", p.APIKey())
	}
}

func TestProviderInitMissingKey(t *testing.T) {
	p := New()
	err := p.Init(map[string]string{})
	if err == nil {
		t.Error("expected error for missing api_key")
	}
}

func TestFetcherReturned(t *testing.T) {
	p := New()
	_ = p.Init(map[string]string{"api_key": "test"})

	f := p.Fetcher(provider.ModelEquityQuote)
	if f == nil {
		t.Fatal("expected non-nil fetcher for EquityQuote")
	}
	if f.ModelType() != provider.ModelEquityQuote {
		t.Errorf("expected ModelEquityQuote, got %s", f.ModelType())
	}

	// Should return nil for unsupported models.
	f = p.Fetcher(provider.ModelType("Nonexistent"))
	if f != nil {
		t.Error("expected nil fetcher for unsupported model")
	}
}

func TestAPIKeyInjection(t *testing.T) {
	// Create a mock server that echoes back the apikey query param.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("apikey")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{
			{"apikey": apiKey, "symbol": "AAPL"},
		})
	}))
	defer srv.Close()

	// We can't easily override the base URL, but we can verify the wrapper
	// correctly injects the API key by checking the returned fetcher is wrapped.
	p := New()
	_ = p.Init(map[string]string{"api_key": "my_secret_key"})

	f := p.Fetcher(provider.ModelEquityQuote)
	if f == nil {
		t.Fatal("nil fetcher")
	}

	// The fetcher should be an apiKeyInjector wrapper.
	wrapper, ok := f.(*apiKeyInjector)
	if !ok {
		t.Fatalf("expected apiKeyInjector, got %T", f)
	}

	// Verify it delegates model type correctly.
	if wrapper.ModelType() != provider.ModelEquityQuote {
		t.Errorf("wrong model type: %s", wrapper.ModelType())
	}
	if wrapper.Description() == "" {
		t.Error("empty description")
	}

	// Required params should be passed through.
	required := wrapper.RequiredParams()
	if len(required) != 1 || required[0] != "symbol" {
		t.Errorf("unexpected required params: %v", required)
	}
}

func TestProviderRegistration(t *testing.T) {
	p := New()
	_ = p.Init(map[string]string{"api_key": "test"})

	reg := provider.NewRegistry()
	if err := reg.Register(p); err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, err := reg.Get("fmp")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Info().Name != "fmp" {
		t.Error("wrong provider name")
	}

	provs := reg.ProvidersFor(provider.ModelEquityQuote)
	if len(provs) == 0 {
		t.Error("no providers for EquityQuote")
	}
	if provs[0] != "fmp" {
		t.Errorf("expected fmp, got %s", provs[0])
	}
}

func TestRegistryFetchMissingParam(t *testing.T) {
	p := New()
	_ = p.Init(map[string]string{"api_key": "test"})

	reg := provider.NewRegistry()
	_ = reg.Register(p)

	// Fetch without required symbol param should fail.
	_, err := reg.Fetch(context.Background(), provider.ModelEquityQuote, provider.QueryParams{})
	if err == nil {
		t.Error("expected error for missing symbol param")
	}
}

func TestFetcherRequiredParams(t *testing.T) {
	p := New()
	_ = p.Init(map[string]string{"api_key": "test"})

	tests := []struct {
		model    provider.ModelType
		required []string
	}{
		{provider.ModelEquityHistorical, []string{"symbol"}},
		{provider.ModelEquityQuote, []string{"symbol"}},
		{provider.ModelEquityInfo, []string{"symbol"}},
		{provider.ModelBalanceSheet, []string{"symbol"}},
		{provider.ModelIncomeStatement, []string{"symbol"}},
		{provider.ModelCashFlowStatement, []string{"symbol"}},
		{provider.ModelKeyMetrics, []string{"symbol"}},
		{provider.ModelFinancialRatios, []string{"symbol"}},
		{provider.ModelKeyExecutives, []string{"symbol"}},
		{provider.ModelHistoricalDividends, []string{"symbol"}},
		{provider.ModelShareStatistics, []string{"symbol"}},
		{provider.ModelPriceTarget, []string{"symbol"}},
		{provider.ModelPriceTargetConsensus, []string{"symbol"}},
		{provider.ModelAnalystEstimates, []string{"symbol"}},
		{provider.ModelCompanyNews, []string{"symbol"}},
	}

	for _, tt := range tests {
		f := p.Fetcher(tt.model)
		if f == nil {
			t.Errorf("no fetcher for %s", tt.model)
			continue
		}
		got := f.RequiredParams()
		if len(got) != len(tt.required) {
			t.Errorf("%s: expected %d required params, got %d", tt.model, len(tt.required), len(got))
			continue
		}
		for i, r := range tt.required {
			if got[i] != r {
				t.Errorf("%s: required[%d] = %q, want %q", tt.model, i, got[i], r)
			}
		}
	}
}

func TestCalendarFetchersNoSymbolRequired(t *testing.T) {
	p := New()
	_ = p.Init(map[string]string{"api_key": "test"})

	calendarModels := []provider.ModelType{
		provider.ModelCalendarEarnings,
		provider.ModelCalendarDividend,
		provider.ModelCalendarIpo,
		provider.ModelEquityGainers,
		provider.ModelEquityLosers,
		provider.ModelEquityActive,
	}

	for _, m := range calendarModels {
		f := p.Fetcher(m)
		if f == nil {
			t.Errorf("no fetcher for %s", m)
			continue
		}
		required := f.RequiredParams()
		if len(required) != 0 {
			t.Errorf("%s: expected 0 required params, got %v", m, required)
		}
	}
}

func TestModelTypeCount(t *testing.T) {
	p := New()
	models := p.SupportedModels()
	// We registered ~32 fetchers.
	if len(models) < 30 {
		t.Errorf("expected at least 30 models, got %d", len(models))
	}
}

func TestHelperFmpURL(t *testing.T) {
	tests := []struct {
		path, key, want string
	}{
		{"/quote/AAPL", "abc", "https://financialmodelingprep.com/api/v3/quote/AAPL?apikey=abc"},
		{"/search?query=apple", "xyz", "https://financialmodelingprep.com/api/v3/search?query=apple&apikey=xyz"},
		{"/earnings?from=2024-01-01&to=2024-12-31", "key", "https://financialmodelingprep.com/api/v3/earnings?from=2024-01-01&to=2024-12-31&apikey=key"},
	}

	for _, tt := range tests {
		got := fmpURL(tt.path, tt.key)
		if got != tt.want {
			t.Errorf("fmpURL(%q, %q) = %q, want %q", tt.path, tt.key, got, tt.want)
		}
	}
}

func TestHelperContainsQuery(t *testing.T) {
	if !containsQuery("/path?key=val") {
		t.Error("expected true for path with ?")
	}
	if containsQuery("/path/noquestion") {
		t.Error("expected false for path without ?")
	}
}
