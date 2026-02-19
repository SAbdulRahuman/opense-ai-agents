package report

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Test Helpers
// ════════════════════════════════════════════════════════════════════

func sampleBars(n int) []models.OHLCV {
	bars := make([]models.OHLCV, n)
	base := 2500.0
	t := time.Date(2025, 1, 1, 9, 15, 0, 0, time.UTC)
	for i := range bars {
		open := base + float64(i)*0.5
		close := open + float64(i%5) - 2
		high := math.Max(open, close) + 5
		low := math.Min(open, close) - 5
		bars[i] = models.OHLCV{
			Timestamp: t.AddDate(0, 0, i),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    int64(1000000 + i*50000),
		}
	}
	return bars
}

func sampleAnalysis() *models.CompositeAnalysis {
	return &models.CompositeAnalysis{
		Ticker: "RELIANCE",
		StockProfile: models.StockProfile{
			Stock: models.Stock{
				Ticker:   "RELIANCE",
				Name:     "Reliance Industries Ltd",
				Exchange: "NSE",
				Sector:   "Oil & Gas",
				Industry: "Refineries",
			},
			Quote: &models.Quote{
				Ticker:        "RELIANCE",
				Name:          "Reliance Industries Ltd",
				LastPrice:     2876.50,
				Change:        42.30,
				ChangePct:     1.49,
				Open:          2840.00,
				High:          2890.00,
				Low:           2835.00,
				PrevClose:     2834.20,
				Volume:        12345678,
				WeekHigh52:    3024.90,
				WeekLow52:     2220.30,
				MarketCap:     1946000000000,
				PE:            28.5,
				PB:            2.8,
				DividendYield: 0.35,
				Timestamp:     time.Now(),
			},
			Historical: sampleBars(60),
			Ratios: &models.FinancialRatios{
				PE:               28.5,
				PB:               2.8,
				EVBITDA:          18.2,
				ROE:              12.5,
				ROCE:             14.3,
				DebtEquity:       0.45,
				CurrentRatio:     1.2,
				InterestCoverage: 8.5,
				DividendYield:    0.35,
				EPS:              100.93,
				BookValue:        1027.32,
				PEGRatio:         1.8,
				GrahamNumber:     1525.00,
			},
		},
		Technical: &models.AnalysisResult{
			Ticker:         "RELIANCE",
			Type:           models.AnalysisTechnical,
			AgentName:      "TechnicalAgent",
			Recommendation: models.ModerateBuy,
			Confidence:     0.72,
			Summary:        "RSI at 58 indicates moderate bullish momentum. MACD crossover is positive. Price above 50-DMA, suggesting uptrend continuation.",
			Signals: []models.Signal{
				{Source: "RSI", Type: models.SignalBuy, Confidence: 0.65, Reason: "RSI at 58, above neutral 50"},
				{Source: "MACD", Type: models.SignalBuy, Confidence: 0.78, Reason: "MACD histogram turning positive"},
				{Source: "Moving Avg", Type: models.SignalBuy, Confidence: 0.70, Reason: "Price above 20 & 50 DMA"},
				{Source: "SuperTrend", Type: models.SignalNeutral, Confidence: 0.55, Reason: "SuperTrend just turned bullish, needs confirmation"},
			},
		},
		Fundamental: &models.AnalysisResult{
			Ticker:         "RELIANCE",
			Type:           models.AnalysisFundamental,
			AgentName:      "FundamentalAgent",
			Recommendation: models.ModerateBuy,
			Confidence:     0.68,
			Summary:        "Strong revenue growth at 15% YoY. Debt-to-equity improving. ROE at 12.5% is above industry average. PE of 28.5x is slightly above historical median.",
			Signals: []models.Signal{
				{Source: "Revenue Growth", Type: models.SignalBuy, Confidence: 0.75, Reason: "15% YoY revenue growth"},
				{Source: "ROE", Type: models.SignalBuy, Confidence: 0.70, Reason: "ROE of 12.5% above industry avg"},
				{Source: "Valuation", Type: models.SignalNeutral, Confidence: 0.50, Reason: "PE 28.5x slightly above 5Y median of 26x"},
			},
		},
		Derivatives: &models.AnalysisResult{
			Ticker:         "RELIANCE",
			Type:           models.AnalysisDerivatives,
			AgentName:      "DerivativesAgent",
			Recommendation: models.ModerateBuy,
			Confidence:     0.65,
			Summary:        "Long buildup observed. PCR at 1.2 is moderately bullish. Max pain at ₹2,850 suggests support.",
			Signals: []models.Signal{
				{Source: "OI Analysis", Type: models.SignalBuy, Confidence: 0.70, Reason: "Long buildup: price up, OI up"},
				{Source: "PCR", Type: models.SignalBuy, Confidence: 0.60, Reason: "PCR 1.2 moderately bullish"},
			},
		},
		Sentiment: &models.AnalysisResult{
			Ticker:         "RELIANCE",
			Type:           models.AnalysisSentiment,
			AgentName:      "SentimentAgent",
			Recommendation: models.ModerateBuy,
			Confidence:     0.62,
			Summary:        "Overall positive sentiment. 7 of 10 recent articles are bullish. Analyst consensus is positive.",
			Signals: []models.Signal{
				{Source: "News", Type: models.SignalBuy, Confidence: 0.65, Reason: "70% positive news articles"},
				{Source: "Social", Type: models.SignalNeutral, Confidence: 0.50, Reason: "Mixed social media sentiment"},
			},
		},
		Risk: &models.AnalysisResult{
			Ticker:         "RELIANCE",
			Type:           models.AnalysisRisk,
			AgentName:      "RiskAgent",
			Recommendation: models.Hold,
			Confidence:     0.60,
			Summary:        "Moderate risk. Beta of 1.1 indicates slightly above-market volatility. Sector headwinds from crude oil prices.",
			Signals: []models.Signal{
				{Source: "Beta", Type: models.SignalNeutral, Confidence: 0.55, Reason: "Beta 1.1 — slightly above market"},
				{Source: "Sector", Type: models.SignalSell, Confidence: 0.45, Reason: "Crude oil price volatility risk"},
			},
		},
		Recommendation:  models.ModerateBuy,
		Confidence:      0.70,
		Summary:         "Overall BUY recommendation with 70% confidence. Technical and fundamental signals align positively. Monitor crude oil prices for sector-level risk.",
		EntryPrice:      2876.50,
		TargetPrice:     3100.00,
		StopLoss:        2750.00,
		RiskRewardRatio: 1.77,
		Timeframe:       "medium-term",
		Timestamp:       time.Now(),
	}
}

