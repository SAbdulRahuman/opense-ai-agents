package imf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
	"github.com/seenimoa/openseai/pkg/models"
)

// ---------------------------------------------------------------------------
// Provider-level tests
// ---------------------------------------------------------------------------

func TestProviderInfo(t *testing.T) {
	p := New()
	info := p.Info()
	if info.Name != "imf" {
		t.Errorf("expected name imf, got %s", info.Name)
	}
	if info.Website == "" {
		t.Error("expected non-empty website")
	}
	if len(info.Credentials) != 0 {
		t.Errorf("imf should have no credentials, got %d", len(info.Credentials))
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

	if len(ms) != 8 {
		t.Errorf("expected 8 supported models, got %d: %v", len(ms), ms)
	}

	expected := []provider.ModelType{
		provider.ModelAvailableIndicators,
		provider.ModelConsumerPriceIndex,
		provider.ModelDirectionOfTrade,
		provider.ModelEconomicIndicators,
		provider.ModelMaritimeChokePointInfo,
		provider.ModelMaritimeChokePointVolume,
		provider.ModelPortInfo,
		provider.ModelPortVolume,
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
	prov, err := reg.Get("imf")
	if err != nil {
		t.Fatalf("get provider: %v", err)
	}
	if prov.Info().Name != "imf" {
		t.Errorf("unexpected name: %s", prov.Info().Name)
	}
}

// ---------------------------------------------------------------------------
// AvailableIndicators fetcher test (static, no HTTP)
// ---------------------------------------------------------------------------

func TestAvailableIndicatorsFetcher(t *testing.T) {
	p := New()
	f := p.Fetcher(provider.ModelAvailableIndicators)
	if f == nil {
		t.Fatal("nil fetcher for AvailableIndicators")
	}

	result, err := f.Fetch(context.Background(), provider.QueryParams{})
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}
	if result == nil {
		t.Fatal("nil result")
	}

	indicators, ok := result.Data.([]models.AvailableEconomicIndicator)
	if !ok {
		t.Fatalf("unexpected data type: %T", result.Data)
	}
	if len(indicators) != 12 {
		t.Errorf("expected 12 known dataflows, got %d", len(indicators))
	}

	// Check that CPI is in the list.
	found := false
	for _, ind := range indicators {
		if ind.ID == "CPI" {
			found = true
			if ind.Source != "IMF" {
				t.Errorf("expected source IMF, got %s", ind.Source)
			}
		}
	}
	if !found {
		t.Error("CPI not found in available indicators")
	}
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

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
		{"-1.5", -1.5},
	}
	for _, tt := range tests {
		got := parseFloat(tt.input)
		if got != tt.expected {
			t.Errorf("parseFloat(%q) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestParseAnyFloat(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
	}{
		{float64(5.33), 5.33},
		{int(42), 42},
		{int64(100), 100},
		{"3.14", 3.14},
		{nil, 0},
		{true, 0},
	}
	for _, tt := range tests {
		got := parseAnyFloat(tt.input)
		if got != tt.expected {
			t.Errorf("parseAnyFloat(%v) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestParseAnyString(t *testing.T) {
	if got := parseAnyString("hello"); got != "hello" {
		t.Errorf("expected hello, got %s", got)
	}
	if got := parseAnyString(nil); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
	if got := parseAnyString(42); got != "42" {
		t.Errorf("expected 42, got %s", got)
	}
}

func TestParseAnyDate(t *testing.T) {
	tests := []struct {
		input string
		year  int
		month int
	}{
		{"2024-01-15", 2024, 1},
		{"2024-01", 2024, 1},
		{"2024", 2024, 1},
		{"invalid", 1, 1}, // zero time
	}
	for _, tt := range tests {
		got := parseAnyDate(tt.input)
		if tt.input == "invalid" {
			if !got.IsZero() {
				t.Errorf("expected zero time for invalid input")
			}
		} else if got.Year() != tt.year || int(got.Month()) != tt.month {
			t.Errorf("parseAnyDate(%q) = %v, want year=%d month=%d", tt.input, got, tt.year, tt.month)
		}
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
		{"all", "*"},
		{"", "*"},
		{"*", "*"},
	}
	for _, tt := range tests {
		got := resolveCountry(tt.input)
		if got != tt.expected {
			t.Errorf("resolveCountry(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestBuildSDMXURL(t *testing.T) {
	url := buildSDMXURL("IMF", "CPI", "USA.CPI._T.IX.M", nil)
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	if !containsStr(url, "api.imf.org") {
		t.Errorf("expected api.imf.org in URL, got %s", url)
	}
	if !containsStr(url, "CPI") {
		t.Errorf("expected CPI in URL, got %s", url)
	}
}

func TestBuildSDMXURLWithParams(t *testing.T) {
	params := map[string]string{
		"lastNObservations": "10",
	}
	url := buildSDMXURL("IMF", "CPI", "USA", params)
	if !containsStr(url, "lastNObservations=10") {
		t.Errorf("expected lastNObservations=10, got %s", url)
	}
}

func TestSplitSymbol(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"CPI::USA", []string{"CPI", "USA"}},
		{"single", []string{"single"}},
		{"DATAFLOW::IND.KEY", []string{"DATAFLOW", "IND.KEY"}},
	}
	for _, tt := range tests {
		got := splitSymbol(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("splitSymbol(%q) = %v, want %v", tt.input, got, tt.expected)
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("splitSymbol(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
			}
		}
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
// ArcGIS helper tests
// ---------------------------------------------------------------------------

func TestBuildDateWhere(t *testing.T) {
	tests := []struct {
		name     string
		params   provider.QueryParams
		expected string
	}{
		{
			"no dates",
			provider.QueryParams{},
			"1=1",
		},
		{
			"start only",
			provider.QueryParams{provider.ParamStartDate: "2024-01-01"},
			"date >= TIMESTAMP '2024-01-01 00:00:00'",
		},
		{
			"both dates",
			provider.QueryParams{
				provider.ParamStartDate: "2024-01-01",
				provider.ParamEndDate:   "2024-06-30",
			},
			"date >= TIMESTAMP '2024-01-01 00:00:00' AND date <= TIMESTAMP '2024-06-30 00:00:00'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDateWhere(tt.params, "date")
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConstructDate(t *testing.T) {
	// Year/month/day construction.
	attrs := map[string]any{
		"year":  float64(2024),
		"month": float64(6),
		"day":   float64(15),
	}
	got := constructDate(attrs)
	expected := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}

	// Epoch millis construction.
	epochMs := float64(1718409600000) // 2024-06-15 00:00:00 UTC
	attrs2 := map[string]any{
		"date": epochMs,
	}
	got2 := constructDate(attrs2)
	expected2 := time.UnixMilli(int64(epochMs)).UTC()
	if !got2.Equal(expected2) {
		t.Errorf("epoch: expected %v, got %v", expected2, got2)
	}

	// Empty attrs — zero time.
	got3 := constructDate(map[string]any{})
	if !got3.IsZero() {
		t.Errorf("expected zero time, got %v", got3)
	}
}

// ---------------------------------------------------------------------------
// SDMX response extraction tests
// ---------------------------------------------------------------------------

func TestExtractCPIFromSDMX(t *testing.T) {
	resp := sdmxDataResponse{
		Data: sdmxData{
			DataSets: []sdmxDataSet{
				{
					Series: map[string]sdmxSeries{
						"0:0:0:0:0": {
							Observations: map[string][]any{
								"2024-01": {310.5},
								"2024-02": {311.2},
							},
						},
					},
				},
			},
		},
	}

	results := extractCPIFromSDMX(resp, "USA")
	if len(results) != 2 {
		t.Fatalf("expected 2 CPI results, got %d", len(results))
	}

	for _, r := range results {
		if r.Country != "USA" {
			t.Errorf("expected USA, got %s", r.Country)
		}
		if r.Value == 0 {
			t.Error("expected non-zero value")
		}
	}
}

func TestExtractTradeFromSDMX(t *testing.T) {
	resp := sdmxDataResponse{
		Data: sdmxData{
			DataSets: []sdmxDataSet{
				{
					Series: map[string]sdmxSeries{
						"0:0:0": {
							Observations: map[string][]any{
								"2024-01": {1234.56},
							},
						},
					},
				},
			},
		},
	}

	results := extractTradeFromSDMX(resp, "USA")
	if len(results) != 1 {
		t.Fatalf("expected 1 trade result, got %d", len(results))
	}
	if results[0].Country != "USA" {
		t.Errorf("expected USA, got %s", results[0].Country)
	}
}

func TestExtractEconomicFromSDMX(t *testing.T) {
	resp := sdmxDataResponse{
		Data: sdmxData{
			DataSets: []sdmxDataSet{
				{
					Series: map[string]sdmxSeries{
						"0:0": {
							Observations: map[string][]any{
								"2024": {42.5},
							},
						},
					},
				},
			},
		},
	}

	results := extractEconomicFromSDMX(resp, "USA")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 42.5 {
		t.Errorf("expected 42.5, got %f", results[0].Value)
	}
}

func TestExtractEconomicFromSDMXEmpty(t *testing.T) {
	resp := sdmxDataResponse{
		Data: sdmxData{
			DataSets: nil,
		},
	}
	results := extractEconomicFromSDMX(resp, "USA")
	if len(results) != 0 {
		t.Errorf("expected 0 results from empty response, got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// ArcGIS mock server test
// ---------------------------------------------------------------------------

func TestChokePointInfoWithMockServer(t *testing.T) {
	mockResp := arcGISResponse{
		Features: []arcGISFeature{
			{Attributes: map[string]any{
				"portname": "Suez Canal",
				"ISO3":     "EGY",
			}},
			{Attributes: map[string]any{
				"portname": "Panama Canal",
				"ISO3":     "PAN",
			}},
		},
		ExceededTransferLimit: false,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer ts.Close()

	// Test the response structure parsing directly.
	data, err := json.Marshal(mockResp)
	if err != nil {
		t.Fatal(err)
	}
	var decoded arcGISResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Features) != 2 {
		t.Errorf("expected 2 features, got %d", len(decoded.Features))
	}
	if decoded.Features[0].Attributes["portname"] != "Suez Canal" {
		t.Errorf("expected Suez Canal, got %v", decoded.Features[0].Attributes["portname"])
	}
}

func TestSDMXResponseUnmarshal(t *testing.T) {
	raw := `{
		"data": {
			"dataSets": [{
				"series": {
					"0:0:0:0:0": {
						"observations": {
							"0": [310.5]
						}
					}
				}
			}]
		}
	}`
	var resp sdmxDataResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Data.DataSets) != 1 {
		t.Fatalf("expected 1 dataset, got %d", len(resp.Data.DataSets))
	}
	series := resp.Data.DataSets[0].Series
	if len(series) != 1 {
		t.Fatalf("expected 1 series, got %d", len(series))
	}
}

// ---------------------------------------------------------------------------
// Context cancellation
// ---------------------------------------------------------------------------

func TestFetcherRespectsContext(t *testing.T) {
	p := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// CPI fetcher uses HTTP, should respect cancelled context.
	f := p.Fetcher(provider.ModelConsumerPriceIndex)
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
	if count := len(p.SupportedModels()); count != 8 {
		t.Errorf("expected exactly 8 models, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Known dataflows test
// ---------------------------------------------------------------------------

func TestKnownDataflows(t *testing.T) {
	if len(knownIMFDataflows) != 12 {
		t.Errorf("expected 12 known dataflows, got %d", len(knownIMFDataflows))
	}
	ids := make(map[string]bool)
	for _, d := range knownIMFDataflows {
		if d.ID == "" {
			t.Error("found empty ID in known dataflows")
		}
		if ids[d.ID] {
			t.Errorf("duplicate dataflow ID: %s", d.ID)
		}
		ids[d.ID] = true
		if d.Source != "IMF" {
			t.Errorf("expected source IMF, got %s", d.Source)
		}
	}
}

// ---------------------------------------------------------------------------
// Min helper
// ---------------------------------------------------------------------------

func TestMin(t *testing.T) {
	if min(5, 10) != 5 {
		t.Error("min(5,10) should be 5")
	}
	if min(10, 5) != 5 {
		t.Error("min(10,5) should be 5")
	}
}

// ---------------------------------------------------------------------------
// Port data struct check
// ---------------------------------------------------------------------------

func TestPortDataStruct(t *testing.T) {
	d := models.PortData{
		Port:     "Shanghai",
		Country:  "CHN",
		Volume:   1000.0,
		Unit:     "metric_tons",
		Category: "port",
	}
	if d.Port != "Shanghai" {
		t.Error("expected Port field")
	}
}

// containsStr is a helper.
func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// Unused import guard — ensure fmt is used (in case tests reuse it).
var _ = fmt.Sprint
