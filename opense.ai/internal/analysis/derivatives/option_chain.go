package derivatives

import (
	"math"
	"sort"

	"github.com/seenimoa/openseai/pkg/models"
)

// OptionChainAnalysis holds derived insights from the option chain.
type OptionChainAnalysis struct {
	Ticker       string       `json:"ticker"`
	SpotPrice    float64      `json:"spot_price"`
	PCR          float64      `json:"pcr"`
	MaxPain      float64      `json:"max_pain"`
	IVSkew       float64      `json:"iv_skew"`       // ATM IV difference (PE-CE)
	ATMStrike    float64      `json:"atm_strike"`
	ATMIV        float64      `json:"atm_iv"`        // average ATM IV
	OISRLevels   OISupportRes `json:"oi_sr_levels"`
	Sentiment    string       `json:"sentiment"`     // "bullish", "bearish", "neutral"
}

// OISupportRes contains OI-based support and resistance levels.
type OISupportRes struct {
	MaxPutOIStrike  float64   `json:"max_put_oi_strike"`  // strongest support
	MaxCallOIStrike float64   `json:"max_call_oi_strike"` // strongest resistance
	TopPutStrikes   []float64 `json:"top_put_strikes"`    // top 3 support levels
	TopCallStrikes  []float64 `json:"top_call_strikes"`   // top 3 resistance levels
}

// AnalyzeOptionChain performs comprehensive analysis on an option chain.
func AnalyzeOptionChain(oc *models.OptionChain) OptionChainAnalysis {
	if oc == nil || len(oc.Contracts) == 0 {
		return OptionChainAnalysis{}
	}

	a := OptionChainAnalysis{
		Ticker:    oc.Ticker,
		SpotPrice: oc.SpotPrice,
		PCR:       oc.PCR,
		MaxPain:   oc.MaxPain,
	}

	// Find ATM strike (closest to spot).
	a.ATMStrike = findATMStrike(oc.Contracts, oc.SpotPrice)

	// IV analysis.
	var atmCEIV, atmPEIV float64
	for _, c := range oc.Contracts {
		if c.StrikePrice == a.ATMStrike {
			if c.OptionType == "CE" {
				atmCEIV = c.IV
			} else {
				atmPEIV = c.IV
			}
		}
	}
	if atmCEIV > 0 && atmPEIV > 0 {
		a.ATMIV = (atmCEIV + atmPEIV) / 2
		a.IVSkew = atmPEIV - atmCEIV
	}

	// OI-based support/resistance.
	a.OISRLevels = computeOISR(oc.Contracts)

	// Sentiment from PCR.
	switch {
	case a.PCR > 1.2:
		a.Sentiment = "bullish" // high PCR → more puts sold → bullish
	case a.PCR < 0.7:
		a.Sentiment = "bearish"
	default:
		a.Sentiment = "neutral"
	}

	return a
}

// ComputeMaxPain calculates the max pain strike from option contracts.
func ComputeMaxPain(contracts []models.OptionContract) float64 {
	if len(contracts) == 0 {
		return 0
	}

	// Collect unique strikes.
	strikeSet := map[float64]bool{}
	for _, c := range contracts {
		strikeSet[c.StrikePrice] = true
	}

	var strikes []float64
	for s := range strikeSet {
		strikes = append(strikes, s)
	}
	sort.Float64s(strikes)

	// Build OI maps.
	ceOI := map[float64]int64{}
	peOI := map[float64]int64{}
	for _, c := range contracts {
		if c.OptionType == "CE" {
			ceOI[c.StrikePrice] = c.OI
		} else {
			peOI[c.StrikePrice] = c.OI
		}
	}

	// Max pain = strike that minimizes total pain (option buyers' loss).
	minPain := math.MaxFloat64
	maxPainStrike := 0.0

	for _, expiry := range strikes {
		totalPain := 0.0

		// CE pain: calls ITM for all strikes below expiry.
		for _, s := range strikes {
			if s < expiry && ceOI[s] > 0 {
				totalPain += (expiry - s) * float64(ceOI[s])
			}
		}

		// PE pain: puts ITM for all strikes above expiry.
		for _, s := range strikes {
			if s > expiry && peOI[s] > 0 {
				totalPain += (s - expiry) * float64(peOI[s])
			}
		}

		if totalPain < minPain {
			minPain = totalPain
			maxPainStrike = expiry
		}
	}

	return maxPainStrike
}

