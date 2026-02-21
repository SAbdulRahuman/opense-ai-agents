package sec

import (
	"testing"

	"github.com/seenimoa/openseai/internal/provider"
)

func TestProviderInfo(t *testing.T) {
	p := New()
	info := p.Info()
	if info.Name != "sec" {
		t.Errorf("expected name sec, got %s", info.Name)
	}
	if info.Website == "" {
		t.Error("expected non-empty website")
	}
	// SEC has no credentials required.
	if len(info.Credentials) != 0 {
		t.Errorf("expected 0 credentials, got %d", len(info.Credentials))
	}
}

func TestProviderSupportedModels(t *testing.T) {
	p := New()
	models := p.SupportedModels()
	if len(models) == 0 {
		t.Fatal("expected at least one supported model")
	}

	expected := []provider.ModelType{
		provider.ModelCompanyFilings,
		provider.ModelSecFiling,
		provider.ModelLatestFinancialReports,
		provider.ModelRssLitigation,
		provider.ModelCikMap,
		provider.ModelSymbolMap,
		provider.ModelSicSearch,
		provider.ModelInstitutionsSearch,
		provider.ModelEquitySearch,
		provider.ModelInsiderTrading,
		provider.ModelInstitutionalOwnership,
		provider.ModelEquityFTD,
		provider.ModelCompareCompanyFacts,
		provider.ModelForm13FHR,
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

	t.Logf("SEC provider supports %d models", len(models))
}

func TestProviderInitNoCredentials(t *testing.T) {
	p := New()
	err := p.Init(nil)
	if err != nil {
		t.Fatalf("Init with nil credentials should succeed: %v", err)
	}

	err = p.Init(map[string]string{})
	if err != nil {
		t.Fatalf("Init with empty credentials should succeed: %v", err)
	}
}

func TestFetcherReturned(t *testing.T) {
	p := New()

	f := p.Fetcher(provider.ModelCompanyFilings)
	if f == nil {
		t.Fatal("expected non-nil fetcher for CompanyFilings")
	}
	if f.ModelType() != provider.ModelCompanyFilings {
		t.Errorf("expected ModelCompanyFilings, got %s", f.ModelType())
	}

	f = p.Fetcher(provider.ModelSecFiling)
	if f == nil {
		t.Fatal("expected non-nil fetcher for SecFiling")
	}

	f = p.Fetcher(provider.ModelInsiderTrading)
	if f == nil {
		t.Fatal("expected non-nil fetcher for InsiderTrading")
	}

	// Should return nil for unsupported models.
	f = p.Fetcher(provider.ModelType("Nonexistent"))
	if f != nil {
		t.Error("expected nil fetcher for unsupported model")
	}
}

func TestFetcherRequiredParams(t *testing.T) {
	p := New()

	tests := []struct {
		model    provider.ModelType
		required []string
	}{
		{provider.ModelCompanyFilings, []string{provider.ParamQuery}},
		{provider.ModelSecFiling, []string{provider.ParamSymbol}},
		{provider.ModelSymbolMap, []string{provider.ParamSymbol}},
		{provider.ModelSicSearch, []string{provider.ParamQuery}},
		{provider.ModelInstitutionsSearch, []string{provider.ParamQuery}},
		{provider.ModelEquitySearch, []string{provider.ParamQuery}},
		{provider.ModelInsiderTrading, []string{provider.ParamSymbol}},
		{provider.ModelInstitutionalOwnership, []string{provider.ParamSymbol}},
		{provider.ModelEquityFTD, []string{provider.ParamSymbol}},
		{provider.ModelCompareCompanyFacts, []string{provider.ParamSymbol}},
		{provider.ModelForm13FHR, []string{provider.ParamSymbol}},
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

func TestPadCIK(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"320193", "0000320193"},
		{"0000320193", "0000320193"},
		{"1", "0000000001"},
		{"12345678901", "12345678901"}, // Already longer
	}
	for _, tt := range tests {
		got := padCIK(tt.input)
		if got != tt.expected {
			t.Errorf("padCIK(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"12345", true},
		{"0", true},
		{"", false},
		{"abc", false},
		{"12a34", false},
	}
	for _, tt := range tests {
		got := isNumeric(tt.input)
		if got != tt.expected {
			t.Errorf("isNumeric(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestParseSECDate(t *testing.T) {
	tests := []struct {
		input string
		year  int
		month int
		day   int
	}{
		{"2024-01-15", 2024, 1, 15},
		{"2023-12-31", 2023, 12, 31},
		{"01/15/2024", 2024, 1, 15},
	}
	for _, tt := range tests {
		got := parseSECDate(tt.input)
		if got.Year() != tt.year || int(got.Month()) != tt.month || got.Day() != tt.day {
			t.Errorf("parseSECDate(%q) = %v, want %d-%02d-%02d", tt.input, got, tt.year, tt.month, tt.day)
		}
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"100", 100},
		{"0", 0},
		{"25abc", 25},
		{"", 0},
	}
	for _, tt := range tests {
		got := parseInt(tt.input)
		if got != tt.expected {
			t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}
