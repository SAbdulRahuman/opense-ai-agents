package cboe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/seenimoa/openseai/internal/provider"
)

// ---------------------------------------------------------------------------
// Provider-level tests
// ---------------------------------------------------------------------------

func TestProviderInfo(t *testing.T) {
	p := New()
	info := p.Info()
	if info.Name != "cboe" {
		t.Errorf("expected name cboe, got %s", info.Name)
	}
	if info.Website == "" {
		t.Error("expected non-empty website")
	}
	if len(info.Credentials) != 0 {
		t.Errorf("cboe should have no credentials, got %d", len(info.Credentials))
	}
}

func TestProviderInit(t *testing.T) {
	p := New()
	// CBOE has no credentials, Init should succeed with nil.
	if err := p.Init(nil); err != nil {
		t.Errorf("Init with nil: %v", err)
	}
	if err := p.Init(map[string]string{}); err != nil {
		t.Errorf("Init with empty: %v", err)
	}
}

func TestProviderSupportedModels(t *testing.T) {
	p := New()
	models := p.SupportedModels()

	// CBOE registers 11 fetchers.
	if len(models) != 11 {
		t.Errorf("expected 11 supported models, got %d: %v", len(models), models)
	}

	expected := []provider.ModelType{
		provider.ModelEquityHistorical,
		provider.ModelEquityQuote,
		provider.ModelEquitySearch,
		provider.ModelEtfHistorical,
		provider.ModelAvailableIndices,
		provider.ModelIndexHistorical,
		provider.ModelIndexSearch,
		provider.ModelIndexSnapshots,
		provider.ModelIndexConstituents,
		provider.ModelOptionsChains,
		provider.ModelFuturesCurve,
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
	f = p.Fetcher(provider.ModelType("NonexistentModel"))
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
		{provider.ModelEquityHistorical, []string{"symbol"}},
		{provider.ModelEquityQuote, []string{"symbol"}},
		{provider.ModelEquitySearch, []string{"query"}},
		{provider.ModelEtfHistorical, []string{"symbol"}},
		{provider.ModelAvailableIndices, nil},
		{provider.ModelIndexHistorical, []string{"symbol"}},
		{provider.ModelIndexSearch, []string{"query"}},
		{provider.ModelIndexSnapshots, nil},
		{provider.ModelIndexConstituents, []string{"symbol"}},
		{provider.ModelOptionsChains, []string{"symbol"}},
		{provider.ModelFuturesCurve, nil},
	}

	for _, tt := range tests {
		f := p.Fetcher(tt.model)
		if f == nil {
			t.Errorf("no fetcher for %s", tt.model)
			continue
		}
		got := f.RequiredParams()
		if len(got) != len(tt.required) {
			t.Errorf("%s: expected %d required params, got %d (%v)",
				tt.model, len(tt.required), len(got), got)
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

	symbolModels := []provider.ModelType{
		provider.ModelEquityHistorical,
		provider.ModelEquityQuote,
		provider.ModelIndexHistorical,
		provider.ModelIndexConstituents,
		provider.ModelOptionsChains,
	}

	for _, model := range symbolModels {
		f := p.Fetcher(model)
		if f == nil {
			t.Errorf("no fetcher for %s", model)
			continue
		}
		_, err := f.Fetch(context.Background(), provider.QueryParams{})
		if err == nil {
			t.Errorf("%s: expected error when fetching without required params", model)
		}
	}
}

func TestProviderRegistration(t *testing.T) {
	p := New()
	_ = p.Init(nil)

	reg := provider.NewRegistry()
	if err := reg.Register(p); err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, err := reg.Get("cboe")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Info().Name != "cboe" {
		t.Error("wrong provider name")
	}

	provs := reg.ProvidersFor(provider.ModelOptionsChains)
	found := false
	for _, pn := range provs {
		if pn == "cboe" {
			found = true
		}
	}
	if !found {
		t.Error("cboe not listed as provider for OptionsChains")
	}
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestSymbolPath(t *testing.T) {
	p := New()

	tests := []struct {
		sym  string
		want string
	}{
		{"AAPL", "AAPL"},       // regular equity — no underscore
		{"NDX", "_NDX"},        // known exception
		{"RUT", "_RUT"},        // known exception
		{"^NDX", "_NDX"},       // strip caret + exception
		{"^SPX", "SPX"},        // caret stripped, not in exceptions nor in directory
	}

	for _, tt := range tests {
		got := p.symbolPath(tt.sym)
		if got != tt.want {
			t.Errorf("symbolPath(%q) = %q, want %q", tt.sym, got, tt.want)
		}
	}
}

func TestSymbolPathWithIndexDirectory(t *testing.T) {
	p := New()
	// Simulate an index directory entry.
	p.indexSymbols["SPX"] = true

	if got := p.symbolPath("SPX"); got != "_SPX" {
		t.Errorf("symbolPath(SPX) with directory = %q, want _SPX", got)
	}
	if got := p.symbolPath("^SPX"); got != "_SPX" {
		t.Errorf("symbolPath(^SPX) with directory = %q, want _SPX", got)
	}
}

func TestParseCBOETime(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"2024-01-19T15:45:00", "2024-01-19 15:45:00"},
		{"2024-01-19", "2024-01-19 00:00:00"},
		{"", "0001-01-01 00:00:00"},
	}

	for _, tt := range tests {
		got := parseCBOETime(tt.in)
		want, _ := time.Parse("2006-01-02 15:04:05", tt.want)
		if !got.Equal(want) {
			t.Errorf("parseCBOETime(%q) = %v, want %v", tt.in, got, want)
		}
	}
}

func TestParseCBOEDate(t *testing.T) {
	got := parseCBOEDate("2024-01-19")
	want := time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("parseCBOEDate(2024-01-19) = %v, want %v", got, want)
	}

	zero := parseCBOEDate("")
	if !zero.IsZero() {
		t.Error("expected zero time for empty string")
	}
}