// FuturesBasisAnalysis analyzes the futures premium/discount.
type FuturesBasisResult struct {
	Ticker    string  `json:"ticker"`
	Basis     float64 `json:"basis"`     // futures - spot
	BasisPct  float64 `json:"basis_pct"`
	Signal    string  `json:"signal"`
	Annualized float64 `json:"annualized"` // annualized basis %
}

// AnalyzeFuturesBasis evaluates futures premium.
func AnalyzeFuturesBasis(fut *models.FuturesContract, spotPrice float64, daysToExpiry int) FuturesBasisResult {
	if fut == nil || spotPrice <= 0 {
		return FuturesBasisResult{}
	}

	basis := fut.LTP - spotPrice
	basisPct := basis / spotPrice * 100
	annualized := 0.0
	if daysToExpiry > 0 {
		annualized = basisPct * 365 / float64(daysToExpiry)
	}

	signal := "neutral"
	if basisPct > 1 {
		signal = "bullish" // strong premium
	} else if basisPct < -0.5 {
		signal = "bearish" // discount
	}

	return FuturesBasisResult{
		Ticker:     fut.Ticker,
		Basis:      basis,
		BasisPct:   basisPct,
		Signal:     signal,
		Annualized: annualized,
	}
}

// --- helpers ---

func findATMStrike(contracts []models.OptionContract, spot float64) float64 {
	if len(contracts) == 0 || spot <= 0 {
		return 0
	}

	closest := contracts[0].StrikePrice
	minDiff := math.Abs(closest - spot)

	for _, c := range contracts {
		diff := math.Abs(c.StrikePrice - spot)
		if diff < minDiff {
			minDiff = diff
			closest = c.StrikePrice
		}
	}

	return closest
}

type oiEntry struct {
	strike float64
	oi     int64
}

func computeOISR(contracts []models.OptionContract) OISupportRes {
	var ceEntries, peEntries []oiEntry

	// Aggregate OI by strike and type.
	ceMap := map[float64]int64{}
	peMap := map[float64]int64{}

	for _, c := range contracts {
		if c.OptionType == "CE" {
			ceMap[c.StrikePrice] += c.OI
		} else {
			peMap[c.StrikePrice] += c.OI
		}
	}

	for s, oi := range ceMap {
		ceEntries = append(ceEntries, oiEntry{s, oi})
	}
	for s, oi := range peMap {
		peEntries = append(peEntries, oiEntry{s, oi})
	}

	// Sort by OI descending.
	sort.Slice(ceEntries, func(i, j int) bool { return ceEntries[i].oi > ceEntries[j].oi })
	sort.Slice(peEntries, func(i, j int) bool { return peEntries[i].oi > peEntries[j].oi })

	sr := OISupportRes{}

	if len(ceEntries) > 0 {
		sr.MaxCallOIStrike = ceEntries[0].strike
		for i := 0; i < 3 && i < len(ceEntries); i++ {
			sr.TopCallStrikes = append(sr.TopCallStrikes, ceEntries[i].strike)
		}
	}

	if len(peEntries) > 0 {
		sr.MaxPutOIStrike = peEntries[0].strike
		for i := 0; i < 3 && i < len(peEntries); i++ {
			sr.TopPutStrikes = append(sr.TopPutStrikes, peEntries[i].strike)
		}
	}

	return sr
}
