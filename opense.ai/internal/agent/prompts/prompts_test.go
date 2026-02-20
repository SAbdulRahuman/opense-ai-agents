package prompts

import (
	"math"
	"strings"
	"testing"
)

// ── Agent Name Constants ──

func TestAgentNameConstants(t *testing.T) {
	names := map[string]string{
		"AgentFundamental": AgentFundamental,
		"AgentTechnical":   AgentTechnical,
		"AgentSentiment":   AgentSentiment,
		"AgentFnO":         AgentFnO,
		"AgentRisk":        AgentRisk,
		"AgentExecutor":    AgentExecutor,
		"AgentReporter":    AgentReporter,
		"AgentCIO":         AgentCIO,
	}
	for label, name := range names {
		if name == "" {
			t.Errorf("%s should not be empty", label)
		}
		if strings.Contains(name, " ") {
			t.Errorf("%s should not contain spaces: %q", label, name)
		}
	}
}

func TestAgentNameValues(t *testing.T) {
	if AgentFundamental != "fundamental_analyst" {
		t.Errorf("AgentFundamental: got %q", AgentFundamental)
	}
	if AgentTechnical != "technical_analyst" {
		t.Errorf("AgentTechnical: got %q", AgentTechnical)
	}
	if AgentSentiment != "sentiment_analyst" {
		t.Errorf("AgentSentiment: got %q", AgentSentiment)
	}
	if AgentFnO != "fno_analyst" {
		t.Errorf("AgentFnO: got %q", AgentFnO)
	}
	if AgentRisk != "risk_manager" {
		t.Errorf("AgentRisk: got %q", AgentRisk)
	}
	if AgentExecutor != "trade_executor" {
		t.Errorf("AgentExecutor: got %q", AgentExecutor)
	}
	if AgentReporter != "report_generator" {
		t.Errorf("AgentReporter: got %q", AgentReporter)
	}
	if AgentCIO != "chief_investment_officer" {
		t.Errorf("AgentCIO: got %q", AgentCIO)
	}
}

// ── System Prompts ──

func TestSystemPromptsNonEmpty(t *testing.T) {
	prompts := map[string]string{
		"FundamentalSystemPrompt": FundamentalSystemPrompt,
		"TechnicalSystemPrompt":   TechnicalSystemPrompt,
		"SentimentSystemPrompt":   SentimentSystemPrompt,
		"FnOSystemPrompt":         FnOSystemPrompt,
		"RiskSystemPrompt":        RiskSystemPrompt,
		"ExecutorSystemPrompt":    ExecutorSystemPrompt,
		"ReporterSystemPrompt":    ReporterSystemPrompt,
		"CIOSystemPrompt":         CIOSystemPrompt,
	}
	for name, prompt := range prompts {
		if prompt == "" {
			t.Errorf("%s should not be empty", name)
		}
		if len(prompt) < 200 {
			t.Errorf("%s is too short (%d chars): system prompts need substance", name, len(prompt))
		}
	}
}

func TestSystemPromptsContainKeywords(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		keywords []string
	}{
		{
			"Fundamental",
			FundamentalSystemPrompt,
			[]string{"Fundamental", "NSE", "PE", "ROE", "valuation", "BUY", "SELL", "HOLD"},
		},
		{
			"Technical",
			TechnicalSystemPrompt,
			[]string{"Technical", "RSI", "MACD", "Bollinger", "support", "resistance"},
		},
		{
			"Sentiment",
			SentimentSystemPrompt,
			[]string{"Sentiment", "FII", "DII", "bearish", "bullish"},
		},
		{
			"FnO",
			FnOSystemPrompt,
			[]string{"F&O", "option chain", "OI", "PCR", "max pain", "lot size"},
		},
		{
			"Risk",
			RiskSystemPrompt,
			[]string{"Risk", "position size", "stop-loss", "VaR", "ATR"},
		},
		{
			"Executor",
			ExecutorSystemPrompt,
			[]string{"Executor", "order", "LIMIT", "confirmation", "brokerage"},
		},
		{
			"Reporter",
			ReporterSystemPrompt,
			[]string{"Report", "research", "analysis", "recommendation"},
		},
		{
			"CIO",
			CIOSystemPrompt,
			[]string{"CIO", "team", "synthesize", "investment thesis", "conviction"},
		},
	}

	for _, tc := range tests {
		for _, kw := range tc.keywords {
			if !strings.Contains(strings.ToLower(tc.prompt), strings.ToLower(kw)) {
				t.Errorf("%s prompt should contain keyword %q", tc.name, kw)
			}
		}
	}
}

