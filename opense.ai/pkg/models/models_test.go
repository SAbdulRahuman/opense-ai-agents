package models

import (
	"encoding/json"
	"testing"
	"time"
)

// ── Stock Tests ──

func TestStockJSONRoundtrip(t *testing.T) {
	s := Stock{
		Ticker:      "RELIANCE",
		NSETicker:   "RELIANCE.NS",
		Name:        "Reliance Industries Limited",
		Exchange:    "NSE",
		Sector:      "Oil & Gas",
		Industry:    "Refineries",
		ISIN:        "INE002A01018",
		MarketCap:   1927345_00_00_000.0, // ₹19,27,345 Cr approx in raw
		FaceValue:   10.0,
		ListingDate: "1995-11-29",
		LotSize:     250,
		TickSize:    0.05,
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal(Stock) error: %v", err)
	}
	var decoded Stock
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(Stock) error: %v", err)
	}
	if decoded.Ticker != s.Ticker {
		t.Errorf("Ticker: got %q, want %q", decoded.Ticker, s.Ticker)
	}
	if decoded.NSETicker != s.NSETicker {
		t.Errorf("NSETicker: got %q, want %q", decoded.NSETicker, s.NSETicker)
	}
	if decoded.MarketCap != s.MarketCap {
		t.Errorf("MarketCap: got %f, want %f", decoded.MarketCap, s.MarketCap)
	}
	if decoded.LotSize != s.LotSize {
		t.Errorf("LotSize: got %d, want %d", decoded.LotSize, s.LotSize)
	}
}

func TestOHLCVTimestamp(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	bar := OHLCV{
		Timestamp: now,
		Open:      2800.0,
		High:      2850.0,
		Low:       2790.0,
		Close:     2847.5,
		Volume:    5_000_000,
	}
	if bar.High < bar.Low {
		t.Error("High should be >= Low")
	}
	if bar.Close < bar.Low || bar.Close > bar.High {
		t.Error("Close should be between Low and High")
	}
	data, err := json.Marshal(bar)
	if err != nil {
		t.Fatalf("json.Marshal(OHLCV) error: %v", err)
	}
	var decoded OHLCV
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(OHLCV) error: %v", err)
	}
	if !decoded.Timestamp.Equal(now) {
		t.Errorf("Timestamp: got %v, want %v", decoded.Timestamp, now)
	}
}

func TestQuoteFields(t *testing.T) {
	q := Quote{
		Ticker:    "TCS",
		Name:      "Tata Consultancy Services Limited",
		LastPrice: 4200.0,
		Change:    45.0,
		ChangePct: 1.08,
		Open:      4180.0,
		High:      4220.0,
		Low:       4170.0,
		PrevClose: 4155.0,
		Volume:    2_000_000,
		Timestamp: time.Now(),
	}
	if q.Change != q.LastPrice-q.PrevClose {
		t.Errorf("Change: got %.2f, want %.2f", q.Change, q.LastPrice-q.PrevClose)
	}
}

func TestTimeframeConstants(t *testing.T) {
	timeframes := map[Timeframe]string{
		Timeframe1Min:  "1m",
		Timeframe5Min:  "5m",
		Timeframe15Min: "15m",
		Timeframe1Hour: "1h",
		Timeframe1Day:  "1d",
		Timeframe1Week: "1w",
		Timeframe1Mon:  "1M",
	}
	for tf, expected := range timeframes {
		if string(tf) != expected {
			t.Errorf("Timeframe %v: got %q, want %q", tf, string(tf), expected)
		}
	}
}

// ── Order Tests ──

func TestOrderSideConstants(t *testing.T) {
	if string(Buy) != "BUY" {
		t.Errorf("Buy: got %q, want %q", Buy, "BUY")
	}
	if string(Sell) != "SELL" {
		t.Errorf("Sell: got %q, want %q", Sell, "SELL")
	}
}

