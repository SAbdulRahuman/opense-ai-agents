package derivatives

import (
	"fmt"
	"math"
	"time"

	"github.com/seenimoa/openseai/pkg/models"
)

// PCRAnalysis holds put-call ratio analysis results.
type PCRAnalysis struct {
	PCR         float64 `json:"pcr"`
	PCRByVolume float64 `json:"pcr_by_volume"`
	Signal      string  `json:"signal"`
	Interpretation string `json:"interpretation"`
}

// ComputePCR calculates PCR from option chain data.
func ComputePCR(oc *models.OptionChain) PCRAnalysis {
	if oc == nil || len(oc.Contracts) == 0 {
		return PCRAnalysis{}
	}

	var totalPutOI, totalCallOI int64
	var totalPutVol, totalCallVol int64

	for _, c := range oc.Contracts {
		if c.OptionType == "PE" {
			totalPutOI += c.OI
			totalPutVol += c.Volume
		} else if c.OptionType == "CE" {
			totalCallOI += c.OI
			totalCallVol += c.Volume
		}
	}

	a := PCRAnalysis{}

	if totalCallOI > 0 {
		a.PCR = float64(totalPutOI) / float64(totalCallOI)
	}
	if totalCallVol > 0 {
		a.PCRByVolume = float64(totalPutVol) / float64(totalCallVol)
	}

	switch {
	case a.PCR > 1.5:
		a.Signal = "strongly_bullish"
		a.Interpretation = "Very high PCR — excessive put writing indicates strong support"
	case a.PCR > 1.2:
		a.Signal = "bullish"
		a.Interpretation = "High PCR — more puts sold, indicating bullish undertone"
	case a.PCR > 0.8:
		a.Signal = "neutral"
		a.Interpretation = "PCR in normal range — no clear directional bias"
	case a.PCR > 0.5:
		a.Signal = "bearish"
		a.Interpretation = "Low PCR — more calls than puts, indicating bearish sentiment"
	default:
		a.Signal = "strongly_bearish"
		a.Interpretation = "Very low PCR — excessive call buying, potential top formation"
	}

	return a
}

// ClassifyOIBuildup classifies the OI buildup for a futures contract.
func ClassifyOIBuildup(priceChange float64, oiChange int64) models.OIBuildupType {
	switch {
	case priceChange > 0 && oiChange > 0:
		return models.LongBuildup
	case priceChange < 0 && oiChange > 0:
		return models.ShortBuildup
	case priceChange < 0 && oiChange < 0:
		return models.LongUnwinding
	case priceChange > 0 && oiChange < 0:
		return models.ShortCovering
	default:
		return models.LongBuildup // price unchanged
	}
}

// OIBuildupAnalysis holds the complete OI buildup analysis.
type OIBuildupAnalysis struct {
	FuturesBuildup models.OIBuildupData `json:"futures_buildup"`
	TopLongBuildup []StrikeBuildup      `json:"top_long_buildup"`
	TopShortBuildup []StrikeBuildup     `json:"top_short_buildup"`
	Interpretation string               `json:"interpretation"`
}

// StrikeBuildup represents OI change at a specific strike.
type StrikeBuildup struct {
	Strike      float64 `json:"strike"`
	OptionType  string  `json:"option_type"`
	OIChange    int64   `json:"oi_change"`
	OIChangePct float64 `json:"oi_change_pct"`
}