// ════════════════════════════════════════════════════════════════════
// Chart Tests
// ════════════════════════════════════════════════════════════════════

func TestCandlestickChart_Basic(t *testing.T) {
	bars := sampleBars(30)
	cfg := DefaultChartConfig()
	cfg.Title = "Test Candlestick"

	svg := CandlestickChart(bars, nil, cfg)
	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG output")
	}
	if !strings.Contains(svg, "Test Candlestick") {
		t.Error("expected title in SVG")
	}
	if !strings.Contains(svg, "rect") {
		t.Error("expected rectangles (candles) in SVG")
	}
}

func TestCandlestickChart_WithOverlays(t *testing.T) {
	bars := sampleBars(20)
	overlays := map[string][]float64{
		"SMA 20": make([]float64, 20),
	}
	for i := range overlays["SMA 20"] {
		overlays["SMA 20"][i] = bars[i].Close + 5
	}

	svg := CandlestickChart(bars, overlays, DefaultChartConfig())
	if !strings.Contains(svg, "SMA 20") {
		t.Error("expected overlay legend in SVG")
	}
	if !strings.Contains(svg, "path") {
		t.Error("expected path element for overlay line")
	}
}

func TestCandlestickChart_Empty(t *testing.T) {
	svg := CandlestickChart(nil, nil, DefaultChartConfig())
	if !strings.Contains(svg, "No data") {
		t.Error("expected empty message for nil bars")
	}
}