func TestOrderTypeConstants(t *testing.T) {
	types := map[OrderType]string{
		Market: "MARKET",
		Limit:  "LIMIT",
		SL:     "SL",
		SLM:    "SL-M",
	}
	for ot, expected := range types {
		if string(ot) != expected {
			t.Errorf("OrderType %v: got %q, want %q", ot, string(ot), expected)
		}
	}
}

func TestOrderProductConstants(t *testing.T) {
	products := map[OrderProduct]string{
		CNC:  "CNC",
		MIS:  "MIS",
		NRML: "NRML",
	}
	for p, expected := range products {
		if string(p) != expected {
			t.Errorf("OrderProduct %v: got %q, want %q", p, string(p), expected)
		}
	}
}

func TestOrderStatusConstants(t *testing.T) {
	statuses := map[OrderStatus]string{
		OrderPending:   "PENDING",
		OrderOpen:      "OPEN",
		OrderComplete:  "COMPLETE",
		OrderCancelled: "CANCELLED",
		OrderRejected:  "REJECTED",
	}
	for s, expected := range statuses {
		if string(s) != expected {
			t.Errorf("OrderStatus %v: got %q, want %q", s, string(s), expected)
		}
	}
}

func TestOrderRequestJSON(t *testing.T) {
	req := OrderRequest{
		Ticker:       "RELIANCE",
		Exchange:     "NSE",
		Side:         Buy,
		OrderType:    Limit,
		Product:      CNC,
		Quantity:     10,
		Price:        2845.50,
		TriggerPrice: 0,
		StopLoss:     2800.0,
		Target:       2950.0,
		Tag:          "test-order",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal(OrderRequest) error: %v", err)
	}
	var decoded OrderRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(OrderRequest) error: %v", err)
	}
	if decoded.Ticker != req.Ticker {
		t.Errorf("Ticker: got %q, want %q", decoded.Ticker, req.Ticker)
	}
	if decoded.Side != Buy {
		t.Errorf("Side: got %q, want %q", decoded.Side, Buy)
	}
	if decoded.Price != req.Price {
		t.Errorf("Price: got %f, want %f", decoded.Price, req.Price)
	}
}

func TestPositionPnLSign(t *testing.T) {
	// Profitable long position
	pos := Position{
		Ticker:   "INFY",
		Quantity: 100,
		AvgPrice: 1500.0,
		LTP:      1550.0,
		PnL:      5000.0,
		PnLPct:   3.33,
	}
	if pos.PnL <= 0 {
		t.Error("Profitable position should have positive PnL")
	}

	// Losing short position
	negPos := Position{
		Ticker:   "INFY",
		Quantity: -100,
		AvgPrice: 1500.0,
		LTP:      1550.0,
		PnL:      -5000.0,
		PnLPct:   -3.33,
	}
	if negPos.PnL >= 0 {
		t.Error("Losing position should have negative PnL")
	}
}

func TestMarginsFields(t *testing.T) {
	m := Margins{
		AvailableCash:   500000.0,
		UsedMargin:      200000.0,
		AvailableMargin: 300000.0,
		Collateral:      100000.0,
		OpeningBalance:  500000.0,
	}
	if m.AvailableMargin != m.AvailableCash-m.UsedMargin {
		t.Errorf("AvailableMargin should = AvailableCash - UsedMargin: got %f, want %f",
			m.AvailableMargin, m.AvailableCash-m.UsedMargin)
	}
}

// ── Financials Tests ──

func TestIncomeStatementJSON(t *testing.T) {
	is := IncomeStatement{
		Period:     "Mar 2025",
		PeriodType: "annual",
		Revenue:    500000.0,
		PAT:        75000.0,
		EPS:        120.5,
		OPMPct:     22.5,
		NPMPct:     15.0,
	}
	data, err := json.Marshal(is)
	if err != nil {
		t.Fatalf("json.Marshal(IncomeStatement) error: %v", err)
	}
	var decoded IncomeStatement
	json.Unmarshal(data, &decoded)
	if decoded.Period != is.Period {
		t.Errorf("Period: got %q, want %q", decoded.Period, is.Period)
	}
	if decoded.NPMPct != is.NPMPct {
		t.Errorf("NPMPct: got %f, want %f", decoded.NPMPct, is.NPMPct)
	}
}

