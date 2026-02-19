# OpeNSE.ai — Agentic AI for NSE Stock Analysis & Trading

> **OpeNSE.ai** (Open + NSE + Agentic AI) — A Go-based multi-agent AI system for comprehensive NSE (National Stock Exchange of India) stock analysis, covering fundamental, technical, derivatives, sentiment analysis, and automated trading.

## Why Go?

- **Concurrency**: Goroutines + channels for parallel data fetching across multiple stocks/indicators
- **Performance**: Low-latency order execution and real-time data processing
- **Single binary**: Easy deployment — no Python dependency hell
- **Strong typing**: Safer financial calculations, compile-time error catching
- **Production-ready**: Built-in HTTP server, excellent stdlib for API services

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                  OpeNSE.ai Web UI (Next.js)                 │
│  ┌─────────────────┬──────────────┬───────────────────────┐ │
│  │  TradingView    │   Chat UI    │  FinanceQL Explorer   │ │
│  │  Charts         │   (Agentic   │  (Prometheus-style    │ │
│  │  (lightweight-  │    Chat +    │   Query Editor +      │ │
│  │   charts)       │    HITL)     │   Result Viewer)      │ │
│  └─────────────────┴──────────────┴───────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                     OpeNSE.ai CLI / API                     │
│              (Quick Query + Deep Analysis modes)             │
├─────────────────────────────────────────────────────────────┤
│                    FinanceQL Engine                          │
│           (PromQL-inspired Financial Query Language)         │
│     Instant queries · Range queries · Screening · Alerts    │
├─────────────────────────────────────────────────────────────┤
│                    Agent Orchestrator                        │
│         ┌──────────┬──────────┬──────────────┐              │
│         │ Single   │ Multi-   │ Trading      │              │
│         │ Agent    │ Agent    │ Agent        │              │
│         │ Mode     │ Team     │ (with HITL)  │              │
│         └──────────┴──────────┴──────────────┘              │
├─────────────────────────────────────────────────────────────┤
│                    LLM Gateway Layer                         │
│     (OpenAI / Ollama / Gemini / Anthropic — configurable)   │
├──────────┬──────────┬───────────┬───────────┬───────────────┤
│ NSE Data │Technical │ F&O       │ Sentiment │ Broker        │
│ Source   │Analysis  │ Analysis  │ Analysis  │ Integration   │
│ Layer    │Engine    │ Engine    │ Engine    │ Layer         │
└──────────┴──────────┴───────────┴───────────┴───────────────┘
```

---

## Project Structure

```
opense.ai/
├── cmd/
│   └── openseai/
│       └── main.go                 # CLI entrypoint (binary: openseai)
├── internal/
│   ├── agent/
│   │   ├── orchestrator.go         # Agent coordination (single + multi-agent)
│   │   ├── agent.go                # Base agent interface & implementation
│   │   ├── fundamental.go          # Fundamental Analyst agent
│   │   ├── technical.go            # Technical Analyst agent
│   │   ├── sentiment.go            # Sentiment Analyst agent
│   │   ├── fno.go                  # F&O / Derivatives Analyst agent
│   │   ├── risk.go                 # Risk Manager agent
│   │   ├── executor.go             # Trade Executor agent (human-in-the-loop)
│   │   ├── reporter.go             # Report Generator agent
│   │   └── prompts/
│   │       ├── system.go           # System prompts for each agent role
│   │       ├── cot.go              # Financial Chain-of-Thought templates
│   │       └── indian_market.go    # India-specific formatting (₹, lakhs, crores)
│   ├── datasource/
│   │   ├── datasource.go           # Common interface for all data sources
│   │   ├── yfinance.go             # YFinance via HTTP API (NSE: RELIANCE.NS)
│   │   ├── nse.go                  # NSE India data (jugaad-trader Go port / NSE APIs)
│   │   ├── nse_derivatives.go      # Option chain, futures, PCR, OI, India VIX
│   │   ├── news.go                 # Indian news scrapers (Moneycontrol, ET, LiveMint)
│   │   ├── screener.go             # Screener.in financial ratios & peer comparison
│   │   └── fii_dii.go              # FII/DII activity data
│   ├── analysis/
│   │   ├── technical/
│   │   │   ├── indicators.go       # RSI, MACD, Bollinger, SuperTrend, ATR
│   │   │   ├── moving_avg.go       # SMA, EMA, WMA, VWAP
│   │   │   ├── patterns.go         # Candlestick pattern detection
│   │   │   ├── support_resistance.go # Pivot points, Fibonacci, S/R levels
│   │   │   └── signals.go          # Consolidated buy/sell/neutral signals
│   │   ├── fundamental/
│   │   │   ├── ratios.go           # PE, PB, ROE, ROCE, D/E, current ratio
│   │   │   ├── financials.go       # Income, balance sheet, cash flow analysis
│   │   │   ├── valuation.go        # DCF, relative valuation, intrinsic value
│   │   │   ├── peer_compare.go     # Sector peer comparison
│   │   │   └── promoter.go         # Promoter holding & pledge analysis
│   │   ├── derivatives/
│   │   │   ├── option_chain.go     # OI analysis, max pain, IV
│   │   │   ├── pcr.go              # Put-Call Ratio analysis
│   │   │   ├── oi_buildup.go       # Long/short buildup detection
│   │   │   └── strategies.go       # Option strategy builder & payoff
│   │   └── sentiment/
│   │       ├── news_scorer.go      # News sentiment scoring via LLM
│   │       └── aggregator.go       # Multi-source sentiment aggregation
│   ├── broker/
│   │   ├── broker.go               # Common broker interface
│   │   ├── zerodha.go              # Zerodha Kite API integration
│   │   ├── ibkr.go                 # Interactive Brokers integration
│   │   ├── paper.go                # Paper trading simulator
│   │   ├── order.go                # Order types, validation, position sizing
│   │   └── risk.go                 # Pre-trade risk checks, daily loss limits
│   ├── llm/
│   │   ├── provider.go             # LLM provider interface
│   │   ├── openai.go               # OpenAI GPT-4 / GPT-4o
│   │   ├── ollama.go               # Local Ollama (Qwen, Llama, etc.)
│   │   ├── gemini.go               # Google Gemini
│   │   ├── anthropic.go            # Anthropic Claude
│   │   ├── tools.go                # Function/tool-calling abstraction
│   │   └── router.go               # Model routing & fallback logic
│   ├── report/
│   │   ├── pdf.go                  # PDF equity research report generation
│   │   ├── chart.go                # Chart generation (go-echarts / go-chart)
│   │   └── templates/              # Report templates (HTML → PDF)
│   ├── backtest/
│   │   ├── engine.go               # Backtesting engine
│   │   ├── strategy.go             # Strategy interface + built-in strategies
│   │   ├── metrics.go              # Sharpe, drawdown, CAGR, win rate
│   │   └── benchmark.go            # Nifty 50 / Sensex benchmark comparison
│   ├── financeql/
│   │   ├── lexer.go                # FinanceQL tokenizer / lexer
│   │   ├── parser.go               # FinanceQL recursive descent parser
│   │   ├── ast.go                  # Abstract Syntax Tree node types
│   │   ├── evaluator.go            # Query evaluator against live/historical data
│   │   ├── functions.go            # Built-in functions (sma, rsi, macd, pe, etc.)
│   │   ├── types.go                # Scalar, Vector, Matrix, Range result types
│   │   └── repl.go                 # Interactive REPL with autocomplete
│   └── config/
│       ├── config.go               # YAML/JSON config loader
│       └── keys.go                 # API key management (env vars / config)
├── pkg/
│   ├── models/
│   │   ├── stock.go                # Stock, OHLCV, Quote structs
│   │   ├── financials.go           # IncomeStatement, BalanceSheet, CashFlow
│   │   ├── option.go               # OptionChain, OptionContract, Greeks
│   │   ├── order.go                # Order, Position, Holding structs
│   │   └── analysis.go             # AnalysisResult, Signal, Recommendation
│   └── utils/
│       ├── indian_format.go        # ₹ formatting, lakhs/crores conversion
│       ├── nse_ticker.go           # Ticker normalization (RELIANCE → RELIANCE.NS)
│       └── timeutil.go             # IST timezone, market hours, trading calendar
├── api/
│   ├── server.go                   # HTTP/gRPC API server
│   ├── handlers.go                 # REST endpoints for analysis
│   └── middleware.go               # Auth, rate-limiting, logging
├── web/                                # Next.js Web UI
│   ├── package.json
│   ├── tsconfig.json
│   ├── next.config.ts
│   ├── tailwind.config.ts
│   ├── postcss.config.mjs
│   ├── .env.local.example
│   ├── public/
│   │   └── favicon.ico
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx              # Root layout (sidebar nav, theme provider)
│   │   │   ├── page.tsx                # Dashboard / home page
│   │   │   ├── globals.css             # Global styles (Tailwind base)
│   │   │   ├── chart/
│   │   │   │   └── page.tsx            # TradingView chart page
│   │   │   ├── chat/
│   │   │   │   └── page.tsx            # Chat UI page
│   │   │   ├── query/
│   │   │   │   └── page.tsx            # FinanceQL Explorer page
│   │   │   ├── portfolio/
│   │   │   │   └── page.tsx            # Portfolio & holdings page
│   │   │   ├── screener/
│   │   │   │   └── page.tsx            # Stock screener page
│   │   │   └── backtest/
│   │   │       └── page.tsx            # Backtest results page
│   │   ├── components/
│   │   │   ├── ui/                     # Shared UI primitives (shadcn/ui)
│   │   │   │   ├── button.tsx
│   │   │   │   ├── input.tsx
│   │   │   │   ├── card.tsx
│   │   │   │   ├── table.tsx
│   │   │   │   ├── tabs.tsx
│   │   │   │   ├── badge.tsx
│   │   │   │   ├── tooltip.tsx
│   │   │   │   └── select.tsx
│   │   │   ├── chart/
│   │   │   │   ├── TradingViewChart.tsx # TradingView lightweight-charts wrapper
│   │   │   │   ├── ChartToolbar.tsx     # Indicator selector, timeframe picker
│   │   │   │   ├── IndicatorOverlay.tsx # RSI, MACD, BB overlay panels
│   │   │   │   ├── VolumePane.tsx       # Volume sub-chart pane
│   │   │   │   └── ChartLegend.tsx      # Price/indicator legend
│   │   │   ├── chat/
│   │   │   │   ├── ChatPanel.tsx        # Main chat container
│   │   │   │   ├── MessageBubble.tsx    # User/agent message rendering
│   │   │   │   ├── AgentBadge.tsx       # Agent role indicator (Fundamental, Technical…)
│   │   │   │   ├── ToolCallCard.tsx     # Expandable tool call/result display
│   │   │   │   ├── TradeConfirm.tsx     # Human-in-the-loop trade approval UI
│   │   │   │   ├── ChatInput.tsx        # Message input with ticker autocomplete
│   │   │   │   └── StreamingText.tsx    # Streaming LLM response renderer
│   │   │   ├── financeql/
│   │   │   │   ├── QueryEditor.tsx      # CodeMirror/Monaco FinanceQL editor
│   │   │   │   ├── QueryHistory.tsx     # Recent queries sidebar
│   │   │   │   ├── ResultTable.tsx      # Tabular result display
│   │   │   │   ├── ResultChart.tsx      # Time-series result as chart
│   │   │   │   ├── ResultScalar.tsx     # Single-value result with formatting
│   │   │   │   ├── ExpressionTree.tsx   # Parsed AST visualization
│   │   │   │   └── AlertManager.tsx     # Active alerts list & management
│   │   │   ├── dashboard/
│   │   │   │   ├── Watchlist.tsx        # Real-time watchlist with sparklines
│   │   │   │   ├── MarketOverview.tsx   # Nifty 50, Bank Nifty, India VIX cards
│   │   │   │   ├── FIIDIIBar.tsx        # FII/DII activity bar chart
│   │   │   │   └── TopMovers.tsx        # Top gainers/losers
│   │   │   ├── layout/
│   │   │   │   ├── Sidebar.tsx          # Navigation sidebar
│   │   │   │   ├── Header.tsx           # Top bar (market status, search)
│   │   │   │   └── ThemeToggle.tsx      # Light/dark mode toggle
│   │   │   └── common/
│   │   │       ├── TickerSearch.tsx     # Global ticker search with autocomplete
│   │   │       ├── IndianNumber.tsx     # ₹ lakhs/crores formatted display
│   │   │       ├── SignalBadge.tsx      # BUY/SELL/NEUTRAL signal badge
│   │   │       └── LoadingSkeleton.tsx  # Skeleton loaders
│   │   ├── hooks/
│   │   │   ├── useWebSocket.ts         # WebSocket connection hook
│   │   │   ├── useFinanceQL.ts         # FinanceQL query execution hook
│   │   │   ├── useChat.ts              # Chat session management hook
│   │   │   ├── useMarketData.ts        # Real-time market data hook
│   │   │   └── useTradingView.ts       # TradingView chart lifecycle hook
│   │   ├── lib/
│   │   │   ├── api.ts                  # API client (fetch wrapper for Go backend)
│   │   │   ├── ws.ts                   # WebSocket client utilities
│   │   │   ├── financeql-lang.ts       # FinanceQL syntax highlighting & autocomplete
│   │   │   └── format.ts               # Indian number formatting (₹, lakhs, crores)
│   │   ├── types/
│   │   │   ├── stock.ts                # Stock, OHLCV, Quote types
│   │   │   ├── analysis.ts             # AnalysisResult, Signal types
│   │   │   ├── chat.ts                 # ChatMessage, AgentRole, ToolCall types
│   │   │   ├── financeql.ts            # QueryResult, Scalar, Vector, Matrix types
│   │   │   └── order.ts               # Order, Position, Holding types
│   │   └── store/
│   │       ├── useAppStore.ts          # Zustand global store
│   │       ├── chatSlice.ts            # Chat state slice
│   │       ├── marketSlice.ts          # Market data state slice
│   │       └── querySlice.ts           # FinanceQL query state slice
│   └── Dockerfile                      # Multi-stage Next.js Docker build
├── config/
│   ├── config.example.yaml         # Example configuration
│   └── agents.yaml                 # Agent roles, toolkits, hierarchy config
├── scripts/
│   ├── setup.sh                    # Dev environment setup
│   └── test_nse.sh                 # NSE data source smoke tests
├── docs/
│   ├── architecture.md             # Detailed architecture doc
│   ├── agents.md                   # Agent roles and responsibilities
│   ├── data-sources.md             # Data source documentation
│   ├── financeql.md                # FinanceQL language reference & examples
│   ├── trading-safety.md           # Trading safety guardrails doc
│   └── web-ui.md                   # Web UI component reference & design system
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yaml             # Orchestrate Go backend + Next.js frontend
├── README.md
└── plan.md                         # This file
```
 .
---

## Implementation Plan

### Phase 1: Foundation (Week 1–2)

#### Step 1.1 — Project Scaffolding
- Initialize Go module: `go mod init github.com/seenimoa/openseai`
- Set up directory structure as above
- Create `Makefile` with targets: `build`, `test`, `lint`, `run`
- Set up CI with GitHub Actions (lint + test)

#### Step 1.2 — Core Models & Utilities
- Define all structs in `pkg/models/`: `Stock`, `OHLCV`, `Quote`, `IncomeStatement`, `BalanceSheet`, `CashFlow`, `OptionChain`, `Order`, `Position`
- Implement `pkg/utils/indian_format.go`: rupee formatting (`₹12,34,567.89`), lakhs/crores conversion
- Implement `pkg/utils/nse_ticker.go`: normalize tickers (add `.NS`, handle aliases like `RELIANCE` → `RELIANCE.NS`)
- Implement `pkg/utils/timeutil.go`: IST timezone, NSE market hours (9:15 AM – 3:30 PM), trading calendar with holidays

#### Step 1.3 — Configuration System
- YAML-based config in `internal/config/config.go`
- Support env vars override: `OPENSEAI_OPENAI_KEY`, `OPENSEAI_ZERODHA_KEY`, etc.
- Agent configuration in `config/agents.yaml`: roles, system prompts, assigned tools, hierarchy

---

### Phase 2: Data Sources (Week 2–3)

#### Step 2.1 — YFinance Data Source
- `internal/datasource/yfinance.go`: HTTP client calling Yahoo Finance API
- Methods: `GetQuote()`, `GetHistoricalData()`, `GetFinancials()`, `GetDividends()`, `GetAnalystRecommendations()`
- Auto-append `.NS` for NSE tickers
- Response parsing into `pkg/models` structs

#### Step 2.2 — NSE Direct Data Source
- `internal/datasource/nse.go`: Scrape/API calls to nseindia.com
- Methods: `GetNSEQuote()`, `GetBhavcopy()`, `GetIndexData()`, `GetPromoterHolding()`, `GetShareholdingPattern()`, `GetCorporateActions()`, `GetBulkDeals()`, `GetBlockDeals()`
- Handle NSE rate-limiting (cookies, headers, delays)
- Cache responses with configurable TTL

#### Step 2.3 — NSE Derivatives Data Source
- `internal/datasource/nse_derivatives.go`
- Methods: `GetOptionChain()`, `GetFuturesData()`, `GetIndiaVIX()`, `GetFIIDIIData()`
- Parse complex NSE JSON responses into `OptionChain` structs

#### Step 2.4 — Indian News Data Source
- `internal/datasource/news.go`: RSS + scraping
- Sources: Moneycontrol, Economic Times Markets, LiveMint, Business Standard
- Methods: `GetStockNews(ticker)`, `GetMarketNews()`, `GetSectorNews(sector)`
- Rate-limited, cached, with fallback sources

#### Step 2.5 — Screener.in Data Source
- `internal/datasource/screener.go`: Scrape Screener.in
- Methods: `GetFinancialRatios()`, `GetPeerComparison()`, `GetQuarterlyResults()`, `GetAnnualResults()`

#### Step 2.6 — Data Source Interface & Parallelism
- Define common `DataSource` interface in `internal/datasource/datasource.go`
- Implement concurrent data fetching using goroutines + `errgroup`
- `FetchAll(ticker) → StockProfile` aggregates from all sources in parallel

---

### Phase 3: Analysis Engines (Week 3–5)

#### Step 3.1 — Technical Analysis Engine
- **Indicators** (`internal/analysis/technical/indicators.go`):
  - RSI (14-period default, configurable), MACD (12,26,9), Bollinger Bands (20,2), SuperTrend (7,3), ATR (14)
- **Moving Averages** (`moving_avg.go`): SMA, EMA, WMA, VWAP for 5/10/20/50/100/200 periods
- **Pattern Detection** (`patterns.go`): Doji, Hammer, Engulfing, Morning/Evening Star, Head & Shoulders
- **Support/Resistance** (`support_resistance.go`): Pivot points (classic, Fibonacci, Camarilla), auto S/R from price action
- **Signal Generator** (`signals.go`): Aggregate multiple indicators → `BUY` / `SELL` / `NEUTRAL` with confidence score

#### Step 3.2 — Fundamental Analysis Engine
- **Ratios** (`internal/analysis/fundamental/ratios.go`): PE, PB, EV/EBITDA, ROE, ROCE, D/E, Current Ratio, Interest Coverage, Dividend Yield — computed from raw financials
- **Financial Analysis** (`financials.go`): Revenue/profit growth rates (QoQ, YoY, 3Y/5Y CAGR), margin trends, working capital analysis
- **Valuation** (`valuation.go`): DCF model (configurable WACC, growth rates), relative valuation vs peers, Graham Number, PEG ratio
- **Peer Comparison** (`peer_compare.go`): Auto-detect sector peers from NSE classification, side-by-side ratio comparison
- **Promoter Analysis** (`promoter.go`): Holding trend, pledge %, institutional holding changes (FII/DII/MF)

#### Step 3.3 — Derivatives Analysis Engine
- **Option Chain Analysis** (`internal/analysis/derivatives/option_chain.go`): Max pain calculation, OI-based support/resistance, unusual OI activity detection
- **PCR Analysis** (`pcr.go`): PCR trend, historical PCR comparison, PCR divergence signals
- **OI Buildup** (`oi_buildup.go`): Long buildup, short buildup, long unwinding, short covering classification
- **Strategy Builder** (`strategies.go`): Bull Call/Put Spread, Bear Call/Put Spread, Straddle, Strangle, Iron Condor, Butterfly — with payoff calculation and breakeven points

#### Step 3.4 — Sentiment Analysis Engine
- **News Scoring** (`internal/analysis/sentiment/news_scorer.go`): Feed news articles to LLM with structured output → sentiment score (-1 to +1) per article
- **Aggregator** (`aggregator.go`): Time-weighted aggregation, source credibility weighting, final sentiment score with confidence

---

### Phase 4: LLM Integration (Week 4–5)

#### Step 4.1 — LLM Provider Abstraction
- `internal/llm/provider.go`: Common interface:
  ```go
  type LLMProvider interface {
      Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error)
      ChatStream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamChunk, error)
  }
  ```
- Implementations: `openai.go`, `ollama.go`, `gemini.go`, `anthropic.go`

#### Step 4.2 — Tool/Function Calling Framework
- `internal/llm/tools.go`: Define tool registration system
  ```go
  type Tool struct {
      Name        string
      Description string
      Parameters  JSONSchema
      Handler     func(ctx context.Context, args json.RawMessage) (string, error)
  }
  ```
- Auto-register analysis functions as LLM tools
- Tool result → string serialization (DataFrames → formatted tables)

#### Step 4.3 — Model Router
- `internal/llm/router.go`: Route requests based on task complexity
  - Simple queries → GPT-4o-mini / local Ollama
  - Complex analysis → GPT-4 / Claude
  - Fallback chain: primary → secondary → tertiary LLM

---

### Phase 5: Agent System (Week 5–7)

#### Step 5.1 — Base Agent Framework
- `internal/agent/agent.go`: Base `Agent` interface and implementation
  ```go
  type Agent interface {
      Name() string
      Role() string
      SystemPrompt() string
      Tools() []Tool
      Process(ctx context.Context, task string) (*AgentResult, error)
  }
  ```
- Conversation memory management (sliding window + summary)
- Tool call execution loop (LLM → tool call → result → LLM)

#### Step 5.2 — Specialized Agents
Create each agent with role-specific system prompts and tool assignments:

| Agent | File | Tools | Responsibility |
|-------|------|-------|---------------|
| **Fundamental Analyst** | `fundamental.go` | YFinance financials, NSE promoter data, Screener ratios, peer comparison | Deep-dive into company financials, valuation, growth prospects |
| **Technical Analyst** | `technical.go` | Historical data, all technical indicators, chart generation | Price action, trend analysis, entry/exit levels |
| **Sentiment Analyst** | `sentiment.go` | Indian news sources, Reddit, sentiment scorer | Market mood, news impact assessment |
| **F&O Analyst** | `fno.go` | Option chain, futures data, PCR, OI analysis, strategy builder | Derivatives analysis, option strategies, OI interpretation |
| **Risk Manager** | `risk.go` | India VIX, FII/DII data, position sizing, risk calculator | Portfolio risk, position sizing, stop-loss levels |
| **Trade Executor** | `executor.go` | Broker APIs (Zerodha/IBKR), order validation | Execute trades with human confirmation |
| **Report Generator** | `reporter.go` | PDF generator, chart tools, all analysis results | Compile comprehensive research reports |

#### Step 5.3 — Agent Orchestrator
- `internal/agent/orchestrator.go`: Two modes:

  **Single Agent Mode** — One agent with all tools for quick queries:
  ```
  User: "What's the RSI of RELIANCE?"
  → Single agent calls GetHistoricalData + CalculateRSI → returns answer
  ```

  **Multi-Agent Mode** — CIO-led hierarchical team for deep analysis:
  ```
  User: "Full analysis of TCS for investment"
  → CIO delegates to:
    → Fundamental Analyst (financials, valuation, peer comparison)
    → Technical Analyst (trend, indicators, levels)
    → Sentiment Analyst (news, market mood)
    → F&O Analyst (option chain, OI signals)
    → Risk Manager (position sizing, risk metrics)
  → CIO synthesizes all inputs → final recommendation
  → Report Generator → PDF report
  ```

- Agent communication via Go channels
- Configurable hierarchy from `config/agents.yaml`

---

### Phase 6: Broker Integration & Trading (Week 7–8)

#### Step 6.1 — Broker Interface
- `internal/broker/broker.go`: Common interface:
  ```go
  type Broker interface {
      GetPositions(ctx) ([]Position, error)
      GetHoldings(ctx) ([]Holding, error)
      GetOrders(ctx) ([]Order, error)
      PlaceOrder(ctx, OrderRequest) (*OrderResponse, error)
      ModifyOrder(ctx, orderID, ModifyRequest) error
      CancelOrder(ctx, orderID) error
      GetMargins(ctx) (*Margins, error)
  }
  ```

#### Step 6.2 — Zerodha Kite Integration
- `internal/broker/zerodha.go`: Full Kite Connect API v3 implementation
- OAuth token management (login flow, token refresh)
- WebSocket streaming for live quotes

#### Step 6.3 — Interactive Brokers Integration
- `internal/broker/ibkr.go`: IB API via Client Portal REST API or TWS API
- Multi-market support for global portfolios

#### Step 6.4 — Paper Trading Simulator
- `internal/broker/paper.go`: In-memory paper trading that mirrors the `Broker` interface
- Simulated order fills, slippage model, brokerage calculation (STT, GST, stamp duty, SEBI charges)
- **Default mode** — all new users start here

#### Step 6.5 — Trading Safety Guardrails
- `internal/broker/risk.go`: Pre-trade validation:
  - Max position size: configurable % of capital (default: 5%)
  - Daily loss limit: stop trading if P&L < -2% of capital
  - Max open positions: configurable (default: 10)
  - Order size sanity check: reject orders > 10% of avg daily volume
  - **Human-in-the-loop**: ALL live orders require explicit `y/N` confirmation in CLI
  - Trade logging: every order (attempted + executed) logged to JSON/SQLite

---

### Phase 7: Report Generation (Week 8–9)

#### Step 7.1 — Chart Generation
- `internal/report/chart.go`: Using `go-echarts` or `gonum/plot`
- Candlestick charts with indicator overlays
- Performance comparison charts (stock vs Nifty 50)
- PE/PB historical band charts
- Option payoff diagrams

#### Step 7.2 — PDF Report Generator
- `internal/report/pdf.go`: Using `go-wkhtmltopdf` (HTML → PDF) or `gofpdf`
- HTML templates in `internal/report/templates/`
- Sections: Executive Summary, Fundamental Analysis, Technical Analysis, F&O View, Risk Assessment, Recommendation
- All charts embedded, Indian formatting throughout

---

### Phase 8: Backtesting Engine (Week 9–10)

#### Step 8.1 — Backtesting Core
- `internal/backtest/engine.go`: Event-driven backtesting engine
- Feed historical OHLCV data bar-by-bar
- Simulated broker with realistic fills, slippage, brokerage

#### Step 8.2 — Strategy Interface
- `internal/backtest/strategy.go`:
  ```go
  type Strategy interface {
      Name() string
      Init(ctx *StrategyContext)
      OnBar(ctx *StrategyContext, bar OHLCV)
      OnOrder(ctx *StrategyContext, order Order)
  }
  ```
- Built-in strategies: SMA Crossover, RSI Mean Reversion, SuperTrend, VWAP Breakout, MACD Crossover

#### Step 8.3 — Performance Metrics
- `internal/backtest/metrics.go`: Sharpe ratio, Sortino ratio, max drawdown, CAGR, win rate, profit factor, average R-multiple
- Benchmark comparison against Nifty 50 / Nifty Bank

---

### Phase 9: FinanceQL — Query Language (Week 10–11)

> **FinanceQL** — A PromQL-inspired domain-specific query language for financial time-series data. Query stocks, indicators, fundamentals, and screening criteria with a concise, composable syntax.

#### FinanceQL Syntax Overview

```
# ── Instant Queries (current/latest value) ──────────────────────────
price(RELIANCE)                              # Latest price → ₹2,847.50
rsi(TCS, 14)                                 # RSI(14) → 62.4
macd(INFY)                                   # MACD line, signal, histogram
pe(HDFCBANK)                                 # Current P/E ratio → 19.8
roe(TCS)                                     # Return on Equity → 48.2%
market_cap(RELIANCE)                         # Market cap → ₹19,27,345 Cr
oi(NIFTY, 24000, CE)                         # Open Interest for strike
vix()                                        # India VIX current value