func TestCandlestickChart_SingleBar(t *testing.T) {
	bars := sampleBars(1)
	svg := CandlestickChart(bars, nil, DefaultChartConfig())
	if !strings.Contains(svg, "<svg") {
		t.Error("expected valid SVG for single bar")
	}
}

func TestCandlestickChart_ZeroConfig(t *testing.T) {
	bars := sampleBars(10)
	svg := CandlestickChart(bars, nil, ChartConfig{})
	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG with zero config (auto-defaults)")
	}
}

func TestLineChart_Basic(t *testing.T) {
	series := []LineChartSeries{
		{Name: "Stock", Values: []float64{100, 105, 102, 110, 108}, Color: "#2196f3"},
		{Name: "Nifty", Values: []float64{100, 103, 101, 106, 104}, Color: "#ff9800"},
	}
	labels := []string{"Mon", "Tue", "Wed", "Thu", "Fri"}

	cfg := DefaultChartConfig()
	cfg.Title = "Performance Comparison"

	svg := LineChart(series, labels, cfg)
	if !strings.Contains(svg, "Performance Comparison") {
		t.Error("expected title")
	}
	if !strings.Contains(svg, "Stock") {
		t.Error("expected series name in legend")
	}
	if !strings.Contains(svg, "Nifty") {
		t.Error("expected second series name")
	}
}

func TestLineChart_Empty(t *testing.T) {
	svg := LineChart(nil, nil, DefaultChartConfig())
	if !strings.Contains(svg, "No data") {
		t.Error("expected empty message")
	}
}

func TestLineChart_SinglePoint(t *testing.T) {
	series := []LineChartSeries{{Name: "A", Values: []float64{42}}}
	svg := LineChart(series, nil, DefaultChartConfig())
	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG for single point")
	}
}

func TestLineChart_NaN(t *testing.T) {
	series := []LineChartSeries{
		{Name: "Test", Values: []float64{10, math.NaN(), 20, math.NaN(), 30}},
	}
	svg := LineChart(series, nil, DefaultChartConfig())
	if !strings.Contains(svg, "path") {
		t.Error("expected path even with NaN values")
	}
}

func TestHorizontalBarChart_Basic(t *testing.T) {
	items := []BarItem{
		{Label: "ROE", Value: 12.5},
		{Label: "ROCE", Value: 14.3},
		{Label: "ROA", Value: 8.2},
	}

	cfg := DefaultChartConfig()
	cfg.Title = "Return Ratios"

	svg := HorizontalBarChart(items, cfg)
	if !strings.Contains(svg, "Return Ratios") {
		t.Error("expected title")
	}
	if !strings.Contains(svg, "ROE") {
		t.Error("expected label 'ROE'")
	}
}

func TestHorizontalBarChart_WithNegative(t *testing.T) {
	items := []BarItem{
		{Label: "Profit", Value: 500},
		{Label: "Loss", Value: -200},
	}
	svg := HorizontalBarChart(items, DefaultChartConfig())
	if !strings.Contains(svg, "Profit") {
		t.Error("expected Profit label")
	}
	// Should contain a zero line for mixed values
	if !strings.Contains(svg, "line") {
		t.Error("expected zero line for mixed positive/negative")
	}
}

func TestHorizontalBarChart_Empty(t *testing.T) {
	svg := HorizontalBarChart(nil, DefaultChartConfig())
	if !strings.Contains(svg, "No data") {
		t.Error("expected empty message")
	}
}

