package technical

import (
	"math"
	"sort"

	"github.com/seenimoa/openseai/pkg/models"
)

// PivotMethod selects the pivot-point formula.
type PivotMethod string

const (
	PivotClassic   PivotMethod = "classic"
	PivotFibonacci PivotMethod = "fibonacci"
	PivotCamarilla PivotMethod = "camarilla"
)

// PivotPoints calculates pivot-based support/resistance from the previous session.
func PivotPoints(candles []models.OHLCV, method PivotMethod) models.SupportResistance {
	if len(candles) == 0 {
		return models.SupportResistance{Method: string(method)}
	}

	last := candles[len(candles)-1]
	h, l, c := last.High, last.Low, last.Close
	rng := h - l

	sr := models.SupportResistance{Method: string(method)}

	switch method {
	case PivotFibonacci:
		pp := (h + l + c) / 3
		sr.PivotPoint = pp
		sr.S1 = pp - 0.382*rng
		sr.S2 = pp - 0.618*rng
		sr.S3 = pp - 1.0*rng
		sr.R1 = pp + 0.382*rng
		sr.R2 = pp + 0.618*rng
		sr.R3 = pp + 1.0*rng

	case PivotCamarilla:
		sr.PivotPoint = (h + l + c) / 3
		sr.S1 = c - rng*1.1/12
		sr.S2 = c - rng*1.1/6
		sr.S3 = c - rng*1.1/4
		sr.R1 = c + rng*1.1/12
		sr.R2 = c + rng*1.1/6
		sr.R3 = c + rng*1.1/4

	default: // classic
		pp := (h + l + c) / 3
		sr.PivotPoint = pp
		sr.S1 = 2*pp - h
		sr.S2 = pp - rng
		sr.S3 = l - 2*(h-pp)
		sr.R1 = 2*pp - l
		sr.R2 = pp + rng
		sr.R3 = h + 2*(pp-l)
	}

	sr.Supports = []float64{sr.S1, sr.S2, sr.S3}
	sr.Resistances = []float64{sr.R1, sr.R2, sr.R3}

	return sr
}

// AutoSupportResistance detects support/resistance levels from price action
// using a peak/trough clustering approach.
func AutoSupportResistance(candles []models.OHLCV, window int, threshold float64) models.SupportResistance {
	if window <= 0 {
		window = 5
	}
	if threshold <= 0 {
		threshold = 0.015 // 1.5% clustering
	}

	n := len(candles)
	if n < window*2+1 {
		return models.SupportResistance{Method: "auto"}
	}

	// Collect local peaks (resistance candidates) and troughs (support candidates).
	var levels []float64

	for i := window; i < n-window; i++ {
		isHigh := true
		isLow := true

		for j := i - window; j <= i+window; j++ {
			if j == i {
				continue
			}
			if candles[j].High >= candles[i].High {
				isHigh = false
			}
			if candles[j].Low <= candles[i].Low {
				isLow = false
			}
		}

		if isHigh {
			levels = append(levels, candles[i].High)
		}
		if isLow {
			levels = append(levels, candles[i].Low)
		}
	}

	if len(levels) == 0 {
		return models.SupportResistance{Method: "auto"}
	}

	// Cluster nearby levels.
	sort.Float64s(levels)
	clusters := clusterLevels(levels, threshold)

	// Current price determines support vs resistance.
	currentPrice := candles[n-1].Close
	var supports, resistances []float64

	for _, level := range clusters {
		if level < currentPrice {
			supports = append(supports, level)
		} else {
			resistances = append(resistances, level)
		}
	}

	// Sort: supports descending (nearest first), resistances ascending.
	sort.Sort(sort.Reverse(sort.Float64Slice(supports)))
	sort.Float64s(resistances)

	sr := models.SupportResistance{
		Method:      "auto",
		Supports:    capSlice(supports, 3),
		Resistances: capSlice(resistances, 3),
	}

	// Populate S1-S3, R1-R3 if available.
	if len(sr.Supports) > 0 {
		sr.S1 = sr.Supports[0]
	}
	if len(sr.Supports) > 1 {
		sr.S2 = sr.Supports[1]
	}
	if len(sr.Supports) > 2 {
		sr.S3 = sr.Supports[2]
	}
	if len(sr.Resistances) > 0 {
		sr.R1 = sr.Resistances[0]
	}
	if len(sr.Resistances) > 1 {
		sr.R2 = sr.Resistances[1]
	}
	if len(sr.Resistances) > 2 {
		sr.R3 = sr.Resistances[2]
	}

	// Pivot as midpoint.
	if sr.S1 > 0 && sr.R1 > 0 {
		sr.PivotPoint = (sr.S1 + sr.R1) / 2
	} else {
		sr.PivotPoint = currentPrice
	}

	return sr
}

