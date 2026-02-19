package fundamental

import (
	"testing"

	"github.com/seenimoa/openseai/pkg/models"
)

func sampleFinancialData() *models.FinancialData {
	return &models.FinancialData{
		Ticker: "RELIANCE",
		AnnualIncome: []models.IncomeStatement{
			{Period: "Mar 2025", Revenue: 250000, EBITDA: 50000, EBIT: 40000, PBT: 35000, PAT: 28000, EPS: 50, OPMPct: 20, NPMPct: 11.2, InterestExpense: 5000},
			{Period: "Mar 2024", Revenue: 220000, EBITDA: 44000, EBIT: 35000, PBT: 30000, PAT: 24000, EPS: 42, OPMPct: 18, NPMPct: 10.9, InterestExpense: 5500},
			{Period: "Mar 2023", Revenue: 200000, EBITDA: 40000, EBIT: 32000, PBT: 27000, PAT: 21000, EPS: 37, OPMPct: 17, NPMPct: 10.5},
			{Period: "Mar 2022", Revenue: 180000, EBITDA: 36000, EBIT: 29000, PBT: 24000, PAT: 19000, EPS: 33, OPMPct: 16, NPMPct: 10.6},
			{Period: "Mar 2021", Revenue: 160000, EBITDA: 32000, EBIT: 26000, PBT: 21000, PAT: 17000, EPS: 30},
			{Period: "Mar 2020", Revenue: 150000, EBITDA: 30000, EBIT: 24000, PBT: 19000, PAT: 15000, EPS: 27},
		},
		QuarterlyIncome: []models.IncomeStatement{
			{Period: "Q4 FY25", Revenue: 65000, PAT: 7500},
			{Period: "Q3 FY25", Revenue: 62000, PAT: 7000},
			{Period: "Q2 FY25", Revenue: 60000, PAT: 6800},
			{Period: "Q1 FY25", Revenue: 58000, PAT: 6500},
			{Period: "Q4 FY24", Revenue: 55000, PAT: 6000},
		},
		AnnualBalanceSheet: []models.BalanceSheet{
			{Period: "Mar 2025", TotalAssets: 500000, TotalEquity: 200000, TotalDebt: 80000, CurrentAssets: 120000, CurrentLiabilities: 70000, ShareCapital: 10000, CashEquivalents: 30000},
			{Period: "Mar 2024", TotalAssets: 450000, TotalEquity: 180000, TotalDebt: 85000, CurrentAssets: 100000, CurrentLiabilities: 65000, ShareCapital: 10000, CashEquivalents: 25000},
		},
		AnnualCashFlow: []models.CashFlow{
			{Period: "Mar 2025", OperatingCashFlow: 40000, FreeCashFlow: 25000},
		},
	}
}

func TestComputeRatios(t *testing.T) {
	fin := sampleFinancialData()
	price := 1250.0
	shares := 560.0 // in crores

	ratios := ComputeRatios(fin, price, shares)

	if ratios.EPS != 50 {
		t.Errorf("expected EPS=50, got %.2f", ratios.EPS)
	}
	if ratios.PE <= 0 {
		t.Errorf("expected positive PE, got %.2f", ratios.PE)
	}
	if ratios.ROE <= 0 {
		t.Errorf("expected positive ROE, got %.2f", ratios.ROE)
	}
	if ratios.ROCE <= 0 {
		t.Errorf("expected positive ROCE, got %.2f", ratios.ROCE)
	}
	if ratios.DebtEquity <= 0 {
		t.Errorf("expected positive D/E, got %.2f", ratios.DebtEquity)
	}
	if ratios.GrahamNumber <= 0 {
		t.Errorf("expected positive Graham Number, got %.2f", ratios.GrahamNumber)
	}
}

func TestComputeRatiosNilData(t *testing.T) {
	ratios := ComputeRatios(nil, 100, 100)
	if ratios.PE != 0 {
		t.Error("expected zero PE for nil data")
	}
}