func TestContainsCI(t *testing.T) {
	tests := []struct {
		s, substr string
		want      bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "xyz", false},
		{"CBOE Volatility", "volatility", true},
		{"", "anything", false},
		{"anything", "", true},
	}

	for _, tt := range tests {
		got := containsCI(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("containsCI(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestThirdWednesdayOfMonth(t *testing.T) {
	tests := []struct {
		year  int
		month time.Month
		want  int
	}{
		{2024, time.January, 17},   // Jan 2024: 1st is Mon, Wed=3, 3rd Wed=17
		{2024, time.February, 21},  // Feb 2024: 1st is Thu, Wed=7, 3rd Wed=21
		{2024, time.March, 20},     // Mar 2024: 1st is Fri, Wed=6, 3rd Wed=20
	}

	for _, tt := range tests {
		got := thirdWednesdayOfMonth(tt.year, tt.month)
		if got != tt.want {
			t.Errorf("thirdWednesdayOfMonth(%d, %s) = %d, want %d",
				tt.year, tt.month, got, tt.want)
		}
	}
}

func TestOptionSymbolParsing(t *testing.T) {
	tests := []struct {
		sym     string
		match   bool
		ticker  string
		expDate string
		typ     string
		strike  float64
	}{
		{"AAPL240119C00150000", true, "AAPL", "240119", "C", 150.0},
		{"SPY240315P00500000", true, "SPY", "240315", "P", 500.0},
		{"TSLA241220C01000000", true, "TSLA", "241220", "C", 1000.0},
		{"invalid", false, "", "", "", 0},
	}

	for _, tt := range tests {
		parts := optionSymbolRE.FindStringSubmatch(tt.sym)
		if tt.match {
			if parts == nil {
				t.Errorf("optionSymbolRE failed to match %q", tt.sym)
				continue
			}
			if parts[1] != tt.ticker {
				t.Errorf("%s: ticker = %q, want %q", tt.sym, parts[1], tt.ticker)
			}
			if parts[2] != tt.expDate {
				t.Errorf("%s: expDate = %q, want %q", tt.sym, parts[2], tt.expDate)
			}
			if parts[3] != tt.typ {
				t.Errorf("%s: type = %q, want %q", tt.sym, parts[3], tt.typ)
			}
		} else {
			if parts != nil {
				t.Errorf("optionSymbolRE should not match %q", tt.sym)
			}
		}
	}
}

func TestParseVXSettlement(t *testing.T) {
	csv := "Symbol,Expiration,Settlement,Volume,OI\nVX/F4,2024-01-17,15.25,1000,5000\nVX/G4,2024-02-14,16.50,800,4000\nES/H4,2024-03-15,4800.00,500,3000\n"
	points, err := parseVXSettlement([]byte(csv))
	if err != nil {
		t.Fatalf("parseVXSettlement: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("expected 2 VX points, got %d", len(points))
	}
	if points[0].Price != 15.25 {
		t.Errorf("point[0].Price = %f, want 15.25", points[0].Price)
	}
	if points[1].Price != 16.50 {
		t.Errorf("point[1].Price = %f, want 16.50", points[1].Price)
	}
}

// ---------------------------------------------------------------------------
// Mock server tests for fetcher behavior
// ---------------------------------------------------------------------------

func TestEquityHistoricalWithMockServer(t *testing.T) {
	chartResp := map[string]interface{}{
		"symbol": "AAPL",
		"data": []map[string]interface{}{
			{"date": "2024-01-15", "open": 150.0, "high": 155.0, "low": 149.0, "close": 154.0, "stock_volume": 1000000},
			{"date": "2024-01-16", "open": 154.0, "high": 156.0, "low": 153.0, "close": 155.5, "stock_volume": 900000},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chartResp)
	}))
	defer srv.Close()

	// We can't easily override the base URL without exporting it,
	// so we verify the provider created correctly and test parsing directly.
	raw, _ := json.Marshal(chartResp)
	bars, err := parseDailyChart(raw, provider.QueryParams{})
	if err != nil {
		t.Fatalf("parseDailyChart: %v", err)
	}
	if len(bars) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(bars))
	}
	if bars[0].Open != 150.0 {
		t.Errorf("bars[0].Open = %f, want 150.0", bars[0].Open)
	}
	if bars[1].Close != 155.5 {
		t.Errorf("bars[1].Close = %f, want 155.5", bars[1].Close)
	}
	if bars[0].Volume != 1000000 {
		t.Errorf("bars[0].Volume = %d, want 1000000", bars[0].Volume)
	}
}

