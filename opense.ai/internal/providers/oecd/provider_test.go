package oecd

import (
	"context"
	"testing"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// Provider-level tests
// ---------------------------------------------------------------------------

func TestProviderInfo(t *testing.T) {
	p := New()
	info := p.Info()
	if info.Name != "oecd" {
		t.Errorf("expected name oecd, got %s", info.Name)
	}
	if info.Website == "" {
		t.Error("expected non-empty website")
	}
	if len(info.Credentials) != 0 {
		t.Errorf("oecd should have no credentials, got %d", len(info.Credentials))
	}
}

func TestProviderInit(t *testing.T) {
	p := New()
	if err := p.Init(nil); err != nil {
		t.Errorf("Init with nil: %v", err)
	}
}

func TestProviderSupportedModels(t *testing.T) {
	p := New()
	ms := p.SupportedModels()

	if len(ms) != 9 {
		t.Errorf("expected 9 supported models, got %d: %v", len(ms), ms)
	}

	expected := []provider.ModelType{
		provider.ModelCompositeLeadingIndicator,
		provider.ModelConsumerPriceIndex,
		provider.ModelCountryInterestRates,
		provider.ModelGdpNominal,
		provider.ModelGdpReal,
		provider.ModelGdpForecast,
		provider.ModelHousePriceIndex,
		provider.ModelSharePriceIndex,
		provider.ModelUnemployment,
	}

	modelSet := make(map[provider.ModelType]bool)
	for _, m := range ms {
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
	for _, m := range p.SupportedModels() {
		f := p.Fetcher(m)
		if f == nil {
			t.Errorf("nil fetcher for model %s", m)
			continue
		}
		if f.ModelType() != m {
			t.Errorf("fetcher model mismatch: expected %s, got %s", m, f.ModelType())
		}
		if f.Description() == "" {
			t.Errorf("empty description for model %s", m)
		}
	}
}

func TestFetcherRegistration(t *testing.T) {
	p := New()
	reg := provider.NewRegistry()
	if err := p.Init(nil); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(p); err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	prov, err := reg.Get("oecd")
	if err != nil {
		t.Fatalf("get provider: %v", err)
	}
	if prov.Info().Name != "oecd" {
		t.Errorf("unexpected name: %s", prov.Info().Name)
	}
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestParseSDMXDate(t *testing.T) {
	tests := []struct {
		input string
		year  int
		month int
		day   int
	}{
		{"2024-01-15", 2024, 1, 15},
		{"2024-01", 2024, 1, 1},
		{"2024", 2024, 1, 1},
		{"2023-Q1", 2023, 1, 1},
		{"2023-Q2", 2023, 4, 1},
		{"2023-Q3", 2023, 7, 1},
		{"2023-Q4", 2023, 10, 1},
	}
	for _, tt := range tests {
		got := parseSDMXDate(tt.input)
		if got.Year() != tt.year || int(got.Month()) != tt.month || got.Day() != tt.day {
			t.Errorf("parseSDMXDate(%q) = %v, want %d-%d-%d", tt.input, got, tt.year, tt.month, tt.day)
		}
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5.33", 5.33},
		{"0", 0},
		{"", 0},
		{"NaN", 0},
		{"NA", 0},
		{"3.14", 3.14},
	}
	for _, tt := range tests {
		got := parseFloat(tt.input)
		if got != tt.expected {
			t.Errorf("parseFloat(%q) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestFindColumn(t *testing.T) {
	header := []string{"REF_AREA", "TIME_PERIOD", "OBS_VALUE", "MEASURE"}
	if idx := findColumn(header, "OBS_VALUE"); idx != 2 {
		t.Errorf("expected 2, got %d", idx)
	}
	if idx := findColumn(header, "obs_value"); idx != 2 {
		t.Errorf("case-insensitive: expected 2, got %d", idx)
	}
	if idx := findColumn(header, "MISSING"); idx != -1 {
		t.Errorf("expected -1 for missing column, got %d", idx)
	}
}

func TestBuildURL(t *testing.T) {
	url := buildURL("OECD.SDD.STES,DSD_STES@DF_CLI,4.1", "USA.M", "2024-01", "2024-06")
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	if !contains(url, "sdmx.oecd.org") {
		t.Errorf("expected sdmx.oecd.org in URL, got %s", url)
	}
	if !contains(url, "startPeriod=2024-01") {
		t.Errorf("expected start period in URL, got %s", url)
	}
	if !contains(url, "endPeriod=2024-06") {
		t.Errorf("expected end period in URL, got %s", url)
	}
}

func TestBuildURLNoPeriods(t *testing.T) {
	url := buildURL("DSD", "KEY", "", "")
	if contains(url, "startPeriod") || contains(url, "endPeriod") {
		t.Errorf("should not include empty periods: %s", url)
	}
}

func TestResolveCountry(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"united_states", "USA"},
		{"germany", "DEU"},
		{"USA", "USA"},
		{"usa", "USA"},
		{"all", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := resolveCountry(tt.input)
		if got != tt.expected {
			t.Errorf("resolveCountry(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCountryName(t *testing.T) {
	if name := countryName("USA"); name != "United States" {
		t.Errorf("expected United States, got %s", name)
	}
	if name := countryName("XXX"); name != "XXX" {
		t.Errorf("expected XXX fallback, got %s", name)
	}
}

func TestDatePeriod(t *testing.T) {
	params := provider.QueryParams{
		provider.ParamStartDate: "2024-01-15",
		provider.ParamEndDate:   "2024-06-30",
	}
	start, end := datePeriod(params)
	if start != "2024-01" {
		t.Errorf("expected 2024-01, got %s", start)
	}
	if end != "2024-06" {
		t.Errorf("expected 2024-06, got %s", end)
	}

	// Short dates should pass through unchanged.
	params2 := provider.QueryParams{
		provider.ParamStartDate: "2024",
		provider.ParamEndDate:   "2024",
	}
	s, e := datePeriod(params2)
	if s != "2024" || e != "2024" {
		t.Errorf("short dates should pass through: got %s, %s", s, e)
	}
}

func TestNewResult(t *testing.T) {
	result := newResult([]string{"test"})
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.FetchedAt.IsZero() {
		t.Error("expected non-zero FetchedAt")
	}
}

// ---------------------------------------------------------------------------
// CSV parsing tests
// ---------------------------------------------------------------------------

func TestParseEconomicCSV(t *testing.T) {
	records := [][]string{
		{"REF_AREA", "TIME_PERIOD", "OBS_VALUE"},
		{"USA", "2024-01", "100.5"},
		{"USA", "2024-02", "101.2"},
	}
	results := parseEconomicCSV(records, "CLI")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Value != 100.5 {
		t.Errorf("expected 100.5, got %f", results[0].Value)
	}
	if results[0].Country != "United States" {
		t.Errorf("expected United States, got %s", results[0].Country)
	}
}

func TestParseCPICSV(t *testing.T) {
	records := [][]string{
		{"REF_AREA", "TIME_PERIOD", "OBS_VALUE"},
		{"USA", "2024-01", "310.5"},
		{"DEU", "2024-01", "118.2"},
	}
	results := parseCPICSV(records)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Value != 310.5 {
		t.Errorf("expected 310.5, got %f", results[0].Value)
	}
}

func TestParseInterestRateCSV(t *testing.T) {
	records := [][]string{
		{"REF_AREA", "TIME_PERIOD", "OBS_VALUE"},
		{"USA", "2024-01", "5.33"},
		{"GBR", "2024-01", "5.25"},
	}
	results := parseInterestRateCSV(records, "short_term")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Interest rates are divided by 100.
	if results[0].Rate != 0.0533 {
		t.Errorf("expected 0.0533, got %f", results[0].Rate)
	}
	if results[0].RateType != "short_term" {
		t.Errorf("expected short_term, got %s", results[0].RateType)
	}
}

func TestParseGDPCSV(t *testing.T) {
	records := [][]string{
		{"REF_AREA", "TIME_PERIOD", "OBS_VALUE"},
		{"USA", "2023-Q4", "27000.5"},
	}
	results := parseGDPCSV(records, "nominal", 1e6)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Type != "nominal" {
		t.Errorf("expected nominal, got %s", results[0].Type)
	}
	if results[0].Value != 27000.5*1e6 {
		t.Errorf("expected %f, got %f", 27000.5*1e6, results[0].Value)
	}
}

func TestParseUnemploymentCSV(t *testing.T) {
	records := [][]string{
		{"REF_AREA", "TIME_PERIOD", "OBS_VALUE"},
		{"USA", "2024-01", "3.7"},
	}
	results := parseUnemploymentCSV(records)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Unemployment rate ÷100.
	diff := results[0].Value - 0.037
	if diff < -1e-9 || diff > 1e-9 {
		t.Errorf("expected ~0.037, got %v", results[0].Value)
	}
	if results[0].Country != "United States" {
		t.Errorf("expected United States, got %s", results[0].Country)
	}
}

func TestParseEconomicCSVEmptyRecords(t *testing.T) {
	records := [][]string{
		{"REF_AREA", "TIME_PERIOD", "OBS_VALUE"},
	}
	results := parseEconomicCSV(records, "CLI")
	if len(results) != 0 {
		t.Errorf("expected 0 results for header-only, got %d", len(results))
	}
}

func TestParseEconomicCSVMissingColumn(t *testing.T) {
	records := [][]string{
		{"COUNTRY", "DATE", "VALUE"},
		{"USA", "2024-01", "100"},
	}
	// Missing REF_AREA/TIME_PERIOD/OBS_VALUE columns — should return empty.
	results := parseEconomicCSV(records, "CLI")
	if len(results) != 0 {
		t.Errorf("expected 0 results for missing columns, got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// Context cancellation
// ---------------------------------------------------------------------------

func TestFetcherRespectsContext(t *testing.T) {
	p := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	f := p.Fetcher(provider.ModelCompositeLeadingIndicator)
	_, err := f.Fetch(ctx, provider.QueryParams{
		provider.ParamCountry: "usa",
	})
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

// ---------------------------------------------------------------------------
// Model type tests
// ---------------------------------------------------------------------------

func TestModelTypeCount(t *testing.T) {
	p := New()
	if count := len(p.SupportedModels()); count != 9 {
		t.Errorf("expected exactly 9 models, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Data model integration checks
// ---------------------------------------------------------------------------

func TestGDPDataStruct(t *testing.T) {
	d := models.GDPData{
		Country: "United States",
		Type:    "nominal",
		Value:   27000000000.0,
	}
	if d.Country != "United States" {
		t.Error("expected country field")
	}
}

func TestUnemploymentDataStruct(t *testing.T) {
	d := models.UnemploymentData{
		Country: "United States",
		Value:   0.037,
	}
	if d.Value != 0.037 {
		t.Error("expected value field")
	}
}

// contains is a helper for string checks.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
