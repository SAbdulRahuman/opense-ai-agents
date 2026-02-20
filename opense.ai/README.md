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
- **Single Binary** — Web UI is embedded into the Go binary via `go:embed`; one file serves everything

## Prerequisites

| Tool       | Version | Purpose                                |
| ---------- | ------- | -------------------------------------- |
| **Go**     | 1.23+   | Backend, CLI, embedded web server      |
| **Node.js**| 20+     | Build the Next.js web frontend         |
| **npm**    | 9+      | Frontend dependency management         |
| **Git**    | any     | Source control                         |
| **Docker** | 24+     | *(optional)* Containerised deployment  |

## Quick Start

```bash
# Clone
git clone https://github.com/seenimoa/openseai.git
cd openseai

# Install frontend dependencies
cd web && npm ci && cd ..

# Configure
cp config/config.example.yaml config/config.yaml
# Edit config/config.yaml with your LLM API key

# Build (compiles web + Go into a single binary)
make build

# Start server — web UI at http://localhost:8080/, API at http://localhost:8080/api/v1
./build/openseai serve
```

## Building

### Build from Source (single binary)

`make build` compiles the Next.js frontend into a static export (`web/out/`), then builds the Go binary with the web assets embedded via `//go:embed`.

```bash
# Full build: web frontend + Go binary
make build

# The resulting binary is at ./build/openseai (~17 MB)
ls -lh build/openseai
```

You can also build each part independently:

```bash
# Build only the web frontend (produces web/out/)
make build-web

# Build only the Go binary (assumes web/out/ already exists)
make build-go
```

### Build with Docker

The multi-stage Dockerfile builds both the frontend and backend, producing a single container image (~30 MB) that serves the web UI and API from one port.

```bash
# Build the Docker image
docker build -t openseai .

# Or use docker compose
docker compose build
```

## Running

### Run the Binary

```bash
# Start server with embedded web UI (default port 8080)
./build/openseai serve

# Custom host/port
./build/openseai serve --host 0.0.0.0 --port 9090

# API-only mode (no web UI)
./build/openseai serve --no-ui

# Check system status
./build/openseai status

# Print version
./build/openseai version
```

Once the server is running:

| URL                              | Description              |
| -------------------------------- | ------------------------ |
| `http://localhost:8080/`         | Web UI (dashboard)       |
| `http://localhost:8080/charts/`  | TradingView charts       |
| `http://localhost:8080/chat/`    | Chat interface           |
| `http://localhost:8080/api/v1/health` | API health check    |
| `http://localhost:8080/api/v1/`  | REST API root            |

### Run with Docker

```bash
# Start the container (web UI + API on port 8080)
docker compose up -d

# View logs
docker compose logs -f

# Stop
docker compose down
```

Or run the image directly:

```bash
docker run -d --name openseai \
  -p 8080:8080 \
  -v ./config:/app/config:ro \
  --env-file .env \
  openseai
```

### Environment Variables

Sensitive values can be set via environment variables instead of the config file:

```bash
export OPENSEAI_LLM_OPENAI_KEY="sk-..."
export OPENSEAI_BROKER_ZERODHA_API_KEY="..."
export OPENSEAI_BROKER_ZERODHA_API_SECRET="..."
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
openseai serve                    # Start HTTP API + web UI server
openseai status                   # Show system status
openseai version                  # Print version info
```

## Testing

### Go Backend Tests

```bash
# Run all tests with race detection and coverage
make test

# Quick run without race detector
make test-short

# Generate HTML coverage report
make coverage
open coverage.html

# Run benchmarks (technical analysis + FinanceQL)
make bench
```

### Frontend Tests

```bash
# Run all vitest unit tests (86 tests across 16 files)
make ui-test

# Watch mode (re-run on file changes)
cd web && npm run test:watch

# With coverage report
cd web && npm run test:coverage
```

### End-to-End Tests (Playwright)

```bash
# Install browsers (first time only)
cd web && npx playwright install chromium

# Run E2E tests
make e2e

# Run headed (see the browser)
make e2e-headed
```

### Smoke Tests

```bash
# Verify the server is working after build
make build
./build/openseai serve &
sleep 2

# Health check
curl http://localhost:8080/health
# → {"success":true,"data":{"status":"ok",...}}

# API endpoint
curl http://localhost:8080/api/v1/alerts
# → {"success":true,"data":[]}

# Web UI
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/
# → 200

# Clean up
kill %1
```

### NSE Data Source Smoke Tests

```bash
# Test connectivity to NSE and Yahoo Finance APIs
bash scripts/test_nse.sh
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

The web UI is compiled into a static export and embedded into the Go binary at build time using `//go:embed`. The `serve` command starts a single HTTP server that serves both the SPA frontend at `/` and the REST API at `/api/v1`.

## Project Structure

```
opense.ai/
├── cmd/openseai/          # CLI entrypoint (Cobra)
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
├── api/                   # HTTP API server + SPA handler
├── web/                   # Next.js frontend
│   ├── embed.go           # go:embed directive for web/out/
│   └── out/               # Static export (build artifact, gitignored)
├── config/                # Configuration files
├── docs/                  # Documentation
├── scripts/               # Dev scripts (setup, NSE smoke tests)
└── Dockerfile             # Multi-stage build (single binary container)
```

## Development

```bash
make build       # Build web + Go binary (full production build)
make build-go    # Build Go binary only (when frontend hasn't changed)
make build-web   # Build frontend only (produces web/out/)
make test        # Run all Go tests with race detection
make ui-test     # Run frontend vitest tests
make bench       # Run Go benchmarks
make lint        # Run golangci-lint
make fmt         # Format Go code
make vet         # Run go vet
make tidy        # Tidy go.mod
make dev         # Run Go with hot reload (requires air)
make ui-dev      # Start Next.js dev server (http://localhost:3000)
make e2e         # Run Playwright E2E tests
make clean       # Remove build/ and web/out/
make docker      # Build Docker image
make docker-up   # Start with docker compose
make docker-down # Stop docker compose
make setup       # Run dev environment setup script
make help        # Show all targets
```

### Development Workflow (split mode)

For frontend development, run the Go API and Next.js dev server separately:

```bash
# Terminal 1: Go API
./build/openseai serve --no-ui -p 8080

# Terminal 2: Next.js dev server (with API proxy)
cd web && NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1 npm run dev
```

## License

MIT

---

*Built with Go, Next.js, and ❤️ for the Indian stock market.*
