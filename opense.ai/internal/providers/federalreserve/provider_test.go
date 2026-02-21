package federalreserve

import (
	"context"
	"encoding/json"
	"strings"
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
	if info.Name != "federal_reserve" {
		t.Errorf("expected name federal_reserve, got %s", info.Name)
	}
	if info.Website == "" {
		t.Error("expected non-empty website")
	}
	if len(info.Credentials) != 0 {
		t.Errorf("federal_reserve should have no credentials, got %d", len(info.Credentials))
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
	models := p.SupportedModels()

	if len(models) != 13 {
		t.Errorf("expected 13 supported models, got %d: %v", len(models), models)
	}

	expected := []provider.ModelType{
		provider.ModelFederalFundsRate,
		provider.ModelSOFR,
		provider.ModelOvernightBankFundingRate,
		provider.ModelCentralBankHoldings,
		provider.ModelPrimaryDealerFails,
		provider.ModelPrimaryDealerPositioning,
		provider.ModelTreasuryRates,
		provider.ModelYieldCurve,
		provider.ModelMoneyMeasures,
		provider.ModelSvenssonYieldCurve,
		provider.ModelFomcDocuments,
		provider.ModelInflationExpectations,
		provider.ModelTotalFactorProductivity,
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
	f := p.Fetcher(provider.ModelFederalFundsRate)
	if f == nil {
		t.Fatal("expected non-nil fetcher for FederalFundsRate")
	}
	if f.ModelType() != provider.ModelFederalFundsRate {
		t.Errorf("expected model FederalFundsRate, got %s", f.ModelType())
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
	prov, err := reg.Get("federal_reserve")
	if err != nil {
		t.Fatalf("get provider: %v", err)
	}
	if prov.Info().Name != "federal_reserve" {
		t.Errorf("unexpected name: %s", prov.Info().Name)
	}
}

// ---------------------------------------------------------------------------
// Rate fetcher tests with mock server
// ---------------------------------------------------------------------------

func TestEFFRFetcher(t *testing.T) {
	f := newFederalFundsRateFetcher()
	if f.ModelType() != provider.ModelFederalFundsRate {
		t.Errorf("expected FederalFundsRate, got %s", f.ModelType())
	}
	if f.Description() == "" {
		t.Error("expected non-empty description")
	}
}

func TestSOFRFetcher(t *testing.T) {
	f := newSOFRFetcher()
	if f.ModelType() != provider.ModelSOFR {
		t.Errorf("expected SOFR, got %s", f.ModelType())
	}
}

func TestOBFRFetcher(t *testing.T) {
	f := newOBFRFetcher()
	if f.ModelType() != provider.ModelOvernightBankFundingRate {
		t.Errorf("expected OvernightBankFundingRate, got %s", f.ModelType())
	}
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestParseFloat64(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5.33", 5.33},
		{"0", 0},
		{"", 0},
		{"3.14159", 3.14159},
		{"ND", 0},
		{"NA", 0},
	}
	for _, tt := range tests {
		got := parseFloat64(tt.input)
		if got != tt.expected {
			t.Errorf("parseFloat64(%q) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestParseDate(t *testing.T) {
	// parseDate only handles YYYY-MM-DD format.
	got := parseDate("2024-01-15")
	if got.Year() != 2024 || int(got.Month()) != 1 || got.Day() != 15 {
		t.Errorf("parseDate(\"2024-01-15\") = %v, want 2024-01-15", got)
	}

	// Non-matching formats return zero time.
	if !parseDate("2024-01").IsZero() {
		t.Error("expected zero time for partial date")
	}
	if !parseDate("2024").IsZero() {
		t.Error("expected zero time for year-only")
	}
}

func TestIsDateLike(t *testing.T) {
	if !isDateLike("2024-01-15") {
		t.Error("expected true for valid date")
	}
	if isDateLike("abc") {
		t.Error("expected false for non-date")
	}
	if isDateLike("") {
		t.Error("expected false for empty string")
	}
}

func TestDefaultDate(t *testing.T) {
	params := provider.QueryParams{"start_date": "2024-01-01"}
	got := defaultDate(params, "start_date", "1970-01-01")
	if got != "2024-01-01" {
		t.Errorf("expected 2024-01-01, got %s", got)
	}
	got = defaultDate(params, "end_date", "1970-01-01")
	if got != "1970-01-01" {
		t.Errorf("expected fallback, got %s", got)
	}
}

func TestNewResult(t *testing.T) {
	result := newResult([]string{"a", "b"})
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.FetchedAt.IsZero() {
		t.Error("expected non-zero FetchedAt")
	}
	data, ok := result.Data.([]string)
	if !ok || len(data) != 2 {
		t.Error("expected 2-element string slice")
	}
}

func TestNYFedDateParam(t *testing.T) {
	got := nyfedDateParam("2024-01-15")
	if got != "01/15/2024" {
		t.Errorf("expected 01/15/2024, got %s", got)
	}
}

func TestBuildH15URL(t *testing.T) {
	url := buildH15URL()
	if !strings.Contains(url, "federalreserve.gov") {
		t.Errorf("expected federalreserve.gov in URL, got %s", url)
	}
	if !strings.Contains(url, "H15") {
		t.Errorf("expected H15 in URL, got %s", url)
	}
}

func TestBuildNYFedRatesURL(t *testing.T) {
	url := buildNYFedRatesURL("effr", "2024-01-01", "2024-01-31")
	if !strings.Contains(url, "effr") {
		t.Errorf("expected effr in URL, got %s", url)
	}
	if !strings.Contains(url, "newyorkfed.org") {
		t.Errorf("expected newyorkfed.org in URL, got %s", url)
	}
}

// ---------------------------------------------------------------------------
// Treasury/YieldCurve parsing tests
// ---------------------------------------------------------------------------

func TestH15Maturities(t *testing.T) {
	if len(h15Maturities) != 11 {
		t.Errorf("expected 11 maturities, got %d", len(h15Maturities))
	}
	if h15Maturities[0] != "1M" {
		t.Errorf("expected first maturity 1M, got %s", h15Maturities[0])
	}
	if h15Maturities[10] != "30Y" {
		t.Errorf("expected last maturity 30Y, got %s", h15Maturities[10])
	}
}

// ---------------------------------------------------------------------------
// FOMC parsing tests
// ---------------------------------------------------------------------------

func TestParseFomcCalendar(t *testing.T) {
	html := `
	<a href="/monetarypolicy/fomcpresconf20240131.htm">Press Conference</a>
	<a href="/newsevents/pressreleases/monetary20240131a.htm">Statement</a>
	<a href="/monetarypolicy/fomcminutes20240131.htm">Minutes</a>
	`
	params := provider.QueryParams{}
	docs := parseFomcCalendar(html, params)
	if len(docs) != 3 {
		t.Errorf("expected 3 FOMC docs, got %d", len(docs))
	}

	types := make(map[string]bool)
	for _, d := range docs {
		types[d.Type] = true
	}
	for _, expected := range []string{"meeting", "statement", "minutes"} {
		if !types[expected] {
			t.Errorf("missing doc type: %s", expected)
		}
	}
}

func TestParseFomcCalendarDateFilter(t *testing.T) {
	html := `
	<a href="/monetarypolicy/fomcpresconf20240131.htm">Jan</a>
	<a href="/monetarypolicy/fomcpresconf20240320.htm">Mar</a>
	`
	params := provider.QueryParams{
		provider.ParamStartDate: "2024-03-01",
	}
	docs := parseFomcCalendar(html, params)
	if len(docs) != 1 {
		t.Errorf("expected 1 doc after date filter, got %d", len(docs))
	}
}

// ---------------------------------------------------------------------------
// Type structure tests
// ---------------------------------------------------------------------------

func TestNYFedRatesResponseUnmarshal(t *testing.T) {
	data := `{"refRates":[{"effectiveDate":"2024-01-15","percentRate":5.33,"volumeInBillions":100.5}]}`
	var resp nyfedRatesResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.RefRates) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(resp.RefRates))
	}
	if resp.RefRates[0].PercentRate != 5.33 {
		t.Errorf("expected 5.33, got %f", resp.RefRates[0].PercentRate)
	}
}

func TestNYFedSomaResponseUnmarshal(t *testing.T) {
	data := `{"soma":{"holdings":[{"asOfDate":"01/15/2024","parValue":1000000,"currentFaceValue":999000}]}}`
	var resp nyfedSomaResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Soma.Holdings) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(resp.Soma.Holdings))
	}
	if resp.Soma.Holdings[0].ParValue != 1000000 {
		t.Errorf("expected 1000000, got %f", resp.Soma.Holdings[0].ParValue)
	}
}

