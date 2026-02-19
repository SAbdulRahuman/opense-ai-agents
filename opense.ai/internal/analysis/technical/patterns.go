package technical

import (
	"math"

	"github.com/seenimoa/openseai/pkg/models"
)

// CandlePattern represents a detected candlestick pattern.
type CandlePattern struct {
	Name       string  // e.g., "Doji", "Hammer", "Bullish Engulfing"
	Type       string  // "bullish", "bearish", "neutral"
	Confidence float64 // 0.0 to 1.0
	Index      int     // candle index where pattern was detected
}

// DetectPatterns scans candles and returns all detected candlestick patterns.
func DetectPatterns(candles []models.OHLCV) []CandlePattern {
	var patterns []CandlePattern
	n := len(candles)

	for i := 0; i < n; i++ {
		c := candles[i]
		body := c.Close - c.Open
		absBody := math.Abs(body)
		upperShadow := c.High - math.Max(c.Open, c.Close)
		lowerShadow := math.Min(c.Open, c.Close) - c.Low
		totalRange := c.High - c.Low

		if totalRange == 0 {
			continue
		}

		// Doji: body is tiny relative to range.
		if absBody/totalRange < 0.1 {
			patterns = append(patterns, CandlePattern{
				Name: "Doji", Type: "neutral", Confidence: 0.7, Index: i,
			})
		}

		// Hammer: small body at top, long lower shadow (bullish reversal).
		if absBody/totalRange < 0.3 && lowerShadow >= 2*absBody && upperShadow < absBody {
			patterns = append(patterns, CandlePattern{
				Name: "Hammer", Type: "bullish", Confidence: 0.65, Index: i,
			})
		}

		// Inverted Hammer: small body at bottom, long upper shadow.
		if absBody/totalRange < 0.3 && upperShadow >= 2*absBody && lowerShadow < absBody {
			patterns = append(patterns, CandlePattern{
				Name: "Inverted Hammer", Type: "bullish", Confidence: 0.6, Index: i,
			})
		}

		// Shooting Star: like inverted hammer but at top of uptrend.
		if i >= 3 && absBody/totalRange < 0.3 && upperShadow >= 2*absBody && lowerShadow < absBody {
			if isUptrend(candles, i, 3) {
				patterns = append(patterns, CandlePattern{
					Name: "Shooting Star", Type: "bearish", Confidence: 0.65, Index: i,
				})
			}
		}

		// Hanging Man: hammer at top of uptrend.
		if i >= 3 && absBody/totalRange < 0.3 && lowerShadow >= 2*absBody && upperShadow < absBody {
			if isUptrend(candles, i, 3) {
				patterns = append(patterns, CandlePattern{
					Name: "Hanging Man", Type: "bearish", Confidence: 0.6, Index: i,
				})
			}
		}

		// Two-candle patterns need at least 2 candles.
		if i < 1 {
			continue
		}
		prev := candles[i-1]
		prevBody := prev.Close - prev.Open

		// Bullish Engulfing: bearish candle followed by larger bullish candle.
		if prevBody < 0 && body > 0 && c.Open <= prev.Close && c.Close >= prev.Open {
			patterns = append(patterns, CandlePattern{
				Name: "Bullish Engulfing", Type: "bullish", Confidence: 0.75, Index: i,
			})
		}

		// Bearish Engulfing: bullish candle followed by larger bearish candle.
		if prevBody > 0 && body < 0 && c.Open >= prev.Close && c.Close <= prev.Open {
			patterns = append(patterns, CandlePattern{
				Name: "Bearish Engulfing", Type: "bearish", Confidence: 0.75, Index: i,
			})
		}

		// Three-candle patterns.
		if i < 2 {
			continue
		}
		prev2 := candles[i-2]
		prev2Body := prev2.Close - prev2.Open

		// Morning Star: bearish → small body (star) → bullish.
		if prev2Body < 0 && math.Abs(prevBody)/math.Max(math.Abs(prev2Body), 0.01) < 0.3 && body > 0 {
			if c.Close > (prev2.Open+prev2.Close)/2 {
				patterns = append(patterns, CandlePattern{
					Name: "Morning Star", Type: "bullish", Confidence: 0.8, Index: i,
				})
			}
		}

		// Evening Star: bullish → small body → bearish.
		if prev2Body > 0 && math.Abs(prevBody)/math.Max(prev2Body, 0.01) < 0.3 && body < 0 {
			if c.Close < (prev2.Open+prev2.Close)/2 {
				patterns = append(patterns, CandlePattern{
					Name: "Evening Star", Type: "bearish", Confidence: 0.8, Index: i,
				})
			}
		}

		// Three White Soldiers: three consecutive bullish candles with progressively higher closes.
		if prev2Body > 0 && prevBody > 0 && body > 0 &&
			prev.Close > prev2.Close && c.Close > prev.Close {
			patterns = append(patterns, CandlePattern{
				Name: "Three White Soldiers", Type: "bullish", Confidence: 0.7, Index: i,
			})
		}

		// Three Black Crows: three consecutive bearish candles with progressively lower closes.
		if prev2Body < 0 && prevBody < 0 && body < 0 &&
			prev.Close < prev2.Close && c.Close < prev.Close {
			patterns = append(patterns, CandlePattern{
				Name: "Three Black Crows", Type: "bearish", Confidence: 0.7, Index: i,
			})
		}
	}

	return patterns
}