func TestComputeGrowth(t *testing.T) {
	fin := sampleFinancialData()
	g := ComputeGrowth(fin)

	if g.RevenueGrowthQoQ <= 0 {
		t.Errorf("expected positive QoQ revenue growth, got %.2f", g.RevenueGrowthQoQ)
	}
	if g.RevenueGrowthYoY <= 0 {
		t.Errorf("expected positive YoY revenue growth, got %.2f", g.RevenueGrowthYoY)
	}
	if g.RevenueCAGR3Y <= 0 {
		t.Errorf("expected positive 3Y CAGR, got %.2f", g.RevenueCAGR3Y)
	}
	if g.EPSGrowthYoY <= 0 {
		t.Errorf("expected positive EPS growth, got %.2f", g.EPSGrowthYoY)
	}
}

func TestComputeGrowthNil(t *testing.T) {
	g := ComputeGrowth(nil)
	if g.RevenueGrowthQoQ != 0 {
		t.Error("expected zero growth for nil data")
	}
}

func TestAssessFinancialHealth(t *testing.T) {
	fin := sampleFinancialData()
	ratios := ComputeRatios(fin, 1250, 560)
	health := AssessFinancialHealth(ratios, fin)

	if health.Score <= 0 || health.Score > 100 {
		t.Errorf("score out of range: %.2f", health.Score)
	}
	if health.Grade == "" {
		t.Error("expected non-empty grade")
	}
	if len(health.Components) == 0 {
		t.Error("expected components to be populated")
	}
}

func TestPiotroskiFScore(t *testing.T) {
	fin := sampleFinancialData()
	qs := PiotroskiFScore(fin)

	if qs.Score < 0 || qs.Score > 9 {
		t.Errorf("F-score out of range: %d", qs.Score)
	}
	if len(qs.Checks) == 0 {
		t.Error("expected checks to be populated")
	}
	// With our sample data, should have at least some passes.
	if qs.Score < 3 {
		t.Errorf("expected F-score >= 3 for good sample data, got %d", qs.Score)
	}
}

func TestPiotroskiFScoreNil(t *testing.T) {
	qs := PiotroskiFScore(nil)
	if qs.Score != 0 {
		t.Errorf("expected 0 for nil, got %d", qs.Score)
	}
}

func TestGrahamNumber(t *testing.T) {
	gn := GrahamNumber(50, 350)
	if gn <= 0 {
		t.Errorf("expected positive Graham Number, got %.2f", gn)
	}
	// sqrt(22.5 * 50 * 350) = sqrt(393750) ≈ 627.5
	if gn < 600 || gn > 650 {
		t.Errorf("Graham Number out of expected range: %.2f", gn)
	}
}

func TestGrahamNumberNegative(t *testing.T) {
	if gn := GrahamNumber(-10, 350); gn != 0 {
		t.Errorf("expected 0 for negative EPS, got %.2f", gn)
	}
}

func TestDCF(t *testing.T) {
	val := DCF(DCFParams{
		FreeCashFlow:      25000,
		GrowthRateYr1_5:   0.15,
		GrowthRateYr6_10:  0.08,
		TerminalGrowth:    0.03,
		DiscountRate:      0.12,
		SharesOutstanding: 560,
	})
	if val <= 0 {
		t.Errorf("expected positive DCF value, got %.2f", val)
	}
}

func TestDCFZeroFCF(t *testing.T) {
	val := DCF(DCFParams{FreeCashFlow: 0, DiscountRate: 0.12, SharesOutstanding: 100})
	if val != 0 {
		t.Errorf("expected 0 for zero FCF, got %.2f", val)
	}
}

