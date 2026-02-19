package fundamental

import (
	"math"

	"github.com/seenimoa/openseai/pkg/models"
)

// ValuationResult contains multiple valuation estimates.
type ValuationResult struct {
	Ticker          string             `json:"ticker"`
	CurrentPrice    float64            `json:"current_price"`
	GrahamNumber    float64            `json:"graham_number"`
	DCFValue        float64            `json:"dcf_value"`
	PERelative      float64            `json:"pe_relative"` // fair value based on relative PE
	EarningsYield   float64            `json:"earnings_yield"`
	MarginOfSafety  float64            `json:"margin_of_safety"` // % below intrinsic value
	Verdict         string             `json:"verdict"`          // "Undervalued", "Fairly Valued", "Overvalued"
	Methods         map[string]float64 `json:"methods"`          // method → estimated fair value
}

// DCFParams holds parameters for a DCF valuation.
type DCFParams struct {
	FreeCashFlow     float64 // most recent annual FCF
	GrowthRateYr1_5  float64 // expected FCF growth years 1-5 (decimal, e.g., 0.15)
	GrowthRateYr6_10 float64 // terminal growth years 6-10
	TerminalGrowth   float64 // perpetual growth rate (typically 0.02-0.04 for India)
	DiscountRate     float64 // WACC or required return (decimal, e.g., 0.12)
	SharesOutstanding float64
}

// DCF performs a two-stage discounted cash flow valuation.
func DCF(p DCFParams) float64 {
	if p.FreeCashFlow <= 0 || p.DiscountRate <= 0 || p.SharesOutstanding <= 0 {
		return 0
	}

	if p.GrowthRateYr1_5 <= 0 {
		p.GrowthRateYr1_5 = 0.10
	}
	if p.GrowthRateYr6_10 <= 0 {
		p.GrowthRateYr6_10 = 0.06
	}
	if p.TerminalGrowth <= 0 {
		p.TerminalGrowth = 0.03
	}

	totalPV := 0.0
	fcf := p.FreeCashFlow

	// Years 1-5: high growth.
	for i := 1; i <= 5; i++ {
		fcf *= (1 + p.GrowthRateYr1_5)
		pv := fcf / math.Pow(1+p.DiscountRate, float64(i))
		totalPV += pv
	}

	// Years 6-10: moderate growth.
	for i := 6; i <= 10; i++ {
		fcf *= (1 + p.GrowthRateYr6_10)
		pv := fcf / math.Pow(1+p.DiscountRate, float64(i))
		totalPV += pv
	}

	// Terminal value using Gordon Growth Model.
	terminalFCF := fcf * (1 + p.TerminalGrowth)
	terminalValue := terminalFCF / (p.DiscountRate - p.TerminalGrowth)
	pvTerminal := terminalValue / math.Pow(1+p.DiscountRate, 10)

	totalPV += pvTerminal

	return totalPV / p.SharesOutstanding
}

// GrahamNumber computes the classic Benjamin Graham intrinsic value.
// Graham Number = sqrt(22.5 × EPS × Book Value per Share)
func GrahamNumber(eps, bookValue float64) float64 {
	if eps <= 0 || bookValue <= 0 {
		return 0
	}
	return math.Sqrt(22.5 * eps * bookValue)
}

// PERelativeValue estimates fair value using sector/peer average PE.
func PERelativeValue(eps, sectorPE float64) float64 {
	if eps <= 0 || sectorPE <= 0 {
		return 0
	}
	return eps * sectorPE
}

// EarningsYield computes earnings yield (inverse of PE).
func EarningsYield(eps, price float64) float64 {
	if price <= 0 {
		return 0
	}
	return eps / price * 100
}

// ComputeValuation runs all valuation methods and provides a verdict.
func ComputeValuation(ticker string, price float64, ratios models.FinancialRatios, fin *models.FinancialData, sharesOutstanding float64, sectorPE float64) ValuationResult {
	v := ValuationResult{
		Ticker:       ticker,
		CurrentPrice: price,
		Methods:      make(map[string]float64),
	}

	// Graham Number.
	gn := GrahamNumber(ratios.EPS, ratios.BookValue)
	if gn > 0 {
		v.GrahamNumber = gn
		v.Methods["graham"] = gn
	}

	// DCF.
	if fin != nil && len(fin.AnnualCashFlow) > 0 && sharesOutstanding > 0 {
		growth := ComputeGrowth(fin)
		growthRate := growth.RevenueCAGR3Y / 100
		if growthRate <= 0 {
			growthRate = 0.10
		}

		dcfVal := DCF(DCFParams{
			FreeCashFlow:      fin.AnnualCashFlow[0].FreeCashFlow,
			GrowthRateYr1_5:   growthRate,
			GrowthRateYr6_10:  growthRate * 0.5,
			TerminalGrowth:    0.03,
			DiscountRate:      0.12,
			SharesOutstanding: sharesOutstanding,
		})
		if dcfVal > 0 {
			v.DCFValue = dcfVal
			v.Methods["dcf"] = dcfVal
		}
	}

	// PE Relative.
	if sectorPE > 0 {
		peRel := PERelativeValue(ratios.EPS, sectorPE)
		if peRel > 0 {
			v.PERelative = peRel
			v.Methods["pe_relative"] = peRel
		}
	}

	// Earnings yield.
	v.EarningsYield = EarningsYield(ratios.EPS, price)

	// Average intrinsic value from all methods.
	var sum float64
	var count int
	for _, val := range v.Methods {
		if val > 0 {
			sum += val
			count++
		}
	}

	if count > 0 && price > 0 {
		avgIntrinsic := sum / float64(count)
		v.MarginOfSafety = (avgIntrinsic - price) / avgIntrinsic * 100

		switch {
		case v.MarginOfSafety > 25:
			v.Verdict = "Undervalued"
		case v.MarginOfSafety > -10:
			v.Verdict = "Fairly Valued"
		default:
			v.Verdict = "Overvalued"
		}
	}

	return v
}
