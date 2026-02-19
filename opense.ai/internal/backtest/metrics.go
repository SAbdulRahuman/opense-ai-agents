package backtest

import (
	"math"
	"sort"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Performance Metrics
// ════════════════════════════════════════════════════════════════════

// ComputeMetrics computes all performance metrics on a BacktestResult in-place.
// riskFreeRate is annual (e.g. 0.065 for 6.5%).
func ComputeMetrics(r *models.BacktestResult, riskFreeRate float64) {
	if r == nil {
		return
	}

	computeTradeStats(r)
	computeCAGR(r)
	computeDrawdown(r)
	computeSharpe(r, riskFreeRate)
	computeSortino(r, riskFreeRate)
}

// ────────────────────────────────────────────────────────────────────
// Trade statistics
// ────────────────────────────────────────────────────────────────────

func computeTradeStats(r *models.BacktestResult) {
	r.TotalTrades = len(r.Trades)
	if r.TotalTrades == 0 {
		return
	}

	var totalWin, totalLoss float64
	for _, t := range r.Trades {
		if t.PnL > 0 {
			r.WinningTrades++
			totalWin += t.PnL
		} else if t.PnL < 0 {
			r.LosingTrades++
			totalLoss += math.Abs(t.PnL)
		}
	}

	r.WinRate = float64(r.WinningTrades) / float64(r.TotalTrades) * 100

	if r.WinningTrades > 0 {
		r.AvgWin = totalWin / float64(r.WinningTrades)
	}
	if r.LosingTrades > 0 {
		r.AvgLoss = totalLoss / float64(r.LosingTrades)
	}

	if totalLoss > 0 {
		r.ProfitFactor = totalWin / totalLoss
	} else if totalWin > 0 {
		r.ProfitFactor = math.Inf(1)
	}
}

// ────────────────────────────────────────────────────────────────────
// CAGR — Compound Annual Growth Rate
// ────────────────────────────────────────────────────────────────────

func computeCAGR(r *models.BacktestResult) {
	if r.InitialCapital <= 0 || r.FinalCapital <= 0 {
		return
	}

	days := r.To.Sub(r.From).Hours() / 24
	if days <= 0 {
		return
	}
	years := days / 365.25

	r.CAGR = (math.Pow(r.FinalCapital/r.InitialCapital, 1.0/years) - 1) * 100
}

// ────────────────────────────────────────────────────────────────────
// Maximum Drawdown
// ────────────────────────────────────────────────────────────────────

func computeDrawdown(r *models.BacktestResult) {
	if len(r.EquityCurve) == 0 {
		return
	}

	peak := r.EquityCurve[0].Value
	maxDD := 0.0
	maxDDPct := 0.0

	for _, ep := range r.EquityCurve {
		if ep.Value > peak {
			peak = ep.Value
		}
		dd := peak - ep.Value
		ddPct := 0.0
		if peak > 0 {
			ddPct = (dd / peak) * 100
		}
		if dd > maxDD {
			maxDD = dd
		}
		if ddPct > maxDDPct {
			maxDDPct = ddPct
		}
	}

	r.MaxDrawdown = maxDD
	r.MaxDrawdownPct = maxDDPct
}

// ────────────────────────────────────────────────────────────────────
// Sharpe Ratio (annualized)
// ────────────────────────────────────────────────────────────────────

func computeSharpe(r *models.BacktestResult, riskFreeRate float64) {
	returns := dailyReturns(r.EquityCurve)
	if len(returns) < 2 {
		return
	}

	dailyRf := riskFreeRate / 252 // trading days
	excessReturns := make([]float64, len(returns))
	for i, ret := range returns {
		excessReturns[i] = ret - dailyRf
	}

	mean := mean(excessReturns)
	sd := stddev(excessReturns)

	if sd > 0 {
		r.SharpeRatio = (mean / sd) * math.Sqrt(252) // annualize
	}
}

// ────────────────────────────────────────────────────────────────────
// Sortino Ratio (annualized, downside deviation only)
// ────────────────────────────────────────────────────────────────────

func computeSortino(r *models.BacktestResult, riskFreeRate float64) {
	returns := dailyReturns(r.EquityCurve)
	if len(returns) < 2 {
		return
	}

	dailyRf := riskFreeRate / 252
	excessReturns := make([]float64, len(returns))
	for i, ret := range returns {
		excessReturns[i] = ret - dailyRf
	}

	meanExcess := mean(excessReturns)

	// Downside deviation: only negative excess returns
	var downsideSqSum float64
	var downsideCount int
	for _, er := range excessReturns {
		if er < 0 {
			downsideSqSum += er * er
			downsideCount++
		}
	}

	if downsideCount > 0 {
		downsideDev := math.Sqrt(downsideSqSum / float64(len(excessReturns)))
		if downsideDev > 0 {
			r.SortinoRatio = (meanExcess / downsideDev) * math.Sqrt(252)
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// Helpers
// ════════════════════════════════════════════════════════════════════

// dailyReturns computes simple returns from equity curve.
func dailyReturns(curve []models.EquityPoint) []float64 {
	if len(curve) < 2 {
		return nil
	}
	returns := make([]float64, len(curve)-1)
	for i := 1; i < len(curve); i++ {
		if curve[i-1].Value > 0 {
			returns[i-1] = (curve[i].Value - curve[i-1].Value) / curve[i-1].Value
		}
	}
	return returns
}

func mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func stddev(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}
	m := mean(data)
	sumSq := 0.0
	for _, v := range data {
		d := v - m
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(data)-1)) // sample stddev
}

// ════════════════════════════════════════════════════════════════════
// Analysis Utilities
// ════════════════════════════════════════════════════════════════════

// MaxConsecutiveWins returns the longest winning streak.
func MaxConsecutiveWins(trades []models.BacktestTrade) int {
	max, current := 0, 0
	for _, t := range trades {
		if t.PnL > 0 {
			current++
			if current > max {
				max = current
			}
		} else {
			current = 0
		}
	}
	return max
}

// MaxConsecutiveLosses returns the longest losing streak.
func MaxConsecutiveLosses(trades []models.BacktestTrade) int {
	max, current := 0, 0
	for _, t := range trades {
		if t.PnL < 0 {
			current++
			if current > max {
				max = current
			}
		} else {
			current = 0
		}
	}
	return max
}

// ExpectancyPerTrade returns the average PnL per trade.
func ExpectancyPerTrade(trades []models.BacktestTrade) float64 {
	if len(trades) == 0 {
		return 0
	}
	sum := 0.0
	for _, t := range trades {
		sum += t.PnL
	}
	return sum / float64(len(trades))
}

// MedianTradePnL returns the median P&L across trades.
func MedianTradePnL(trades []models.BacktestTrade) float64 {
	if len(trades) == 0 {
		return 0
	}
	pnls := make([]float64, len(trades))
	for i, t := range trades {
		pnls[i] = t.PnL
	}
	sort.Float64s(pnls)
	n := len(pnls)
	if n%2 == 0 {
		return (pnls[n/2-1] + pnls[n/2]) / 2
	}
	return pnls[n/2]
}

// AverageHoldingPeriod returns the mean holding period in days.
func AverageHoldingPeriod(trades []models.BacktestTrade) float64 {
	if len(trades) == 0 {
		return 0
	}
	totalDays := 0.0
	for _, t := range trades {
		totalDays += t.ExitDate.Sub(t.EntryDate).Hours() / 24
	}
	return totalDays / float64(len(trades))
}
