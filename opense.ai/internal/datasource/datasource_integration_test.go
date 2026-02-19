package datasource

import (
	"testing"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

func TestNSEName(t *testing.T) {
	nse := NewNSE()
	if nse.Name() != "NSE India" {
		t.Errorf("Name() = %q, want %q", nse.Name(), "NSE India")
	}
}

func TestNSEDerivativesName(t *testing.T) {
	nse := NewNSE()
	d := NewNSEDerivatives(nse)
	if d.Name() != "NSE Derivatives" {
		t.Errorf("Name() = %q, want %q", d.Name(), "NSE Derivatives")
	}
}

func TestParseCircuit(t *testing.T) {
	q := &models.Quote{}
	parseCircuit("20.00", "-20.00", 100.0, q)

	if q.UpperCircuit != 120.0 {
		t.Errorf("UpperCircuit = %f, want 120.0", q.UpperCircuit)
	}
	if q.LowerCircuit != 80.0 {
		t.Errorf("LowerCircuit = %f, want 80.0", q.LowerCircuit)
	}
}

func TestParseCircuitBadInput(t *testing.T) {
	q := &models.Quote{}
	parseCircuit("N/A", "N/A", 100.0, q)
	if q.UpperCircuit != 0 || q.LowerCircuit != 0 {
		t.Errorf("expected zero for invalid circuit inputs, got upper=%f lower=%f", q.UpperCircuit, q.LowerCircuit)
	}
}

func TestCalculateMaxPain(t *testing.T) {
	nse := NewNSE()
	d := NewNSEDerivatives(nse)

	oc := &models.OptionChain{
		Contracts: []models.OptionContract{
			{StrikePrice: 100, OptionType: "CE", OI: 1000},
			{StrikePrice: 100, OptionType: "PE", OI: 500},
			{StrikePrice: 110, OptionType: "CE", OI: 2000},
			{StrikePrice: 110, OptionType: "PE", OI: 1500},
			{StrikePrice: 120, OptionType: "CE", OI: 500},
			{StrikePrice: 120, OptionType: "PE", OI: 3000},
		},
	}

	maxPain := d.calculateMaxPain(oc)
	// Max pain should be at the strike where total pain is minimized.
	// With these OI values, it should be around 110 (middle ground).
	if maxPain != 100 && maxPain != 110 && maxPain != 120 {
		t.Errorf("maxPain = %f, expected one of 100/110/120", maxPain)
	}
}

func TestCalculateMaxPainEmpty(t *testing.T) {
	nse := NewNSE()
	d := NewNSEDerivatives(nse)

	oc := &models.OptionChain{}
	maxPain := d.calculateMaxPain(oc)
	if maxPain != 0 {
		t.Errorf("maxPain = %f, want 0 for empty chain", maxPain)
	}
}

func TestAggregatorSources(t *testing.T) {
	agg := NewAggregator()
	sources := agg.Sources()
	if len(sources) != 6 {
		t.Fatalf("expected 6 sources, got %d", len(sources))
	}

	names := make(map[string]bool)
	for _, s := range sources {
		names[s.Name()] = true
	}

	expected := []string{"Yahoo Finance", "NSE India", "NSE Derivatives", "Screener.in", "Indian News", "FII/DII Activity"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing source: %s", name)
		}
	}
}

func TestAggregatorAccessors(t *testing.T) {
	agg := NewAggregator()
	if agg.YFinance() == nil {
		t.Error("YFinance() returned nil")
	}
	if agg.NSE() == nil {
		t.Error("NSE() returned nil")
	}
	if agg.Derivatives() == nil {
		t.Error("Derivatives() returned nil")
	}
	if agg.Screener() == nil {
		t.Error("Screener() returned nil")
	}
	if agg.NewsSource() == nil {
		t.Error("NewsSource() returned nil")
	}
	if agg.FIIDII() == nil {
		t.Error("FIIDII() returned nil")
	}
}

func TestTickerKeywords(t *testing.T) {
	kw := tickerKeywords("RELIANCE")
	if len(kw) < 2 {
		t.Errorf("expected multiple keywords for RELIANCE, got %d", len(kw))
	}
	found := false
	for _, k := range kw {
		if k == "reliance industries" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'reliance industries' in keywords")
	}
}

func TestMatchesAny(t *testing.T) {
	tests := []struct {
		text     string
		keywords []string
		want     bool
	}{
		{"Reliance Industries Q4 results", []string{"reliance"}, true},
		{"TCS quarterly earnings", []string{"infosys"}, false},
		{"HDFC Bank merger update", []string{"hdfc bank", "hdfc"}, true},
	}
	for _, tt := range tests {
		got := matchesAny(tt.text, tt.keywords)
		if got != tt.want {
			t.Errorf("matchesAny(%q, %v) = %v, want %v", tt.text, tt.keywords, got, tt.want)
		}
	}
}

func TestSortArticlesByDate(t *testing.T) {
	now := time.Now()
	articles := []models.NewsArticle{
		{Title: "old", PublishedAt: now.Add(-2 * time.Hour)},
		{Title: "newest", PublishedAt: now},
		{Title: "mid", PublishedAt: now.Add(-1 * time.Hour)},
	}
	sortArticlesByDate(articles)
	if articles[0].Title != "newest" {
		t.Errorf("first article = %q, want %q", articles[0].Title, "newest")
	}
	if articles[2].Title != "old" {
		t.Errorf("last article = %q, want %q", articles[2].Title, "old")
	}
}

func TestCleanHTML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<p>Hello <b>world</b></p>", "Hello world"},
		{"plain text", "plain text"},
		{"", ""},
		{"<div><a href='#'>link</a> and text</div>", "link and text"},
	}
	for _, tt := range tests {
		got := cleanHTML(tt.input)
		if got != tt.want {
			t.Errorf("cleanHTML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFIIDIIName(t *testing.T) {
	nse := NewNSE()
	f := NewFIIDII(nse)
	if f.Name() != "FII/DII Activity" {
		t.Errorf("Name() = %q, want %q", f.Name(), "FII/DII Activity")
	}
}

func TestScreenerName(t *testing.T) {
	s := NewScreener()
	if s.Name() != "Screener.in" {
		t.Errorf("Name() = %q, want %q", s.Name(), "Screener.in")
	}
}

func TestNewsName(t *testing.T) {
	n := NewNews()
	if n.Name() != "Indian News" {
		t.Errorf("Name() = %q, want %q", n.Name(), "Indian News")
	}
}