# ── Range Queries (time-series) ──────────────────────────────────────
price(RELIANCE)[30d]                         # 30-day price series
rsi(TCS, 14)[90d]                            # RSI over 90 days
volume(INFY)[1w]                             # Volume last 7 days
sma(RELIANCE, 50)[200d]                      # 50-SMA over 200 days

# ── Aggregation & Math ───────────────────────────────────────────────
avg(price(RELIANCE)[30d])                    # 30-day average price
max(rsi(TCS, 14)[90d])                       # Max RSI in 90 days
stddev(returns(INFY)[252d])                  # Annualized volatility
change_pct(price(RELIANCE), 30d)             # 30-day % change
correlation(price(TCS), price(INFY), 90d)    # 90-day price correlation

# ── Technical Indicator Functions ────────────────────────────────────
sma(RELIANCE, 200)                           # Simple Moving Average
ema(TCS, 21)                                 # Exponential Moving Average
vwap(INFY)                                   # VWAP (intraday)
bollinger(RELIANCE, 20, 2)                   # Bollinger Bands
supertrend(TCS, 7, 3)                        # SuperTrend
atr(RELIANCE, 14)                            # Average True Range
macd(INFY, 12, 26, 9)                        # MACD with custom params

# ── Fundamental Functions ────────────────────────────────────────────
pe(TCS)                                      # Price-to-Earnings
pb(HDFCBANK)                                 # Price-to-Book
eve_ebitda(RELIANCE)                         # EV/EBITDA
roe(TCS)                                     # Return on Equity
roce(INFY)                                   # Return on Capital Employed
debt_equity(TATAMOTORS)                      # Debt-to-Equity ratio
dividend_yield(ITC)                          # Dividend Yield
promoter_holding(RELIANCE)                   # Promoter holding %

