package technical

import (
	"math/rand"
	"testing"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// benchCandles creates synthetic OHLCV data for benchmarks.
func benchCandles(n int) []models.OHLCV {
	candles := make([]models.OHLCV, n)
	rng := rand.New(rand.NewSource(42))
	price := 2500.0
	t := time.Date(2023, 1, 1, 9, 15, 0, 0, time.UTC)

	for i := range candles {
		change := (rng.Float64() - 0.48) * 50 // slight upward bias
		open := price
		close := price + change
		high := open + rng.Float64()*30
		low := open - rng.Float64()*30
		if high < close {
			high = close + rng.Float64()*10
		}
		if low > close {
			low = close - rng.Float64()*10
		}
		if low > open {
			low = open - rng.Float64()*5
		}

		candles[i] = models.OHLCV{
			Timestamp: t,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    int64(rng.Intn(5_000_000) + 100_000),
		}
		price = close
		t = t.Add(24 * time.Hour)
	}
	return candles
}

// benchCloses extracts close prices from candles.
func benchCloses(candles []models.OHLCV) []float64 {
	data := make([]float64, len(candles))
	for i, c := range candles {
		data[i] = c.Close
	}
	return data
}

// ── Moving Average Benchmarks ──

func BenchmarkSMA20_200(b *testing.B) {
	data := benchCloses(benchCandles(200))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SMA(data, 20)
	}
}

func BenchmarkSMA50_500(b *testing.B) {
	data := benchCloses(benchCandles(500))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SMA(data, 50)
	}
}

func BenchmarkSMA200_1000(b *testing.B) {
	data := benchCloses(benchCandles(1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SMA(data, 200)
	}
}

func BenchmarkEMA20_200(b *testing.B) {
	data := benchCloses(benchCandles(200))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EMA(data, 20)
	}
}

func BenchmarkEMA50_500(b *testing.B) {
	data := benchCloses(benchCandles(500))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EMA(data, 50)
	}
}

func BenchmarkWMA20_200(b *testing.B) {
	data := benchCloses(benchCandles(200))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WMA(data, 20)
	}
}

func BenchmarkMultiSMA(b *testing.B) {
	data := benchCloses(benchCandles(500))
	periods := []int{5, 10, 20, 50, 100, 200}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MultiSMA(data, periods)
	}
}

func BenchmarkMultiEMA(b *testing.B) {
	data := benchCloses(benchCandles(500))
	periods := []int{12, 26, 50, 100, 200}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MultiEMA(data, periods)
	}
}

func BenchmarkVWAP(b *testing.B) {
	candles := benchCandles(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VWAP(candles)
	}
}

// ── Technical Indicator Benchmarks ──

func BenchmarkRSI14_200(b *testing.B) {
	candles := benchCandles(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RSI(candles, 14)
	}
}

func BenchmarkRSI14_500(b *testing.B) {
	candles := benchCandles(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RSI(candles, 14)
	}
}

func BenchmarkRSILatest(b *testing.B) {
	candles := benchCandles(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RSILatest(candles, 14)
	}
}

func BenchmarkMACD_200(b *testing.B) {
	candles := benchCandles(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MACD(candles, 12, 26, 9)
	}
}

func BenchmarkMACD_500(b *testing.B) {
	candles := benchCandles(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MACD(candles, 12, 26, 9)
	}
}

func BenchmarkMACDLatest(b *testing.B) {
	candles := benchCandles(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MACDLatest(candles, 12, 26, 9)
	}
}

func BenchmarkBollingerBands_200(b *testing.B) {
	candles := benchCandles(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BollingerBands(candles, 20, 2.0)
	}
}

func BenchmarkATR14_200(b *testing.B) {
	candles := benchCandles(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ATR(candles, 14)
	}
}

func BenchmarkSuperTrend_200(b *testing.B) {
	candles := benchCandles(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SuperTrend(candles, 10, 3.0)
	}
}

// ── Pattern Detection ──

func BenchmarkDetectPatterns_200(b *testing.B) {
	candles := benchCandles(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectPatterns(candles)
	}
}

func BenchmarkDetectLatestPatterns(b *testing.B) {
	candles := benchCandles(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectLatestPatterns(candles, 20)
	}
}

// ── Support/Resistance ──

func BenchmarkPivotPointsClassic(b *testing.B) {
	candles := benchCandles(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PivotPoints(candles, PivotClassic)
	}
}

func BenchmarkAutoSupportResistance(b *testing.B) {
	candles := benchCandles(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AutoSupportResistance(candles, 20, 1.0)
	}
}

func BenchmarkVolumeProfile(b *testing.B) {
	candles := benchCandles(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VolumeProfile(candles, 50)
	}
}

// ── Signal Generation ──

func BenchmarkGenerateSignals(b *testing.B) {
	candles := benchCandles(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateSignals(candles)
	}
}

// ── Composite (heaviest) ──

func BenchmarkComputeAll(b *testing.B) {
	candles := benchCandles(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeAll("RELIANCE", candles)
	}
}

func BenchmarkFullTechnicalAnalysis(b *testing.B) {
	candles := benchCandles(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FullTechnicalAnalysis("RELIANCE", candles)
	}
}

// ── Scaling benchmark: vary N for SMA ──

func BenchmarkSMA20_100(b *testing.B) {
	data := benchCloses(benchCandles(100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SMA(data, 20)
	}
}

func BenchmarkSMA20_2000(b *testing.B) {
	data := benchCloses(benchCandles(2000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SMA(data, 20)
	}
}

func BenchmarkSMA20_5000(b *testing.B) {
	data := benchCloses(benchCandles(5000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SMA(data, 20)
	}
}
