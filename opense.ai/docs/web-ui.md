# OpeNSE.ai — Web UI

> Component reference and design system for the Next.js frontend.

## Tech Stack

| Technology | Version | Purpose |
|-----------|---------|---------|
| Next.js | 16.x | React framework (App Router) |
| React | 19.x | UI library |
| TypeScript | 5.x | Type safety |
| Tailwind CSS | 4.x | Utility-first styling |
| Zustand | latest | State management |
| Recharts | latest | Charts and data visualization |
| Vitest | 4.x | Unit testing |
| Playwright | latest | E2E testing |

## Application Routes

| Route | Page | Description |
|-------|------|-------------|
| `/` | Dashboard | Market overview, key indices, top movers |
| `/charts` | Charts | Interactive stock charts with indicators |
| `/chat` | Chat | AI conversation interface |
| `/financeql` | FinanceQL | Query REPL with syntax highlighting |
| `/portfolio` | Portfolio | Holdings, PnL tracking, alerts |
| `/screener` | Screener | Multi-criterion stock filtering |
| `/backtest` | Backtest | Strategy testing with equity curves |

## Component Architecture

```
src/
├── app/                    # App Router pages
│   ├── layout.tsx          # Root layout (Sidebar + TopBar)
│   ├── page.tsx            # Dashboard
│   ├── charts/page.tsx
│   ├── chat/page.tsx
│   ├── financeql/page.tsx
│   ├── portfolio/page.tsx
│   ├── screener/page.tsx
│   └── backtest/page.tsx
├── components/
│   ├── layout/             # App shell components
│   │   ├── Sidebar.tsx     # Navigation sidebar
│   │   └── TopBar.tsx      # Header with search, theme toggle
│   ├── dashboard/          # Dashboard-specific components
│   │   ├── MarketOverview.tsx
│   │   ├── IndexCards.tsx
│   │   └── TopMovers.tsx
│   ├── chart/              # Chart components
│   │   ├── StockChart.tsx  # Main interactive chart
│   │   ├── OHLCVChart.tsx  # Candlestick chart
│   │   └── IndicatorOverlay.tsx
│   ├── chat/               # Chat components
│   │   ├── ChatInterface.tsx
│   │   ├── MessageBubble.tsx
│   │   └── ChatInput.tsx
│   ├── financeql/          # FinanceQL components
│   │   ├── QueryEditor.tsx # Code editor with highlighting
│   │   ├── ResultsPane.tsx # Query results display
│   │   └── FunctionDocs.tsx
│   └── ui/                 # Shared UI primitives
│       ├── Button.tsx
│       ├── Card.tsx
│       ├── Input.tsx
│       ├── Badge.tsx
│       ├── Tooltip.tsx
│       └── LoadingSpinner.tsx
├── store/                  # Zustand stores
│   ├── useMarketStore.ts   # Market data state
│   ├── useChatStore.ts     # Chat messages and sessions
│   ├── usePortfolioStore.ts
│   └── useSettingsStore.ts
├── lib/                    # Utilities
│   ├── api.ts              # API client (fetch wrapper)
│   ├── format.ts           # Indian number formatting
│   └── constants.ts        # App-wide constants
└── hooks/                  # Custom hooks
    ├── useWebSocket.ts     # Real-time data
    ├── useDebounce.ts
    └── useLocalStorage.ts
```

## UI Components

### Layout

**Sidebar** (`components/layout/Sidebar.tsx`):
- Collapsible navigation with icons + labels
- Active route highlighting
- Links: Dashboard, Charts, Chat, FinanceQL, Portfolio, Screener, Backtest

**TopBar** (`components/layout/TopBar.tsx`):
- Stock search with autocomplete
- Theme toggle (light/dark)
- Connection status indicator

### Dashboard Components

**MarketOverview**: Key market indices (Nifty 50, Bank Nifty, India VIX) with real-time updates.

**IndexCards**: Card grid showing index values, change %, daily range bars.

**TopMovers**: Tables for top gainers/losers with ticker, price, change, volume.

### Chart Components

**StockChart**: Interactive chart with:
- Candlestick / line / area chart types
- Timeframe selector (1D, 1W, 1M, 3M, 6M, 1Y, 5Y)
- Technical indicator overlays (SMA, EMA, Bollinger)
- Volume bars
- Support/resistance level lines

### Chat Components

**ChatInterface**: AI conversation with:
- Message history with user/assistant bubbles
- Inline stock cards and chart previews
- Analysis result rendering
- Markdown support in responses

**ChatInput**: Text input with:
- Send button and keyboard shortcut (Enter)
- File attachment for screenshots
- Quick action buttons

### FinanceQL Components

**QueryEditor**: Code editor with:
- Syntax highlighting for FinanceQL tokens
- Auto-complete for function names and tickers
- Query history (up/down arrow navigation)
- Multi-line query support

**ResultsPane**: Results display with:
- Scalar values with formatting
- Vector data as interactive charts
- Boolean results with visual indicators
- Error messages with suggestions

## State Management

### Zustand Stores

```typescript
// Market data
useMarketStore: {
  indices: IndexData[]
  watchlist: Quote[]
  selectedTicker: string | null
  fetchIndices: () => Promise<void>
  setSelectedTicker: (ticker: string) => void
}

// Chat
useChatStore: {
  messages: Message[]
  isLoading: boolean
  sendMessage: (content: string) => Promise<void>
  clearMessages: () => void
}

// Portfolio
usePortfolioStore: {
  holdings: Holding[]
  positions: Position[]
  totalPnL: number
  fetchPortfolio: () => Promise<void>
}
```

## API Client

The API client (`lib/api.ts`) communicates with the Go backend:

```typescript
const api = {
  // Stock data
  getQuote: (ticker: string) => fetch(`/api/v1/quotes/${ticker}`),
  getOHLCV: (ticker: string, tf: string, range: string) => ...,
  
  // Analysis
  analyze: (ticker: string) => fetch(`/api/v1/analyze/${ticker}`),
  
  // FinanceQL
  query: (q: string) => fetch('/api/v1/query', { body: { query: q } }),
  
  // Chat
  chat: (message: string) => fetch('/api/v1/chat', { body: { message } }),
  
  // Portfolio
  getPortfolio: () => fetch('/api/v1/portfolio'),
}
```

## Indian Number Formatting

The `lib/format.ts` module handles Indian conventions:

```typescript
formatINR(1234567)    // "₹12,34,567"
formatCrores(192734500000) // "₹19,273.45 Cr"
formatLakhs(500000)   // "₹5.00L"
formatPct(15.5)       // "15.50%"
formatDate(date)      // "19-Feb-2026"
```

## Theming

- **Dark mode** (default): Dark backgrounds, green/red for gain/loss
- **Light mode**: White backgrounds, matching color scheme
- Theme persisted in localStorage
- Tailwind CSS dark mode via class strategy

## Testing

### Unit Tests (Vitest)

```bash
cd web && npm test           # Run all tests
cd web && npm test -- --ui   # Interactive UI
```

Test locations: `src/__tests__/` mirroring the component structure.

### E2E Tests (Playwright)

```bash
cd web && npx playwright test
```

E2E tests cover:
- Page navigation and rendering
- Chart rendering and interaction
- Chat message flow
- FinanceQL query execution
- Portfolio data display

## Development

```bash
# Start frontend dev server
cd web && npm run dev        # http://localhost:3000

# Start backend API (separate terminal)
make serve                   # http://localhost:8080

# Or use make targets
make ui-dev                  # Frontend only
make dev                     # Backend only
```
