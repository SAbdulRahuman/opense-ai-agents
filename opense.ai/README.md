# OpeNSE.ai

> **OpeNSE.ai** (Open + NSE + Agentic AI) — A Go-based multi-agent AI system for comprehensive NSE (National Stock Exchange of India) stock analysis, covering fundamental, technical, derivatives, sentiment analysis, and automated trading.

⚠️ **DISCLAIMER**: OpeNSE.ai is for educational and research purposes only. Not financial advice. Always do your own research before making investment decisions.

## Features

- **Multi-Agent AI System** — Specialized agents for fundamental, technical, derivatives, sentiment, and risk analysis
- **FinanceQL** — PromQL-inspired query language for financial data (`rsi(RELIANCE, 14)`, `screener(pe < 15 AND roe > 20)`)
- **TradingView Charts** — Interactive charts powered by TradingView's lightweight-charts
- **Chat Interface** — Conversational AI with agent transparency and human-in-the-loop trade confirmation
- **Broker Integration** — Zerodha Kite + Interactive Brokers (paper trading by default)
- **Indian Market Focus** — NSE/BSE data, ₹ formatting (lakhs/crores), IST timezone, trading holidays

## Quick Start

### Prerequisites

- Go 1.23+
- Node.js 20+ (for web UI)
- An LLM API key (OpenAI, Gemini, Anthropic, or local Ollama)

### Install & Run

```bash
# Clone
git clone https://github.com/seenimoa/openseai.git
cd openseai

# Configure
cp config/config.example.yaml config/config.yaml
# Edit config/config.yaml with your API keys

# Build & run
make build
./build/openseai status        # check system status
./build/openseai analyze RELIANCE
./build/openseai technical TCS
./build/openseai query 'rsi(RELIANCE, 14)'
```

### Environment Variables

Sensitive values can be set via environment variables instead of the config file:

```bash
export OPENSEAI_LLM_OPENAI_KEY="sk-..."
export OPENSEAI_BROKER_ZERODHA_API_KEY="..."
export OPENSEAI_BROKER_ZERODHA_API_SECRET="..."
```

### Web UI

```bash
cd web
npm install
npm run dev    # → http://localhost:3000
```

### Docker

```bash
docker compose up -d   # starts Go API + Next.js frontend
```

## CLI Commands

```
openseai analyze RELIANCE         # Quick single-agent analysis
openseai analyze RELIANCE --deep  # Multi-agent deep analysis
openseai technical TCS            # Technical analysis only
openseai fundamental INFY         # Fundamental analysis only
openseai fno NIFTY                # F&O / option chain analysis
openseai query 'rsi(TCS, 14)'    # FinanceQL instant query
openseai query --repl             # FinanceQL interactive REPL
openseai chat                     # Free-form chat mode
openseai serve                    # Start HTTP API server
openseai status                   # Show system status
openseai version                  # Print version info
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  OpeNSE.ai Web UI (Next.js)                 │
│     TradingView Charts  │  Chat UI  │  FinanceQL Explorer   │
├─────────────────────────────────────────────────────────────┤
│                     CLI / REST API (Go)                     │
├─────────────────────────────────────────────────────────────┤
│  FinanceQL Engine  │  Agent Orchestrator  │  LLM Gateway    │
├─────────────────────────────────────────────────────────────┤
│  NSE Data  │  Technical  │  F&O  │  Sentiment  │  Broker   │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
opense.ai/
├── cmd/openseai/          # CLI entrypoint
├── internal/
│   ├── agent/             # AI agents (fundamental, technical, sentiment, F&O, risk)
│   ├── analysis/          # Analysis engines (technical, fundamental, derivatives, sentiment)
│   ├── broker/            # Broker integration (Zerodha, IBKR, paper)
│   ├── config/            # Configuration system
│   ├── datasource/        # Data sources (YFinance, NSE, news, Screener.in)
│   ├── financeql/         # FinanceQL query language (lexer, parser, evaluator)
│   ├── llm/               # LLM provider abstraction (OpenAI, Ollama, Gemini, Anthropic)
│   └── report/            # PDF report generation
├── pkg/
│   ├── models/            # Shared data models
│   └── utils/             # Utilities (Indian formatting, tickers, time)
├── api/                   # HTTP API server
├── web/                   # Next.js frontend
├── config/                # Configuration files
└── docs/                  # Documentation
```

## Development

```bash
make build       # Build binary
make test        # Run all tests
make lint        # Run linter
make fmt         # Format code
make dev         # Run with hot reload (requires air)
make ui-dev      # Start Next.js dev server
make help        # Show all targets
```

## License

MIT

---

*Built with Go, Next.js, and ❤️ for the Indian stock market.*
