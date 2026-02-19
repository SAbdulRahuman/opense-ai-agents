package technical

import (
	"fmt"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// GenerateSignals produces trading signals from technical analysis on the given candles.
func GenerateSignals(candles []models.OHLCV) []models.Signal {
	if len(candles) < 30 {
		return nil
	}

	var signals []models.Signal
	last := candles[len(candles)-1]

	// Compute indicators using *Latest helpers.
	rsi := RSILatest(candles, 14)
	macd := MACDLatest(candles, 12, 26, 9)
	bb := BollingerLatest(candles, 20, 2)
	st := SuperTrendLatest(candles, 7, 3)

	closes := extractCloses(candles)
	smaVals := map[int]float64{}
	emaVals := map[int]float64{}
	for _, p := range StandardPeriods {
		if s := SMALatest(closes, p); s > 0 {
			smaVals[p] = s
		}
		if e := EMALatest(closes, p); e > 0 {
			emaVals[p] = e
		}
	}

	// --- RSI signals ---
	if rsi > 0 {
		if rsi < 30 {
			signals = append(signals, models.Signal{
				Source:     "RSI",
				Type:       models.SignalBuy,
				Confidence: models.Confidence(0.5 + (30-rsi)/100),
				Reason:     fmt.Sprintf("RSI oversold at %.1f", rsi),
				Price:      last.Close,
			})
		} else if rsi > 70 {
			signals = append(signals, models.Signal{
				Source:     "RSI",
				Type:       models.SignalSell,
				Confidence: models.Confidence(0.5 + (rsi-70)/100),
				Reason:     fmt.Sprintf("RSI overbought at %.1f", rsi),
				Price:      last.Close,
			})
		}
	}

	// --- MACD signals ---
	if macd.MACDLine != 0 && macd.SignalLine != 0 {
		if macd.Histogram > 0 && macd.MACDLine > macd.SignalLine {
			signals = append(signals, models.Signal{
				Source:     "MACD",
				Type:       models.SignalBuy,
				Confidence: models.Confidence(clampf(0.5+macd.Histogram/last.Close*100, 0, 1)),
				Reason:     fmt.Sprintf("MACD bullish crossover (histogram: %.2f)", macd.Histogram),
				Price:      last.Close,
			})
		} else if macd.Histogram < 0 && macd.MACDLine < macd.SignalLine {
			signals = append(signals, models.Signal{
				Source:     "MACD",
				Type:       models.SignalSell,
				Confidence: models.Confidence(clampf(0.5-macd.Histogram/last.Close*100, 0, 1)),
				Reason:     fmt.Sprintf("MACD bearish crossover (histogram: %.2f)", macd.Histogram),
				Price:      last.Close,
			})
		}
	}

	// --- Bollinger Band signals ---
	if bb.Upper > 0 && bb.Lower > 0 {
		if last.Close < bb.Lower {
			signals = append(signals, models.Signal{
				Source:     "Bollinger",
				Type:       models.SignalBuy,
				Confidence: 0.6,
				Reason:     fmt.Sprintf("Price (%.2f) below lower Bollinger Band (%.2f)", last.Close, bb.Lower),
				Price:      last.Close,
			})
		} else if last.Close > bb.Upper {
			signals = append(signals, models.Signal{
				Source:     "Bollinger",
				Type:       models.SignalSell,
				Confidence: 0.6,
				Reason:     fmt.Sprintf("Price (%.2f) above upper Bollinger Band (%.2f)", last.Close, bb.Upper),
				Price:      last.Close,
			})
		}
	}

	// --- SuperTrend signal ---
	if st.Value > 0 {
		if st.Trend == "UP" {
			signals = append(signals, models.Signal{
				Source:     "SuperTrend",
				Type:       models.SignalBuy,
				Confidence: 0.65,
				Reason:     fmt.Sprintf("SuperTrend bullish, support at %.2f", st.Value),
				Price:      last.Close,
			})
		} else {
			signals = append(signals, models.Signal{
				Source:     "SuperTrend",
				Type:       models.SignalSell,
				Confidence: 0.65,
				Reason:     fmt.Sprintf("SuperTrend bearish, resistance at %.2f", st.Value),
				Price:      last.Close,
			})
		}
	}

	// --- Moving Average crossover signals ---
	if sma50, ok := smaVals[50]; ok {
		if sma200, ok2 := smaVals[200]; ok2 {
			if sma50 > sma200 && last.Close > sma50 {
				signals = append(signals, models.Signal{
					Source:     "MA_Golden_Cross",
					Type:       models.SignalBuy,
					Confidence: 0.7,
					Reason:     fmt.Sprintf("SMA50 (%.2f) above SMA200 (%.2f), golden cross", sma50, sma200),
					Price:      last.Close,
				})
			} else if sma50 < sma200 && last.Close < sma50 {
				signals = append(signals, models.Signal{
					Source:     "MA_Death_Cross",
					Type:       models.SignalSell,
					Confidence: 0.7,
					Reason:     fmt.Sprintf("SMA50 (%.2f) below SMA200 (%.2f), death cross", sma50, sma200),
					Price:      last.Close,
				})
			}
		}
	}

	// --- Price vs EMA20 signals ---
	if ema20, ok := emaVals[20]; ok {
		pctDiff := (last.Close - ema20) / ema20 * 100
		if pctDiff < -3 {
			signals = append(signals, models.Signal{
				Source:     "EMA20",
				Type:       models.SignalBuy,
				Confidence: models.Confidence(clampf(0.4+(-pctDiff)/20, 0, 0.9)),
				Reason:     fmt.Sprintf("Price %.1f%% below EMA20 (%.2f)", pctDiff, ema20),
				Price:      last.Close,
			})
		} else if pctDiff > 5 {
			signals = append(signals, models.Signal{
				Source:     "EMA20",
				Type:       models.SignalSell,
				Confidence: models.Confidence(clampf(0.4+pctDiff/20, 0, 0.9)),
				Reason:     fmt.Sprintf("Price %.1f%% above EMA20 (%.2f)", pctDiff, ema20),
				Price:      last.Close,
			})
		}
	}

	// --- Pattern signals ---
	patterns := DetectLatestPatterns(candles, 3)
	for _, p := range patterns {
		sig := models.Signal{
			Source:     "Pattern",
			Confidence: models.Confidence(p.Confidence),
			Reason:     fmt.Sprintf("Candlestick pattern: %s", p.Name),
			Price:      last.Close,
		}
		switch p.Type {
		case "bullish":
			sig.Type = models.SignalBuy
		case "bearish":
			sig.Type = models.SignalSell
		default:
			sig.Type = models.SignalNeutral
		}
		signals = append(signals, sig)
	}

	return signals
}

// AggregateSignal computes a weighted aggregate from multiple signals.
func AggregateSignal(signals []models.Signal) (models.SignalType, models.Confidence, models.Recommendation) {
	if len(signals) == 0 {
		return models.SignalNeutral, 0, models.Hold
	}

	// Source weights for aggregation.
	weights := map[string]float64{
		"RSI":              1.0,
		"MACD":             1.2,
		"Bollinger":        0.8,
		"SuperTrend":       1.1,
		"MA_Golden_Cross":  1.3,
		"MA_Death_Cross":   1.3,
		"EMA20":            0.7,
		"Pattern":          0.6,
	}

	var buyScore, sellScore, totalWeight float64

	for _, sig := range signals {
		w := weights[sig.Source]
		if w == 0 {
			w = 1.0
		}
		conf := float64(sig.Confidence)

		switch sig.Type {
		case models.SignalBuy:
			buyScore += w * conf
		case models.SignalSell:
			sellScore += w * conf
		}
		totalWeight += w
	}

	if totalWeight == 0 {
		return models.SignalNeutral, 0, models.Hold
	}

	buyNorm := buyScore / totalWeight
	sellNorm := sellScore / totalWeight
	netScore := buyNorm - sellNorm // -1 to +1 range

	var sigType models.SignalType
	var rec models.Recommendation
	var conf models.Confidence

	switch {
	case netScore > 0.3:
		sigType = models.SignalBuy
		rec = models.StrongBuy
		conf = models.Confidence(clampf(0.7+netScore*0.3, 0, 1))
	case netScore > 0.1:
		sigType = models.SignalBuy
		rec = models.ModerateBuy
		conf = models.Confidence(clampf(0.5+netScore*0.3, 0, 1))
	case netScore < -0.3:
		sigType = models.SignalSell
		rec = models.StrongSell
		conf = models.Confidence(clampf(0.7+(-netScore)*0.3, 0, 1))
	case netScore < -0.1:
		sigType = models.SignalSell
		rec = models.ModerateSell
		conf = models.Confidence(clampf(0.5+(-netScore)*0.3, 0, 1))
	default:
		sigType = models.SignalNeutral
		rec = models.Hold
		conf = models.Confidence(0.4)
	}

	return sigType, conf, rec
}

// FullTechnicalAnalysis runs complete technical analysis and returns an AnalysisResult.
func FullTechnicalAnalysis(ticker string, candles []models.OHLCV) *models.AnalysisResult {
	signals := GenerateSignals(candles)
	sigType, conf, rec := AggregateSignal(signals)

	// Compute indicators for details.
	indicators := ComputeAll(ticker, candles)

	// Support/Resistance.
	pivotSR := PivotPoints(candles, PivotClassic)
	pivotSR.Ticker = ticker
	autoSR := AutoSupportResistance(candles, 5, 0.015)
	autoSR.Ticker = ticker

	// Patterns in recent candles.
	patterns := DetectLatestPatterns(candles, 5)
	patternNames := make([]string, len(patterns))
	for i, p := range patterns {
		patternNames[i] = p.Name
	}

	// Volume profile.
	vp := VolumeProfile(candles, 50)

	details := map[string]any{
		"indicators":       indicators,
		"pivot_sr":         pivotSR,
		"auto_sr":          autoSR,
		"patterns":         patternNames,
		"volume_profile":   vp,
		"signal_type":      sigType,
		"signal_count_buy": countSignals(signals, models.SignalBuy),
		"signal_count_sell": countSignals(signals, models.SignalSell),
	}

	summary := fmt.Sprintf("Technical analysis for %s: %s signal with %.0f%% confidence. %s",
		ticker, sigType, float64(conf)*100, summarizeSignals(signals))

	return &models.AnalysisResult{
		Ticker:         ticker,
		Type:           models.AnalysisTechnical,
		AgentName:      "technical-analysis",
		Signals:        signals,
		Recommendation: rec,
		Confidence:     conf,
		Summary:        summary,
		Details:        details,
		Timestamp:      time.Now(),
	}
}

// --- helpers ---

func clampf(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func countSignals(signals []models.Signal, t models.SignalType) int {
	count := 0
	for _, s := range signals {
		if s.Type == t {
			count++
		}
	}
	return count
}

func summarizeSignals(signals []models.Signal) string {
	buy, sell, neutral := 0, 0, 0
	for _, s := range signals {
		switch s.Type {
		case models.SignalBuy:
			buy++
		case models.SignalSell:
			sell++
		default:
			neutral++
		}
	}
	return fmt.Sprintf("%d buy, %d sell, %d neutral signals from %d indicators",
		buy, sell, neutral, len(signals))
}