func TestFinancialDataAggregation(t *testing.T) {
	fd := FinancialData{
		Ticker: "RELIANCE",
		AnnualIncome: []IncomeStatement{
			{Period: "Mar 2025", PeriodType: "annual", Revenue: 500000},
			{Period: "Mar 2024", PeriodType: "annual", Revenue: 450000},
		},
		QuarterlyIncome: []IncomeStatement{
			{Period: "Q3 FY26", PeriodType: "quarterly", Revenue: 130000},
		},
	}
	if len(fd.AnnualIncome) != 2 {
		t.Errorf("AnnualIncome: got %d items, want 2", len(fd.AnnualIncome))
	}
	if fd.AnnualIncome[0].Revenue <= fd.AnnualIncome[1].Revenue {
		// Expected: latest year has higher revenue (growth)
		t.Log("Revenue declining — flagging for analysis")
	}
}

func TestGrowthRatesFields(t *testing.T) {
	gr := GrowthRates{
		RevenueGrowthYoY: 15.5,
		ProfitGrowthYoY:  22.3,
		RevenueCAGR3Y:    18.0,
		ProfitCAGR3Y:     25.0,
	}
	if gr.ProfitGrowthYoY < gr.RevenueGrowthYoY {
		t.Log("Profit growing slower than revenue — margin compression")
	}
	data, _ := json.Marshal(gr)
	var decoded GrowthRates
	json.Unmarshal(data, &decoded)
	if decoded.RevenueCAGR3Y != gr.RevenueCAGR3Y {
		t.Errorf("RevenueCAGR3Y: got %f, want %f", decoded.RevenueCAGR3Y, gr.RevenueCAGR3Y)
	}
}

// ── Option Tests ──

func TestOptionChainPCR(t *testing.T) {
	oc := OptionChain{
		Ticker:    "NIFTY",
		SpotPrice: 22500.0,
		TotalCEOI: 10_000_000,
		TotalPEOI: 12_000_000,
		PCR:       1.2,
		MaxPain:   22400.0,
	}
	expectedPCR := float64(oc.TotalPEOI) / float64(oc.TotalCEOI)
	if oc.PCR != expectedPCR {
		t.Errorf("PCR: got %f, want %f", oc.PCR, expectedPCR)
	}
}

func TestOIBuildupTypeConstants(t *testing.T) {
	types := map[OIBuildupType]string{
		LongBuildup:   "long_buildup",
		ShortBuildup:  "short_buildup",
		LongUnwinding: "long_unwinding",
		ShortCovering: "short_covering",
	}
	for bt, expected := range types {
		if string(bt) != expected {
			t.Errorf("OIBuildupType %v: got %q, want %q", bt, string(bt), expected)
		}
	}
}

func TestOptionStrategyJSON(t *testing.T) {
	strategy := OptionStrategy{
		Name: "Bull Call Spread",
		Legs: []OptionLeg{
			{OptionType: "CE", StrikePrice: 22500, Action: "BUY", Lots: 1, Premium: 250},
			{OptionType: "CE", StrikePrice: 22700, Action: "SELL", Lots: 1, Premium: 120},
		},
		MaxProfit:  200 * 25, // (strike diff - net premium) * lot size
		MaxLoss:    130 * 25, // net premium * lot size
		Breakevens: []float64{22630},
		NetPremium: -130,
	}
	data, err := json.Marshal(strategy)
	if err != nil {
		t.Fatalf("json.Marshal(OptionStrategy) error: %v", err)
	}
	var decoded OptionStrategy
	json.Unmarshal(data, &decoded)
	if len(decoded.Legs) != 2 {
		t.Errorf("Legs: got %d, want 2", len(decoded.Legs))
	}
	if decoded.Name != "Bull Call Spread" {
		t.Errorf("Name: got %q, want %q", decoded.Name, "Bull Call Spread")
	}
}

