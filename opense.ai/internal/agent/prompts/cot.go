package prompts

import "fmt"

// ── Chain-of-Thought Templates ──
//
// These templates guide agents to think step-by-step through financial analysis,
// producing more reliable and explainable results.

// CoTAnalysis wraps a task with a chain-of-thought reasoning template.
func CoTAnalysis(ticker, task string) string {
	return fmt.Sprintf(`Analyze %s: %s

Think step-by-step:

1. **Data Gathering**: What data do I need? Use tools to fetch real numbers.
2. **Key Metrics**: What are the most important metrics for this analysis?
3. **Interpretation**: What do these numbers tell us? Are they above/below normal ranges?
4. **Context**: How does this compare to the sector, market, and historical averages?
5. **Conclusion**: Based on the evidence, what is the clear signal?

Important: Always fetch real data using your tools before drawing conclusions. Never estimate or fabricate numbers.`, ticker, task)
}

// CoTFundamental is a chain-of-thought template for fundamental analysis.
func CoTFundamental(ticker string) string {
	return fmt.Sprintf(`Perform a comprehensive fundamental analysis of %s.

Think step-by-step:

**Step 1 — Financial Health**
- Fetch the latest financial data (income statement, balance sheet, cash flow)
- Calculate key ratios: PE, PB, ROE, ROCE, D/E, Current Ratio
- Is the company profitable? Is profitability improving?

**Step 2 — Growth Assessment**
- Revenue growth: QoQ, YoY, 3Y CAGR
- Profit growth: QoQ, YoY, 3Y CAGR
- Is growth accelerating or decelerating?

**Step 3 — Valuation**
- Current PE vs 5-year average PE
- PEG ratio (PE / earnings growth)
- Estimated intrinsic value using DCF or relative valuation
- Is the stock overvalued, fairly valued, or undervalued?

**Step 4 — Ownership & Governance**
- Promoter holding trend (increasing = positive, decreasing = negative)
- Pledge percentage (high pledge = red flag)
- FII/DII holding changes

**Step 5 — Peer Comparison**
- Compare key ratios against 3-5 sector peers
- Where does this company rank in its peer group?

**Step 6 — Risk Assessment**
- Identify top 3 risks (debt, sector headwinds, regulatory, etc.)
- Any red flags from auditor reports or corporate governance?

**Step 7 — Final Verdict**
- Synthesize all evidence into a clear BUY/SELL/HOLD
- Set target price and stop-loss based on valuation
- Assign confidence level (0-100%%)`, ticker)
}

// CoTTechnical is a chain-of-thought template for technical analysis.
func CoTTechnical(ticker string) string {
	return fmt.Sprintf(`Perform a comprehensive technical analysis of %s.

Think step-by-step:

**Step 1 — Trend Identification**
- Fetch historical price data (at least 200 candles for daily charts)
- Determine the primary trend using SMA(50) vs SMA(200)
- Golden Cross (50 > 200) = bullish, Death Cross (50 < 200) = bearish
- Is the stock above or below its key moving averages?

**Step 2 — Momentum Indicators**
- RSI(14): Is the stock overbought (>70) or oversold (<30)?
- MACD: Is the MACD line above/below signal line? Histogram expanding or contracting?
- Is momentum confirming the trend or diverging?

**Step 3 — Volatility Assessment**
- Bollinger Bands: Is price near upper band (resistance) or lower band (support)?
- ATR(14): What is the current volatility? Use for stop-loss calculation
- SuperTrend: What direction is SuperTrend indicating?

**Step 4 — Support & Resistance**
- Calculate pivot points and Fibonacci levels
- Identify key support and resistance from price action
- Where is the stock relative to these levels?

**Step 5 — Pattern Recognition**
- Any active candlestick patterns (doji, hammer, engulfing)?
- Any chart patterns (head & shoulders, double top/bottom)?

**Step 6 — Volume Analysis**
- Is volume confirming the price move?
- Any unusual volume spikes?

**Step 7 — Trade Setup**
- Entry price, target price, stop-loss (ATR-based)
- Risk-reward ratio
- BUY/SELL/NEUTRAL signal with confidence`, ticker)
}