func TestValuationBandChart_Basic(t *testing.T) {
	data := make([]BandDataPoint, 12)
	for i := range data {
		data[i] = BandDataPoint{
			Date:  time.Date(2025, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC),
			Value: 25 + float64(i)*0.5,
			Price: 2500 + float64(i)*50,
		}
	}

	svg := ValuationBandChart(data, "PE", DefaultChartConfig())
	if !strings.Contains(svg, "PE") {
		t.Error("expected band name in chart")
	}
	if !strings.Contains(svg, "Price") {
		t.Error("expected 'Price' series legend")
	}
}

func TestValuationBandChart_Empty(t *testing.T) {
	svg := ValuationBandChart(nil, "PE", DefaultChartConfig())
	if !strings.Contains(svg, "No valuation") {
		t.Error("expected empty message")
	}
}

func TestOptionPayoffChart_Basic(t *testing.T) {
	payoff := make([]models.OptionPayoff, 50)
	for i := range payoff {
		price := 2500 + float64(i)*20
		payoff[i] = models.OptionPayoff{
			UnderlyingPrice: price,
			PnL:             price - 2800,
		}
	}

	cfg := DefaultChartConfig()
	svg := OptionPayoffChart(payoff, "Bull Call Spread", cfg)
	if !strings.Contains(svg, "Bull Call Spread") {
		t.Error("expected strategy name in title")
	}
	if !strings.Contains(svg, "path") {
		t.Error("expected payoff line path")
	}
}

func TestOptionPayoffChart_Empty(t *testing.T) {
	svg := OptionPayoffChart(nil, "Test", DefaultChartConfig())
	if !strings.Contains(svg, "No payoff") {
		t.Error("expected empty message")
	}
}

func TestGaugeChart_Values(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		label string
	}{
		{"low", 15, "Low"},
		{"medium_low", 40, "Medium Low"},
		{"medium", 55, "Medium"},
		{"high", 85, "High"},
		{"clamped_zero", -10, "Below Zero"},
		{"clamped_max", 150, "Above Max"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg := GaugeChart(tt.value, tt.label, 200)
			if !strings.Contains(svg, "<svg") {
				t.Error("expected SVG output")
			}
			if !strings.Contains(svg, tt.label) {
				t.Errorf("expected label '%s' in output", tt.label)
			}
		})
	}
}

func TestGaugeChart_ZeroWidth(t *testing.T) {
	svg := GaugeChart(50, "Test", 0)
	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG with auto-width")
	}
}

// ════════════════════════════════════════════════════════════════════
// Report Generator Tests
// ════════════════════════════════════════════════════════════════════

func TestGenerateHTML_Basic(t *testing.T) {
	analysis := sampleAnalysis()
	cfg := DefaultReportConfig()

	html, err := GenerateHTML(analysis, cfg)
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}

	checks := []struct {
		name    string
		substr  string
	}{
		{"html tag", "<html"},
		{"ticker", "RELIANCE"},
		{"company name", "Reliance Industries"},
		{"sector", "Oil &amp; Gas"},
		{"recommendation", "Buy"},
		{"confidence", "70%"},
		{"entry price", "2,876"},
		{"target price", "3,100"},
		{"technical section", "Technical Analysis"},
		{"fundamental section", "Fundamental Analysis"},
		{"derivatives section", "Derivatives"},
		{"sentiment section", "Sentiment"},
		{"risk section", "Risk Assessment"},
		{"disclaimer", "Disclaimer"},
		{"CSS", "font-family"},
		{"signal badge", "signal-badge"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(html, c.substr) {
				t.Errorf("expected '%s' in HTML output", c.substr)
			}
		})
	}
}

func TestGenerateHTML_NilAnalysis(t *testing.T) {
	_, err := GenerateHTML(nil, DefaultReportConfig())
	if err == nil {
		t.Error("expected error for nil analysis")
	}
}