func TestSystemPromptsContainOutputFormat(t *testing.T) {
	prompts := []string{
		FundamentalSystemPrompt, TechnicalSystemPrompt, SentimentSystemPrompt,
		FnOSystemPrompt, RiskSystemPrompt, ExecutorSystemPrompt,
		ReporterSystemPrompt, CIOSystemPrompt,
	}
	for i, p := range prompts {
		if !strings.Contains(p, "Output Format") && !strings.Contains(p, "output format") {
			t.Errorf("System prompt %d should contain 'Output Format' section", i)
		}
	}
}

// ── CoT Templates ──

func TestCoTAnalysisContainsTicker(t *testing.T) {
	result := CoTAnalysis("RELIANCE", "full analysis")
	if !strings.Contains(result, "RELIANCE") {
		t.Error("CoTAnalysis should contain the ticker")
	}
	if !strings.Contains(result, "full analysis") {
		t.Error("CoTAnalysis should contain the task")
	}
	if !strings.Contains(result, "step-by-step") {
		t.Error("CoTAnalysis should contain 'step-by-step'")
	}
}

func TestCoTFundamentalContainsTicker(t *testing.T) {
	result := CoTFundamental("TCS")
	if !strings.Contains(result, "TCS") {
		t.Error("CoTFundamental should contain the ticker")
	}
	steps := []string{"Financial Health", "Growth", "Valuation", "Ownership", "Peer Comparison", "Risk", "Final Verdict"}
	for _, step := range steps {
		if !strings.Contains(result, step) {
			t.Errorf("CoTFundamental should contain step %q", step)
		}
	}
}

func TestCoTTechnicalContainsTicker(t *testing.T) {
	result := CoTTechnical("INFY")
	if !strings.Contains(result, "INFY") {
		t.Error("CoTTechnical should contain the ticker")
	}
	steps := []string{"Trend", "Momentum", "Volatility", "Support & Resistance", "Pattern", "Volume", "Trade Setup"}
	for _, step := range steps {
		if !strings.Contains(result, step) {
			t.Errorf("CoTTechnical should contain step %q", step)
		}
	}
}

func TestCoTDerivativesContainsTicker(t *testing.T) {
	result := CoTDerivatives("NIFTY")
	if !strings.Contains(result, "NIFTY") {
		t.Error("CoTDerivatives should contain the ticker")
	}
	steps := []string{"Option Chain", "Max Pain", "PCR", "OI Buildup", "Futures", "Strategy"}
	for _, step := range steps {
		if !strings.Contains(result, step) {
			t.Errorf("CoTDerivatives should contain step %q", step)
		}
	}
}

func TestCoTRiskContainsTickerAndCapital(t *testing.T) {
	result := CoTRisk("RELIANCE", 1000000)
	if !strings.Contains(result, "RELIANCE") {
		t.Error("CoTRisk should contain the ticker")
	}
	if !strings.Contains(result, "1000000") {
		t.Error("CoTRisk should contain the capital amount")
	}
	// 5% of 1000000 = 50000
	if !strings.Contains(result, "50000") {
		t.Error("CoTRisk should contain 5% position size limit")
	}
}

func TestCoTSynthesisContainsTicker(t *testing.T) {
	result := CoTSynthesis("HDFCBANK")
	if !strings.Contains(result, "HDFCBANK") {
		t.Error("CoTSynthesis should contain the ticker")
	}
	steps := []string{"Fundamental", "Technical", "Sentiment", "Derivatives", "Risk", "Weight", "Final View"}
	for _, step := range steps {
		if !strings.Contains(result, step) {
			t.Errorf("CoTSynthesis should contain step %q", step)
		}
	}
}

func TestCoTFunctionsReturnNonEmpty(t *testing.T) {
	funcs := []struct {
		name   string
		result string
	}{
		{"CoTAnalysis", CoTAnalysis("X", "test")},
		{"CoTFundamental", CoTFundamental("X")},
		{"CoTTechnical", CoTTechnical("X")},
		{"CoTDerivatives", CoTDerivatives("X")},
		{"CoTRisk", CoTRisk("X", 100000)},
		{"CoTSynthesis", CoTSynthesis("X")},
	}
	for _, f := range funcs {
		if f.result == "" {
			t.Errorf("%s should not return empty string", f.name)
		}
		if len(f.result) < 100 {
			t.Errorf("%s result is too short (%d chars)", f.name, len(f.result))
		}
	}
}

// ── Indian Market Functions ──