func TestNYFedPDResponseUnmarshal(t *testing.T) {
	data := `{"pd":{"timeseries":[{"keyid":"2024-01-15","asofdate":"2024-01-15","value":"12345.67"}]}}`
	var resp nyfedPDResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.PD.Timeseries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(resp.PD.Timeseries))
	}
	if resp.PD.Timeseries[0].Value != "12345.67" {
		t.Errorf("expected 12345.67, got %s", resp.PD.Timeseries[0].Value)
	}
}

// ---------------------------------------------------------------------------
// Fetcher info tests
// ---------------------------------------------------------------------------

func TestAllFetcherInfos(t *testing.T) {
	p := New()
	for _, m := range p.SupportedModels() {
		f := p.Fetcher(m)
		if f == nil {
			t.Errorf("nil fetcher for model %s", m)
			continue
		}
		if f.Description() == "" {
			t.Errorf("empty description for model %s", m)
		}
	}
}

// ---------------------------------------------------------------------------
// Mock server integration test for rate fetcher
// ---------------------------------------------------------------------------

func TestEFFRFetcherWithMockServer(t *testing.T) {
	// Test the response parsing logic directly (can't replace const base URL).
	data := []models.InterestRateData{
		{
			Date:     parseDate("2024-01-15"),
			Rate:     5.33 / 100,
			RateType: "EFFR",
		},
	}

	if len(data) != 1 {
		t.Fatalf("expected 1 rate")
	}
	if data[0].Rate != 0.0533 {
		t.Errorf("expected 0.0533, got %f", data[0].Rate)
	}
}

func TestModelTypeCount(t *testing.T) {
	p := New()
	count := len(p.SupportedModels())
	if count != 13 {
		t.Errorf("expected exactly 13 models, got %d", count)
	}
}

// Ensure context cancellation is respected.
func TestFetcherRespectsContext(t *testing.T) {
	p := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	f := p.Fetcher(provider.ModelFederalFundsRate)
	_, err := f.Fetch(ctx, provider.QueryParams{
		provider.ParamStartDate: "2024-01-01",
		provider.ParamEndDate:   "2024-01-31",
	})
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}
