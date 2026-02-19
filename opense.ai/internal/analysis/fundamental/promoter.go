package fundamental

import (
	"fmt"
	"math"

	"github.com/seenimoa/openseai/pkg/models"
)

// PromoterAnalysis contains insights from promoter holding data.
type PromoterAnalysis struct {
	CurrentHolding  float64  `json:"current_holding"`
	PledgePct       float64  `json:"pledge_pct"`
	TrendDirection  string   `json:"trend_direction"` // "increasing", "decreasing", "stable"
	TrendChangePct  float64  `json:"trend_change_pct"` // change over available history
	Signals         []string `json:"signals"`          // notable observations
	RiskLevel       string   `json:"risk_level"`       // "low", "medium", "high"
}

// AnalyzePromoterHolding evaluates promoter-related signals.
func AnalyzePromoterHolding(p *models.PromoterData) PromoterAnalysis {
	if p == nil {
		return PromoterAnalysis{RiskLevel: "unknown"}
	}

	a := PromoterAnalysis{
		CurrentHolding: p.PromoterHolding,
		PledgePct:      p.PromoterPledge,
	}

	// Pledge risk assessment.
	switch {
	case p.PromoterPledge > 50:
		a.RiskLevel = "high"
		a.Signals = append(a.Signals, fmt.Sprintf("⚠ Very high promoter pledge: %.1f%%", p.PromoterPledge))
	case p.PromoterPledge > 20:
		a.RiskLevel = "medium"
		a.Signals = append(a.Signals, fmt.Sprintf("Elevated promoter pledge: %.1f%%", p.PromoterPledge))
	case p.PromoterPledge > 0:
		a.RiskLevel = "low"
		a.Signals = append(a.Signals, fmt.Sprintf("Minimal promoter pledge: %.1f%%", p.PromoterPledge))
	default:
		a.RiskLevel = "low"
		a.Signals = append(a.Signals, "No promoter pledge — positive")
	}

	// Holding level assessment.
	switch {
	case p.PromoterHolding > 70:
		a.Signals = append(a.Signals, "High promoter holding indicates strong conviction")
	case p.PromoterHolding > 50:
		a.Signals = append(a.Signals, "Majority promoter holding")
	case p.PromoterHolding > 30:
		a.Signals = append(a.Signals, "Moderate promoter holding")
	case p.PromoterHolding > 0:
		a.Signals = append(a.Signals, "Low promoter holding — watch for potential governance issues")
	}

	// Institutional interest.
	totalInst := p.FIIHolding + p.DIIHolding
	if totalInst > 40 {
		a.Signals = append(a.Signals, fmt.Sprintf("Strong institutional interest: FII %.1f%% + DII %.1f%%", p.FIIHolding, p.DIIHolding))
	}

	if p.MFHolding > 10 {
		a.Signals = append(a.Signals, fmt.Sprintf("Good mutual fund holding: %.1f%%", p.MFHolding))
	}

	// Trend analysis.
	if len(p.PromoterTrend) >= 2 {
		latest := p.PromoterTrend[0].Pct
		oldest := p.PromoterTrend[len(p.PromoterTrend)-1].Pct
		change := latest - oldest

		a.TrendChangePct = change

		switch {
		case math.Abs(change) < 1:
			a.TrendDirection = "stable"
			a.Signals = append(a.Signals, "Promoter holding stable over recent quarters")
		case change > 0:
			a.TrendDirection = "increasing"
			a.Signals = append(a.Signals, fmt.Sprintf("Promoter buying: holding up %.1f%% — bullish", change))
		default:
			a.TrendDirection = "decreasing"
			a.Signals = append(a.Signals, fmt.Sprintf("Promoter selling: holding down %.1f%%", change))
			if change < -5 {
				a.RiskLevel = "high"
				a.Signals = append(a.Signals, "⚠ Significant promoter selling — red flag")
			}
		}
	}

	return a
}

// PromoterSignal generates a trading signal from promoter holding analysis.
func PromoterSignal(p *models.PromoterData) models.Signal {
	a := AnalyzePromoterHolding(p)

	sig := models.Signal{
		Source: "Promoter",
	}

	switch {
	case a.RiskLevel == "high":
		sig.Type = models.SignalSell
		sig.Confidence = 0.5
		sig.Reason = "High promoter risk: " + a.Signals[0]
	case a.TrendDirection == "increasing" && a.PledgePct < 5:
		sig.Type = models.SignalBuy
		sig.Confidence = 0.55
		sig.Reason = "Promoter increasing stake with low pledge"
	case a.TrendDirection == "decreasing":
		sig.Type = models.SignalSell
		sig.Confidence = 0.45
		sig.Reason = fmt.Sprintf("Promoter reducing stake by %.1f%%", math.Abs(a.TrendChangePct))
	default:
		sig.Type = models.SignalNeutral
		sig.Confidence = 0.3
		sig.Reason = "Promoter holding stable"
	}

	return sig
}
