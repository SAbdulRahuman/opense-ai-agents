package fundamental

import (
	"fmt"
	"strings"

	"github.com/seenimoa/openseai/pkg/models"
)

// FinancialHealth scores the overall financial robustness of a company.
type FinancialHealth struct {
	Score       float64            // 0-100 composite score
	Grade       string             // "A+", "A", "B+", "B", "C", "D"
	Strengths   []string           // positive factors
	Weaknesses  []string           // negative factors
	Components  map[string]float64 // individual component scores
}

// AssessFinancialHealth evaluates financial health from ratios and financial data.
func AssessFinancialHealth(ratios models.FinancialRatios, fin *models.FinancialData) FinancialHealth {
	h := FinancialHealth{
		Components: make(map[string]float64),
	}

	totalScore := 0.0
	totalWeight := 0.0

	// Profitability (30 points).
	profScore := 0.0
	if ratios.ROE > 20 {
		profScore += 10
		h.Strengths = append(h.Strengths, fmt.Sprintf("High ROE: %.1f%%", ratios.ROE))
	} else if ratios.ROE > 12 {
		profScore += 6
	} else if ratios.ROE > 0 {
		profScore += 3
	} else {
		h.Weaknesses = append(h.Weaknesses, "Negative or zero ROE")
	}

	if ratios.ROCE > 20 {
		profScore += 10
		h.Strengths = append(h.Strengths, fmt.Sprintf("High ROCE: %.1f%%", ratios.ROCE))
	} else if ratios.ROCE > 12 {
		profScore += 6
	} else if ratios.ROCE > 0 {
		profScore += 3
	}

	metrics := ComputeOperatingMetrics(fin)
	if metrics.OPM > 20 {
		profScore += 10
		h.Strengths = append(h.Strengths, fmt.Sprintf("Strong OPM: %.1f%%", metrics.OPM))
	} else if metrics.OPM > 10 {
		profScore += 6
	} else if metrics.OPM > 0 {
		profScore += 3
	} else {
		h.Weaknesses = append(h.Weaknesses, "Negative operating margin")
	}

	h.Components["profitability"] = profScore
	totalScore += profScore
	totalWeight += 30

	// Solvency (25 points).
	solvScore := 0.0
	if ratios.DebtEquity < 0.5 {
		solvScore += 12.5
		h.Strengths = append(h.Strengths, "Low debt-to-equity ratio")
	} else if ratios.DebtEquity < 1 {
		solvScore += 8
	} else if ratios.DebtEquity < 2 {
		solvScore += 4
	} else {
		h.Weaknesses = append(h.Weaknesses, fmt.Sprintf("High D/E ratio: %.2f", ratios.DebtEquity))
	}

	if ratios.InterestCoverage > 5 {
		solvScore += 12.5
	} else if ratios.InterestCoverage > 2 {
		solvScore += 8
	} else if ratios.InterestCoverage > 1 {
		solvScore += 4
	} else if ratios.InterestCoverage > 0 {
		solvScore += 2
		h.Weaknesses = append(h.Weaknesses, "Low interest coverage")
	}

	h.Components["solvency"] = solvScore
	totalScore += solvScore
	totalWeight += 25

	// Liquidity (15 points).
	liqScore := 0.0
	if ratios.CurrentRatio > 2 {
		liqScore += 15
		h.Strengths = append(h.Strengths, "Strong current ratio")
	} else if ratios.CurrentRatio > 1.5 {
		liqScore += 12
	} else if ratios.CurrentRatio > 1 {
		liqScore += 7
	} else {
		h.Weaknesses = append(h.Weaknesses, fmt.Sprintf("Weak current ratio: %.2f", ratios.CurrentRatio))
	}

	h.Components["liquidity"] = liqScore
	totalScore += liqScore
	totalWeight += 15

	// Growth (20 points).
	growthScore := 0.0
	growth := ComputeGrowth(fin)
	if growth.RevenueGrowthYoY > 20 {
		growthScore += 10
		h.Strengths = append(h.Strengths, fmt.Sprintf("Strong revenue growth: %.1f%% YoY", growth.RevenueGrowthYoY))
	} else if growth.RevenueGrowthYoY > 10 {
		growthScore += 6
	} else if growth.RevenueGrowthYoY > 0 {
		growthScore += 3
	} else {
		h.Weaknesses = append(h.Weaknesses, "Declining revenue")
	}

	if growth.ProfitGrowthYoY > 20 {
		growthScore += 10
	} else if growth.ProfitGrowthYoY > 10 {
		growthScore += 6
	} else if growth.ProfitGrowthYoY > 0 {
		growthScore += 3
	} else {
		h.Weaknesses = append(h.Weaknesses, "Declining profits")
	}

	h.Components["growth"] = growthScore
	totalScore += growthScore
	totalWeight += 20

	// Cash Flow (10 points).
	cfScore := 0.0
	if metrics.FCFE > 0 {
		cfScore += 10
		h.Strengths = append(h.Strengths, "Positive free cash flow")
	} else if len(fin.AnnualCashFlow) > 0 && fin.AnnualCashFlow[0].OperatingCashFlow > 0 {
		cfScore += 5
	} else {
		h.Weaknesses = append(h.Weaknesses, "Negative operating cash flow")
	}

	h.Components["cash_flow"] = cfScore
	totalScore += cfScore
	totalWeight += 10

	// Compute final score (0-100).
	if totalWeight > 0 {
		h.Score = totalScore / totalWeight * 100
	}

	// Grade.
	switch {
	case h.Score >= 85:
		h.Grade = "A+"
	case h.Score >= 70:
		h.Grade = "A"
	case h.Score >= 55:
		h.Grade = "B+"
	case h.Score >= 40:
		h.Grade = "B"
	case h.Score >= 25:
		h.Grade = "C"
	default:
		h.Grade = "D"
	}

	return h
}