// AnalyzeOIBuildup classifies OI changes across the option chain.
func AnalyzeOIBuildup(oc *models.OptionChain, fut *models.FuturesContract) OIBuildupAnalysis {
	a := OIBuildupAnalysis{}

	// Futures buildup.
	if fut != nil {
		a.FuturesBuildup = models.OIBuildupData{
			Ticker:      fut.Ticker,
			Buildup:     ClassifyOIBuildup(fut.Change, fut.OIChange),
			PriceChange: fut.Change,
			OIChange:    fut.OIChange,
		}
		if fut.OI > 0 {
			a.FuturesBuildup.OIChangePct = float64(fut.OIChange) / float64(fut.OI) * 100
		}

		switch a.FuturesBuildup.Buildup {
		case models.LongBuildup:
			a.Interpretation = "Long buildup in futures — fresh longs being added, bullish"
		case models.ShortBuildup:
			a.Interpretation = "Short buildup in futures — fresh shorts being added, bearish"
		case models.LongUnwinding:
			a.Interpretation = "Long unwinding in futures — longs exiting, weak sentiment"
		case models.ShortCovering:
			a.Interpretation = "Short covering in futures — shorts exiting, potential bounce"
		}
	}

	// Option chain strike-level buildup.
	if oc != nil {
		var longBuild, shortBuild []StrikeBuildup

		for _, c := range oc.Contracts {
			if c.OIChange == 0 {
				continue
			}
			sb := StrikeBuildup{
				Strike:     c.StrikePrice,
				OptionType: c.OptionType,
				OIChange:   c.OIChange,
				OIChangePct: c.OIChangePct,
			}

			// Put OI increase = support being built (bullish).
			// Call OI increase = resistance being built (bearish).
			if c.OptionType == "PE" && c.OIChange > 0 {
				longBuild = append(longBuild, sb)
			} else if c.OptionType == "CE" && c.OIChange > 0 {
				shortBuild = append(shortBuild, sb)
			}
		}

		// Sort by absolute OI change.
		sortBuildups(longBuild)
		sortBuildups(shortBuild)

		a.TopLongBuildup = capBuildups(longBuild, 5)
		a.TopShortBuildup = capBuildups(shortBuild, 5)
	}

	return a
}

// StrategyBuilder builds common option strategies from the chain.

// BuildBullCallSpread creates a bull call spread near the money.
func BuildBullCallSpread(oc *models.OptionChain, lotSize int) models.OptionStrategy {
	if oc == nil || len(oc.Contracts) == 0 || lotSize <= 0 {
		return models.OptionStrategy{Name: "Bull Call Spread"}
	}

	atm := findATMStrike(oc.Contracts, oc.SpotPrice)

	var buyCall, sellCall models.OptionContract
	var foundBuy, foundSell bool

	// Buy ATM call, sell next OTM call.
	for _, c := range oc.Contracts {
		if c.OptionType == "CE" && c.StrikePrice == atm {
			buyCall = c
			foundBuy = true
		}
		if c.OptionType == "CE" && c.StrikePrice > atm && !foundSell {
			sellCall = c
			foundSell = true
		}
	}

	if !foundBuy || !foundSell {
		return models.OptionStrategy{Name: "Bull Call Spread"}
	}

	netPremium := (buyCall.LTP - sellCall.LTP) * float64(lotSize)
	maxProfit := (sellCall.StrikePrice - buyCall.StrikePrice) * float64(lotSize) - math.Abs(netPremium)
	maxLoss := math.Abs(netPremium)
	breakeven := buyCall.StrikePrice + buyCall.LTP - sellCall.LTP

	strat := models.OptionStrategy{
		Name:       "Bull Call Spread",
		NetPremium: -netPremium, // debit strategy
		MaxProfit:  maxProfit,
		MaxLoss:    maxLoss,
		Breakevens: []float64{breakeven},
		Legs: []models.OptionLeg{
			{OptionType: "CE", StrikePrice: buyCall.StrikePrice, Action: "BUY", Lots: 1, Premium: buyCall.LTP},
			{OptionType: "CE", StrikePrice: sellCall.StrikePrice, Action: "SELL", Lots: 1, Premium: sellCall.LTP},
		},
	}

	// Compute payoff.
	strat.Payoff = computeSpreadPayoff(strat, oc.SpotPrice, lotSize)

	return strat
}

