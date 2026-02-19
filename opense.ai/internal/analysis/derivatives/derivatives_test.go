package derivatives

import (
	"testing"

	"github.com/seenimoa/openseai/pkg/models"
)

func sampleOptionChain() *models.OptionChain {
	return &models.OptionChain{
		Ticker:    "NIFTY",
		SpotPrice: 25000,
		ExpiryDate: "2025-07-24",
		TotalCEOI: 500000,
		TotalPEOI: 600000,
		PCR:       1.2,
		MaxPain:   25000,
		Contracts: []models.OptionContract{
			{StrikePrice: 24500, OptionType: "PE", LTP: 45, OI: 120000, OIChange: 15000, OIChangePct: 14, Volume: 50000, IV: 15},
			{StrikePrice: 24800, OptionType: "PE", LTP: 80, OI: 100000, OIChange: 10000, OIChangePct: 11, Volume: 40000, IV: 14},
			{StrikePrice: 25000, OptionType: "PE", LTP: 150, OI: 180000, OIChange: 20000, OIChangePct: 12.5, Volume: 60000, IV: 13.5},
			{StrikePrice: 25000, OptionType: "CE", LTP: 160, OI: 170000, OIChange: -5000, OIChangePct: -2.9, Volume: 55000, IV: 12},
			{StrikePrice: 25200, OptionType: "CE", LTP: 80, OI: 200000, OIChange: 25000, OIChangePct: 14.3, Volume: 70000, IV: 13},
			{StrikePrice: 25500, OptionType: "CE", LTP: 30, OI: 150000, OIChange: 18000, OIChangePct: 13.6, Volume: 45000, IV: 14.5},
		},
	}
}

func TestAnalyzeOptionChain(t *testing.T) {
	oc := sampleOptionChain()
	a := AnalyzeOptionChain(oc)

	if a.Ticker != "NIFTY" {
		t.Errorf("expected NIFTY, got %s", a.Ticker)
	}
	if a.ATMStrike != 25000 {
		t.Errorf("expected ATM 25000, got %.0f", a.ATMStrike)
	}
	if a.Sentiment == "" {
		t.Error("expected sentiment to be set")
	}
	if a.ATMIV <= 0 {
		t.Errorf("expected positive ATM IV, got %.2f", a.ATMIV)
	}
}

func TestAnalyzeOptionChainNil(t *testing.T) {
	a := AnalyzeOptionChain(nil)
	if a.Ticker != "" {
		t.Error("expected empty for nil input")
	}
}

func TestComputeMaxPain(t *testing.T) {
	oc := sampleOptionChain()
	mp := ComputeMaxPain(oc.Contracts)
	if mp <= 0 {
		t.Errorf("expected positive max pain, got %.0f", mp)
	}
}

func TestComputePCR(t *testing.T) {
	oc := sampleOptionChain()
	pcr := ComputePCR(oc)

	if pcr.PCR <= 0 {
		t.Errorf("expected positive PCR, got %.4f", pcr.PCR)
	}
	if pcr.Signal == "" {
		t.Error("expected signal to be set")
	}
	if pcr.Interpretation == "" {
		t.Error("expected interpretation")
	}
}

func TestComputePCRNil(t *testing.T) {
	pcr := ComputePCR(nil)
	if pcr.PCR != 0 {
		t.Error("expected zero PCR for nil")
	}
}

func TestClassifyOIBuildup(t *testing.T) {
	tests := []struct {
		price    float64
		oi       int64
		expected models.OIBuildupType
	}{
		{10, 5000, models.LongBuildup},
		{-10, 5000, models.ShortBuildup},
		{-10, -5000, models.LongUnwinding},
		{10, -5000, models.ShortCovering},
	}

	for _, tt := range tests {
		result := ClassifyOIBuildup(tt.price, tt.oi)
		if result != tt.expected {
			t.Errorf("OI buildup for price=%.0f oi=%d: expected %s, got %s",
				tt.price, tt.oi, tt.expected, result)
		}
	}
}

func TestAnalyzeOIBuildup(t *testing.T) {
	oc := sampleOptionChain()
	fut := &models.FuturesContract{
		Ticker:   "NIFTY",
		LTP:      25050,
		Change:   50,
		OI:       1000000,
		OIChange: 50000,
	}

	a := AnalyzeOIBuildup(oc, fut)

	if a.FuturesBuildup.Buildup != models.LongBuildup {
		t.Errorf("expected long buildup, got %s", a.FuturesBuildup.Buildup)
	}
	if len(a.TopLongBuildup) == 0 {
		t.Error("expected top long buildups")
	}
	if len(a.TopShortBuildup) == 0 {
		t.Error("expected top short buildups")
	}
}

func TestBuildBullCallSpread(t *testing.T) {
	oc := sampleOptionChain()
	strat := BuildBullCallSpread(oc, 25)

	if strat.Name != "Bull Call Spread" {
		t.Errorf("expected Bull Call Spread, got %s", strat.Name)
	}
	if len(strat.Legs) != 2 {
		t.Errorf("expected 2 legs, got %d", len(strat.Legs))
	}
	if len(strat.Breakevens) == 0 {
		t.Error("expected breakevens")
	}
}

func TestBuildIronCondor(t *testing.T) {
	oc := sampleOptionChain()
	strat := BuildIronCondor(oc, 25, 200)

	if strat.Name != "Iron Condor" {
		t.Errorf("expected Iron Condor, got %s", strat.Name)
	}
	if len(strat.Legs) != 4 {
		t.Errorf("expected 4 legs, got %d", len(strat.Legs))
	}
}

func TestAnalyzeFuturesBasis(t *testing.T) {
	fut := &models.FuturesContract{
		Ticker: "RELIANCE",
		LTP:    1280,
	}
	result := AnalyzeFuturesBasis(fut, 1250, 20)

	if result.Basis != 30 {
		t.Errorf("expected basis=30, got %.2f", result.Basis)
	}
	if result.BasisPct <= 0 {
		t.Error("expected positive basis pct")
	}
	if result.Annualized <= 0 {
		t.Error("expected positive annualized basis")
	}
}

func TestAnalyzeFuturesBasisNil(t *testing.T) {
	result := AnalyzeFuturesBasis(nil, 1250, 20)
	if result.Ticker != "" {
		t.Error("expected empty for nil")
	}
}

func TestFullDerivativesAnalysis(t *testing.T) {
	oc := sampleOptionChain()
	fut := &models.FuturesContract{
		Ticker:   "NIFTY",
		LTP:      25050,
		Change:   50,
		OI:       1000000,
		OIChange: 50000,
	}

	result := FullDerivativesAnalysis("NIFTY", oc, fut)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Type != models.AnalysisDerivatives {
		t.Errorf("expected derivatives type, got %s", result.Type)
	}
	if len(result.Signals) == 0 {
		t.Error("expected signals")
	}
}

func TestFullDerivativesAnalysisNil(t *testing.T) {
	result := FullDerivativesAnalysis("NIFTY", nil, nil)
	if result != nil {
		t.Error("expected nil for nil chain")
	}
}