// QualityCheck performs a quick quality screen (Piotroski-style F-score simplified).
type QualityScore struct {
	Score  int      // 0-9
	Checks []string // descriptions of each check passed/failed
}

// PiotroskiFScore computes a simplified Piotroski F-Score.
func PiotroskiFScore(fin *models.FinancialData) QualityScore {
	qs := QualityScore{}
	if fin == nil || len(fin.AnnualIncome) < 2 || len(fin.AnnualBalanceSheet) < 2 {
		return qs
	}

	curr := fin.AnnualIncome[0]
	prev := fin.AnnualIncome[1]
	currBS := fin.AnnualBalanceSheet[0]
	prevBS := fin.AnnualBalanceSheet[1]

	// 1. Positive PAT.
	if curr.PAT > 0 {
		qs.Score++
		qs.Checks = append(qs.Checks, "✓ Positive net income")
	} else {
		qs.Checks = append(qs.Checks, "✗ Negative net income")
	}

	// 2. Positive operating cash flow.
	if len(fin.AnnualCashFlow) > 0 && fin.AnnualCashFlow[0].OperatingCashFlow > 0 {
		qs.Score++
		qs.Checks = append(qs.Checks, "✓ Positive operating cash flow")
	} else {
		qs.Checks = append(qs.Checks, "✗ Negative operating cash flow")
	}

	// 3. ROA improving.
	if currBS.TotalAssets > 0 && prevBS.TotalAssets > 0 {
		currROA := curr.PAT / currBS.TotalAssets
		prevROA := prev.PAT / prevBS.TotalAssets
		if currROA > prevROA {
			qs.Score++
			qs.Checks = append(qs.Checks, "✓ Improving ROA")
		} else {
			qs.Checks = append(qs.Checks, "✗ Declining ROA")
		}
	}

	// 4. OCF > PAT (accrual check).
	if len(fin.AnnualCashFlow) > 0 && fin.AnnualCashFlow[0].OperatingCashFlow > curr.PAT {
		qs.Score++
		qs.Checks = append(qs.Checks, "✓ Cash flow > Net income (quality earnings)")
	} else {
		qs.Checks = append(qs.Checks, "✗ Cash flow < Net income")
	}

	// 5. Declining leverage.
	if currBS.TotalEquity > 0 && prevBS.TotalEquity > 0 {
		currDE := currBS.TotalDebt / currBS.TotalEquity
		prevDE := prevBS.TotalDebt / prevBS.TotalEquity
		if currDE < prevDE {
			qs.Score++
			qs.Checks = append(qs.Checks, "✓ Declining leverage")
		} else {
			qs.Checks = append(qs.Checks, "✗ Increasing leverage")
		}
	}

	// 6. Improving current ratio.
	if currBS.CurrentLiabilities > 0 && prevBS.CurrentLiabilities > 0 {
		currCR := currBS.CurrentAssets / currBS.CurrentLiabilities
		prevCR := prevBS.CurrentAssets / prevBS.CurrentLiabilities
		if currCR > prevCR {
			qs.Score++
			qs.Checks = append(qs.Checks, "✓ Improving current ratio")
		} else {
			qs.Checks = append(qs.Checks, "✗ Declining current ratio")
		}
	}

	// 7. No dilution (share capital not increased).
	if currBS.ShareCapital <= prevBS.ShareCapital {
		qs.Score++
		qs.Checks = append(qs.Checks, "✓ No equity dilution")
	} else {
		qs.Checks = append(qs.Checks, "✗ Equity diluted")
	}

	// 8. Improving gross margin.
	if curr.OPMPct > prev.OPMPct {
		qs.Score++
		qs.Checks = append(qs.Checks, "✓ Improving operating margin")
	} else {
		qs.Checks = append(qs.Checks, "✗ Declining operating margin")
	}

	// 9. Improving asset turnover.
	if currBS.TotalAssets > 0 && prevBS.TotalAssets > 0 {
		currAT := curr.Revenue / currBS.TotalAssets
		prevAT := prev.Revenue / prevBS.TotalAssets
		if currAT > prevAT {
			qs.Score++
			qs.Checks = append(qs.Checks, "✓ Improving asset turnover")
		} else {
			qs.Checks = append(qs.Checks, "✗ Declining asset turnover")
		}
	}

	return qs
}

