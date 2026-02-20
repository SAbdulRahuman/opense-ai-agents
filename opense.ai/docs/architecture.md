# OpeNSE.ai — Architecture

> System architecture, data flow, and design decisions for the OpeNSE.ai AI-powered Indian stock analysis platform.

## High-Level Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────────────┐
│   Next.js    │────▶│  Go REST API │────▶│   Agent Orchestrator  │
│  Frontend    │◀────│  (port 8080) │◀────│   (CIO → N agents)   │
│  (port 3000) │     └──────┬───────┘     └──────────┬───────────┘
└──────────────┘            │                        │
                            │                        ▼
                     ┌──────▼───────┐     ┌──────────────────────┐
                     │   FinanceQL  │     │    LLM Providers     │
                     │  Query Engine│     │ OpenAI│Gemini│Ollama  │
                     └──────┬───────┘     │ Anthropic             │
                            │             └──────────────────────┘
                     ┌──────▼───────┐
                     │  Data Sources │
                     │  (NSE/Yahoo)  │
                     └──────────────┘
```

## Project Structure

```
opense.ai/
├── cmd/openseai/          # CLI entrypoint (Cobra, 14 commands)
├── api/                   # REST API server (Gin framework)
├── internal/
│   ├── agent/             # Multi-agent orchestration
│   │   └── prompts/       # System prompts, CoT templates, Indian market context
│   ├── analysis/
│   │   ├── technical/     # RSI, MACD, Bollinger, SuperTrend, S/R, patterns
│   │   ├── fundamental/   # Financial ratios, growth, valuation
│   │   ├── derivatives/   # Option chain, OI analysis, PCR, max pain
│   │   └── sentiment/     # News sentiment, market mood
│   ├── backtest/          # Strategy backtesting engine
│   ├── broker/            # Broker integrations (Paper, Zerodha, IBKR)
│   ├── config/            # Configuration (Viper, YAML + env vars)
│   ├── datasource/        # Data aggregator, NSE/Yahoo adapters
│   ├── financeql/         # FinanceQL query language (lexer→parser→evaluator)
│   ├── llm/               # LLM provider abstraction
│   └── report/            # Report generation with Go templates
├── pkg/
│   ├── models/            # Shared data types (Stock, Order, OHLCV, Analysis)
│   └── utils/             # Utility functions (formatting, validation)
├── web/                   # Next.js 16 frontend
│   ├── src/app/           # App Router pages (7 routes)
│   ├── src/components/    # React components (chart, chat, dashboard, UI)
│   ├── src/store/         # Zustand state management
│   ├── src/lib/           # API client, data formatting
│   └── src/hooks/         # Custom React hooks
├── config/                # Configuration files (YAML)
├── docs/                  # Documentation
└── scripts/               # Development & deployment scripts
```

## Component Details

### 1. CLI Layer (`cmd/openseai`)

Single-binary Go CLI built with Cobra. Commands:

| Command | Purpose |
|---------|---------|
| `analyze` | Run comprehensive multi-agent analysis |
| `technical` | Technical analysis only |
| `fundamental` | Fundamental analysis only |
| `fno` | F&O / derivatives analysis |
| `report` | Generate equity research report |
| `backtest` | Run strategy backtests |
| `trade` | Execute trades (paper/live) |
| `watch` | Real-time price monitoring |
| `portfolio` | Portfolio management |
| `query` | Execute FinanceQL queries |
| `chat` | Interactive chat mode |
| `serve` | Start API server |
| `status` | System health check |
| `version` | Build info |

### 2. Agent Orchestration (`internal/agent`)

Multi-agent system coordinated by a Chief Investment Officer (CIO) agent:

```
                    ┌─────────┐
                    │   CIO   │  (orchestrator)
                    └────┬────┘
          ┌──────────────┼──────────────┐
          ▼              ▼              ▼
    ┌───────────┐ ┌───────────┐ ┌───────────┐
    │Fundamental│ │ Technical │ │ Sentiment │
    │  Analyst  │ │  Analyst  │ │  Analyst  │
    └───────────┘ └───────────┘ └───────────┘
          ┌──────────────┼──────────────┐
          ▼              ▼              ▼
    ┌───────────┐ ┌───────────┐ ┌───────────┐
    │F&O Analyst│ │Risk Manager│ │  Reporter │
    └───────────┘ └───────────┘ └───────────┘
                         │
                         ▼
                  ┌──────────────┐
                  │Trade Executor│
                  └──────────────┘