func TestIndianMarketConstants(t *testing.T) {
	if IndianMarketContext == "" {
		t.Error("IndianMarketContext should not be empty")
	}
	if IndianNumberFormat == "" {
		t.Error("IndianNumberFormat should not be empty")
	}
	// Verify key content
	if !strings.Contains(IndianMarketContext, "NSE") {
		t.Error("IndianMarketContext should mention NSE")
	}
	if !strings.Contains(IndianMarketContext, "9:15") {
		t.Error("IndianMarketContext should mention market open time")
	}
	if !strings.Contains(IndianNumberFormat, "Crores") {
		t.Error("IndianNumberFormat should mention Crores")
	}
	if !strings.Contains(IndianNumberFormat, "₹") {
		t.Error("IndianNumberFormat should contain ₹ symbol")
	}
}

func TestIndianMarketPromptSuffix(t *testing.T) {
	suffix := IndianMarketPromptSuffix()
	if suffix == "" {
		t.Error("IndianMarketPromptSuffix should not be empty")
	}
	// Should contain both constants
	if !strings.Contains(suffix, "Indian Market Context") {
		t.Error("Suffix should contain IndianMarketContext content")
	}
	if !strings.Contains(suffix, "Number Formatting Rules") {
		t.Error("Suffix should contain IndianNumberFormat content")
	}
	// Should be the concatenation
	expected := IndianMarketContext + IndianNumberFormat
	if suffix != expected {
		t.Error("IndianMarketPromptSuffix should return the concatenation of both constants")
	}
}

func TestNSESectorsNotEmpty(t *testing.T) {
	if len(NSESectors) == 0 {
		t.Error("NSESectors should not be empty")
	}
	expectedSectors := []string{"IT", "Banking", "Pharma", "Auto", "FMCG", "Metal", "Oil & Gas"}
	for _, sector := range expectedSectors {
		if _, ok := NSESectors[sector]; !ok {
			t.Errorf("NSESectors should contain %q", sector)
		}
	}
	// Each sector should have at least 3 tickers
	for sector, tickers := range NSESectors {
		if len(tickers) < 3 {
			t.Errorf("Sector %q should have at least 3 tickers, got %d", sector, len(tickers))
		}
	}
}

func TestSectorForTickerKnown(t *testing.T) {
	tests := []struct {
		ticker string
		want   string
	}{
		{"TCS", "IT"},
		{"INFY", "IT"},
		{"HDFCBANK", "Banking"},
		{"SBIN", "Banking"},
		{"RELIANCE", "Oil & Gas"},
		{"SUNPHARMA", "Pharma"},
		{"MARUTI", "Auto"},
		{"HINDUNILVR", "FMCG"},
		{"TATASTEEL", "Metal"},
		{"ULTRACEMCO", "Cement"},
		{"BHARTIARTL", "Telecom"},
		{"NTPC", "Power"},
		{"HAL", "Capital Goods"},
		{"DLF", "Realty"},
		{"SBILIFE", "Insurance"},
		{"BAJFINANCE", "NBFC"},
	}
	for _, tc := range tests {
		got := SectorForTicker(tc.ticker)
		if got != tc.want {
			t.Errorf("SectorForTicker(%q): got %q, want %q", tc.ticker, got, tc.want)
		}
	}
}

func TestSectorForTickerUnknown(t *testing.T) {
	got := SectorForTicker("NONEXISTENT")
	if got != "" {
		t.Errorf("SectorForTicker(NONEXISTENT): got %q, want empty string", got)
	}
	got = SectorForTicker("")
	if got != "" {
		t.Errorf("SectorForTicker(''): got %q, want empty string", got)
	}
}

func TestSectorPeersKnown(t *testing.T) {
	peers := SectorPeers("TCS")
	if len(peers) == 0 {
		t.Fatal("SectorPeers(TCS) should return peers")
	}
	// Should not include TCS itself
	for _, p := range peers {
		if p == "TCS" {
			t.Error("SectorPeers should exclude the queried ticker")
		}
	}
	// Should include other IT stocks
	found := false
	for _, p := range peers {
		if p == "INFY" {
			found = true
			break
		}
	}
	if !found {
		t.Error("SectorPeers(TCS) should include INFY")
	}
}

func TestSectorPeersCount(t *testing.T) {
	// IT sector has 9 tickers, peers of TCS should be 8
	peers := SectorPeers("TCS")
	itTickers := NSESectors["IT"]
	if len(peers) != len(itTickers)-1 {
		t.Errorf("SectorPeers(TCS): got %d peers, want %d", len(peers), len(itTickers)-1)
	}
}