func TestGenerateHTML_MinimalAnalysis(t *testing.T) {
	analysis := &models.CompositeAnalysis{
		Ticker:         "TCS",
		Recommendation: models.Hold,
		Confidence:     0.50,
		Summary:        "Hold recommendation.",
	}

	html, err := GenerateHTML(analysis, DefaultReportConfig())
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}
	if !strings.Contains(html, "TCS") {
		t.Error("expected ticker in output")
	}
	if !strings.Contains(html, "Hold") {
		t.Error("expected recommendation")
	}
}

func TestGenerateHTML_SelectedSections(t *testing.T) {
	analysis := sampleAnalysis()
	cfg := DefaultReportConfig()
	cfg.Sections = []ReportSection{SectionSummary, SectionTechnical}

	html, err := GenerateHTML(analysis, cfg)
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}

	if !strings.Contains(html, "Technical Analysis") {
		t.Error("expected technical section")
	}
	// Fundamental should NOT appear (not in sections)
	if strings.Contains(html, "Fundamental Analysis") {
		t.Error("did not expect fundamental section when not selected")
	}
}

func TestGenerateHTML_CustomTitle(t *testing.T) {
	analysis := sampleAnalysis()
	cfg := DefaultReportConfig()
	cfg.Title = "Custom Title Report"

	html, err := GenerateHTML(analysis, cfg)
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}
	if !strings.Contains(html, "Custom Title Report") {
		t.Error("expected custom title in HTML")
	}
}

func TestGenerateHTML_WithPriceChart(t *testing.T) {
	analysis := sampleAnalysis()
	cfg := DefaultReportConfig()

	html, err := GenerateHTML(analysis, cfg)
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}
	if !strings.Contains(html, "<svg") {
		t.Error("expected embedded SVG chart in HTML")
	}
	if !strings.Contains(html, "Price Chart") {
		t.Error("expected price chart section heading")
	}
}

func TestGenerateHTML_WithOptionStrategy(t *testing.T) {
	analysis := sampleAnalysis()
	analysis.Derivatives.Details = map[string]any{
		"strategy": &models.OptionStrategy{
			Name:       "Bull Call Spread",
			MaxProfit:  15000,
			MaxLoss:    5000,
			Breakevens: []float64{2900},
			Payoff: []models.OptionPayoff{
				{UnderlyingPrice: 2700, PnL: -5000},
				{UnderlyingPrice: 2800, PnL: -2000},
				{UnderlyingPrice: 2900, PnL: 0},
				{UnderlyingPrice: 3000, PnL: 10000},
				{UnderlyingPrice: 3100, PnL: 15000},
			},
		},
	}

	cfg := DefaultReportConfig()
	html, err := GenerateHTML(analysis, cfg)
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}
	if !strings.Contains(html, "Bull Call Spread") {
		t.Error("expected option strategy name")
	}
}

func TestGenerateText_Basic(t *testing.T) {
	analysis := sampleAnalysis()
	cfg := DefaultReportConfig()

	text, err := GenerateText(analysis, cfg)
	if err != nil {
		t.Fatalf("GenerateText failed: %v", err)
	}

	checks := []string{
		"RELIANCE",
		"Reliance Industries",
		"RECOMMENDATION",
		"FUNDAMENTAL",
		"TECHNICAL",
		"DERIVATIVES",
		"SENTIMENT",
		"RISK",
		"Disclaimer",
	}

	for _, c := range checks {
		if !strings.Contains(text, c) {
			t.Errorf("expected '%s' in text report", c)
		}
	}
}

func TestGenerateText_NilAnalysis(t *testing.T) {
	_, err := GenerateText(nil, DefaultReportConfig())
	if err == nil {
		t.Error("expected error for nil analysis")
	}
}

