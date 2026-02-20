# OpeNSE.ai — Trading Safety

> Safety guardrails, risk management rules, and capital protection mechanisms.

## Safety Philosophy

OpeNSE.ai is designed with a **safety-first approach**. The system defaults to paper trading and requires explicit, deliberate actions to execute live trades. Every layer of the system includes safeguards against accidental capital loss.

## Default Safety Configuration

```yaml
trading:
  mode: "paper"                   # NEVER live by default
  max_position_pct: 5.0           # Max 5% of capital per position
  daily_loss_limit_pct: 2.0       # Stop trading if 2% daily loss hit
  max_open_positions: 10          # Max 10 concurrent positions
  require_confirmation: true      # Human must confirm every trade
  confirm_timeout_sec: 60         # Confirmation expires after 60s
  initial_capital: 1000000        # ₹10 lakh default paper capital
```

## Guardrails

### 1. Paper Trading by Default

- Fresh installations default to `mode: paper`
- Switching to `mode: live` requires explicit config file change
- Paper mode simulates fills at realistic prices with slippage modeling
- All PnL tracking works identically in paper and live modes

### 2. Human-in-the-Loop Execution

The Trade Executor agent **never executes a live trade without explicit human confirmation**:

```
⚠️ TRADE CONFIRMATION REQUIRED
─────────────────────────────
  Action:   BUY
  Ticker:   RELIANCE
  Exchange: NSE
  Type:     LIMIT @ ₹2,845.50
  Qty:      10 shares
  Product:  CNC (Delivery)
  Cost:     ₹28,455.00
  STT:      ₹28.46
  Total:    ₹28,512.34
─────────────────────────────
  Confirm? [Y/N] (expires in 60s)
```

### 3. Position Size Limits

| Rule | Default | Purpose |
|------|---------|---------|
| Max position % | 5% of capital | Prevents over-concentration |
| Max open positions | 10 | Limits total portfolio risk |
| Minimum R:R ratio | 1:2 | Ensures favorable risk-reward |
| Max sector exposure | 25% | Prevents sector concentration |

**Example**: With ₹10,00,000 capital:
- Max per position: ₹50,000
- Max total invested: ₹5,00,000 (10 positions × ₹50,000)

### 4. Daily Loss Limit

If cumulative daily losses exceed 2% of capital, the system:
1. Stops placing new buy orders
2. Alerts the user
3. Requires explicit override to continue trading that day

### 5. Order Validation

Before any order reaches the broker API:

| Check | Rule |
|-------|------|
| Price sanity | Must be within ±20% of last traded price |
| Quantity limits | Must respect NSE lot size and exchange limits |
| Margin check | Sufficient margin must be available |
| Market hours | Validates NSE trading hours (9:15 AM – 3:30 PM IST) |
| Circuit limits | Respects 5%/10%/20% circuit breakers |

### 6. Risk Assessment

The Risk Manager agent evaluates every trade against:

- **Volatility**: India VIX level (>15 = elevated, >20 = high)
- **Stock volatility**: ATR(14) relative to price
- **Liquidity**: Average daily volume vs. position size
- **Correlation**: With existing portfolio holdings
- **Portfolio beta**: After adding the proposed position

### 7. Audit Trail

Every order attempt is logged with full context:

```json
{
  "timestamp": "2026-02-20T14:30:15+05:30",
  "action": "ORDER_SUBMITTED",
  "ticker": "RELIANCE",
  "side": "BUY",
  "quantity": 10,
  "price": 2845.50,
  "order_type": "LIMIT",
  "status": "COMPLETE",
  "agent": "trade_executor",
  "analysis_id": "abc-123",
  "confirmation": "user_approved"
}
```

## Brokerage Cost Awareness

All trade evaluations include Indian brokerage costs:

| Component | Rate |
|-----------|------|
| **STT** (Securities Transaction Tax) | 0.1% delivery (buy+sell), 0.025% intraday (sell) |
| **Exchange charges** | ~0.00345% of turnover |
| **SEBI charges** | ₹10 per crore |
| **Stamp duty** | 0.015% (buy side) |
| **GST** | 18% on brokerage + exchange charges |

The Risk Manager ensures that after costs, the trade still has adequate profit potential.

## Broker Safety

### Zerodha (Kite Connect)

- API keys stored in environment variables (never in config files committed to git)
- Access token refresh handled automatically
- Order placement uses idempotency keys to prevent duplicate orders
- All errors logged with Kite error codes

### Interactive Brokers (IBKR)

- Connects via TWS/IB Gateway (localhost only)
- Read-only mode available for analysis without trade capability
- Paper trading account support for testing

### Paper Broker

- Default broker — no real money at risk
- Simulates realistic fills: market orders at last price ± slippage
- Tracks positions, PnL, and portfolio metrics identically to live

## Configuration for Live Trading

**Prerequisites** before switching to live mode:

1. ✅ Thoroughly tested strategy in paper mode
2. ✅ API keys configured via environment variables
3. ✅ Understood all cost components
4. ✅ Set appropriate position limits
5. ✅ Reviewed the Risk Manager's assessment

**Steps**:

```yaml
# config/config.yaml
trading:
  mode: "live"                    # Changed from "paper"
  require_confirmation: true      # Keep this TRUE
  max_position_pct: 3.0           # Consider reducing from 5%
```

```bash
# Set API keys via environment (NEVER in config file)
export OPENSEAI_BROKER_ZERODHA_API_KEY="your_key"
export OPENSEAI_BROKER_ZERODHA_API_SECRET="your_secret"
```

## Emergency Procedures

| Situation | Action |
|-----------|--------|
| Unexpected large loss | System triggers daily loss limit; review positions |
| Broker API error | System logs error, does NOT retry blindly |
| Market circuit halt | System pauses all pending orders |
| Config override needed | Edit config + restart; no hot-reload for safety |

## Disclaimer

OpeNSE.ai is a research and analysis tool. It does not constitute financial advice. All trading decisions should be made with full understanding of the risks involved. Past performance of any analysis or strategy does not guarantee future results. Always consult a registered financial advisor before making investment decisions.
