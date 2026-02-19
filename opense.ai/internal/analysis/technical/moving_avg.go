package technical

import (
	"github.com/seenimoa/openseai/pkg/models"
)

// SMA calculates Simple Moving Average for the given period.
func SMA(data []float64, period int) []float64 {
	n := len(data)
	if n < period || period <= 0 {
		return nil
	}

	result := make([]float64, n)
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[period-1] = sum / float64(period)

	for i := period; i < n; i++ {
		sum += data[i] - data[i-period]
		result[i] = sum / float64(period)
	}

	return result
}

// SMALatest returns the most recent SMA value.
func SMALatest(data []float64, period int) float64 {
	vals := SMA(data, period)
	if len(vals) == 0 {
		return 0
	}
	return vals[len(vals)-1]
}

// EMA calculates Exponential Moving Average for the given period.
func EMA(data []float64, period int) []float64 {
	return emaCalc(data, period)
}

// EMALatest returns the most recent EMA value.
func EMALatest(data []float64, period int) float64 {
	vals := EMA(data, period)
	if len(vals) == 0 {
		return 0
	}
	return vals[len(vals)-1]
}

// WMA calculates Weighted Moving Average for the given period.
// More recent prices get higher weight.
func WMA(data []float64, period int) []float64 {
	n := len(data)
	if n < period || period <= 0 {
		return nil
	}

	result := make([]float64, n)
	denominator := float64(period * (period + 1) / 2)

	for i := period - 1; i < n; i++ {
		weightedSum := 0.0
		for j := 0; j < period; j++ {
			weight := float64(j + 1)
			weightedSum += data[i-period+1+j] * weight
		}
		result[i] = weightedSum / denominator
	}

	return result
}

// WMALatest returns the most recent WMA value.
func WMALatest(data []float64, period int) float64 {
	vals := WMA(data, period)
	if len(vals) == 0 {
		return 0
	}
	return vals[len(vals)-1]
}

// VWAP calculates Volume Weighted Average Price for the candle series.
// Typically computed intraday — resets daily. This computes a running VWAP
// across the entire series.
func VWAP(candles []models.OHLCV) []float64 {
	n := len(candles)
	if n == 0 {
		return nil
	}

	result := make([]float64, n)
	cumVolume := 0.0
	cumTPV := 0.0 // cumulative (typical price * volume)

	for i := 0; i < n; i++ {
		tp := (candles[i].High + candles[i].Low + candles[i].Close) / 3
		vol := float64(candles[i].Volume)
		cumTPV += tp * vol
		cumVolume += vol

		if cumVolume > 0 {
			result[i] = cumTPV / cumVolume
		}
	}

	return result
}

// VWAPLatest returns the most recent VWAP value.
func VWAPLatest(candles []models.OHLCV) float64 {
	vals := VWAP(candles)
	if len(vals) == 0 {
		return 0
	}
	return vals[len(vals)-1]
}

// MultiSMA computes SMA for multiple periods at once.
func MultiSMA(data []float64, periods []int) map[int]float64 {
	result := make(map[int]float64, len(periods))
	for _, p := range periods {
		if v := SMALatest(data, p); v > 0 {
			result[p] = v
		}
	}
	return result
}

// MultiEMA computes EMA for multiple periods at once.
func MultiEMA(data []float64, periods []int) map[int]float64 {
	result := make(map[int]float64, len(periods))
	for _, p := range periods {
		if v := EMALatest(data, p); v > 0 {
			result[p] = v
		}
	}
	return result
}

// StandardPeriods are the commonly used MA periods for Indian market analysis.
var StandardPeriods = []int{5, 10, 20, 50, 100, 200}
