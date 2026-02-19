package technical

import (
	"testing"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// makeCandles generates synthetic OHLCV data for testing.
func makeCandles(n int, basePrice float64, trend float64) []models.OHLCV {
	candles := make([]models.OHLCV, n)
	price := basePrice
	for i := 0; i < n; i++ {
		open := price
		close := open + trend
		high := open + 5
		low := open - 5
		if close > open {
			high = close + 3
		} else {
			low = close - 3
		}
		candles[i] = models.OHLCV{
			Timestamp: time.Now().Add(time.Duration(-n+i) * 24 * time.Hour),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    1000000 + int64(i*10000),
		}
		price = close
	}
	return candles
}

func TestRSI(t *testing.T) {
	candles := makeCandles(50, 100, 1.5)
	vals := RSI(candles, 14)
	if vals == nil {
		t.Fatal("RSI returned nil for sufficient data")
	}
	if len(vals) != 50 {
		t.Fatalf("expected 50 RSI values, got %d", len(vals))
	}
	// In a strong uptrend RSI should be high.
	latest := vals[len(vals)-1]
	if latest < 50 {
		t.Errorf("expected RSI > 50 in uptrend, got %.2f", latest)
	}
}

func TestRSIInsufficientData(t *testing.T) {
	candles := makeCandles(5, 100, 1)
	vals := RSI(candles, 14)
	if vals != nil {
		t.Error("RSI should return nil for insufficient data")
	}
}

func TestRSILatest(t *testing.T) {
	candles := makeCandles(50, 100, 1)
	val := RSILatest(candles, 14)
	if val <= 0 {
		t.Errorf("RSILatest should return positive value, got %.2f", val)
	}
}

func TestMACD(t *testing.T) {
	candles := makeCandles(50, 100, 0.5)
	results := MACD(candles, 12, 26, 9)
	if results == nil {
		t.Fatal("MACD returned nil")
	}
	if len(results) != 50 {
		t.Fatalf("expected 50 MACD results, got %d", len(results))
	}
}

func TestMACDLatest(t *testing.T) {
	candles := makeCandles(50, 100, 1)
	macd := MACDLatest(candles, 12, 26, 9)
	// In uptrend MACD line should be positive.
	if macd.MACDLine < 0 {
		t.Errorf("expected positive MACD line in uptrend, got %.4f", macd.MACDLine)
	}
}

func TestBollingerBands(t *testing.T) {
	candles := makeCandles(50, 100, 0.3)
	bands := BollingerBands(candles, 20, 2)
	if bands == nil {
		t.Fatal("BollingerBands returned nil")
	}
	latest := bands[len(bands)-1]
	if latest.Upper <= latest.Middle || latest.Middle <= latest.Lower {
		t.Errorf("invalid Bollinger bands: upper=%.2f, middle=%.2f, lower=%.2f",
			latest.Upper, latest.Middle, latest.Lower)
	}
}

func TestATR(t *testing.T) {
	candles := makeCandles(30, 100, 1)
	vals := ATR(candles, 14)
	if vals == nil {
		t.Fatal("ATR returned nil")
	}
	latest := ATRLatest(candles, 14)
	if latest <= 0 {
		t.Errorf("expected positive ATR, got %.2f", latest)
	}
}

func TestSuperTrend(t *testing.T) {
	candles := makeCandles(50, 100, 1)
	results := SuperTrend(candles, 7, 3)
	if results == nil {
		t.Fatal("SuperTrend returned nil")
	}
	latest := SuperTrendLatest(candles, 7, 3)
	if latest.Value <= 0 {
		t.Errorf("expected positive SuperTrend value, got %.2f", latest.Value)
	}
	if latest.Trend != "UP" && latest.Trend != "DOWN" {
		t.Errorf("expected UP or DOWN trend, got %q", latest.Trend)
	}
}

func TestComputeAll(t *testing.T) {
	candles := makeCandles(250, 100, 0.2)
	ti := ComputeAll("RELIANCE", candles)
	if ti == nil {
		t.Fatal("ComputeAll returned nil")
	}
	if ti.Ticker != "RELIANCE" {
		t.Errorf("expected ticker RELIANCE, got %s", ti.Ticker)
	}
	if ti.RSI <= 0 {
		t.Error("expected positive RSI")
	}
	if len(ti.SMA) == 0 {
		t.Error("expected SMA map to be populated")
	}
	if len(ti.EMA) == 0 {
		t.Error("expected EMA map to be populated")
	}
}

// --- Moving Average tests ---

func TestSMA(t *testing.T) {
	data := []float64{10, 20, 30, 40, 50}
	vals := SMA(data, 3)
	if vals == nil {
		t.Fatal("SMA returned nil")
	}
	// SMA(3) at index 2 = (10+20+30)/3 = 20
	if vals[2] != 20 {
		t.Errorf("expected SMA[2]=20, got %.2f", vals[2])
	}
	// SMA(3) at index 4 = (30+40+50)/3 = 40
	if vals[4] != 40 {
		t.Errorf("expected SMA[4]=40, got %.2f", vals[4])
	}
}

func TestEMA(t *testing.T) {
	data := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	vals := EMA(data, 5)
	if vals == nil {
		t.Fatal("EMA returned nil")
	}
	if vals[4] == 0 {
		t.Error("EMA seed value should not be zero")
	}
}

func TestWMA(t *testing.T) {
	data := []float64{10, 20, 30}
	vals := WMA(data, 3)
	if vals == nil {
		t.Fatal("WMA returned nil")
	}
	// WMA(3) at index 2 = (10*1 + 20*2 + 30*3) / 6 = 140/6 ≈ 23.33
	expected := 140.0 / 6.0
	if diff := vals[2] - expected; diff > 0.01 || diff < -0.01 {
		t.Errorf("expected WMA[2]≈%.4f, got %.4f", expected, vals[2])
	}
}

func TestVWAP(t *testing.T) {
	candles := makeCandles(10, 100, 1)
	vals := VWAP(candles)
	if vals == nil {
		t.Fatal("VWAP returned nil")
	}
	if vals[len(vals)-1] <= 0 {
		t.Error("expected positive VWAP")
	}
}

// --- Pattern tests ---

func TestDetectDoji(t *testing.T) {
	candles := []models.OHLCV{
		{Open: 100, High: 105, Low: 95, Close: 100.1, Volume: 1000},
	}
	patterns := DetectPatterns(candles)
	found := false
	for _, p := range patterns {
		if p.Name == "Doji" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Doji pattern to be detected")
	}
}

func TestDetectBullishEngulfing(t *testing.T) {
	candles := []models.OHLCV{
		{Open: 110, High: 112, Low: 98, Close: 100, Volume: 1000}, // bearish
		{Open: 98, High: 115, Low: 97, Close: 112, Volume: 2000},  // bullish engulfs
	}
	patterns := DetectPatterns(candles)
	found := false
	for _, p := range patterns {
		if p.Name == "Bullish Engulfing" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Bullish Engulfing pattern")
	}
}

func TestDetectLatestPatterns(t *testing.T) {
	candles := makeCandles(20, 100, 0.5)
	// Should not panic and may or may not find patterns.
	patterns := DetectLatestPatterns(candles, 5)
	_ = patterns // just ensure no panic
}

func TestHeadAndShoulders(t *testing.T) {
	// Construct a basic H&S pattern.
	candles := makeCandles(30, 100, 0)
	// Left shoulder.
	candles[5].High = 115
	// Head.
	candles[15].High = 125
	// Right shoulder.
	candles[25].High = 115

	neckline, conf := HeadAndShoulders(candles, 30)
	// Pattern may or may not be detected depending on exact geometry;
	// main test is that it doesn't panic.
	_ = neckline
	_ = conf
}

// --- Support/Resistance tests ---

func TestPivotPointsClassic(t *testing.T) {
	candles := []models.OHLCV{
		{High: 110, Low: 90, Close: 100},
	}
	sr := PivotPoints(candles, PivotClassic)
	if sr.PivotPoint != 100 {
		t.Errorf("expected pivot=100, got %.2f", sr.PivotPoint)
	}
	if sr.R1 <= sr.PivotPoint || sr.S1 >= sr.PivotPoint {
		t.Error("R1 should be above pivot, S1 below")
	}
}

func TestPivotPointsFibonacci(t *testing.T) {
	candles := []models.OHLCV{
		{High: 110, Low: 90, Close: 100},
	}
	sr := PivotPoints(candles, PivotFibonacci)
	if sr.Method != "fibonacci" {
		t.Errorf("expected method fibonacci, got %s", sr.Method)
	}
	if len(sr.Supports) != 3 || len(sr.Resistances) != 3 {
		t.Error("expected 3 support and 3 resistance levels")
	}
}

func TestPivotPointsCamarilla(t *testing.T) {
	candles := []models.OHLCV{
		{High: 110, Low: 90, Close: 100},
	}
	sr := PivotPoints(candles, PivotCamarilla)
	if sr.Method != "camarilla" {
		t.Errorf("expected method camarilla, got %s", sr.Method)
	}
}

func TestAutoSupportResistance(t *testing.T) {
	candles := makeCandles(100, 100, 0.1)
	sr := AutoSupportResistance(candles, 5, 0.015)
	if sr.Method != "auto" {
		t.Errorf("expected method auto, got %s", sr.Method)
	}
}

func TestVolumeProfile(t *testing.T) {
	candles := makeCandles(50, 100, 0.5)
	vp := VolumeProfile(candles, 20)
	if vp.PointOfControl <= 0 {
		t.Error("expected positive point of control")
	}
	if vp.ValueAreaHigh <= vp.ValueAreaLow {
		t.Error("value area high should be > value area low")
	}
}

// --- Signal tests ---

func TestGenerateSignals(t *testing.T) {
	candles := makeCandles(250, 100, 0.2)
	signals := GenerateSignals(candles)
	if len(signals) == 0 {
		t.Error("expected at least one signal from 250 candles")
	}
}

func TestAggregateSignal(t *testing.T) {
	signals := []models.Signal{
		{Source: "RSI", Type: models.SignalBuy, Confidence: 0.7},
		{Source: "MACD", Type: models.SignalBuy, Confidence: 0.6},
		{Source: "SuperTrend", Type: models.SignalSell, Confidence: 0.5},
	}
	sigType, conf, rec := AggregateSignal(signals)
	if sigType != models.SignalBuy {
		t.Errorf("expected BUY signal, got %s", sigType)
	}
	if conf <= 0 || conf > 1 {
		t.Errorf("confidence out of range: %.2f", conf)
	}
	_ = rec
}

func TestAggregateSignalEmpty(t *testing.T) {
	sigType, conf, rec := AggregateSignal(nil)
	if sigType != models.SignalNeutral {
		t.Errorf("expected NEUTRAL, got %s", sigType)
	}
	if conf != 0 {
		t.Errorf("expected 0 confidence for empty, got %.2f", conf)
	}
	if rec != models.Hold {
		t.Errorf("expected HOLD, got %s", rec)
	}
}

func TestFullTechnicalAnalysis(t *testing.T) {
	candles := makeCandles(250, 100, 0.2)
	result := FullTechnicalAnalysis("RELIANCE", candles)
	if result == nil {
		t.Fatal("FullTechnicalAnalysis returned nil")
	}
	if result.Ticker != "RELIANCE" {
		t.Errorf("expected ticker RELIANCE, got %s", result.Ticker)
	}
	if result.Type != models.AnalysisTechnical {
		t.Errorf("expected type technical, got %s", result.Type)
	}
	if result.Summary == "" {
		t.Error("expected non-empty summary")
	}
}
