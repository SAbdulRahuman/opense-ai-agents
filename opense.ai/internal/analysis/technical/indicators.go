// Package technical implements technical analysis indicators and signal generation
// for NSE stock price data. All functions operate on []models.OHLCV candle slices.
package technical

import (
	"math"

	"github.com/seenimoa/openseai/pkg/models"
)

// RSI calculates the Relative Strength Index for the given period.
// Default period is 14. Returns values 0–100.
func RSI(candles []models.OHLCV, period int) []float64 {
	if period <= 0 {
		period = 14
	}
	n := len(candles)
	if n < period+1 {
		return nil
	}

	rsi := make([]float64, n)
	// Calculate initial gains and losses.
	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		change := candles[i].Close - candles[i-1].Close
		if change > 0 {
			avgGain += change
		} else {
			avgLoss += -change
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[period] = 100 - (100 / (1 + rs))
	}

	// Wilder's smoothing for subsequent values.
	for i := period + 1; i < n; i++ {
		change := candles[i].Close - candles[i-1].Close
		gain, loss := 0.0, 0.0
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return rsi
}

// RSILatest returns only the most recent RSI value.
func RSILatest(candles []models.OHLCV, period int) float64 {
	vals := RSI(candles, period)
	if len(vals) == 0 {
		return 0
	}
	return vals[len(vals)-1]
}

// MACDResult holds a single MACD computation point.
type MACDResult struct {
	MACD      float64
	Signal    float64
	Histogram float64
}

// MACD calculates the Moving Average Convergence Divergence.
// Default parameters: fast=12, slow=26, signal=9.
func MACD(candles []models.OHLCV, fast, slow, signal int) []MACDResult {
	if fast <= 0 {
		fast = 12
	}
	if slow <= 0 {
		slow = 26
	}
	if signal <= 0 {
		signal = 9
	}

	closes := extractCloses(candles)
	if len(closes) < slow {
		return nil
	}

	fastEMA := emaCalc(closes, fast)
	slowEMA := emaCalc(closes, slow)

	n := len(closes)
	macdLine := make([]float64, n)
	for i := 0; i < n; i++ {
		macdLine[i] = fastEMA[i] - slowEMA[i]
	}

	signalLine := emaCalc(macdLine, signal)

	results := make([]MACDResult, n)
	for i := 0; i < n; i++ {
		results[i] = MACDResult{
			MACD:      macdLine[i],
			Signal:    signalLine[i],
			Histogram: macdLine[i] - signalLine[i],
		}
	}

	return results
}

// MACDLatest returns the most recent MACD values.
func MACDLatest(candles []models.OHLCV, fast, slow, signal int) models.MACDData {
	results := MACD(candles, fast, slow, signal)
	if len(results) == 0 {
		return models.MACDData{}
	}
	r := results[len(results)-1]
	return models.MACDData{
		MACDLine:   r.MACD,
		SignalLine:  r.Signal,
		Histogram:  r.Histogram,
	}
}

// BollingerBands calculates Bollinger Bands (upper, middle, lower).
// Default: period=20, stddev multiplier=2.
func BollingerBands(candles []models.OHLCV, period int, mult float64) []models.BollingerData {
	if period <= 0 {
		period = 20
	}
	if mult <= 0 {
		mult = 2.0
	}

	closes := extractCloses(candles)
	n := len(closes)
	if n < period {
		return nil
	}

	result := make([]models.BollingerData, n)
	for i := period - 1; i < n; i++ {
		window := closes[i-period+1 : i+1]
		mean := avg(window)
		sd := stddev(window, mean)
		result[i] = models.BollingerData{
			Upper:  mean + mult*sd,
			Middle: mean,
			Lower:  mean - mult*sd,
		}
	}

	return result
}

// BollingerLatest returns the most recent Bollinger Bands values.
func BollingerLatest(candles []models.OHLCV, period int, mult float64) models.BollingerData {
	vals := BollingerBands(candles, period, mult)
	if len(vals) == 0 {
		return models.BollingerData{}
	}
	return vals[len(vals)-1]
}

// ATR calculates the Average True Range for the given period.
func ATR(candles []models.OHLCV, period int) []float64 {
	if period <= 0 {
		period = 14
	}
	n := len(candles)
	if n < 2 {
		return nil
	}

	// Calculate True Range.
	tr := make([]float64, n)
	tr[0] = candles[0].High - candles[0].Low
	for i := 1; i < n; i++ {
		hl := candles[i].High - candles[i].Low
		hc := math.Abs(candles[i].High - candles[i-1].Close)
		lc := math.Abs(candles[i].Low - candles[i-1].Close)
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}

	// Wilder's smoothing ATR.
	atr := make([]float64, n)
	if n < period {
		return atr
	}

	// First ATR = simple average of first `period` true ranges.
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += tr[i]
	}
	atr[period-1] = sum / float64(period)

	// Subsequent ATR uses smoothing.
	for i := period; i < n; i++ {
		atr[i] = (atr[i-1]*float64(period-1) + tr[i]) / float64(period)
	}

	return atr
}