// BuildIronCondor creates an iron condor around the current strike.
func BuildIronCondor(oc *models.OptionChain, lotSize int, width float64) models.OptionStrategy {
	if oc == nil || len(oc.Contracts) == 0 || lotSize <= 0 {
		return models.OptionStrategy{Name: "Iron Condor"}
	}

	if width <= 0 {
		width = 200 // default width in points for NIFTY
	}

	spot := oc.SpotPrice
	atm := findATMStrike(oc.Contracts, spot)

	// Iron Condor: sell OTM put + OTM call, buy further OTM for protection.
	sellPutStrike := findNearestStrike(oc.Contracts, "PE", atm-width)
	buyPutStrike := findNearestStrike(oc.Contracts, "PE", atm-width*2)
	sellCallStrike := findNearestStrike(oc.Contracts, "CE", atm+width)
	buyCallStrike := findNearestStrike(oc.Contracts, "CE", atm+width*2)

	if sellPutStrike == 0 || buyPutStrike == 0 || sellCallStrike == 0 || buyCallStrike == 0 {
		return models.OptionStrategy{Name: "Iron Condor"}
	}

	sellPut := findContract(oc.Contracts, "PE", sellPutStrike)
	buyPut := findContract(oc.Contracts, "PE", buyPutStrike)
	sellCall := findContract(oc.Contracts, "CE", sellCallStrike)
	buyCall := findContract(oc.Contracts, "CE", buyCallStrike)

	netCredit := (sellPut.LTP - buyPut.LTP + sellCall.LTP - buyCall.LTP) * float64(lotSize)
	putWidth := sellPutStrike - buyPutStrike
	callWidth := buyCallStrike - sellCallStrike
	maxWidth := putWidth
	if callWidth > maxWidth {
		maxWidth = callWidth
	}
	maxLoss := maxWidth*float64(lotSize) - netCredit

	return models.OptionStrategy{
		Name:       "Iron Condor",
		NetPremium: netCredit,
		MaxProfit:  netCredit,
		MaxLoss:    maxLoss,
		Breakevens: []float64{sellPutStrike - netCredit/float64(lotSize), sellCallStrike + netCredit/float64(lotSize)},
		Legs: []models.OptionLeg{
			{OptionType: "PE", StrikePrice: buyPutStrike, Action: "BUY", Lots: 1, Premium: buyPut.LTP},
			{OptionType: "PE", StrikePrice: sellPutStrike, Action: "SELL", Lots: 1, Premium: sellPut.LTP},
			{OptionType: "CE", StrikePrice: sellCallStrike, Action: "SELL", Lots: 1, Premium: sellCall.LTP},
			{OptionType: "CE", StrikePrice: buyCallStrike, Action: "BUY", Lots: 1, Premium: buyCall.LTP},
		},
	}
}

// DerivativesAnalysisResult returns a full derivatives analysis.
func FullDerivativesAnalysis(ticker string, oc *models.OptionChain, fut *models.FuturesContract) *models.AnalysisResult {
	if oc == nil {
		return nil
	}

	chainAnalysis := AnalyzeOptionChain(oc)
	pcrAnalysis := ComputePCR(oc)
	oiAnalysis := AnalyzeOIBuildup(oc, fut)

	var signals []models.Signal

	// PCR signal.
	pcrSignal := models.Signal{
		Source:     "PCR",
		Confidence: 0.6,
		Reason:     pcrAnalysis.Interpretation,
	}
	switch pcrAnalysis.Signal {
	case "strongly_bullish", "bullish":
		pcrSignal.Type = models.SignalBuy
	case "bearish", "strongly_bearish":
		pcrSignal.Type = models.SignalSell
	default:
		pcrSignal.Type = models.SignalNeutral
	}
	signals = append(signals, pcrSignal)

	// Max pain signal.
	if oc.MaxPain > 0 && oc.SpotPrice > 0 {
		diff := (oc.SpotPrice - oc.MaxPain) / oc.MaxPain * 100
		mpSignal := models.Signal{
			Source:     "MaxPain",
			Confidence: 0.5,
		}
		if diff > 2 {
			mpSignal.Type = models.SignalSell
			mpSignal.Reason = fmt.Sprintf("Price %.1f%% above max pain (%.0f) — gravitational pull downward", diff, oc.MaxPain)
		} else if diff < -2 {
			mpSignal.Type = models.SignalBuy
			mpSignal.Reason = fmt.Sprintf("Price %.1f%% below max pain (%.0f) — gravitational pull upward", diff, oc.MaxPain)
		} else {
			mpSignal.Type = models.SignalNeutral
			mpSignal.Reason = fmt.Sprintf("Price near max pain (%.0f)", oc.MaxPain)
		}
		signals = append(signals, mpSignal)
	}

	// Futures OI buildup signal.
	if fut != nil {
		oiSignal := models.Signal{
			Source:     "FuturesOI",
			Confidence: 0.6,
			Reason:     oiAnalysis.Interpretation,
		}
		switch oiAnalysis.FuturesBuildup.Buildup {
		case models.LongBuildup:
			oiSignal.Type = models.SignalBuy
		case models.ShortBuildup:
			oiSignal.Type = models.SignalSell
		case models.ShortCovering:
			oiSignal.Type = models.SignalBuy
			oiSignal.Confidence = 0.45
		case models.LongUnwinding:
			oiSignal.Type = models.SignalSell
			oiSignal.Confidence = 0.45
		}
		signals = append(signals, oiSignal)
	}

	// IV skew signal.
	if chainAnalysis.IVSkew > 5 {
		signals = append(signals, models.Signal{
			Source:     "IVSkew",
			Type:       models.SignalSell,
			Confidence: 0.4,
			Reason:     fmt.Sprintf("High IV skew (%.1f) — elevated put demand, hedging activity", chainAnalysis.IVSkew),
		})
	}

	// Aggregate.
	_, conf, rec := aggregateDerivativeSignals(signals)

	details := map[string]any{
		"chain_analysis": chainAnalysis,
		"pcr_analysis":   pcrAnalysis,
		"oi_analysis":    oiAnalysis,
	}

	summary := fmt.Sprintf("Derivatives analysis for %s: PCR %.2f (%s), Max Pain %.0f, %s",
		ticker, pcrAnalysis.PCR, pcrAnalysis.Signal, oc.MaxPain, oiAnalysis.Interpretation)

	return &models.AnalysisResult{
		Ticker:         ticker,
		Type:           models.AnalysisDerivatives,
		AgentName:      "derivatives-analysis",
		Signals:        signals,
		Recommendation: rec,
		Confidence:     conf,
		Summary:        summary,
		Details:        details,
		Timestamp:      time.Now(),
	}
}

