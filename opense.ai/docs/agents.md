# OpeNSE.ai — AI Agents

> Architecture, roles, and prompt engineering for the multi-agent stock analysis system.

## Overview

OpeNSE.ai employs a **multi-agent architecture** where specialized AI agents collaborate under a Chief Investment Officer (CIO) orchestrator. Each agent has deep domain expertise, structured reasoning templates, and access to specific tools.

## Agent Roster

| Agent | Internal Name | Role | Weight (Long-term) | Weight (Short-term) |
|-------|--------------|------|--------------------|--------------------|
| Fundamental Analyst | `fundamental_analyst` | Company financials, valuation | 40% | 10% |
| Technical Analyst | `technical_analyst` | Price action, indicators | 20% | 35% |
| Sentiment Analyst | `sentiment_analyst` | News, market mood | 15% | 20% |
| F&O Analyst | `fno_analyst` | Derivatives, options | 15% | 25% |
| Risk Manager | `risk_manager` | Position sizing, risk | 10% | 10% |
| Trade Executor | `trade_executor` | Order execution | — | — |
| Report Generator | `report_generator` | Research reports | — | — |
| CIO | `chief_investment_officer` | Orchestration | — | — |

## Agent Details

### 1. Fundamental Analyst

**Expertise**: Indian company financials, ratios (PE, PB, ROE, ROCE, D/E), growth metrics, DCF valuation, promoter holding analysis, institutional flows.

**Key capabilities**:
- Income Statement, Balance Sheet, Cash Flow analysis
- Revenue/profit growth: QoQ, YoY, 3-year CAGR, 5-year CAGR
- Relative valuation (sector peers), Graham Number, PEG ratio
- Red flag detection: declining promoter holding, pledge increases, auditor qualifications

**Output**: Structured analysis with Company Overview, Financial Health, Valuation, Peer Comparison, Red Flags, Recommendation.

### 2. Technical Analyst

**Expertise**: Price action analysis for NSE stocks using indicators, patterns, and volume analysis.

**Key capabilities**:
- Indicators: RSI(14), MACD(12,26,9), Bollinger Bands(20,2), SuperTrend, ATR, VWAP
- Moving averages: SMA, EMA, WMA across 5/10/20/50/100/200 periods
- Candlestick patterns: Doji, Hammer, Engulfing, Morning/Evening Star, H&S
- Support/Resistance: Classic, Fibonacci, Camarilla pivot points
- Multi-timeframe: daily (swing), weekly (positional), monthly (investment)

**Output**: Trend, Key Indicators, Support/Resistance, Patterns, Signal (BUY/SELL/NEUTRAL with entry/target/stop-loss).

### 3. Sentiment Analyst

**Expertise**: Market sentiment from Indian financial media, FII/DII flows, corporate events, macro factors.

**Key capabilities**:
- News sentiment scoring (-1.0 to +1.0 scale)
- Source weighting: official filings > tier-1 media > forums
- Catalyst identification (earnings, M&A, regulatory changes)
- Market-wide vs company-specific sentiment separation

**Output**: Overall Sentiment score, Key Drivers, Upcoming Catalysts, Market Context.

### 4. F&O Analyst

**Expertise**: Derivatives analysis on NSE — option chains, futures, OI analysis.

**Key capabilities**:
- Option chain analysis: OI distribution, IV analysis, max pain
- PCR trend analysis (PCR > 1.3 = oversold, PCR < 0.7 = overbought)
- OI buildup classification: Long Buildup, Short Buildup, Long Unwinding, Short Covering
- Strategy design: spreads, straddles, iron condors with payoff calculation
- NSE-specific: weekly/monthly expiry, lot sizes, SEBI margin rules

**Output**: OI Analysis, PCR Analysis, Futures View, Strategy Suggestion, Signal.

### 5. Risk Manager

**Expertise**: Portfolio risk assessment and position sizing for Indian market investors.