func TestDailyChartDateFiltering(t *testing.T) {
	chartResp := map[string]interface{}{
		"symbol": "AAPL",
		"data": []map[string]interface{}{
			{"date": "2024-01-10", "open": 148.0, "high": 150.0, "low": 147.0, "close": 149.0, "stock_volume": 800000},
			{"date": "2024-01-15", "open": 150.0, "high": 155.0, "low": 149.0, "close": 154.0, "stock_volume": 1000000},
			{"date": "2024-01-20", "open": 155.0, "high": 158.0, "low": 154.0, "close": 157.0, "stock_volume": 1100000},
		},
	}
	raw, _ := json.Marshal(chartResp)

	// With date range filter.
	params := provider.QueryParams{
		provider.ParamStartDate: "2024-01-14",
		provider.ParamEndDate:   "2024-01-16",
	}
	bars, err := parseDailyChart(raw, params)
	if err != nil {
		t.Fatalf("parseDailyChart with dates: %v", err)
	}
	if len(bars) != 1 {
		t.Fatalf("expected 1 bar after filtering, got %d", len(bars))
	}
	if bars[0].Close != 154.0 {
		t.Errorf("filtered bar Close = %f, want 154.0", bars[0].Close)
	}
}

func TestIntradayChartParsing(t *testing.T) {
	chartResp := map[string]interface{}{
		"symbol": "AAPL",
		"data": []map[string]interface{}{
			{
				"datetime": "2024-01-15T09:30:00",
				"price":    map[string]float64{"open": 150.0, "high": 150.5, "low": 149.8, "close": 150.3},
				"volume":   map[string]int64{"stock_volume": 500000},
			},
		},
	}
	raw, _ := json.Marshal(chartResp)

	bars, err := parseIntradayChart(raw)
	if err != nil {
		t.Fatalf("parseIntradayChart: %v", err)
	}
	if len(bars) != 1 {
		t.Fatalf("expected 1 intraday bar, got %d", len(bars))
	}
	if bars[0].Open != 150.0 {
		t.Errorf("Open = %f, want 150.0", bars[0].Open)
	}
	if bars[0].High != 150.5 {
		t.Errorf("High = %f, want 150.5", bars[0].High)
	}
}

func TestNewResultFields(t *testing.T) {
	data := []string{"a", "b", "c"}
	result := newResult(data)

	if result.Data == nil {
		t.Error("expected non-nil data")
	}
	if result.FetchedAt.IsZero() {
		t.Error("expected non-zero FetchedAt")
	}
	// Verify data is the same reference.
	got, ok := result.Data.([]string)
	if !ok {
		t.Fatal("data type mismatch")
	}
	if len(got) != 3 {
		t.Errorf("expected 3 items, got %d", len(got))
	}
}

func TestURLBuilders(t *testing.T) {
	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{"quotesURL", func() string { return quotesURL("AAPL") }, "https://cdn.cboe.com/api/global/delayed_quotes/quotes/AAPL.json"},
		{"quotesURL underscore", func() string { return quotesURL("_NDX") }, "https://cdn.cboe.com/api/global/delayed_quotes/quotes/_NDX.json"},
		{"chartURL daily", func() string { return chartURL("AAPL", "1d") }, "https://cdn.cboe.com/api/global/delayed_quotes/charts/historical/AAPL.json"},
		{"chartURL intraday", func() string { return chartURL("AAPL", "1m") }, "https://cdn.cboe.com/api/global/delayed_quotes/charts/intraday/AAPL.json"},
		{"optionsURL", func() string { return optionsURL("_SPX") }, "https://cdn.cboe.com/api/global/delayed_quotes/options/_SPX.json"},
	}

	for _, tt := range tests {
		got := tt.fn()
		if got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestModelTypeCount(t *testing.T) {
	p := New()
	models := p.SupportedModels()
	// We registered 11 fetchers.
	if len(models) != 11 {
		t.Errorf("expected 11 models, got %d", len(models))
	}
}