func TestFIIDIINetCalculation(t *testing.T) {
	data := FIIDIIData{
		Date:    "20-Feb-2026",
		FIIBuy:  5000.0,
		FIISell: 3000.0,
		FIINet:  2000.0,
		DIIBuy:  4000.0,
		DIISell: 4500.0,
		DIINet:  -500.0,
	}
	expectedFIINet := data.FIIBuy - data.FIISell
	if data.FIINet != expectedFIINet {
		t.Errorf("FIINet: got %f, want %f", data.FIINet, expectedFIINet)
	}
	expectedDIINet := data.DIIBuy - data.DIISell
	if data.DIINet != expectedDIINet {
		t.Errorf("DIINet: got %f, want %f", data.DIINet, expectedDIINet)
	}
}

// ── Analysis Tests ──

func TestSignalTypeConstants(t *testing.T) {
	if string(SignalBuy) != "BUY" {
		t.Errorf("SignalBuy: got %q", SignalBuy)
	}
	if string(SignalSell) != "SELL" {
		t.Errorf("SignalSell: got %q", SignalSell)
	}
	if string(SignalNeutral) != "NEUTRAL" {
		t.Errorf("SignalNeutral: got %q", SignalNeutral)
	}
}

func TestAnalysisTypeConstants(t *testing.T) {
	types := []AnalysisType{
		AnalysisTechnical, AnalysisFundamental, AnalysisDerivatives,
		AnalysisSentiment, AnalysisRisk, AnalysisComposite,
	}
	for _, at := range types {
		if string(at) == "" {
			t.Errorf("AnalysisType should not be empty: %v", at)
		}
	}
}

func TestRecommendationConstants(t *testing.T) {
	recs := map[Recommendation]string{
		StrongBuy:    "STRONG_BUY",
		ModerateBuy:  "BUY",
		Hold:         "HOLD",
		ModerateSell: "SELL",
		StrongSell:   "STRONG_SELL",
	}
	for r, expected := range recs {
		if string(r) != expected {
			t.Errorf("Recommendation %v: got %q, want %q", r, string(r), expected)
		}
	}
}

func TestAnalysisResultJSON(t *testing.T) {
	result := AnalysisResult{
		Ticker:         "RELIANCE",
		Type:           AnalysisTechnical,
		AgentName:      "technical_analyst",
		Recommendation: ModerateBuy,
		Confidence:     0.78,
		Summary:        "RSI at 45 with bullish MACD crossover",
		Signals: []Signal{
			{Source: "RSI", Type: SignalBuy, Confidence: 0.7, Reason: "RSI at 45, rising"},
			{Source: "MACD", Type: SignalBuy, Confidence: 0.85, Reason: "Bullish crossover"},
		},
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal(AnalysisResult) error: %v", err)
	}
	var decoded AnalysisResult
	json.Unmarshal(data, &decoded)
	if len(decoded.Signals) != 2 {
		t.Errorf("Signals: got %d, want 2", len(decoded.Signals))
	}
	if decoded.Confidence < 0 || decoded.Confidence > 1 {
		t.Errorf("Confidence should be 0-1: got %f", decoded.Confidence)
	}
}

func TestCompositeAnalysisJSON(t *testing.T) {
	comp := CompositeAnalysis{
		Ticker:          "TCS",
		Recommendation:  StrongBuy,
		Confidence:      0.85,
		Summary:         "Strong fundamentals with bullish technical setup",
		EntryPrice:      4200.0,
		TargetPrice:     4600.0,
		StopLoss:        4050.0,
		RiskRewardRatio: 2.67,
		Timeframe:       "medium-term",
		Timestamp:       time.Now(),
	}
	data, err := json.Marshal(comp)
	if err != nil {
		t.Fatalf("json.Marshal(CompositeAnalysis) error: %v", err)
	}
	var decoded CompositeAnalysis
	json.Unmarshal(data, &decoded)
	if decoded.RiskRewardRatio < 1 {
		t.Error("RiskRewardRatio should be >= 1 for a recommended trade")
	}
}