// --- helpers ---

func sortBuildups(b []StrikeBuildup) {
	for i := 0; i < len(b); i++ {
		for j := i + 1; j < len(b); j++ {
			if abs64(b[j].OIChange) > abs64(b[i].OIChange) {
				b[i], b[j] = b[j], b[i]
			}
		}
	}
}

func capBuildups(b []StrikeBuildup, max int) []StrikeBuildup {
	if len(b) > max {
		return b[:max]
	}
	return b
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

func findNearestStrike(contracts []models.OptionContract, optType string, target float64) float64 {
	best := 0.0
	bestDiff := math.MaxFloat64

	for _, c := range contracts {
		if c.OptionType == optType {
			diff := math.Abs(c.StrikePrice - target)
			if diff < bestDiff {
				bestDiff = diff
				best = c.StrikePrice
			}
		}
	}

	return best
}

func findContract(contracts []models.OptionContract, optType string, strike float64) models.OptionContract {
	for _, c := range contracts {
		if c.OptionType == optType && c.StrikePrice == strike {
			return c
		}
	}
	return models.OptionContract{}
}

func computeSpreadPayoff(strat models.OptionStrategy, spot float64, lotSize int) []models.OptionPayoff {
	lower := spot * 0.9
	upper := spot * 1.1
	step := (upper - lower) / 50

	var payoff []models.OptionPayoff
	for p := lower; p <= upper; p += step {
		pnl := 0.0
		for _, leg := range strat.Legs {
			intrinsic := 0.0
			if leg.OptionType == "CE" {
				intrinsic = math.Max(0, p-leg.StrikePrice)
			} else {
				intrinsic = math.Max(0, leg.StrikePrice-p)
			}

			mult := float64(lotSize * leg.Lots)
			if leg.Action == "BUY" {
				pnl += (intrinsic - leg.Premium) * mult
			} else {
				pnl += (leg.Premium - intrinsic) * mult
			}
		}
		payoff = append(payoff, models.OptionPayoff{UnderlyingPrice: p, PnL: pnl})
	}

	return payoff
}

func aggregateDerivativeSignals(signals []models.Signal) (models.SignalType, models.Confidence, models.Recommendation) {
	if len(signals) == 0 {
		return models.SignalNeutral, 0, models.Hold
	}

	buyScore, sellScore := 0.0, 0.0
	total := 0.0

	for _, s := range signals {
		w := float64(s.Confidence)
		total += w
		switch s.Type {
		case models.SignalBuy:
			buyScore += w
		case models.SignalSell:
			sellScore += w
		}
	}

	if total == 0 {
		return models.SignalNeutral, 0, models.Hold
	}

	net := (buyScore - sellScore) / total

	switch {
	case net > 0.3:
		return models.SignalBuy, models.Confidence(0.6 + net*0.3), models.ModerateBuy
	case net > 0.1:
		return models.SignalBuy, models.Confidence(0.5), models.ModerateBuy
	case net < -0.3:
		return models.SignalSell, models.Confidence(0.6 + (-net)*0.3), models.ModerateSell
	case net < -0.1:
		return models.SignalSell, models.Confidence(0.5), models.ModerateSell
	default:
		return models.SignalNeutral, models.Confidence(0.4), models.Hold
	}
}