func TestComputeValuation(t *testing.T) {
	fin := sampleFinancialData()
	ratios := ComputeRatios(fin, 1250, 560)
	v := ComputeValuation("RELIANCE", 1250, ratios, fin, 560, 25)

	if v.Ticker != "RELIANCE" {
		t.Errorf("expected RELIANCE, got %s", v.Ticker)
	}
	if v.GrahamNumber <= 0 {
		t.Errorf("expected positive Graham number, got %.2f", v.GrahamNumber)
	}
	if v.Verdict == "" {
		t.Error("expected a verdict")
	}
}

func TestComparePeers(t *testing.T) {
	target := PeerEntry{
		Ticker: "RELIANCE",
		Ratios: models.FinancialRatios{ROE: 18, ROCE: 16, PE: 25, DebtEquity: 0.4, DividendYield: 1.5},
	}
	peers := []PeerEntry{
		{Ticker: "TCS", Ratios: models.FinancialRatios{ROE: 25, ROCE: 28, PE: 30, DebtEquity: 0, DividendYield: 2}},
		{Ticker: "INFY", Ratios: models.FinancialRatios{ROE: 22, ROCE: 24, PE: 28, DebtEquity: 0, DividendYield: 3}},
		{Ticker: "HDFCBANK", Ratios: models.FinancialRatios{ROE: 15, ROCE: 12, PE: 20, DebtEquity: 5, DividendYield: 1}},
	}

	pc := ComparePeers(target, peers)
	if pc.Target.Rank <= 0 {
		t.Error("expected target to have a rank")
	}
	if len(pc.Peers) != 3 {
		t.Errorf("expected 3 peers, got %d", len(pc.Peers))
	}
	if pc.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestRelativeValuationMetrics(t *testing.T) {
	target := models.FinancialRatios{PE: 25, ROE: 18, ROCE: 16}
	peers := []models.FinancialRatios{
		{PE: 30, ROE: 25, ROCE: 28},
		{PE: 20, ROE: 15, ROCE: 12},
		{PE: 28, ROE: 22, ROCE: 24},
	}
	metrics := RelativeValuationMetrics(target, peers)
	if len(metrics) == 0 {
		t.Error("expected relative metrics")
	}
}

func TestAnalyzePromoterHolding(t *testing.T) {
	p := &models.PromoterData{
		PromoterHolding: 55,
		PromoterPledge:  5,
		FIIHolding:      25,
		DIIHolding:      20,
		MFHolding:       12,
		PromoterTrend: []models.HoldingPoint{
			{Quarter: "Dec 2025", Pct: 55},
			{Quarter: "Sep 2025", Pct: 53},
			{Quarter: "Jun 2025", Pct: 52},
		},
	}

	a := AnalyzePromoterHolding(p)
	if a.CurrentHolding != 55 {
		t.Errorf("expected 55, got %.2f", a.CurrentHolding)
	}
	if a.TrendDirection != "increasing" {
		t.Errorf("expected increasing trend, got %s", a.TrendDirection)
	}
	if len(a.Signals) == 0 {
		t.Error("expected signals")
	}
}

func TestPromoterSignal(t *testing.T) {
	p := &models.PromoterData{
		PromoterHolding: 55,
		PromoterPledge:  2,
		PromoterTrend: []models.HoldingPoint{
			{Quarter: "Dec 2025", Pct: 55},
			{Quarter: "Sep 2025", Pct: 53},
		},
	}
	sig := PromoterSignal(p)
	if sig.Source != "Promoter" {
		t.Errorf("expected source Promoter, got %s", sig.Source)
	}
	if sig.Type != models.SignalBuy {
		t.Errorf("expected BUY for increasing + low pledge, got %s", sig.Type)
	}
}

func TestFormatFinancialSummary(t *testing.T) {
	fin := sampleFinancialData()
	ratios := ComputeRatios(fin, 1250, 560)
	growth := ComputeGrowth(fin)
	health := AssessFinancialHealth(ratios, fin)
	summary := FormatFinancialSummary(ratios, growth, health)
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}
