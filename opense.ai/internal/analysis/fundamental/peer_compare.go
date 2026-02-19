package fundamental

import (
	"sort"

	"github.com/seenimoa/openseai/pkg/models"
)

// PeerEntry represents a single peer in comparison.
type PeerEntry struct {
	Ticker      string             `json:"ticker"`
	Name        string             `json:"name"`
	Ratios      models.FinancialRatios `json:"ratios"`
	MarketCap   float64            `json:"market_cap"`
	Price       float64            `json:"price"`
	Rank        int                `json:"rank"`        // computed rank
	Score       float64            `json:"score"`       // composite score
}

// PeerComparison holds comparative analysis across peers.
type PeerComparison struct {
	Target  PeerEntry   `json:"target"`
	Peers   []PeerEntry `json:"peers"`
	Metrics []string    `json:"metrics"` // metrics used for comparison
	Summary string      `json:"summary"`
}

// ComparePeers ranks a target stock against its sector peers.
func ComparePeers(target PeerEntry, peers []PeerEntry) PeerComparison {
	pc := PeerComparison{
		Target:  target,
		Peers:   peers,
		Metrics: []string{"PE", "PB", "ROE", "ROCE", "D/E", "EPS Growth"},
	}

	// Merge target + peers for ranking.
	all := make([]PeerEntry, 0, len(peers)+1)
	all = append(all, target)
	all = append(all, peers...)

	// Score each peer on multiple factors.
	for i := range all {
		all[i].Score = scorePeer(all[i].Ratios)
	}

	// Sort by score descending.
	sort.Slice(all, func(i, j int) bool {
		return all[i].Score > all[j].Score
	})

	// Assign ranks.
	for i := range all {
		all[i].Rank = i + 1
	}

	// Find target's rank.
	targetRank := 0
	for _, p := range all {
		if p.Ticker == target.Ticker {
			targetRank = p.Rank
			pc.Target = p
			break
		}
	}

	// Rebuild peers without target.
	pc.Peers = make([]PeerEntry, 0, len(peers))
	for _, p := range all {
		if p.Ticker != target.Ticker {
			pc.Peers = append(pc.Peers, p)
		}
	}

	pc.Summary = buildPeerSummary(target.Ticker, targetRank, len(all))

	return pc
}

// RelativeValuation computes whether a stock is cheap/expensive vs peers.
type RelativeMetric struct {
	Metric      string  `json:"metric"`
	TargetValue float64 `json:"target_value"`
	PeerAvg     float64 `json:"peer_avg"`
	PeerMedian  float64 `json:"peer_median"`
	Percentile  float64 `json:"percentile"` // 0-100
}

// RelativeValuationMetrics computes relative valuation against peers.
func RelativeValuationMetrics(target models.FinancialRatios, peers []models.FinancialRatios) []RelativeMetric {
	type metricExtractor struct {
		name string
		fn   func(models.FinancialRatios) float64
		// lower is better? (for PE, PB, D/E); false means higher is better (ROE, ROCE)
		lowerBetter bool
	}

	extractors := []metricExtractor{
		{"PE", func(r models.FinancialRatios) float64 { return r.PE }, true},
		{"PB", func(r models.FinancialRatios) float64 { return r.PB }, true},
		{"EV/EBITDA", func(r models.FinancialRatios) float64 { return r.EVBITDA }, true},
		{"ROE", func(r models.FinancialRatios) float64 { return r.ROE }, false},
		{"ROCE", func(r models.FinancialRatios) float64 { return r.ROCE }, false},
		{"D/E", func(r models.FinancialRatios) float64 { return r.DebtEquity }, true},
		{"Dividend Yield", func(r models.FinancialRatios) float64 { return r.DividendYield }, false},
	}

	var results []RelativeMetric

	for _, ext := range extractors {
		tv := ext.fn(target)
		if tv == 0 {
			continue
		}

		var vals []float64
		for _, p := range peers {
			v := ext.fn(p)
			if v > 0 {
				vals = append(vals, v)
			}
		}

		if len(vals) == 0 {
			continue
		}

		avg := avgFloat(vals)
		med := medianFloat(vals)

		// Percentile rank.
		below := 0
		for _, v := range vals {
			if ext.lowerBetter {
				// For lower-is-better, count how many peers are above (worse).
				if v > tv {
					below++
				}
			} else {
				// For higher-is-better, count how many peers are below (worse).
				if v < tv {
					below++
				}
			}
		}
		pctile := float64(below) / float64(len(vals)) * 100

		results = append(results, RelativeMetric{
			Metric:      ext.name,
			TargetValue: tv,
			PeerAvg:     avg,
			PeerMedian:  med,
			Percentile:  pctile,
		})
	}

	return results
}

// --- helpers ---

func scorePeer(r models.FinancialRatios) float64 {
	score := 0.0

	// Higher ROE is better (max 30 points).
	if r.ROE > 0 {
		s := r.ROE
		if s > 30 {
			s = 30
		}
		score += s
	}

	// Higher ROCE is better (max 30 points).
	if r.ROCE > 0 {
		s := r.ROCE
		if s > 30 {
			s = 30
		}
		score += s
	}

	// Lower PE is better (max 20 points).
	if r.PE > 0 && r.PE < 100 {
		score += (100 - r.PE) / 5
	}

	// Lower D/E is better (max 10 points).
	if r.DebtEquity >= 0 && r.DebtEquity < 5 {
		score += (5 - r.DebtEquity) * 2
	}

	// Dividend yield bonus (max 10 points).
	if r.DividendYield > 0 {
		s := r.DividendYield * 2
		if s > 10 {
			s = 10
		}
		score += s
	}

	return score
}

func buildPeerSummary(ticker string, rank, total int) string {
	pctile := (1 - float64(rank-1)/float64(total)) * 100
	switch {
	case pctile >= 80:
		return ticker + " ranks in the top quintile among peers — strong fundamentals"
	case pctile >= 60:
		return ticker + " ranks above average among peers"
	case pctile >= 40:
		return ticker + " ranks average among peers"
	case pctile >= 20:
		return ticker + " ranks below average among peers"
	default:
		return ticker + " ranks in the bottom quintile among peers — weak relative position"
	}
}

func avgFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func medianFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}