func TestTechnicalIndicatorsJSON(t *testing.T) {
	ti := TechnicalIndicators{
		Ticker: "INFY",
		RSI:    62.4,
		MACD: MACDData{
			MACDLine:   15.3,
			SignalLine:  12.1,
			Histogram:  3.2,
		},
		SMA: map[int]float64{20: 1850.0, 50: 1830.0, 200: 1780.0},
		EMA: map[int]float64{12: 1855.0, 26: 1840.0},
		Bollinger: BollingerData{Upper: 1900.0, Middle: 1850.0, Lower: 1800.0},
		SuperTrend: SuperTrendData{Value: 1810.0, Trend: "UP"},
		ATR:       35.5,
		VWAP:      1848.0,
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(ti)
	if err != nil {
		t.Fatalf("json.Marshal(TechnicalIndicators) error: %v", err)
	}
	var decoded TechnicalIndicators
	json.Unmarshal(data, &decoded)
	if decoded.RSI != 62.4 {
		t.Errorf("RSI: got %f, want 62.4", decoded.RSI)
	}
	if decoded.MACD.Histogram != 3.2 {
		t.Errorf("MACD Histogram: got %f, want 3.2", decoded.MACD.Histogram)
	}
	if decoded.SMA[200] != 1780.0 {
		t.Errorf("SMA[200]: got %f, want 1780.0", decoded.SMA[200])
	}
}

func TestBacktestResultJSON(t *testing.T) {
	br := BacktestResult{
		StrategyName:   "SMA Crossover",
		Ticker:         "RELIANCE",
		From:           time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		To:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		InitialCapital: 1000000,
		FinalCapital:   1250000,
		TotalReturn:    250000,
		TotalReturnPct: 25.0,
		WinRate:        0.55,
		TotalTrades:    20,
		WinningTrades:  11,
		LosingTrades:   9,
	}
	if br.WinningTrades+br.LosingTrades != br.TotalTrades {
		t.Errorf("WinningTrades+LosingTrades should = TotalTrades: %d+%d != %d",
			br.WinningTrades, br.LosingTrades, br.TotalTrades)
	}
	data, err := json.Marshal(br)
	if err != nil {
		t.Fatalf("json.Marshal(BacktestResult) error: %v", err)
	}
	var decoded BacktestResult
	json.Unmarshal(data, &decoded)
	if decoded.StrategyName != "SMA Crossover" {
		t.Errorf("StrategyName: got %q", decoded.StrategyName)
	}
}

func TestSentimentScoreRange(t *testing.T) {
	scores := []float64{-1.0, -0.5, 0.0, 0.5, 1.0}
	for _, s := range scores {
		ss := SentimentScore{Score: s, Source: "test"}
		if ss.Score < -1.0 || ss.Score > 1.0 {
			t.Errorf("Score %f out of range [-1, 1]", ss.Score)
		}
	}
}

func TestSupportResistanceJSON(t *testing.T) {
	sr := SupportResistance{
		Ticker:      "RELIANCE",
		Supports:    []float64{2800.0, 2750.0, 2700.0},
		Resistances: []float64{2900.0, 2950.0, 3000.0},
		PivotPoint:  2850.0,
		S1:          2800.0,
		S2:          2750.0,
		S3:          2700.0,
		R1:          2900.0,
		R2:          2950.0,
		R3:          3000.0,
		Method:      "classic",
	}
	data, _ := json.Marshal(sr)
	var decoded SupportResistance
	json.Unmarshal(data, &decoded)
	if len(decoded.Supports) != 3 {
		t.Errorf("Supports: got %d, want 3", len(decoded.Supports))
	}
	if decoded.R1 <= decoded.PivotPoint {
		t.Error("R1 should be above PivotPoint")
	}
	if decoded.S1 >= decoded.PivotPoint {
		t.Error("S1 should be below PivotPoint")
	}
}