// CoTDerivatives is a chain-of-thought template for F&O analysis.
func CoTDerivatives(ticker string) string {
	return fmt.Sprintf(`Perform a comprehensive derivatives analysis of %s.

Think step-by-step:

**Step 1 — Option Chain Analysis**
- Fetch the current option chain
- Identify strikes with highest Call OI and Put OI
- These represent key resistance and support levels from OI perspective

**Step 2 — Max Pain Calculation**
- Calculate the max pain level
- The market tends to expire near max pain (option seller advantage)

**Step 3 — PCR Analysis**
- Current PCR value and trend (last 5 sessions)
- PCR > 1.0 suggests bearish positioning (but can be supportive for spot)
- PCR < 0.7 suggests bullish positioning (but could mean complacency)

**Step 4 — OI Buildup Analysis**
- Classify the current buildup: Long Buildup, Short Buildup, Long Unwinding, Short Covering
- Price ↑ + OI ↑ = Long Buildup (Bullish)
- Price ↓ + OI ↑ = Short Buildup (Bearish)
- Price ↓ + OI ↓ = Long Unwinding (Bearish)
- Price ↑ + OI ↓ = Short Covering (Neutral to Bullish)

**Step 5 — Futures Analysis**
- Futures premium/discount to spot (basis)
- Cost of carry (positive = bullish, negative = bearish)
- Rollover percentage if near expiry

**Step 6 — Strategy Suggestion**
- Based on the analysis, suggest an appropriate option strategy
- Include specific strikes, expiry, expected payoff, max loss, breakeven

**Step 7 — Conclusion**
- Bullish/Bearish/Neutral derivatives outlook with confidence`, ticker)
}

// CoTRisk is a chain-of-thought template for risk assessment.
func CoTRisk(ticker string, capitalINR float64) string {
	return fmt.Sprintf(`Perform a risk assessment for a potential trade in %s with capital ₹%.0f.

Think step-by-step:

**Step 1 — Volatility Assessment**
- Check India VIX level (>15 = elevated, >20 = high volatility)
- Calculate stock-specific volatility using ATR(14) and daily returns
- Compare to sector average volatility

**Step 2 — Position Sizing**
- Apply 5%% max position size rule: max ₹%.0f in this stock
- Calculate ATR-based position size: capital_at_risk / ATR
- Choose the more conservative of the two

**Step 3 — Stop-Loss Calculation**
- ATR-based stop-loss: entry price - (2 × ATR)
- Support-level stop-loss: just below nearest support
- Choose the tighter of the two (capital preservation first)

**Step 4 — Risk-Reward Evaluation**
- Calculate risk (entry - stop) and reward (target - entry)
- R:R must be at least 1:2 to approve
- Factor in brokerage costs for realistic returns

**Step 5 — Portfolio Impact**
- Check sector concentration (max 25%% in one sector)
- Check correlation with existing holdings
- Assess overall portfolio beta after this addition

**Step 6 — Decision**
- Approve/Reject/Modify the trade with specific conditions
- If approved: exact quantity, entry range, stop-loss, targets`, ticker, capitalINR, capitalINR*0.05)
}

// CoTSynthesis is a chain-of-thought template for the CIO synthesizing all analyses.
func CoTSynthesis(ticker string) string {
	return fmt.Sprintf(`Synthesize all analyst reports to form a final investment view on %s.

Think step-by-step:

**Step 1 — Review Each Analyst's Finding**
- Fundamental: What is the business quality and valuation?
- Technical: What is the price trend and timing?
- Sentiment: What is the market mood?
- Derivatives: What is smart money signaling?
- Risk: Is the risk acceptable?

**Step 2 — Identify Agreement & Conflict**
- Where do analysts agree? (High-conviction signal)
- Where do they disagree? (Requires judgment call)
- Any strong contrarian signals?

**Step 3 — Weight the Evidence**
- For long-term investment: Fundamental 40%%, Technical 20%%, Sentiment 15%%, Derivatives 15%%, Risk 10%%
- For short-term trading: Technical 35%%, Derivatives 25%%, Sentiment 20%%, Fundamental 10%%, Risk 10%%
- Adjust based on current market regime

**Step 4 — Form Final View**
- Overall: BUY / SELL / HOLD
- Conviction: HIGH / MEDIUM / LOW
- Timeframe: Short / Medium / Long term

**Step 5 — Set Parameters**
- Entry price or range
- Target price(s)
- Stop-loss level
- Position size recommendation

**Step 6 — Risk Caveat**
- Top 3 risks that could invalidate this thesis
- Under what conditions should this view be revised?`, ticker)
}