// ATRLatest returns the most recent ATR value.
func ATRLatest(candles []models.OHLCV, period int) float64 {
	vals := ATR(candles, period)
	if len(vals) == 0 {
		return 0
	}
	return vals[len(vals)-1]
}

// SuperTrend calculates the SuperTrend indicator.
// Default: period=7, multiplier=3.
func SuperTrend(candles []models.OHLCV, period int, mult float64) []models.SuperTrendData {
	if period <= 0 {
		period = 7
	}
	if mult <= 0 {
		mult = 3.0
	}

	n := len(candles)
	atr := ATR(candles, period)
	if len(atr) == 0 || n < period {
		return nil
	}

	result := make([]models.SuperTrendData, n)
	upperBand := make([]float64, n)
	lowerBand := make([]float64, n)

	for i := period - 1; i < n; i++ {
		mid := (candles[i].High + candles[i].Low) / 2
		upperBand[i] = mid + mult*atr[i]
		lowerBand[i] = mid - mult*atr[i]
	}

	// Adjust bands based on previous values.
	for i := period; i < n; i++ {
		if lowerBand[i] < lowerBand[i-1] && candles[i-1].Close > lowerBand[i-1] {
			lowerBand[i] = lowerBand[i-1]
		}
		if upperBand[i] > upperBand[i-1] && candles[i-1].Close < upperBand[i-1] {
			upperBand[i] = upperBand[i-1]
		}
	}

	// Determine trend direction.
	for i := period - 1; i < n; i++ {
		if i == period-1 {
			if candles[i].Close > upperBand[i] {
				result[i] = models.SuperTrendData{Value: lowerBand[i], Trend: "UP"}
			} else {
				result[i] = models.SuperTrendData{Value: upperBand[i], Trend: "DOWN"}
			}
			continue
		}

		prevTrend := result[i-1].Trend
		if prevTrend == "UP" {
			if candles[i].Close < lowerBand[i] {
				result[i] = models.SuperTrendData{Value: upperBand[i], Trend: "DOWN"}
			} else {
				result[i] = models.SuperTrendData{Value: lowerBand[i], Trend: "UP"}
			}
		} else {
			if candles[i].Close > upperBand[i] {
				result[i] = models.SuperTrendData{Value: lowerBand[i], Trend: "UP"}
			} else {
				result[i] = models.SuperTrendData{Value: upperBand[i], Trend: "DOWN"}
			}
		}
	}

	return result
}

// SuperTrendLatest returns the most recent SuperTrend value.
func SuperTrendLatest(candles []models.OHLCV, period int, mult float64) models.SuperTrendData {
	vals := SuperTrend(candles, period, mult)
	if len(vals) == 0 {
		return models.SuperTrendData{}
	}
	return vals[len(vals)-1]
}

// ComputeAll calculates all major indicators and returns a TechnicalIndicators struct.
func ComputeAll(ticker string, candles []models.OHLCV) *models.TechnicalIndicators {
	if len(candles) == 0 {
		return nil
	}

	ti := &models.TechnicalIndicators{
		Ticker: ticker,
		RSI:    RSILatest(candles, 14),
		MACD:   MACDLatest(candles, 12, 26, 9),
		SMA:    make(map[int]float64),
		EMA:    make(map[int]float64),
		Bollinger:  BollingerLatest(candles, 20, 2),
		SuperTrend: SuperTrendLatest(candles, 7, 3),
		ATR:    ATRLatest(candles, 14),
	}

	closes := extractCloses(candles)
	for _, p := range []int{5, 10, 20, 50, 100, 200} {
		if sma := SMALatest(closes, p); sma > 0 {
			ti.SMA[p] = sma
		}
		if ema := EMALatest(closes, p); ema > 0 {
			ti.EMA[p] = ema
		}
	}

	ti.VWAP = VWAPLatest(candles)

	return ti
}

// --- helper functions ---

func extractCloses(candles []models.OHLCV) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	return closes
}

func avg(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func stddev(data []float64, mean float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sumSq := 0.0
	for _, v := range data {
		d := v - mean
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(data)))
}

func emaCalc(data []float64, period int) []float64 {
	n := len(data)
	if n == 0 || period <= 0 {
		return make([]float64, n)
	}

	ema := make([]float64, n)
	k := 2.0 / float64(period+1)

	// Seed with SMA of first `period` values.
	if n < period {
		return ema
	}
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	ema[period-1] = sum / float64(period)

	for i := period; i < n; i++ {
		ema[i] = data[i]*k + ema[i-1]*(1-k)
	}

	return ema
}
