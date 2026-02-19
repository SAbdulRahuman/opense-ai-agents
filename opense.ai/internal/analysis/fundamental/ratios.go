package fundamental

import (
	"math"

	"github.com/seenimoa/openseai/pkg/models"
)

// ComputeRatios calculates financial ratios from raw financial data and current price.
func ComputeRatios(fin *models.FinancialData, price float64, sharesOutstanding float64) models.FinancialRatios {
	ratios := models.FinancialRatios{}

	if fin == nil || sharesOutstanding <= 0 {
		return ratios
	}

	// Use latest annual income statement.
	if len(fin.AnnualIncome) > 0 {
		latest := fin.AnnualIncome[0]

		// EPS
		if latest.EPS != 0 {
			ratios.EPS = latest.EPS
		} else if sharesOutstanding > 0 {
			ratios.EPS = latest.PAT / sharesOutstanding
		}

		// PE
		if ratios.EPS > 0 && price > 0 {
			ratios.PE = price / ratios.EPS
		}
	}

	// Use latest balance sheet.
	if len(fin.AnnualBalanceSheet) > 0 {
		bs := fin.AnnualBalanceSheet[0]

		// Book Value per share
		if sharesOutstanding > 0 && bs.TotalEquity > 0 {
			ratios.BookValue = bs.TotalEquity / sharesOutstanding
		}

		// PB
		if ratios.BookValue > 0 && price > 0 {
			ratios.PB = price / ratios.BookValue
		}

		// Debt/Equity
		if bs.TotalEquity > 0 {
			ratios.DebtEquity = bs.TotalDebt / bs.TotalEquity
		}

		// Current Ratio
		if bs.CurrentLiabilities > 0 {
			ratios.CurrentRatio = bs.CurrentAssets / bs.CurrentLiabilities
		}

		// ROE = PAT / Total Equity
		if len(fin.AnnualIncome) > 0 && bs.TotalEquity > 0 {
			ratios.ROE = fin.AnnualIncome[0].PAT / bs.TotalEquity * 100
		}

		// ROCE = EBIT / (Total Assets - Current Liabilities)
		if len(fin.AnnualIncome) > 0 {
			capitalEmployed := bs.TotalAssets - bs.CurrentLiabilities
			if capitalEmployed > 0 {
				ratios.ROCE = fin.AnnualIncome[0].EBIT / capitalEmployed * 100
			}
		}

		// Interest Coverage = EBIT / Interest Expense
		if len(fin.AnnualIncome) > 0 && fin.AnnualIncome[0].InterestExpense > 0 {
			ratios.InterestCoverage = fin.AnnualIncome[0].EBIT / fin.AnnualIncome[0].InterestExpense
		}
	}

	// EV/EBITDA
	if len(fin.AnnualIncome) > 0 && len(fin.AnnualBalanceSheet) > 0 {
		ebitda := fin.AnnualIncome[0].EBITDA
		bs := fin.AnnualBalanceSheet[0]
		if ebitda > 0 {
			marketCap := price * sharesOutstanding
			ev := marketCap + bs.TotalDebt - bs.CashEquivalents
			ratios.EVBITDA = ev / ebitda
		}
	}

	// Graham Number = sqrt(22.5 * EPS * BookValue)
	if ratios.EPS > 0 && ratios.BookValue > 0 {
		ratios.GrahamNumber = math.Sqrt(22.5 * ratios.EPS * ratios.BookValue)
	}

	// PEG Ratio = PE / EPS growth rate
	growth := ComputeGrowth(fin)
	if growth.EPSGrowthYoY > 0 && ratios.PE > 0 {
		ratios.PEGRatio = ratios.PE / growth.EPSGrowthYoY
	}

	return ratios
}