```

Each agent has:
- **System prompt**: Domain expertise definition (see `prompts/system.go`)
- **CoT template**: Step-by-step reasoning framework (see `prompts/cot.go`)
- **Tool access**: Functions it can call via the LLM function-calling API

### 3. Analysis Engine (`internal/analysis`)

Four specialized analysis modules:

| Module | Key Functions | Data Dependencies |
|--------|--------------|-------------------|
| **Technical** | RSI, MACD, Bollinger, SuperTrend, S/R, patterns, signals | OHLCV price data |
| **Fundamental** | Ratios, DCF, growth rates, peer comparison | Financial statements |
| **Derivatives** | Option chain, OI analysis, PCR, max pain, strategies | Live option data |
| **Sentiment** | News scoring, market mood, FII/DII flows | News feeds, flow data |

### 4. FinanceQL Engine (`internal/financeql`)

Custom query language pipeline:

```
Input Query ──▶ Lexer ──▶ Tokens ──▶ Parser ──▶ AST ──▶ Evaluator ──▶ Result
                                                              │
                                                    ┌─────────▼────────┐
                                                    │  40+ built-in    │
                                                    │  functions (SMA,  │
                                                    │  RSI, corr, etc.) │
                                                    └──────────────────┘
```

### 5. Data Layer (`internal/datasource`)

Aggregator pattern with pluggable adapters:

- **NSE adapter**: Direct NSE data (quotes, option chains, deliverables)
- **Yahoo Finance adapter**: Historical OHLCV, financials, fundamentals
- **Caching**: In-memory TTL cache (configurable per source)

### 6. Broker Integration (`internal/broker`)

Provider interface with implementations:

| Provider | Status | Features |
|----------|--------|----------|
| **Paper** | ✅ Production | Simulated trading, PnL tracking |
| **Zerodha** | ✅ Production | Kite Connect API, CNC/MIS/NRML |
| **IBKR** | ✅ Production | Interactive Brokers TWS |

### 7. Web Frontend (`web/`)

Next.js 16 with App Router:

| Route | Purpose |
|-------|---------|
| `/` | Dashboard with market overview |
| `/charts` | Interactive stock charts |
| `/chat` | AI chat interface |
| `/financeql` | FinanceQL query REPL |
| `/portfolio` | Portfolio tracker |
| `/screener` | Stock screener |
| `/backtest` | Strategy backtesting UI |

**Tech stack**: React 19, TypeScript 5, Tailwind CSS 4, Zustand (state), Recharts (charts)

## Data Flow

### Analysis Request

```
1. User → CLI/API/Web: "analyze RELIANCE"
2. CIO agent receives request
3. CIO delegates to specialized agents in parallel:
   - Fundamental → fetches financials, computes ratios
   - Technical → fetches OHLCV, computes indicators
   - Sentiment → fetches news, scores sentiment
   - F&O → fetches option chain, analyzes OI
4. Each agent returns structured AnalysisResult
5. Risk Manager evaluates combined risk
6. CIO synthesizes all results → CompositeAnalysis
7. Reporter generates formatted research report
8. Result returned to user
```

### FinanceQL Query

```
1. User: rsi(close("TCS", "1d", "365d"), 14) < 30
2. Lexer tokenizes → identifiers, strings, numbers, operators
3. Parser builds AST → FunctionCall(rsi, FunctionCall(close, ...), 14) < 30
4. Evaluator:
   a. Resolves close("TCS", ...) → fetches OHLCV from datasource
   b. Computes rsi(..., 14) → calls technical.RSI()
   c. Evaluates < 30 → boolean result
5. Result returned as typed Value
```

## Configuration

Three-layer configuration with precedence:

1. **Defaults** (built-in) → safety-first values
2. **Config file** (`config.yaml`) → project settings
3. **Environment variables** (`OPENSEAI_*`) → secrets & overrides

Key safety defaults:
- Trading mode: `paper` (never live by default)
- Max position: 5% of capital
- Daily loss limit: 2%
- Require confirmation: `true`

## Build & Deployment

```bash
# Development
make dev          # Run with hot-reload
make ui-dev       # Frontend dev server

# Testing
make test         # All Go tests
make ui-test      # Frontend tests
make bench        # Go benchmarks
make e2e          # Playwright E2E tests

# Production
make build        # CGO_ENABLED=0 static binary
make docker       # Docker image
make docker-up    # Docker Compose (API + frontend)
```

## Design Principles

1. **Safety First**: Paper trading by default, mandatory confirmation for live trades
2. **Indian Market Native**: ₹ formatting, NSE conventions, SEBI regulations
3. **Multi-Agent Intelligence**: Specialized agents with domain expertise
4. **Composable Queries**: FinanceQL for flexible data exploration
5. **Offline Capable**: Ollama support for fully local LLM inference
6. **Observable**: Structured logging, health checks, comprehensive testing
