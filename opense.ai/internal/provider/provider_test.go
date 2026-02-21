package provider

import (
	"context"
	"testing"
	"time"
)

// mockFetcher implements the Fetcher interface for testing.
type mockFetcher struct {
	BaseFetcher
	fetchFn func(ctx context.Context, params QueryParams) (*FetchResult, error)
}

func newMockFetcher(model ModelType, required []string) *mockFetcher {
	return &mockFetcher{
		BaseFetcher: NewBaseFetcher(model, "mock fetcher for "+string(model), required, nil),
	}
}

func (m *mockFetcher) Fetch(ctx context.Context, params QueryParams) (*FetchResult, error) {
	if m.fetchFn != nil {
		return m.fetchFn(ctx, params)
	}
	return &FetchResult{
		Data:      "mock-data",
		FetchedAt: time.Now(),
	}, nil
}

// mockProvider implements the Provider interface for testing.
type mockProvider struct {
	BaseProvider
}

func newMockProvider(name string, models ...ModelType) *mockProvider {
	mp := &mockProvider{
		BaseProvider: NewBaseProvider(name, "Mock "+name, "https://example.com", nil),
	}
	for _, m := range models {
		mp.RegisterFetcher(newMockFetcher(m, []string{ParamSymbol}))
	}
	return mp
}

// --- Registry Tests ---

func TestRegistryRegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	p := newMockProvider("test-provider", ModelEquityQuote, ModelEquityHistorical)

	if err := p.Init(nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := reg.Register(p); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	got, err := reg.Get("test-provider")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Info().Name != "test-provider" {
		t.Errorf("expected name test-provider, got %s", got.Info().Name)
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
	if _, ok := err.(*ErrProviderNotFound); !ok {
		t.Errorf("expected ErrProviderNotFound, got %T", err)
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(newMockProvider("beta", ModelEquityQuote))
	_ = reg.Register(newMockProvider("alpha", ModelEquityHistorical))

	list := reg.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(list))
	}
	// Should be sorted alphabetically.
	if list[0].Name != "alpha" {
		t.Errorf("expected first provider 'alpha', got %s", list[0].Name)
	}
	if list[1].Name != "beta" {
		t.Errorf("expected second provider 'beta', got %s", list[1].Name)
	}
}

func TestRegistryProvidersFor(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(newMockProvider("p1", ModelEquityQuote, ModelBalanceSheet))
	_ = reg.Register(newMockProvider("p2", ModelEquityQuote))
	_ = reg.Register(newMockProvider("p3", ModelBalanceSheet))

	provs := reg.ProvidersFor(ModelEquityQuote)
	if len(provs) != 2 {
		t.Fatalf("expected 2 providers for EquityQuote, got %d", len(provs))
	}

	provs = reg.ProvidersFor(ModelBalanceSheet)
	if len(provs) != 2 {
		t.Fatalf("expected 2 providers for BalanceSheet, got %d", len(provs))
	}

	provs = reg.ProvidersFor(ModelCryptoHistorical)
	if len(provs) != 0 {
		t.Fatalf("expected 0 providers for CryptoHistorical, got %d", len(provs))
	}
}

func TestRegistrySetDefault(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(newMockProvider("p1", ModelEquityQuote))
	_ = reg.Register(newMockProvider("p2", ModelEquityQuote))

	// Default should be p1 (first registered).
	def, ok := reg.DefaultProvider(ModelEquityQuote)
	if !ok || def != "p1" {
		t.Errorf("expected default p1, got %s (ok=%v)", def, ok)
	}

	// Change default.
	if err := reg.SetDefault(ModelEquityQuote, "p2"); err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}
	def, ok = reg.DefaultProvider(ModelEquityQuote)
	if !ok || def != "p2" {
		t.Errorf("expected default p2, got %s (ok=%v)", def, ok)
	}

	// Set default to non-existent provider.
	if err := reg.SetDefault(ModelEquityQuote, "nope"); err == nil {
		t.Error("expected error setting default to non-existent provider")
	}
}

func TestRegistryUnregister(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(newMockProvider("p1", ModelEquityQuote))
	_ = reg.Register(newMockProvider("p2", ModelEquityQuote))

	reg.Unregister("p1")

	_, err := reg.Get("p1")
	if err == nil {
		t.Error("expected error after unregister")
	}

	provs := reg.ProvidersFor(ModelEquityQuote)
	if len(provs) != 1 || provs[0] != "p2" {
		t.Errorf("expected only p2 after unregister, got %v", provs)
	}

	// Default should have shifted to p2.
	def, _ := reg.DefaultProvider(ModelEquityQuote)
	if def != "p2" {
		t.Errorf("expected default to shift to p2, got %s", def)
	}
}