func TestGenerateText_MinimalAnalysis(t *testing.T) {
	analysis := &models.CompositeAnalysis{
		Ticker:         "INFY",
		Recommendation: models.StrongBuy,
		Confidence:     0.85,
		Summary:        "Strong buy.",
	}

	text, err := GenerateText(analysis, DefaultReportConfig())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(text, "INFY") {
		t.Error("expected ticker")
	}
	if !strings.Contains(text, "Strong Buy") {
		t.Error("expected recommendation")
	}
}

// ════════════════════════════════════════════════════════════════════
// Report Config Tests
// ════════════════════════════════════════════════════════════════════

func TestDefaultReportConfig(t *testing.T) {
	cfg := DefaultReportConfig()
	if cfg.Format != FormatHTML {
		t.Errorf("expected HTML format, got %s", cfg.Format)
	}
	if cfg.Author != "OpeNSE.ai Agent" {
		t.Errorf("expected default author, got %s", cfg.Author)
	}
	if len(cfg.Sections) != 7 {
		t.Errorf("expected 7 sections, got %d", len(cfg.Sections))
	}
}

func TestHasSection(t *testing.T) {
	cfg := DefaultReportConfig()
	if !cfg.hasSection(SectionTechnical) {
		t.Error("expected technical section in default config")
	}

	cfg.Sections = []ReportSection{SectionSummary}
	if cfg.hasSection(SectionTechnical) {
		t.Error("did not expect technical section with only summary")
	}
	if !cfg.hasSection(SectionSummary) {
		t.Error("expected summary section")
	}
}

func TestAllSections(t *testing.T) {
	sections := AllSections()
	if len(sections) != 7 {
		t.Errorf("expected 7 sections, got %d", len(sections))
	}
	// Check all unique
	seen := make(map[ReportSection]bool)
	for _, s := range sections {
		if seen[s] {
			t.Errorf("duplicate section: %s", s)
		}
		seen[s] = true
	}
}

// ════════════════════════════════════════════════════════════════════
// Data Building Tests
// ════════════════════════════════════════════════════════════════════

func TestFlattenSignals(t *testing.T) {
	signals := []models.Signal{
		{Source: "RSI", Type: models.SignalBuy, Confidence: 0.75, Reason: "Bullish"},
		{Source: "MACD", Type: models.SignalSell, Confidence: 0.60, Reason: "Bearish"},
		{Source: "Vol", Type: models.SignalNeutral, Confidence: 0.50, Reason: "Average"},
	}

	rows := flattenSignals(signals)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	if rows[0].TypeClass != "buy" {
		t.Errorf("expected 'buy' class, got %s", rows[0].TypeClass)
	}
	if rows[1].TypeClass != "sell" {
		t.Errorf("expected 'sell' class, got %s", rows[1].TypeClass)
	}
	if rows[2].TypeClass != "neutral" {
		t.Errorf("expected 'neutral' class, got %s", rows[2].TypeClass)
	}
	if rows[0].Confidence != "75%" {
		t.Errorf("expected '75%%', got %s", rows[0].Confidence)
	}
}