# ── Screening / Filtering ───────────────────────────────────────────
screener(pe < 15 AND roe > 20 AND market_cap > 10000cr)      # Value picks
screener(rsi(*, 14) < 30 AND sector == "IT")                 # Oversold IT
screener(sma(*, 50) > sma(*, 200) AND volume_avg(*, 20) > 1000000)  # Golden cross

# ── Pipes / Composition ─────────────────────────────────────────────
price(RELIANCE)[90d] | sma(20) | trend()     # Pipe: get data → apply SMA → get trend
screener(pe < 20) | sort(roe, desc) | top(10)  # Screen → sort → top 10
nifty50() | where(rsi(*, 14) < 30) | sort(change_pct(*, 1d))   # Oversold Nifty 50

# ── Alerts ───────────────────────────────────────────────────────────
alert(rsi(RELIANCE, 14) < 30, "RELIANCE oversold")           # RSI alert
alert(price(TCS) > 4500, "TCS breakout above 4500")          # Price alert
alert(crossover(sma(INFY, 50), sma(INFY, 200)), "INFY golden cross")  # Crossover
```

#### Step 9.1 — Language Design & Specification
- Define formal grammar (BNF/EBNF) for FinanceQL
- Data types: `Scalar` (single value), `Vector` (time-series), `Matrix` (multi-stock), `Range` (time range selector)
- Operators: arithmetic (`+`, `-`, `*`, `/`), comparison (`>`, `<`, `>=`, `<=`, `==`, `!=`), logical (`AND`, `OR`, `NOT`), pipe (`|`)
- Function categories: price, technical, fundamental, screening, aggregation, alert
- Document syntax in `docs/financeql.md` with comprehensive examples

#### Step 9.2 — Lexer & Parser
- `internal/financeql/lexer.go`: Hand-written tokenizer (identifiers, numbers, strings, operators, brackets, pipes)
- `internal/financeql/parser.go`: Recursive descent parser → AST
- `internal/financeql/ast.go`: AST node types — `FunctionCall`, `RangeSelector`, `BinaryExpr`, `PipeExpr`, `ScreenerExpr`, `AlertExpr`
- Full error reporting with line/column info and suggestions

#### Step 9.3 — Evaluator & Data Resolution
- `internal/financeql/evaluator.go`: Walk AST → resolve data from `DataSource` layer → compute results
- Auto-resolve tickers to NSE symbols (append `.NS`)
- Cache-aware: reuse cached market data within TTL
- Concurrent evaluation: range queries fire parallel goroutines for each time point
- Result types map to Go types: `Scalar → float64`, `Vector → []TimePoint`, `Matrix → map[string][]TimePoint`

#### Step 9.4 — Built-in Functions Library
- `internal/financeql/functions.go`: Register all built-in functions:
  - **Price**: `price()`, `open()`, `high()`, `low()`, `close()`, `volume()`, `returns()`
  - **Technical**: `sma()`, `ema()`, `rsi()`, `macd()`, `bollinger()`, `supertrend()`, `atr()`, `vwap()`
  - **Fundamental**: `pe()`, `pb()`, `roe()`, `roce()`, `debt_equity()`, `market_cap()`, `dividend_yield()`, `promoter_holding()`, `eve_ebitda()`
  - **Aggregation**: `avg()`, `sum()`, `min()`, `max()`, `stddev()`, `percentile()`, `change_pct()`, `correlation()`
  - **Screening**: `screener()`, `where()`, `sort()`, `top()`, `bottom()`, `nifty50()`, `niftybank()`, `sector()`
  - **Alerts**: `alert()`, `crossover()`, `crossunder()`
- Extensible: users can register custom functions via plugin system

#### Step 9.5 — Interactive REPL
- `internal/financeql/repl.go`: Interactive FinanceQL shell
- Features: tab-completion (tickers, functions, keywords), syntax highlighting, history, multi-line input
- Output formatting: tables for vectors, sparklines for time-series, colored signals
- Use `chzyer/readline` for readline support

#### Step 9.6 — LLM ↔ FinanceQL Translation
- Natural language → FinanceQL: LLM translates user queries to FinanceQL expressions
  ```
  User: "Show me oversold large-cap IT stocks"
  → LLM generates: screener(rsi(*, 14) < 30 AND sector == "IT" AND market_cap > 50000cr)
  ```
- FinanceQL → Natural language: explain query results in plain English
- Register FinanceQL as an LLM tool: agents can compose and execute queries programmatically

---

### Phase 10: CLI & API (Week 11–12)

#### Step 9.1 — CLI Interface
- `cmd/openseai/main.go`: Using `cobra` CLI framework
- Commands:
  ```
  openseai analyze RELIANCE         # Quick single-agent analysis
  openseai analyze RELIANCE --deep  # Multi-agent deep analysis
  openseai technical TCS             # Technical analysis only
  openseai fundamental INFY          # Fundamental analysis only
  openseai fno NIFTY                 # F&O / option chain analysis
  openseai report TCS --pdf          # Generate PDF research report
  openseai backtest --strategy sma_crossover --ticker RELIANCE --from 2023-01-01
  openseai trade                     # Interactive trading mode
  openseai watch RELIANCE TCS INFY   # Real-time watchlist with alerts
  openseai portfolio                 # Portfolio analysis from broker
  openseai chat                      # Free-form chat mode
  openseai query 'rsi(RELIANCE, 14)'                      # FinanceQL instant query
  openseai query 'price(TCS)[30d] | sma(20) | trend()'    # FinanceQL piped query
  openseai query 'screener(pe < 15 AND roe > 20)'         # FinanceQL screening
  openseai query --repl                                    # FinanceQL interactive REPL
  openseai query --nl "oversold IT stocks"                  # Natural language → FinanceQL
  ```

#### Step 9.2 — HTTP API Server
- `api/server.go`: REST API using `gin` or `chi`
- Endpoints:
  - `POST /api/v1/analyze` — run analysis
  - `GET /api/v1/quote/:ticker` — live quote
  - `POST /api/v1/backtest` — run backtest
  - `GET /api/v1/portfolio` — portfolio summary
  - `POST /api/v1/chat` — conversational interface
  - `POST /api/v1/query` — execute FinanceQL query
  - `POST /api/v1/query/explain` — parse and explain a FinanceQL expression
  - `POST /api/v1/query/nl` — natural language → FinanceQL translation + execution
  - `GET /api/v1/alerts` — list active FinanceQL alerts
- WebSocket endpoint for streaming analysis updates

---

### Phase 12: Web UI — Next.js Frontend (Week 12–14)

> **Stack**: Next.js 15 (App Router) + TypeScript + Tailwind CSS + shadcn/ui + Zustand

#### Step 12.1 — Project Scaffolding & Layout
- Initialize Next.js app in `web/`: `npx create-next-app@latest web --typescript --tailwind --app --src-dir`
- Install shadcn/ui: `npx shadcn@latest init` + core components (button, input, card, table, tabs, badge, tooltip, select)
- Set up Zustand store in `web/src/store/` with slices: `chatSlice`, `marketSlice`, `querySlice`
- Create responsive layout:
  - `Sidebar.tsx`: collapsible nav — Dashboard, Charts, Chat, FinanceQL, Portfolio, Screener, Backtest
  - `Header.tsx`: market status indicator (open/closed), global ticker search, theme toggle
  - `ThemeToggle.tsx`: light/dark mode with `next-themes`
- API client (`web/src/lib/api.ts`): typed fetch wrapper pointing to Go backend (`/api/v1/*`)
- WebSocket client (`web/src/lib/ws.ts`): reconnecting WS for streaming data
- Environment config: `NEXT_PUBLIC_API_URL`, `NEXT_PUBLIC_WS_URL`

#### Step 12.2 — TradingView Charts Component
- Install `lightweight-charts` npm package (TradingView's open-source charting)
- **`TradingViewChart.tsx`** — Core chart component:
  - Candlestick series as primary view (OHLCV data from Go backend)
  - Volume histogram as overlay or separate pane (`VolumePane.tsx`)
  - Responsive container with auto-resize on window change
  - Real-time data updates via WebSocket subscription
  - Indian formatting: ₹ price axis, volume in lakhs/crores
- **`ChartToolbar.tsx`** — Toolbar above chart:
  - Timeframe selector: 1m, 5m, 15m, 1h, 1D, 1W, 1M
  - Indicator toggles: SMA, EMA, Bollinger Bands, SuperTrend (rendered as line series overlays)
  - Drawing tools: trendline, horizontal line, Fibonacci retracement
  - Crosshair mode toggle, screenshot/export
- **`IndicatorOverlay.tsx`** — Sub-chart indicator panels:
  - RSI panel (with overbought/oversold zones at 70/30)
  - MACD panel (MACD line, signal line, histogram)
  - Separable panes below the main candlestick chart
- **`ChartLegend.tsx`** — Floating legend showing OHLCV + indicator values at crosshair position
- **`useTradingView.ts`** hook — Manages chart instance lifecycle, data fetching, series updates, and cleanup
- Data flow: `useMarketData(ticker, timeframe)` → REST for historical + WS for live ticks → chart `update()`

#### Step 12.3 — Chat UI Component
- **`ChatPanel.tsx`** — Full chat interface:
  - Scrollable message list with auto-scroll to latest
  - Supports both single-agent (quick queries) and multi-agent (deep analysis) modes
  - Mode toggle: "Quick" (single agent) vs "Deep Analysis" (multi-agent team)
  - Agent activity indicator: shows which agents are currently working
- **`MessageBubble.tsx`** — Message rendering:
  - User messages: right-aligned, simple text
  - Agent messages: left-aligned, with `AgentBadge.tsx` showing role (e.g., "Technical Analyst", "Risk Manager")
  - Markdown rendering for agent responses (tables, lists, bold, code blocks)
  - Inline charts: embed mini TradingView charts when agent references price data
  - Indian number formatting: ₹ values auto-formatted with lakhs/crores
- **`ToolCallCard.tsx`** — Expandable cards showing agent tool invocations:
  - Collapsed: tool name + brief result summary (e.g., "GetQuote(RELIANCE) → ₹2,847.50")
  - Expanded: full tool arguments (JSON) + complete result
  - Visual distinction for different tool categories (data fetch, analysis, broker)
- **`TradeConfirm.tsx`** — Human-in-the-loop trade approval:
  - Modal/inline card when Trade Executor agent proposes an order
  - Shows: ticker, action (BUY/SELL), quantity, price, order type, estimated cost
  - Risk summary: position size % of capital, current exposure
  - Approve / Reject / Modify buttons
  - Timeout: auto-reject if no response in configurable time (default: 60s)
- **`ChatInput.tsx`** — Message input:
  - Ticker autocomplete (type `$` to trigger, e.g., `$RELI` → `RELIANCE`)
  - Slash commands: `/analyze`, `/technical`, `/fno`, `/report`, `/trade`
  - File/image attachment support (for chart screenshots)
  - Send on Enter, Shift+Enter for multiline
- **`StreamingText.tsx`** — Streams LLM responses token-by-token via WebSocket/SSE
- **`useChat.ts`** hook — Manages chat session state, message history, WS connection, and agent events
- Backend integration:
  - `POST /api/v1/chat` for new messages
  - `WS /api/v1/chat/stream` for streaming responses + agent events
  - `POST /api/v1/trade/confirm` for trade approval/rejection

#### Step 12.4 — FinanceQL Explorer (Prometheus-style UI)
- **Modeled after Prometheus UI** — familiar query-execute-visualize workflow:

- **`QueryEditor.tsx`** — FinanceQL query input:
  - Monaco Editor (or CodeMirror 6) with custom FinanceQL language mode (`financeql-lang.ts`):
    - Syntax highlighting: functions (blue), tickers (green), operators (red), numbers (orange), strings (yellow)
    - Autocomplete: function names, ticker symbols (fetched from backend), keywords (`AND`, `OR`, `NOT`)
    - Inline error markers from backend parse errors (red squiggles with tooltip)
    - Bracket matching, auto-close for `()`, `[]`, `""`
  - "Execute" button (or Ctrl+Enter) — sends query to `POST /api/v1/query`
  - "Explain" button — sends to `POST /api/v1/query/explain` for human-readable breakdown
  - "Natural Language" toggle — switch input mode to plain English, auto-translates to FinanceQL via `POST /api/v1/query/nl`
  - Time range selector (like Prometheus): relative (last 30d, 90d, 1y) or absolute (date pickers)
  - Evaluation step/resolution control for range queries

- **`QueryHistory.tsx`** — Recent queries sidebar:
  - Persisted to localStorage + backend
  - Click to re-execute, star to bookmark
  - Query duration + result type indicator (Scalar, Vector, Matrix)

- **Result Display** — Auto-selects best visualization based on result type:
  - **`ResultScalar.tsx`** — Single value display:
    - Large formatted number (e.g., `₹2,847.50`, `RSI: 62.4`, `PE: 19.8`)
    - Context badges: stock name, metric name, timestamp
    - Sparkline mini-chart where applicable
  - **`ResultTable.tsx`** — Tabular data (screening results, multi-stock queries):
    - Sortable columns, pagination
    - Inline sparklines for time-series columns
    - Color-coded cells (green for positive, red for negative)
    - Export to CSV
  - **`ResultChart.tsx`** — Time-series visualization:
    - Uses TradingView lightweight-charts for financial time-series
    - Line chart for single series (e.g., `rsi(TCS, 14)[90d]`)
    - Multi-line overlay for comparisons (e.g., `price(TCS)[90d]` vs `sma(TCS, 50)[90d]`)
    - Stacked area for composition data
    - Automatic axis formatting: ₹ for prices, % for ratios
    - Zoom, pan, crosshair with data point tooltip
  - **Tab layout** (like Prometheus): "Table" | "Graph" tabs to switch views

- **`ExpressionTree.tsx`** — Parsed AST visualization:
  - Collapsible tree view of the parsed FinanceQL expression
  - Shows function calls, arguments, pipes, ranges, operators
  - Useful for debugging complex queries

- **`AlertManager.tsx`** — Active alerts management:
  - List all active `alert()` expressions with status (pending/triggered/expired)
  - Alert history with trigger timestamps
  - Create/edit/delete alerts from UI (generates FinanceQL `alert()` expression)
  - Browser push notifications on alert trigger

- **`useFinanceQL.ts`** hook — Manages query execution, result caching, history, and polling for alerts

#### Step 12.5 — Dashboard & Supporting Pages
- **Dashboard** (`web/src/app/page.tsx`):
  - `MarketOverview.tsx`: Nifty 50, Bank Nifty, India VIX live cards with change %
  - `Watchlist.tsx`: user-configurable watchlist with real-time prices, sparklines, signals
  - `FIIDIIBar.tsx`: today's FII/DII net buy/sell bar chart
  - `TopMovers.tsx`: top gainers/losers from Nifty 50
  - Quick FinanceQL query bar embedded at top
- **Portfolio Page** (`web/src/app/portfolio/page.tsx`):
  - Holdings table with current value, P&L, allocation %
  - Position summary and margin utilization
  - Performance chart (portfolio value over time vs Nifty 50)
- **Screener Page** (`web/src/app/screener/page.tsx`):
  - Pre-built FinanceQL screener queries (value picks, momentum, oversold)
  - Custom filter builder that generates FinanceQL `screener()` expressions
  - Results grid with drill-down to individual stock analysis
- **Backtest Page** (`web/src/app/backtest/page.tsx`):
  - Strategy selector, parameter config, date range
  - Equity curve chart, drawdown chart, benchmark overlay
  - Metrics dashboard: Sharpe, CAGR, max drawdown, win rate

#### Step 12.6 — Real-time & WebSocket Integration
- Go backend WebSocket endpoints:
  - `WS /api/v1/ws/market` — live price ticks for subscribed tickers
  - `WS /api/v1/ws/chat` — streaming chat responses + agent events
  - `WS /api/v1/ws/alerts` — real-time alert notifications
- `useWebSocket.ts` hook: auto-reconnect, heartbeat, subscription management
- Optimistic UI updates for trade confirmations

#### Step 12.7 — Build & Deployment
- `web/Dockerfile`: multi-stage build (Node build → `nginx` or `next start`)
- `docker-compose.yaml`: orchestrate Go backend + Next.js frontend + optional Redis for WS pub/sub
- `Makefile` targets: `ui-dev` (Next dev server), `ui-build` (production build), `ui-test`, `ui-lint`
- CORS configuration in Go backend for local dev (`localhost:3000` → `localhost:8080`)
- Production: Next.js serves static + SSR, Go backend as API

---

### Phase 13: Testing & Documentation (Week 14–15)

#### Step 13.1 — Testing
- **Unit tests**: Every analysis function with known inputs/outputs (e.g., RSI of known price series)
- **Integration tests**: Data source connectivity (NSE, YFinance, Screener)
- **Agent tests**: Mock LLM responses, verify tool selection and orchestration
- **Backtest validation**: Compare results against known strategy outcomes
- **Paper trading tests**: End-to-end order flow in paper mode
- **FinanceQL tests**: Lexer/parser unit tests, evaluator tests with mock data, REPL integration tests, NL→FinanceQL accuracy tests
- **Benchmark tests**: Performance benchmarks for data fetching and analysis
- **Frontend tests**:
  - Component tests: Vitest + React Testing Library for all components
  - E2E tests: Playwright for critical flows (chart rendering, chat, FinanceQL query execution)
  - Visual regression: Chromatic or Percy for UI consistency
  - WebSocket mock tests: verify real-time update rendering
  - Accessibility: axe-core audit for WCAG 2.1 AA compliance

#### Step 13.2 — Documentation
- `README.md`: Quick start, installation, usage examples
- `docs/architecture.md`: System design, data flow diagrams
- `docs/agents.md`: Agent roles, capabilities, prompt engineering
- `docs/data-sources.md`: Available data sources, rate limits, caching
- `docs/financeql.md`: FinanceQL language reference, syntax, built-in functions, examples, REPL usage
- `docs/trading-safety.md`: Safety guardrails, risk management, disclaimers
- `docs/web-ui.md`: Web UI component reference, design system, page layouts, WebSocket protocol

---

## Key Dependencies

### Go Modules

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/spf13/viper` | Configuration management |
| `github.com/go-chi/chi` or `github.com/gin-gonic/gin` | HTTP server |
| `github.com/sashabaranov/go-openai` | OpenAI API client |
| `github.com/ollama/ollama/api` | Ollama local LLM client |
| `github.com/google/generative-ai-go` | Gemini API client |
| `github.com/PuerkitoBio/goquery` | HTML scraping (NSE, Screener, Moneycontrol) |
| `github.com/mmcdole/gofeed` | RSS feed parsing for news |
| `gonum.org/v1/gonum` | Numerical computing (statistics, matrix ops) |
| `github.com/go-echarts/go-echarts` | Chart generation |
| `github.com/SebastiaanKlipworking/go-wkhtmltopdf` | PDF generation |
| `github.com/gorilla/websocket` | WebSocket for streaming |
| `github.com/mattn/go-sqlite3` | Local storage for trade logs, cache |
| `golang.org/x/sync/errgroup` | Concurrent data fetching |
| `github.com/chzyer/readline` | REPL readline support (FinanceQL interactive mode) |
| `github.com/alecthomas/participle` | Parser combinator (optional, for FinanceQL grammar) |

### Frontend (npm — `web/`)

| Package | Purpose |
|---------|---------|
| `next` (v15) | React framework with App Router, SSR, API routes |
| `react` / `react-dom` (v19) | UI library |
| `typescript` | Type safety across all frontend code |
| `tailwindcss` | Utility-first CSS framework |
| `@shadcn/ui` | Accessible, composable UI component primitives |
| `lightweight-charts` | TradingView open-source financial charting library |
| `@monaco-editor/react` | Monaco Editor for FinanceQL query editor with syntax highlighting |
| `zustand` | Lightweight state management (global store with slices) |
| `next-themes` | Dark/light theme toggle with system preference |
| `react-markdown` + `remark-gfm` | Markdown rendering for agent chat responses |
| `lucide-react` | Icon library (consistent with shadcn/ui) |
| `date-fns` | Date formatting and manipulation (IST support) |
| `vitest` + `@testing-library/react` | Unit & component testing |
| `playwright` | End-to-end browser testing |
| `class-variance-authority` + `clsx` | Conditional CSS class utilities (shadcn/ui dependency) |

---

## API Keys & Configuration

```yaml
# config/config.example.yaml
llm:
  primary: openai          # openai | ollama | gemini | anthropic
  openai_key: ""           # env: OPENSEAI_OPENAI_KEY
  ollama_url: "http://localhost:11434"
  model: "gpt-4o"          # or "qwen2.5:32b" for Ollama

broker:
  provider: paper          # paper | zerodha | ibkr
  zerodha:
    api_key: ""            # env: OPENSEAI_ZERODHA_KEY
    api_secret: ""         # env: OPENSEAI_ZERODHA_SECRET
  ibkr:
    host: "127.0.0.1"
    port: 7497

trading:
  mode: paper              # paper | live
  max_position_pct: 5.0    # max 5% of capital per trade
  daily_loss_limit_pct: 2.0
  max_open_positions: 10
  require_confirmation: true  # human-in-the-loop for live trades

analysis:
  cache_ttl: 300           # 5 min cache for market data
  concurrent_fetches: 5    # parallel goroutines for data fetching

financeql:
  cache_ttl: 60            # 1 min cache for FinanceQL query results
  max_range: 365d          # max range selector (1 year)
  alert_check_interval: 30 # alert re-evaluation interval in seconds
  repl_history_file: "~/.openseai/financeql_history"
```

---

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | **Go** | Concurrency for parallel data fetching, single binary deployment, low-latency for trading |
| NSE data | **Direct NSE scraping + YFinance** | Go needs native HTTP scraping of nseindia.com |
| Technical analysis | **Custom Go implementation** | No mature Go TA library; implement core indicators (RSI, MACD, BB, SuperTrend) from formulas |
| LLM framework | **Custom agent framework** | No Go equivalent of AutoGen; build lightweight agent loop (LLM → tool → LLM) |
| Broker primary | **Zerodha Kite** | Most popular Indian retail broker, well-documented REST API |
| Trading default | **Paper trading** | Safety first — all users start in simulation mode |
| Human-in-the-loop | **Mandatory for live** | Every live order requires explicit confirmation |
| Number formatting | **Indian system** | ₹ symbol, lakhs (1,00,000), crores (1,00,00,000) throughout |
| Query language | **FinanceQL (PromQL-inspired)** | Composable, expressive DSL for financial data; familiar syntax for monitoring/DevOps users; enables NL→FinanceQL via LLM |
| Web UI framework | **Next.js 15 + TypeScript** | App Router for file-based routing, SSR for SEO, React Server Components for performance, strong typing |
| Charting library | **TradingView lightweight-charts** | Open-source, high-performance, purpose-built for financial data; candlestick, volume, overlays |
| FinanceQL UI | **Prometheus-style Explorer** | Familiar query→execute→visualize pattern for DevOps/SRE users; Monaco Editor for rich editing experience |
| Component library | **shadcn/ui + Tailwind CSS** | Copy-paste composable components, no runtime CSS-in-JS overhead, full customization |
| State management | **Zustand** | Minimal boilerplate, TypeScript-native, slice pattern for modular state |
| Config format | **YAML + env vars** | Human-readable config with secure env var override for secrets |

---

## Risk & Safety

- **DISCLAIMER**: OpeNSE.ai is for educational and research purposes. Not financial advice.
- All AI-generated recommendations include confidence scores and reasoning
- Live trading requires: (1) explicit opt-in, (2) valid broker credentials, (3) human confirmation per order
- Daily loss circuit breaker auto-disables trading
- All trades logged with full audit trail
- Paper trading mode for strategy validation before going live

---

## Milestones

| Week | Milestone | Deliverable |
|------|-----------|-------------|
| 1–2 | Foundation | Project structure, models, config, utilities |
| 2–3 | Data Layer | All 5 data sources working with tests |
| 3–5 | Analysis | Technical + Fundamental + F&O + Sentiment engines |
| 4–5 | LLM Layer | Multi-provider LLM with tool calling |
| 5–7 | Agents | All 7 agents + orchestrator (single + multi mode) |
| 7–8 | Trading | Broker integration + paper trading |
| 8–9 | Reports | PDF generation with charts |
| 9–10 | Backtest | Backtesting engine with strategies |
| 10–11 | FinanceQL | Query language: lexer, parser, evaluator, REPL, LLM integration |
| 11–12 | Interface | CLI + HTTP API |
| 12–14 | Web UI | Next.js frontend: TradingView charts, Chat UI, FinanceQL Explorer, Dashboard |
| 14–15 | Polish | Tests (backend + frontend), docs, README, examples |
