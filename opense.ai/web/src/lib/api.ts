// ============================================================================
// OpeNSE.ai — API Client (typed fetch wrapper for Go backend)
// ============================================================================

import type {
  AnalysisResult,
  BacktestParams,
  BacktestResult,
  ChatMessage,
  Quote,
  QueryResult,
  PortfolioSummary,
  Alert,
  ScreenerResult,
  OHLCV,
  MarketIndex,
  TopMover,
  FIIDIIData,
  AppConfig,
  ConfigResponse,
  KeyStatus,
  Order,
  OrderRequest,
  OrderResponse,
  Position,
  Margins,
} from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "/api/v1";

class APIError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "APIError";
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const url = `${API_BASE}${path}`;
  const res = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  if (!res.ok) {
    const body = await res.text();
    throw new APIError(res.status, body || res.statusText);
  }

  return res.json() as Promise<T>;
}

// --- Health ---

export async function getHealth(): Promise<{ status: string; version: string }> {
  return request("/health");
}

// --- Market Data ---

export async function getQuote(ticker: string): Promise<Quote> {
  return request(`/quote/${encodeURIComponent(ticker)}`);
}

export async function getOHLCV(
  ticker: string,
  timeframe: string = "1D",
  days: number = 365,
): Promise<OHLCV[]> {
  return request(`/ohlcv/${encodeURIComponent(ticker)}?timeframe=${timeframe}&days=${days}`);
}

export async function getMarketIndices(): Promise<MarketIndex[]> {
  return request("/market/indices");
}

export async function getTopMovers(direction: "gainers" | "losers"): Promise<TopMover[]> {
  return request(`/market/movers?direction=${direction}`);
}

export async function getFIIDII(): Promise<FIIDIIData> {
  return request("/market/fiidii");
}

// --- Analysis ---

export async function analyze(
  ticker: string,
  mode: "quick" | "deep" = "quick",
): Promise<AnalysisResult> {
  return request("/analyze", {
    method: "POST",
    body: JSON.stringify({ ticker, mode }),
  });
}

// --- Chat ---

export async function sendChatMessage(
  message: string,
  mode: "quick" | "deep" = "quick",
  history: ChatMessage[] = [],
): Promise<ChatMessage> {
  return request("/chat", {
    method: "POST",
    body: JSON.stringify({ message, mode, history }),
  });
}

// --- FinanceQL ---

export async function executeQuery(
  query: string,
  timeRange?: { start?: string; end?: string; relative?: string },
): Promise<QueryResult> {
  return request("/query", {
    method: "POST",
    body: JSON.stringify({ query, timeRange }),
  });
}

export async function explainQuery(query: string): Promise<{ explanation: string; ast: unknown }> {
  return request("/query/explain", {
    method: "POST",
    body: JSON.stringify({ query }),
  });
}

export async function naturalLanguageQuery(
  text: string,
): Promise<{ financeql: string; result: QueryResult }> {
  return request("/query/nl", {
    method: "POST",
    body: JSON.stringify({ text }),
  });
}

// --- Portfolio ---

export async function getPortfolio(): Promise<PortfolioSummary> {
  return request("/portfolio");
}

// --- Backtest ---

export async function runBacktest(params: BacktestParams): Promise<BacktestResult> {
  return request("/backtest", {
    method: "POST",
    body: JSON.stringify(params),
  });
}

// --- Alerts ---

export async function getAlerts(): Promise<Alert[]> {
  return request("/alerts");
}

export async function createAlert(expression: string): Promise<Alert> {
  return request("/alerts", {
    method: "POST",
    body: JSON.stringify({ expression }),
  });
}

export async function deleteAlert(id: string): Promise<void> {
  return request(`/alerts/${id}`, { method: "DELETE" });
}

// --- Trade Confirmation ---

export async function confirmTrade(
  proposalId: string,
  action: "approve" | "reject" | "modify",
  modifications?: Record<string, unknown>,
): Promise<{ status: string }> {
  return request("/trade/confirm", {
    method: "POST",
    body: JSON.stringify({ proposalId, action, modifications }),
  });
}

// --- Screener ---

export async function runScreener(query: string): Promise<ScreenerResult[]> {
  return request("/screener", {
    method: "POST",
    body: JSON.stringify({ query }),
  });
}

// --- Ticker Search ---

export async function searchTickers(q: string): Promise<Array<{ ticker: string; name: string }>> {
  return request(`/search/tickers?q=${encodeURIComponent(q)}`);
}

// --- Configuration ---

export async function getConfig(): Promise<ConfigResponse> {
  return request("/config");
}

export async function updateConfig(config: Partial<AppConfig>): Promise<ConfigResponse> {
  return request("/config", {
    method: "PUT",
    body: JSON.stringify(config),
  });
}

export async function getConfigKeys(): Promise<KeyStatus[]> {
  return request("/config/keys");
}

// --- Orders ---

export async function getOrders(): Promise<Order[]> {
  return request("/orders");
}

export async function getOrderById(id: string): Promise<Order> {
  return request(`/orders/${encodeURIComponent(id)}`);
}

export async function placeOrder(req: OrderRequest): Promise<OrderResponse> {
  return request("/orders", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function modifyOrder(id: string, req: OrderRequest): Promise<OrderResponse> {
  return request(`/orders/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(req),
  });
}

export async function cancelOrder(id: string): Promise<void> {
  return request(`/orders/${encodeURIComponent(id)}`, { method: "DELETE" });
}

// --- Positions ---

export async function getPositions(): Promise<Position[]> {
  return request("/positions");
}

// --- Funds / Margins ---

export async function getFunds(): Promise<Margins> {
  return request("/funds");
}
