// ============================================================================
// OpeNSE.ai — Shared TypeScript Types
// ============================================================================

// --- Market Data ---

export interface OHLCV {
  time: number; // Unix timestamp (seconds)
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface Quote {
  ticker: string;
  name: string;
  price: number;
  change: number;
  changePercent: number;
  open: number;
  high: number;
  low: number;
  prevClose: number;
  volume: number;
  timestamp: string;
}

export interface MarketIndex {
  name: string;
  value: number;
  change: number;
  changePercent: number;
}

export interface TopMover {
  ticker: string;
  name: string;
  price: number;
  changePercent: number;
}

export interface FIIDIIData {
  date: string;
  fiiBuy: number;
  fiiSell: number;
  fiiNet: number;
  diiBuy: number;
  diiSell: number;
  diiNet: number;
}

// --- Analysis ---

export interface AnalysisResult {
  ticker: string;
  technical: TechnicalSummary;
  fundamental: FundamentalSummary;
  sentiment: SentimentSummary;
  recommendation: string;
  confidence: number;
  reasoning: string;
}

export interface TechnicalSummary {
  rsi: number;
  macd: { macd: number; signal: number; histogram: number };
  sma20: number;
  sma50: number;
  sma200: number;
  supertrend: { value: number; direction: string };
  bollingerBands: { upper: number; middle: number; lower: number };
  signal: string;
}

export interface FundamentalSummary {
  pe: number;
  pb: number;
  marketCap: number;
  eps: number;
  dividendYield: number;
  roe: number;
  debtToEquity: number;
  sector: string;
  industry: string;
}

export interface SentimentSummary {
  score: number;
  label: string;
  newsCount: number;
  topHeadlines: string[];
}

// --- Chat ---

export interface ChatMessage {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  agent?: string;
  toolCalls?: ToolCall[];
  timestamp: string;
  streaming?: boolean;
}

export interface ToolCall {
  id: string;
  name: string;
  arguments: Record<string, unknown>;
  result?: string;
  status: "pending" | "running" | "completed" | "failed";
}

export interface TradeProposal {
  id: string;
  ticker: string;
  action: "BUY" | "SELL";
  quantity: number;
  price: number;
  orderType: "MARKET" | "LIMIT" | "SL" | "SL-M";
  estimatedCost: number;
  positionPct: number;
  currentExposure: number;
  timeout: number; // seconds
}

// --- FinanceQL ---

export type QueryResultType = "scalar" | "vector" | "matrix" | "table";

export interface QueryResult {
  type: QueryResultType;
  data: ScalarResult | VectorResult | MatrixResult | TableResult;
  query: string;
  duration: number; // ms
  timestamp: string;
}

export interface ScalarResult {
  value: number;
  label: string;
  ticker?: string;
  metric?: string;
}

export interface VectorResult {
  items: Array<{
    ticker: string;
    value: number;
    labels: Record<string, string>;
  }>;
}

export interface MatrixResult {
  series: Array<{
    label: string;
    data: Array<{ time: number; value: number }>;
  }>;
}

export interface TableResult {
  columns: string[];
  rows: Array<Record<string, unknown>>;
}

export interface QueryHistoryEntry {
  id: string;
  query: string;
  resultType: QueryResultType;
  duration: number;
  timestamp: string;
  starred: boolean;
}

// --- Portfolio ---

export interface Holding {
  ticker: string;
  name: string;
  quantity: number;
  avgPrice: number;
  currentPrice: number;
  value: number;
  pnl: number;
  pnlPercent: number;
  allocationPercent: number;
}

export interface PortfolioSummary {
  totalValue: number;
  totalInvested: number;
  totalPnl: number;
  totalPnlPercent: number;
  dayPnl: number;
  dayPnlPercent: number;
  holdings: Holding[];
  marginUsed: number;
  marginAvailable: number;
}

// --- Backtest ---

export interface BacktestParams {
  strategy: string;
  tickers: string[];
  startDate: string;
  endDate: string;
  capital: number;
  params: Record<string, number>;
}

export interface BacktestResult {
  equityCurve: Array<{ time: number; value: number }>;
  benchmarkCurve: Array<{ time: number; value: number }>;
  drawdownCurve: Array<{ time: number; value: number }>;
  trades: BacktestTrade[];
  metrics: BacktestMetrics;
}

export interface BacktestTrade {
  ticker: string;
  entryDate: string;
  exitDate: string;
  entryPrice: number;
  exitPrice: number;
  quantity: number;
  pnl: number;
  pnlPercent: number;
  side: "LONG" | "SHORT";
}

export interface BacktestMetrics {
  totalReturn: number;
  cagr: number;
  sharpeRatio: number;
  maxDrawdown: number;
  winRate: number;
  totalTrades: number;
  profitFactor: number;
  avgWin: number;
  avgLoss: number;
}

// --- Screener ---

export interface ScreenerResult {
  ticker: string;
  name: string;
  price: number;
  changePercent: number;
  pe: number;
  marketCap: number;
  rsi: number;
  signal: string;
  sector: string;
}

// --- WebSocket ---

export interface WSMessage {
  type: string;
  data: unknown;
}

// --- API Response ---

export interface APIResponse<T> {
  data?: T;
  error?: string;
  status: number;
}

// --- Alerts ---

export interface Alert {
  id: string;
  expression: string;
  name?: string;
  query?: string;
  condition?: string;
  threshold?: number;
  status: "pending" | "triggered" | "expired";
  createdAt: string;
  triggeredAt?: string;
  value?: number;
}