func TestRegistryFetch(t *testing.T) {
	reg := NewRegistry()
	mp := newMockProvider("test", ModelEquityQuote)
	_ = reg.Register(mp)

	ctx := context.Background()
	params := QueryParams{ParamSymbol: "AAPL"}

	result, err := reg.Fetch(ctx, ModelEquityQuote, params)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if result.Provider != "test" {
		t.Errorf("expected provider 'test', got %s", result.Provider)
	}
	if result.Model != ModelEquityQuote {
		t.Errorf("expected model EquityQuote, got %s", result.Model)
	}
	if result.Data != "mock-data" {
		t.Errorf("unexpected data: %v", result.Data)
	}
}

func TestRegistryFetchMissingParam(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(newMockProvider("test", ModelEquityQuote))

	ctx := context.Background()
	params := QueryParams{} // Missing required "symbol" param.

	_, err := reg.Fetch(ctx, ModelEquityQuote, params)
	if err == nil {
		t.Fatal("expected error for missing param")
	}
	if _, ok := err.(*ErrMissingParam); !ok {
		t.Errorf("expected ErrMissingParam, got %T: %v", err, err)
	}
}

func TestRegistryFetchUnsupportedModel(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(newMockProvider("test", ModelEquityQuote))

	ctx := context.Background()
	params := QueryParams{ParamSymbol: "AAPL"}

	_, err := reg.Fetch(ctx, ModelCryptoHistorical, params)
	if err == nil {
		t.Fatal("expected error for unsupported model")
	}
}

func TestRegistryFetchWithProviderOverride(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(newMockProvider("p1", ModelEquityQuote))

	mp2 := newMockProvider("p2", ModelEquityQuote)
	f := newMockFetcher(ModelEquityQuote, []string{ParamSymbol})
	f.fetchFn = func(ctx context.Context, params QueryParams) (*FetchResult, error) {
		return &FetchResult{Data: "from-p2"}, nil
	}
	mp2.BaseProvider.fetchers[ModelEquityQuote] = f
	_ = reg.Register(mp2)

	ctx := context.Background()
	params := QueryParams{
		ParamSymbol:   "AAPL",
		ParamProvider: "p2", // Force provider p2.
	}

	result, err := reg.Fetch(ctx, ModelEquityQuote, params)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if result.Data != "from-p2" {
		t.Errorf("expected data from p2, got %v", result.Data)
	}
}

func TestRegistryFetchWithFallback(t *testing.T) {
	reg := NewRegistry()

	// p1 always fails.
	mp1 := newMockProvider("p1", ModelEquityQuote)
	f1 := newMockFetcher(ModelEquityQuote, []string{ParamSymbol})
	f1.fetchFn = func(ctx context.Context, params QueryParams) (*FetchResult, error) {
		return nil, &ErrModelNotSupported{Provider: "p1", Model: ModelEquityQuote}
	}
	mp1.BaseProvider.fetchers[ModelEquityQuote] = f1
	_ = reg.Register(mp1)

	// p2 succeeds.
	mp2 := newMockProvider("p2", ModelEquityQuote)
	f2 := newMockFetcher(ModelEquityQuote, []string{ParamSymbol})
	f2.fetchFn = func(ctx context.Context, params QueryParams) (*FetchResult, error) {
		return &FetchResult{Data: "fallback-data"}, nil
	}
	mp2.BaseProvider.fetchers[ModelEquityQuote] = f2
	_ = reg.Register(mp2)

	ctx := context.Background()
	params := QueryParams{ParamSymbol: "AAPL"}

	result, err := reg.FetchWithFallback(ctx, ModelEquityQuote, params)
	if err != nil {
		t.Fatalf("FetchWithFallback failed: %v", err)
	}
	if result.Data != "fallback-data" {
		t.Errorf("expected fallback-data, got %v", result.Data)
	}
}

func TestModelCoverage(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(newMockProvider("p1", ModelEquityQuote, ModelBalanceSheet))
	_ = reg.Register(newMockProvider("p2", ModelEquityQuote, ModelCryptoHistorical))

	coverage := reg.ModelCoverage()

	if len(coverage[ModelEquityQuote]) != 2 {
		t.Errorf("expected 2 providers for EquityQuote, got %d", len(coverage[ModelEquityQuote]))
	}
	if len(coverage[ModelBalanceSheet]) != 1 {
		t.Errorf("expected 1 provider for BalanceSheet, got %d", len(coverage[ModelBalanceSheet]))
	}
	if len(coverage[ModelCryptoHistorical]) != 1 {
		t.Errorf("expected 1 provider for CryptoHistorical, got %d", len(coverage[ModelCryptoHistorical]))
	}
}

