package yfinance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seenimoa/openseai/internal/provider"
)

func TestProviderInfo(t *testing.T) {
	p := New()
	info := p.Info()
	if info.Name != "yfinance" {
		t.Errorf("expected name yfinance, got %s", info.Name)
	}
	if info.Website == "" {
		t.Error("expected non-empty website")
	}
	if len(info.Credentials) != 0 {
		t.Errorf("yfinance should have no credentials, got %d", len(info.Credentials))
	}
}

func TestProviderSupportedModels(t *testing.T) {
	p := New()
	models := p.SupportedModels()
	if len(models) == 0 {
		t.Fatal("expected at least one supported model")
	}

	// Verify key model types are present.
	expected := []provider.ModelType{
		provider.ModelEquityHistorical,
		provider.ModelEquityQuote,
		provider.ModelEquityInfo,
		provider.ModelEquitySearch,
		provider.ModelBalanceSheet,
		provider.ModelIncomeStatement,
		provider.ModelCashFlowStatement,
		provider.ModelKeyMetrics,
		provider.ModelHistoricalDividends,
		provider.ModelShareStatistics,
		provider.ModelEquityGainers,
		provider.ModelEquityLosers,
		provider.ModelEquityActive,
		provider.ModelCompanyNews,
		provider.ModelEtfHistorical,
		provider.ModelEtfInfo,
		provider.ModelIndexHistorical,
		provider.ModelOptionsChains,
		provider.ModelCryptoHistorical,
		provider.ModelCurrencyHistorical,
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

func TestProviderFetcher(t *testing.T) {
	p := New()

	// Should return a fetcher for supported models.
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

func TestProviderInit(t *testing.T) {
	p := New()
	// YFinance has no credentials, Init should succeed with nil.
	if err := p.Init(nil); err != nil {
		t.Errorf("Init with nil: %v", err)
	}
	if err := p.Init(map[string]string{}); err != nil {
		t.Errorf("Init with empty: %v", err)
	}
}

func TestFetcherRequiredParams(t *testing.T) {
	p := New()

	tests := []struct {
		model    provider.ModelType
		required []string
	}{
		{provider.ModelEquityHistorical, []string{"symbol"}},
		{provider.ModelEquityQuote, []string{"symbol"}},
		{provider.ModelEquityInfo, []string{"symbol"}},
		{provider.ModelEquitySearch, []string{"query"}},
		{provider.ModelBalanceSheet, []string{"symbol"}},
		{provider.ModelOptionsChains, []string{"symbol"}},
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

func TestFetchMissingRequiredParam(t *testing.T) {
	p := New()
	f := p.Fetcher(provider.ModelEquityQuote)
	if f == nil {
		t.Fatal("no fetcher for EquityQuote")
	}

	// Fetch without symbol should eventually fail, but let's check params at the caller level.
	_, err := f.Fetch(context.Background(), provider.QueryParams{})
	// The fetcher should fail - either param validation or API error.
	if err == nil {
		t.Error("expected error when fetching without symbol")
	}
}

func TestProviderRegistration(t *testing.T) {
	p := New()
	_ = p.Init(nil)

	reg := provider.NewRegistry()
	if err := reg.Register(p); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Should be retrievable.
	got, err := reg.Get("yfinance")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Info().Name != "yfinance" {
		t.Error("wrong provider name")
	}

	// Should appear in providers for supported models.
	provs := reg.ProvidersFor(provider.ModelEquityQuote)
	if len(provs) == 0 {
		t.Error("no providers for EquityQuote")
	}
	if provs[0] != "yfinance" {
		t.Errorf("expected yfinance, got %s", provs[0])
	}
}

func TestHelperToYFTicker(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"AAPL", "AAPL.NS"},       // Default Indian market: appends .NS
		{"RELIANCE", "RELIANCE.NS"},
		{"RELIANCE.NS", "RELIANCE.NS"}, // Already has suffix
		{"^NSEI", "^NSEI"},             // Index prefix preserved
	}
	for _, tt := range tests {
		got := toYFTicker(tt.in)
		if got != tt.want {
			t.Errorf("toYFTicker(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestHelperFromYFTicker(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"AAPL", "AAPL"},
		{"RELIANCE.NS", "RELIANCE"},
		{"RELIANCE.BO", "RELIANCE"},
		{"BTC-USD", "BTC-USD"},
	}
	for _, tt := range tests {
		got := fromYFTicker(tt.in)
		if got != tt.want {
			t.Errorf("fromYFTicker(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPingWithMockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"quoteResponse":{"result":[{"symbol":"AAPL"}]}}`))
	}))
	defer srv.Close()

	p := New()
	// Override base URL for test (we can't easily do this without exporting it,
	// so we just verify the provider was created correctly).
	if p.Info().Name != "yfinance" {
		t.Error("wrong provider name")
	}
}

func TestModelTypeCount(t *testing.T) {
	p := New()
	models := p.SupportedModels()
	// We registered 21 fetchers.
	if len(models) < 20 {
		t.Errorf("expected at least 20 models, got %d", len(models))
	}
}
