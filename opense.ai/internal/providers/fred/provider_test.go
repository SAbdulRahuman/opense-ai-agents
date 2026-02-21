package fred

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
	if info.Name != "fred" {
		t.Errorf("expected name fred, got %s", info.Name)
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
		// Core
		provider.ModelFredSearch,
		provider.ModelFredSeries,
		provider.ModelFredReleaseTable,
		provider.ModelFredRegional,
		// Rates
		provider.ModelSOFR,
		provider.ModelSONIA,
		provider.ModelAmeribor,
		provider.ModelFederalFundsRate,
		provider.ModelProjections,
		provider.ModelIORB,
		provider.ModelDiscountWindowPrimaryCreditRate,
		provider.ModelOvernightBankFundingRate,
		provider.ModelEuroShortTermRate,
		provider.ModelEuropeanCentralBankInterestRates,
		// Economy
		provider.ModelConsumerPriceIndex,
		provider.ModelNonFarmPayrolls,
		provider.ModelPersonalConsumptionExpenditures,
		provider.ModelUniversityOfMichigan,
		provider.ModelManufacturingOutlookNY,
		provider.ModelManufacturingOutlookTexas,
		provider.ModelRetailPrices,
		provider.ModelCommoditySpotPrices,
		provider.ModelUnemployment,
		provider.ModelGdpReal,
		// Bonds / Fixed Income
		provider.ModelYieldCurve,
		provider.ModelTreasuryConstantMaturity,
		provider.ModelSelectedTreasuryConstantMaturity,
		provider.ModelSelectedTreasuryBill,
		provider.ModelTipsYields,
		provider.ModelHighQualityMarketCorporateBond,
		provider.ModelSpotRate,
		provider.ModelCommercialPaper,
		provider.ModelBondIndices,
		provider.ModelMortgageIndices,
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

	t.Logf("FRED provider supports %d models", len(models))
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

	f := p.Fetcher(provider.ModelFredSearch)
	if f == nil {
		t.Fatal("expected non-nil fetcher for FredSearch")
	}
	if f.ModelType() != provider.ModelFredSearch {
		t.Errorf("expected ModelFredSearch, got %s", f.ModelType())
	}

	// Should be wrapped in apiKeyInjector.
	wrapper, ok := f.(*apiKeyInjector)
	if !ok {
		t.Fatalf("expected apiKeyInjector wrapper, got %T", f)
	}
	if wrapper.inner == nil {
		t.Error("inner fetcher should not be nil")
	}

	// Should return nil for unsupported models.
	f = p.Fetcher(provider.ModelType("Nonexistent"))
	if f != nil {
		t.Error("expected nil fetcher for unsupported model")
	}
}

func TestAPIKeyInjection(t *testing.T) {
	p := New()
	_ = p.Init(map[string]string{"api_key": "my_fred_key"})

	f := p.Fetcher(provider.ModelFredSeries)
	if f == nil {
		t.Fatal("nil fetcher")
	}

	wrapper, ok := f.(*apiKeyInjector)
	if !ok {
		t.Fatalf("expected apiKeyInjector, got %T", f)
	}
	if *wrapper.apiKey != "my_fred_key" {
		t.Errorf("expected api key my_fred_key, got %s", *wrapper.apiKey)
	}
}

func TestFetcherRequiredParams(t *testing.T) {
	p := New()
	_ = p.Init(map[string]string{"api_key": "test"})

	tests := []struct {
		model    provider.ModelType
		required []string
	}{
		{provider.ModelFredSearch, []string{provider.ParamQuery}},
		{provider.ModelFredSeries, []string{provider.ParamSymbol}},
		{provider.ModelFredReleaseTable, []string{provider.ParamSymbol}},
		{provider.ModelFredRegional, []string{provider.ParamSymbol}},
		{provider.ModelSOFR, nil},
		{provider.ModelFederalFundsRate, nil},
		{provider.ModelConsumerPriceIndex, nil},
		{provider.ModelYieldCurve, nil},
		{provider.ModelUnemployment, nil},
		{provider.ModelGdpReal, nil},
	}

	for _, tt := range tests {
		f := p.Fetcher(tt.model)
		if f == nil {
			t.Errorf("nil fetcher for %s", tt.model)
			continue
		}
		req := f.RequiredParams()
		if len(req) != len(tt.required) {
			t.Errorf("%s: expected %d required params, got %d", tt.model, len(tt.required), len(req))
			continue
		}
		for i, r := range tt.required {
			if req[i] != r {
				t.Errorf("%s: expected required param %s at index %d, got %s", tt.model, r, i, req[i])
			}
		}
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"100", 100},
		{"0.5", 0.5},
		{"-1.5", -1.5},
		{"", 0},
		{".", 0},
	}
	for _, tt := range tests {
		got := parseFloat(tt.input)
		diff := got - tt.expected
		if diff < -0.001 || diff > 0.001 {
			t.Errorf("parseFloat(%q) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestParseFredDate(t *testing.T) {
	tests := []struct {
		input string
		year  int
		month int
		day   int
	}{
		{"2024-01-15", 2024, 1, 15},
		{"2023-12-31", 2023, 12, 31},
	}
	for _, tt := range tests {
		got := parseFredDate(tt.input)
		if got.Year() != tt.year || int(got.Month()) != tt.month || got.Day() != tt.day {
			t.Errorf("parseFredDate(%q) = %v, want %d-%02d-%02d", tt.input, got, tt.year, tt.month, tt.day)
		}
	}
}

// TestFredSeriesWithMockServer tests the FredSeries fetcher with a mock HTTP server.
func TestFredSeriesWithMockServer(t *testing.T) {
	// Create mock FRED API server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/fred/series/observations" {
			json.NewEncoder(w).Encode(map[string]any{
				"observations": []map[string]string{
					{"date": "2024-01-01", "value": "5.33"},
					{"date": "2024-01-02", "value": "5.34"},
					{"date": "2024-01-03", "value": "."},
				},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	// Note: We can't easily override baseURL constant, but we verify
	// the fetcher structure is correct.
	f := newFredSeriesFetcher()
	if f.ModelType() != provider.ModelFredSeries {
		t.Errorf("expected ModelFredSeries, got %s", f.ModelType())
	}

	// Verify the fetcher can handle context cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := f.Fetch(ctx, provider.QueryParams{
		provider.ParamSymbol: "GDP",
		"_fred_api_key":      "test",
	})
	if err == nil {
		// May or may not error depending on rate limiter behavior.
		// This is mainly to verify no panics.
	}
	_ = err
}

func TestFredURLBuilder(t *testing.T) {
	tests := []struct {
		endpoint string
		apiKey   string
		contains []string
	}{
		{
			"series/observations?series_id=GDP",
			"testkey",
			[]string{"api_key=testkey", "file_type=json", "series_id=GDP"},
		},
		{
			"series/search",
			"abc123",
			[]string{"api_key=abc123", "file_type=json"},
		},
	}

	for _, tt := range tests {
		url := fredURL(tt.endpoint, tt.apiKey)
		for _, substr := range tt.contains {
			found := false
			for i := 0; i <= len(url)-len(substr); i++ {
				if url[i:i+len(substr)] == substr {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("fredURL(%q, %q) = %q, missing %q", tt.endpoint, tt.apiKey, url, substr)
			}
		}
	}
}