// FormatFinancialSummary generates a readable summary of key financials.
func FormatFinancialSummary(ratios models.FinancialRatios, growth models.GrowthRates, health FinancialHealth) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Financial Health: %s (%.0f/100)\n", health.Grade, health.Score))
	b.WriteString(fmt.Sprintf("PE: %.1f | PB: %.1f | EV/EBITDA: %.1f\n", ratios.PE, ratios.PB, ratios.EVBITDA))
	b.WriteString(fmt.Sprintf("ROE: %.1f%% | ROCE: %.1f%% | D/E: %.2f\n", ratios.ROE, ratios.ROCE, ratios.DebtEquity))
	b.WriteString(fmt.Sprintf("Revenue Growth YoY: %.1f%% | Profit Growth YoY: %.1f%%\n", growth.RevenueGrowthYoY, growth.ProfitGrowthYoY))

	if ratios.GrahamNumber > 0 {
		b.WriteString(fmt.Sprintf("Graham Number: ₹%.2f\n", ratios.GrahamNumber))
	}

	if len(health.Strengths) > 0 {
		b.WriteString("Strengths: ")
		b.WriteString(strings.Join(health.Strengths, "; "))
		b.WriteString("\n")
	}
	if len(health.Weaknesses) > 0 {
		b.WriteString("Weaknesses: ")
		b.WriteString(strings.Join(health.Weaknesses, "; "))
		b.WriteString("\n")
	}

	return b.String()
}