// ComputeGrowth calculates growth rates from financial data.
func ComputeGrowth(fin *models.FinancialData) models.GrowthRates {
	g := models.GrowthRates{}
	if fin == nil {
		return g
	}

	// Quarterly growth (QoQ).
	if len(fin.QuarterlyIncome) >= 2 {
		curr := fin.QuarterlyIncome[0]
		prev := fin.QuarterlyIncome[1]
		g.RevenueGrowthQoQ = pctChange(prev.Revenue, curr.Revenue)
		g.ProfitGrowthQoQ = pctChange(prev.PAT, curr.PAT)
	}

	// YoY: compare Q with same Q last year (index 4 back in quarterly).
	if len(fin.QuarterlyIncome) >= 5 {
		curr := fin.QuarterlyIncome[0]
		prev := fin.QuarterlyIncome[4]
		g.RevenueGrowthYoY = pctChange(prev.Revenue, curr.Revenue)
		g.ProfitGrowthYoY = pctChange(prev.PAT, curr.PAT)
	}

	// Annual CAGR.
	if len(fin.AnnualIncome) >= 4 {
		g.RevenueCAGR3Y = cagr(fin.AnnualIncome[3].Revenue, fin.AnnualIncome[0].Revenue, 3)
		g.ProfitCAGR3Y = cagr(fin.AnnualIncome[3].PAT, fin.AnnualIncome[0].PAT, 3)
	}
	if len(fin.AnnualIncome) >= 6 {
		g.RevenueCAGR5Y = cagr(fin.AnnualIncome[5].Revenue, fin.AnnualIncome[0].Revenue, 5)
		g.ProfitCAGR5Y = cagr(fin.AnnualIncome[5].PAT, fin.AnnualIncome[0].PAT, 5)
	}

	// EPS growth.
	if len(fin.AnnualIncome) >= 2 {
		g.EPSGrowthYoY = pctChange(fin.AnnualIncome[1].EPS, fin.AnnualIncome[0].EPS)
	}
	if len(fin.AnnualIncome) >= 4 {
		g.EPSCAGR3Y = cagr(fin.AnnualIncome[3].EPS, fin.AnnualIncome[0].EPS, 3)
	}

	return g
}

// OperatingMetrics computes additional operating metrics.
type OperatingMetrics struct {
	OPM            float64 // Operating Profit Margin %
	NPM            float64 // Net Profit Margin %
	AssetTurnover  float64 // Revenue / Total Assets
	Efficiency     float64 // Revenue / Employees (if available)
	WorkingCapital float64 // Current Assets - Current Liabilities
	FCFE           float64 // Free Cash Flow to Equity
}

// ComputeOperatingMetrics calculates operational efficiency metrics.
func ComputeOperatingMetrics(fin *models.FinancialData) OperatingMetrics {
	m := OperatingMetrics{}
	if fin == nil {
		return m
	}

	if len(fin.AnnualIncome) > 0 {
		inc := fin.AnnualIncome[0]
		if inc.Revenue > 0 {
			m.OPM = inc.OPMPct
			m.NPM = inc.NPMPct
			if m.OPM == 0 && inc.EBITDA > 0 {
				m.OPM = inc.EBITDA / inc.Revenue * 100
			}
			if m.NPM == 0 && inc.PAT != 0 {
				m.NPM = inc.PAT / inc.Revenue * 100
			}
		}
	}

	if len(fin.AnnualBalanceSheet) > 0 && len(fin.AnnualIncome) > 0 {
		bs := fin.AnnualBalanceSheet[0]
		inc := fin.AnnualIncome[0]
		if bs.TotalAssets > 0 {
			m.AssetTurnover = inc.Revenue / bs.TotalAssets
		}
		m.WorkingCapital = bs.CurrentAssets - bs.CurrentLiabilities
	}

	if len(fin.AnnualCashFlow) > 0 {
		cf := fin.AnnualCashFlow[0]
		m.FCFE = cf.FreeCashFlow
	}

	return m
}

// --- helpers ---

func pctChange(old, new_ float64) float64 {
	if old == 0 {
		return 0
	}
	return (new_ - old) / math.Abs(old) * 100
}

func cagr(start, end float64, years float64) float64 {
	if start <= 0 || end <= 0 || years <= 0 {
		return 0
	}
	return (math.Pow(end/start, 1/years) - 1) * 100
}