// --- Base Provider Tests ---

func TestBaseProviderInit(t *testing.T) {
	creds := []ProviderCredential{
		{Name: "api_key", Required: true, EnvVar: "TEST_KEY"},
	}
	bp := NewBaseProvider("test", "desc", "https://test.com", creds)

	// Missing required credential.
	if err := bp.Init(map[string]string{}); err == nil {
		t.Error("expected error for missing required credential")
	}

	// With credential.
	if err := bp.Init(map[string]string{"api_key": "secret123"}); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if bp.Credential("api_key") != "secret123" {
		t.Error("credential not stored")
	}
}

func TestBaseProviderRegisterFetcher(t *testing.T) {
	bp := NewBaseProvider("test", "desc", "https://test.com", nil)
	f := newMockFetcher(ModelEquityQuote, nil)
	bp.RegisterFetcher(f)

	if bp.Fetcher(ModelEquityQuote) == nil {
		t.Error("fetcher not registered")
	}
	if bp.Fetcher(ModelBalanceSheet) != nil {
		t.Error("fetcher should be nil for unregistered model")
	}
	if len(bp.SupportedModels()) != 1 {
		t.Errorf("expected 1 supported model, got %d", len(bp.SupportedModels()))
	}
}

// --- CacheKey Tests ---

func TestCacheKey(t *testing.T) {
	params := QueryParams{
		ParamSymbol:   "AAPL",
		ParamInterval: "1d",
		ParamProvider: "fmp", // Should be excluded.
	}

	key := CacheKey(ModelEquityHistorical, params)

	if key == "" {
		t.Error("cache key should not be empty")
	}
	// Provider should not be in key.
	if contains(key, "fmp") {
		t.Error("cache key should not contain provider name")
	}
	// Should contain model and params.
	if !contains(key, "EquityHistorical") {
		t.Error("cache key should contain model type")
	}
	if !contains(key, "AAPL") {
		t.Error("cache key should contain symbol")
	}
}

// --- ValidateParams Tests ---

func TestValidateParams(t *testing.T) {
	err := ValidateParams(QueryParams{ParamSymbol: "AAPL"}, []string{ParamSymbol})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = ValidateParams(QueryParams{}, []string{ParamSymbol})
	if err == nil {
		t.Error("expected error for missing param")
	}

	err = ValidateParams(QueryParams{ParamSymbol: ""}, []string{ParamSymbol})
	if err == nil {
		t.Error("expected error for empty param")
	}
}

// --- AllModels Tests ---

func TestAllModels(t *testing.T) {
	all := AllModels()
	if len(all) < 160 {
		t.Errorf("expected at least 160 models, got %d", len(all))
	}

	// Check no duplicates.
	seen := make(map[ModelType]bool)
	for _, m := range all {
		if seen[m] {
			t.Errorf("duplicate model type: %s", m)
		}
		seen[m] = true
	}
}

func TestModelCategory(t *testing.T) {
	tests := []struct {
		model    ModelType
		category string
	}{
		{ModelEquityQuote, "Equity / Price"},
		{ModelBalanceSheet, "Equity / Fundamentals"},
		{ModelPriceTarget, "Equity / Estimates"},
		{ModelCalendarEarnings, "Equity / Calendar"},
		{ModelEquityGainers, "Equity / Discovery"},
		{ModelInsiderTrading, "Equity / Ownership"},
		{ModelEquityFTD, "Equity / Shorts"},
		{ModelOptionsChains, "Derivatives / Options"},
		{ModelFuturesHistorical, "Derivatives / Futures"},
		{ModelEtfSearch, "ETF"},
		{ModelIndexHistorical, "Index"},
		{ModelCryptoHistorical, "Crypto"},
		{ModelCurrencyHistorical, "Currency"},
		{ModelCompanyNews, "News"},
		{ModelSOFR, "Fixed Income / Rates"},
		{ModelYieldCurve, "Fixed Income / Government"},
		{ModelCompanyFilings, "Regulators / SEC"},
		{ModelCOT, "Regulators / CFTC"},
		{ModelCommoditySpotPrices, "Commodity"},
	}

	for _, tt := range tests {
		cat := ModelCategory(tt.model)
		if cat != tt.category {
			t.Errorf("ModelCategory(%s) = %q, want %q", tt.model, cat, tt.category)
		}
	}
}

// --- Global Registry Tests ---

func TestGlobalRegistry(t *testing.T) {
	g := Global()
	if g == nil {
		t.Fatal("Global() returned nil")
	}
}

// helper for string containment check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