// VolumeProfile identifies high-volume price zones (point of control, value area).
type VolumeProfileResult struct {
	PointOfControl float64 // price level with highest volume
	ValueAreaHigh  float64 // upper boundary of 70% volume zone
	ValueAreaLow   float64 // lower boundary
}

// VolumeProfile computes a simplified volume profile.
func VolumeProfile(candles []models.OHLCV, bins int) VolumeProfileResult {
	if bins <= 0 {
		bins = 50
	}
	n := len(candles)
	if n == 0 {
		return VolumeProfileResult{}
	}

	// Find price range.
	minPrice := candles[0].Low
	maxPrice := candles[0].High
	for _, c := range candles {
		if c.Low < minPrice {
			minPrice = c.Low
		}
		if c.High > maxPrice {
			maxPrice = c.High
		}
	}

	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		return VolumeProfileResult{PointOfControl: candles[n-1].Close}
	}

	binSize := priceRange / float64(bins)
	volumes := make([]float64, bins)
	totalVol := 0.0

	for _, c := range candles {
		tp := (c.High + c.Low + c.Close) / 3
		idx := int((tp - minPrice) / binSize)
		if idx >= bins {
			idx = bins - 1
		}
		if idx < 0 {
			idx = 0
		}
		volumes[idx] += float64(c.Volume)
		totalVol += float64(c.Volume)
	}

	// Point of Control: bin with max volume.
	pocIdx := 0
	maxVol := volumes[0]
	for i, v := range volumes {
		if v > maxVol {
			maxVol = v
			pocIdx = i
		}
	}

	poc := minPrice + (float64(pocIdx)+0.5)*binSize

	// Value area: expand from POC until 70% volume captured.
	target := totalVol * 0.70
	captured := volumes[pocIdx]
	lo, hi := pocIdx, pocIdx

	for captured < target && (lo > 0 || hi < bins-1) {
		expandLo := false
		expandHi := false

		if lo > 0 && hi < bins-1 {
			if volumes[lo-1] >= volumes[hi+1] {
				expandLo = true
			} else {
				expandHi = true
			}
		} else if lo > 0 {
			expandLo = true
		} else if hi < bins-1 {
			expandHi = true
		}

		if expandLo {
			lo--
			captured += volumes[lo]
		}
		if expandHi {
			hi++
			captured += volumes[hi]
		}
	}

	vaLow := minPrice + float64(lo)*binSize
	vaHigh := minPrice + float64(hi+1)*binSize

	return VolumeProfileResult{
		PointOfControl: math.Round(poc*100) / 100,
		ValueAreaHigh:  math.Round(vaHigh*100) / 100,
		ValueAreaLow:   math.Round(vaLow*100) / 100,
	}
}

// clusterLevels groups nearby price levels and returns midpoints.
func clusterLevels(sorted []float64, threshold float64) []float64 {
	if len(sorted) == 0 {
		return nil
	}

	var clusters []float64
	clusterSum := sorted[0]
	clusterCount := 1

	for i := 1; i < len(sorted); i++ {
		clusterMid := clusterSum / float64(clusterCount)
		if (sorted[i]-clusterMid)/clusterMid <= threshold {
			clusterSum += sorted[i]
			clusterCount++
		} else {
			clusters = append(clusters, clusterSum/float64(clusterCount))
			clusterSum = sorted[i]
			clusterCount = 1
		}
	}
	clusters = append(clusters, clusterSum/float64(clusterCount))

	return clusters
}

func capSlice(s []float64, max int) []float64 {
	if len(s) > max {
		return s[:max]
	}
	return s
}