func TestFormatRecommendation(t *testing.T) {
	tests := []struct {
		input    models.Recommendation
		expected string
	}{
		{models.StrongBuy, "Strong Buy"},
		{models.ModerateBuy, "Buy"},
		{models.Hold, "Hold"},
		{models.ModerateSell, "Sell"},
		{models.StrongSell, "Strong Sell"},
		{models.Recommendation("UNKNOWN"), "UNKNOWN"},
	}

	for _, tt := range tests {
		result := formatRecommendation(tt.input)
		if result != tt.expected {
			t.Errorf("formatRecommendation(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestRecommendationClass(t *testing.T) {
	tests := []struct {
		input    models.Recommendation
		expected string
	}{
		{models.StrongBuy, "strong-buy"},
		{models.ModerateBuy, "buy"},
		{models.Hold, "hold"},
		{models.ModerateSell, "sell"},
		{models.StrongSell, "strong-sell"},
	}

	for _, tt := range tests {
		result := recommendationClass(tt.input)
		if result != tt.expected {
			t.Errorf("recommendationClass(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestBuildRatioRows(t *testing.T) {
	ratios := &models.FinancialRatios{PE: 25.5, PB: 3.2, ROE: 15.0}
	rows := buildRatioRows(ratios)
	if len(rows) != 13 {
		t.Fatalf("expected 13 ratio rows, got %d", len(rows))
	}
	if rows[0].Label != "P/E Ratio" {
		t.Errorf("expected 'P/E Ratio', got '%s'", rows[0].Label)
	}
	if rows[0].Value != "25.50" {
		t.Errorf("expected '25.50', got '%s'", rows[0].Value)
	}
}

func TestBuildOverlaysFromDetails(t *testing.T) {
	// nil input
	if overlays := buildOverlaysFromDetails(nil); overlays != nil {
		t.Error("expected nil overlays for nil input")
	}

	// With SMA data
	result := &models.AnalysisResult{
		Details: map[string]any{
			"sma_20": []float64{100, 101, 102},
			"sma_50": []float64{98, 99, 100},
		},
	}
	overlays := buildOverlaysFromDetails(result)
	if len(overlays) != 2 {
		t.Errorf("expected 2 overlays, got %d", len(overlays))
	}
	if _, ok := overlays["SMA 20"]; !ok {
		t.Error("expected SMA 20 overlay")
	}
}

// ════════════════════════════════════════════════════════════════════
// PDF Tests
// ════════════════════════════════════════════════════════════════════

func TestDetectPDFEngine(t *testing.T) {
	engine := DetectPDFEngine()
	// Just verify it returns a valid engine type (could be EngineNone)
	switch engine {
	case EngineWKHTML, EngineChromium, EngineNone:
		// OK
	default:
		t.Errorf("unexpected engine: %s", engine)
	}
}

func TestIsPDFSupported(t *testing.T) {
	// Just verify it doesn't panic
	_ = IsPDFSupported()
}

func TestGeneratePDF_NoOutputPath(t *testing.T) {
	err := GeneratePDF("<html></html>", PDFConfig{})
	if err == nil {
		t.Error("expected error for empty output path")
	}
}

func TestGeneratePDF_HTMLFallback(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test.pdf")

	cfg := PDFConfig{
		Engine:     EngineNone,
		OutputPath: outPath,
	}

	html := "<html><body>Test Report</body></html>"
	err := GeneratePDF(html, cfg)
	if err != nil {
		t.Fatalf("GeneratePDF fallback failed: %v", err)
	}

	// Should have written .html instead of .pdf
	htmlPath := filepath.Join(tmpDir, "test.html")
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("reading fallback file: %v", err)
	}
	if string(data) != html {
		t.Error("fallback HTML content mismatch")
	}
}

func TestDefaultPDFConfig(t *testing.T) {
	cfg := DefaultPDFConfig()
	if cfg.PageSize != "A4" {
		t.Errorf("expected A4, got %s", cfg.PageSize)
	}
	if cfg.Orientation != "portrait" {
		t.Errorf("expected portrait, got %s", cfg.Orientation)
	}
}

// ════════════════════════════════════════════════════════════════════
// Utility Tests
// ════════════════════════════════════════════════════════════════════

func TestReportTimestamp(t *testing.T) {
	ts := ReportTimestamp()
	if ts == "" {
		t.Error("expected non-empty timestamp")
	}
	if !strings.Contains(ts, "IST") {
		t.Errorf("expected IST in timestamp, got %s", ts)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		contains string
	}{
		{5 * time.Second, "s"},
		{3 * time.Minute, "m"},
		{2 * time.Hour, "h"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.input)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("FormatDuration(%v) = %s, expected to contain '%s'", tt.input, result, tt.contains)
		}
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"a & b", "a &amp; b"},
		{"<b>test</b>", "&lt;b&gt;test&lt;/b&gt;"},
		{`"quoted"`, "&quot;quoted&quot;"},
	}

	for _, tt := range tests {
		result := escapeXML(tt.input)
		if result != tt.expected {
			t.Errorf("escapeXML(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDefaultChartConfig(t *testing.T) {
	cfg := DefaultChartConfig()
	if cfg.Width != 800 {
		t.Errorf("expected width 800, got %d", cfg.Width)
	}
	if cfg.Height != 400 {
		t.Errorf("expected height 400, got %d", cfg.Height)
	}
	if cfg.BgColor != "#ffffff" {
		t.Errorf("expected white bg, got %s", cfg.BgColor)
	}
}

func TestPlotArea(t *testing.T) {
	cfg := DefaultChartConfig()
	x, y, w, h := cfg.plotArea()
	if x != cfg.MarginLeft {
		t.Errorf("expected x=%d, got %d", cfg.MarginLeft, x)
	}
	if y != cfg.MarginTop {
		t.Errorf("expected y=%d, got %d", cfg.MarginTop, y)
	}
	expectedW := cfg.Width - cfg.MarginLeft - cfg.MarginRight
	if w != expectedW {
		t.Errorf("expected w=%d, got %d", expectedW, w)
	}
	expectedH := cfg.Height - cfg.MarginTop - cfg.MarginBottom
	if h != expectedH {
		t.Errorf("expected h=%d, got %d", expectedH, h)
	}
}

func TestEmptySVG(t *testing.T) {
	svg := emptySVG(ChartConfig{}, "Test message")
	if !strings.Contains(svg, "Test message") {
		t.Error("expected message in empty SVG")
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("expected valid SVG")
	}
}

// ════════════════════════════════════════════════════════════════════
// Integration: Full Report Pipeline
// ════════════════════════════════════════════════════════════════════

func TestFullReportPipeline_HTML(t *testing.T) {
	analysis := sampleAnalysis()
	cfg := DefaultReportConfig()

	html, err := GenerateHTML(analysis, cfg)
	if err != nil {
		t.Fatalf("GenerateHTML: %v", err)
	}

	// Verify it's valid HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE")
	}
	if !strings.Contains(html, "</html>") {
		t.Error("expected closing HTML tag")
	}

	// Verify all sections are present
	sections := []string{
		"Recommendation", "Price Chart",
		"Fundamental Analysis", "Technical Analysis",
		"Derivatives", "Sentiment", "Risk Assessment",
	}
	for _, s := range sections {
		if !strings.Contains(html, s) {
			t.Errorf("missing section: %s", s)
		}
	}

	// Verify SVG charts embedded
	svgCount := strings.Count(html, "<svg")
	if svgCount < 2 {
		t.Errorf("expected at least 2 SVG charts, found %d", svgCount)
	}
}

func TestFullReportPipeline_Text(t *testing.T) {
	analysis := sampleAnalysis()
	cfg := DefaultReportConfig()

	text, err := GenerateText(analysis, cfg)
	if err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	// Verify structure
	if !strings.Contains(text, "═") {
		t.Error("expected section separators")
	}
	if !strings.Contains(text, "RECOMMENDATION") {
		t.Error("expected recommendation section")
	}
	if !strings.Contains(text, "KEY FINANCIAL RATIOS") {
		t.Error("expected financial ratios section")
	}
}

func TestFullReportPipeline_WriteToDisk(t *testing.T) {
	analysis := sampleAnalysis()
	cfg := DefaultReportConfig()

	html, err := GenerateHTML(analysis, cfg)
	if err != nil {
		t.Fatalf("GenerateHTML: %v", err)
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "report.html")
	if err := os.WriteFile(outPath, []byte(html), 0644); err != nil {
		t.Fatalf("writing report: %v", err)
	}

	// Verify file was written and is non-empty
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat report file: %v", err)
	}
	if info.Size() < 1000 {
		t.Errorf("report file suspiciously small: %d bytes", info.Size())
	}
}
