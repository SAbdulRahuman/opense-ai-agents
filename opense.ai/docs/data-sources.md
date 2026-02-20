# OpeNSE.ai — Data Sources

> Available data sources, adapters, rate limits, and caching strategy.

## Architecture

```
┌──────────────────────┐
│   Data Aggregator    │  ← unified interface
│  (internal/datasource)│
├──────────────────────┤
│  ┌────────────────┐  │
│  │  NSE Adapter   │  │  ← quotes, option chains, corporate actions
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │  Yahoo Finance │  │  ← historical OHLCV, fundamentals
│  │    Adapter     │  │
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │  In-Memory     │  │  ← TTL-based caching layer
│  │    Cache       │  │
│  └────────────────┘  │
└──────────────────────┘
```

## Data Sources

### NSE (National Stock Exchange)

| Data Type | Endpoint | Update Frequency | Cache TTL |
|-----------|----------|-----------------|-----------|
| Live Quotes | NSE website API | Real-time (market hours) | 15s |
| Option Chain | NSE option chain API | Every minute | 60s |
| Deliverables | NSE delivery data | End of day | 300s |
| Corporate Actions | NSE corporate filings | As published | 3600s |
| Index Data | NSE index API | Real-time | 15s |

**Rate limits**: NSE enforces informal rate limits. The adapter implements:
- Request throttling: max 3 requests/second
- Exponential backoff on 429 responses
- Session management with cookie rotation

### Yahoo Finance

| Data Type | Coverage | Cache TTL |
|-----------|----------|-----------|
| Historical OHLCV | All NSE stocks, up to 20 years | 300s |
| Quarterly Financials | Income statement, balance sheet, cash flow | 3600s |
| Annual Financials | 5-year history | 3600s |
| Stock Profile | Company info, sector, industry | 86400s |
| Key Statistics | Market cap, PE, PB, beta | 300s |

**Rate limits**: Yahoo Finance API has liberal limits but:
- The adapter batches requests where possible
- Uses `.NS` suffix for NSE tickers (e.g., `TCS.NS`, `RELIANCE.NS`)
- Handles currency conversion (all values in ₹)

## Data Models

### OHLCV (Price Data)

```go
type OHLCV struct {
    Timestamp time.Time
    Open      float64
    High      float64
    Low       float64
    Close     float64
    Volume    int64
}
```

Supported timeframes: `1m`, `5m`, `15m`, `1h`, `1d`, `1w`, `1M`

### Quote (Live)

```go
type Quote struct {
    Ticker    string
    LastPrice float64
    Change    float64
    ChangePct float64
    Open, High, Low, PrevClose float64
    Volume    int64
    Timestamp time.Time
}
```

### Financial Data

```go
type FinancialData struct {
    Ticker            string
    AnnualIncome      []IncomeStatement
    QuarterlyIncome   []IncomeStatement
    AnnualBalance     []BalanceSheet
    QuarterlyBalance  []BalanceSheet
    AnnualCashFlow    []CashFlow
    QuarterlyCashFlow []CashFlow
    Ratios            FinancialRatios
    Growth            GrowthRates
}
```

### Option Chain

```go
type OptionChain struct {
    Ticker     string
    SpotPrice  float64
    Expiry     time.Time
    Contracts  []OptionContract
    Futures    []FuturesContract
    TotalCEOI  int64
    TotalPEOI  int64
    PCR        float64
    MaxPain    float64
}
```

## Caching Strategy

### In-Memory Cache

- **Implementation**: TTL-based concurrent map with lazy expiration
- **Hit behavior**: Returns cached data immediately (zero latency)
- **Miss behavior**: Fetches from source, stores result, returns
- **Eviction**: Time-based (TTL per data type)

### Default TTLs

| Data Type | TTL | Rationale |
|-----------|-----|-----------|
| Live quotes | 15s | Near-real-time but prevents hammering |
| Option chains | 60s | OI changes slowly within a minute |
| Historical OHLCV | 300s | Doesn't change after market close |
| Financials | 3600s | Quarterly updates only |
| Stock profiles | 86400s | Rarely changes |

### Configuration

```yaml
analysis:
  cache_ttl: 300        # default cache TTL in seconds
  concurrent_fetches: 5  # max parallel data fetches

financeql:
  cache_ttl: 60         # FinanceQL query cache TTL
```

Override via environment:
```bash
export OPENSEAI_ANALYSIS_CACHE_TTL=600
export OPENSEAI_ANALYSIS_CONCURRENT_FETCHES=10
```

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Source timeout | Retry with exponential backoff (3 attempts) |
| Rate limit (429) | Wait + retry with jitter |
| Invalid ticker | Return structured error with suggestion |
| Partial data | Return available data with warning flag |
| Source down | Fallback to cache (even if stale) if available |

## Adding a New Data Source

1. Implement the `DataSource` interface in `internal/datasource/`
2. Register the adapter in the Aggregator
3. Add TTL configuration to `config.go`
4. Write tests following the existing adapter test patterns
5. Document the source in this file