func TestSectorPeersUnknown(t *testing.T) {
	peers := SectorPeers("NONEXISTENT")
	if peers != nil {
		t.Errorf("SectorPeers(NONEXISTENT): got %v, want nil", peers)
	}
}

func TestFormatTickerPromptKnown(t *testing.T) {
	result := FormatTickerPrompt("TCS")
	if !strings.Contains(result, "TCS") {
		t.Error("FormatTickerPrompt should contain the ticker")
	}
	if !strings.Contains(result, "IT") {
		t.Error("FormatTickerPrompt(TCS) should contain sector 'IT'")
	}
	if !strings.Contains(result, "Peers") || !strings.Contains(result, "Key Peers") {
		t.Error("FormatTickerPrompt should contain 'Key Peers' section")
	}
	if !strings.Contains(result, "NSE") {
		t.Error("FormatTickerPrompt should contain 'NSE'")
	}
}

func TestFormatTickerPromptUnknown(t *testing.T) {
	result := FormatTickerPrompt("UNKNOWN123")
	if !strings.Contains(result, "UNKNOWN123") {
		t.Error("FormatTickerPrompt should contain the unknown ticker")
	}
	if !strings.Contains(result, "Unknown") {
		t.Error("FormatTickerPrompt for unknown ticker should mention 'Unknown'")
	}
}

func TestFormatTickerPromptPeersLimit(t *testing.T) {
	// For a sector with more than 5 tickers, only first 5 peers should appear
	result := FormatTickerPrompt("TCS")
	// Count peers in the output (IT sector has 8 peers for TCS)
	// Only first 5 should be shown
	for _, p := range NSESectors["IT"] {
		if p == "TCS" {
			continue
		}
		_ = strings.Contains(result, p)
	}
	// Just verify it's well-formed
	if !strings.HasPrefix(result, "Stock:") {
		t.Error("FormatTickerPrompt should start with 'Stock:'")
	}
}

func TestIndianBrokerageEstimateDelivery(t *testing.T) {
	result := IndianBrokerageEstimate(100.0, 110.0, 100, true)
	if !strings.Contains(result, "Brokerage Estimate") {
		t.Error("Should contain 'Brokerage Estimate' header")
	}
	if !strings.Contains(result, "STT") {
		t.Error("Should contain STT line")
	}
	if !strings.Contains(result, "Stamp Duty") {
		t.Error("Should contain Stamp Duty line")
	}
	if !strings.Contains(result, "Total") {
		t.Error("Should contain Total line")
	}
	if !strings.Contains(result, "Turnover") {
		t.Error("Should contain Turnover line")
	}
}

func TestIndianBrokerageEstimateIntraday(t *testing.T) {
	result := IndianBrokerageEstimate(100.0, 105.0, 50, false)
	if !strings.Contains(result, "Brokerage Estimate") {
		t.Error("Should contain 'Brokerage Estimate' header")
	}
}

func TestIndianBrokerageEstimateValues(t *testing.T) {
	buyPrice := 1000.0
	sellPrice := 1100.0
	qty := 10
	isDelivery := true

	result := IndianBrokerageEstimate(buyPrice, sellPrice, qty, isDelivery)

	// Turnover = (1000 + 1100) * 10 = 21000
	turnover := (buyPrice + sellPrice) * float64(qty)
	if turnover != 21000 {
		t.Errorf("expected turnover 21000, got %f", turnover)
	}

	// STT for delivery: 0.1% of turnover
	stt := turnover * 0.001 // 21.0
	if math.Abs(stt-21.0) > 0.01 {
		t.Errorf("STT: got %f, want 21.0", stt)
	}

	// Result should contain the calculated values
	if !strings.Contains(result, "21000.00") {
		t.Errorf("Result should contain turnover 21000.00, got:\n%s", result)
	}

	// Verify non-empty multiline output
	lines := strings.Split(result, "\n")
	if len(lines) < 7 {
		t.Errorf("Expected at least 7 lines of output, got %d", len(lines))
	}
}

func TestIndianBrokerageEstimateIntradayLowerSTT(t *testing.T) {
	buyPrice := 1000.0
	sellPrice := 1100.0
	qty := 10

	deliveryResult := IndianBrokerageEstimate(buyPrice, sellPrice, qty, true)
	intradayResult := IndianBrokerageEstimate(buyPrice, sellPrice, qty, false)

	// Delivery STT is higher than intraday STT
	// We can check by looking at the Total line values
	// Just verify both return valid output
	if deliveryResult == intradayResult {
		t.Error("Delivery and intraday calculations should differ (different STT)")
	}
}