**Key capabilities**:
- Position sizing: Kelly criterion, fixed fractional, ATR-based
- Risk metrics: VaR, Beta, correlation, drawdown analysis
- India VIX-based volatility assessment
- Cost-aware: STT, GST, stamp duty factored into returns
- Safety rules: 5% max position, 2% daily loss limit, 10 max open positions

**Output**: Position Sizing (₹), Risk Assessment, Stop-Loss, R:R Ratio, Portfolio Impact, Approve/Reject.

### 6. Trade Executor

**Expertise**: Order execution through broker APIs with human-in-the-loop.

**Key safety rules**:
1. **Never executes live trades without explicit human confirmation**
2. Validates: price within 20% of LTP, quantity within limits
3. Uses LIMIT orders by default (avoid market orders during volatility)
4. Calculates complete brokerage including STT, GST, stamp duty
5. Logs every order attempt for audit trail

### 7. Report Generator

**Expertise**: Compiling multi-agent analysis into professional equity research reports.

**Output format**: Executive Summary → Company Profile → Fundamental Analysis → Technical Analysis → Derivatives Outlook → Sentiment → Risk Assessment → Recommendation → Disclaimer.

### 8. CIO (Orchestrator)

**Expertise**: Coordinating the team and synthesizing diverse perspectives.

**Decision framework**:
1. Business quality (fundamental) → Is this a good company?
2. Timing (technical) → Is now the right time?
3. Market mood (sentiment) → Is the market supportive?
4. Smart money (derivatives) → What do informed traders signal?
5. Risk (risk manager) → Is the trade appropriately sized?
6. Synthesize → Weighted final recommendation

**Conflict resolution**: When agents disagree, CIO explains the divergence and states which view is favored with reasoning.

## Prompt Engineering

### System Prompts

Each agent receives a detailed system prompt (`internal/agent/prompts/system.go`) that defines:
- Role and expertise areas
- Domain-specific guidelines (8-10 rules)
- Output format specification
- Indian market conventions

### Chain-of-Thought (CoT) Templates

Six CoT templates (`internal/agent/prompts/cot.go`) guide step-by-step reasoning:

| Template | Steps | Use Case |
|----------|-------|----------|
| `CoTAnalysis(ticker, task)` | 5 steps | Generic analysis task |
| `CoTFundamental(ticker)` | 7 steps | Full fundamental analysis |
| `CoTTechnical(ticker)` | 7 steps | Full technical analysis |
| `CoTDerivatives(ticker)` | 7 steps | Full derivatives analysis |
| `CoTRisk(ticker, capital)` | 6 steps | Risk assessment |
| `CoTSynthesis(ticker)` | 6 steps | CIO final synthesis |

### Indian Market Context

All agents receive India-specific context (`internal/agent/prompts/indian_market.go`):

- **Market Structure**: NSE/BSE, 9:15-3:30 IST, T+1 settlement, circuit limits
- **Number Formatting**: ₹ prefix, Indian comma grouping (₹12,34,567), Lakhs/Crores
- **Sector Classification**: 15 sectors with ticker mappings (IT, Banking, NBFC, Pharma, etc.)
- **Brokerage Calculation**: STT, exchange charges, SEBI charges, stamp duty, GST

## Agent Configuration

Agents are configured in `config/agents.yaml`:

```yaml
agents:
  fundamental_analyst:
    model: gpt-4o
    temperature: 0.1
    max_tokens: 4096
    
teams:
  full_analysis:
    - fundamental_analyst
    - technical_analyst
    - sentiment_analyst
    - fno_analyst
    - risk_manager
    - report_generator
```

## Extension Points

To add a new agent:

1. Define the agent name constant in `prompts/system.go`
2. Write the system prompt with expertise, guidelines, and output format
3. Create a CoT template in `prompts/cot.go` (optional but recommended)
4. Register the agent in `config/agents.yaml`
5. Add to relevant team compositions