// DetectLatestPatterns returns patterns detected in the last N candles.
func DetectLatestPatterns(candles []models.OHLCV, lookback int) []CandlePattern {
	if lookback <= 0 {
		lookback = 5
	}
	n := len(candles)
	if n == 0 {
		return nil
	}

	all := DetectPatterns(candles)
	threshold := n - lookback
	if threshold < 0 {
		threshold = 0
	}

	var recent []CandlePattern
	for _, p := range all {
		if p.Index >= threshold {
			recent = append(recent, p)
		}
	}
	return recent
}

// --- Head & Shoulders detection (simplified) ---

// HeadAndShoulders detects a basic head-and-shoulders topping pattern.
// Returns the neckline level and confidence, or 0 if not found.
func HeadAndShoulders(candles []models.OHLCV, window int) (neckline float64, confidence float64) {
	if window <= 0 {
		window = 20
	}
	n := len(candles)
	if n < window {
		return 0, 0
	}

	// Find local peaks in the window.
	start := n - window
	peaks := findPeaks(candles[start:], 3)

	if len(peaks) < 3 {
		return 0, 0
	}

	// Check if middle peak is highest (head).
	for i := 1; i < len(peaks)-1; i++ {
		left := peaks[i-1]
		head := peaks[i]
		right := peaks[i+1]

		leftHigh := candles[start+left].High
		headHigh := candles[start+head].High
		rightHigh := candles[start+right].High

		if headHigh > leftHigh && headHigh > rightHigh {
			// Shoulders should be roughly at the same level (within 5%).
			shoulderDiff := math.Abs(leftHigh-rightHigh) / math.Max(leftHigh, rightHigh)
			if shoulderDiff < 0.05 {
				neckline = math.Min(
					candles[start+left].Low,
					candles[start+right].Low,
				)
				conf := 0.6 + (1-shoulderDiff)*0.2
				return neckline, math.Min(conf, 1.0)
			}
		}
	}

	return 0, 0
}

// --- helpers ---

func isUptrend(candles []models.OHLCV, idx, lookback int) bool {
	if idx < lookback {
		return false
	}
	gains := 0
	for i := idx - lookback; i < idx; i++ {
		if candles[i].Close > candles[i].Open {
			gains++
		}
	}
	return gains > lookback/2
}

func findPeaks(candles []models.OHLCV, minDist int) []int {
	n := len(candles)
	var peaks []int
	lastPeak := -minDist

	for i := 1; i < n-1; i++ {
		if candles[i].High > candles[i-1].High && candles[i].High > candles[i+1].High {
			if i-lastPeak >= minDist {
				peaks = append(peaks, i)
				lastPeak = i
			}
		}
	}

	return peaks
}
