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

// --- Trading / Orders ---

export type OrderSide = "BUY" | "SELL";
export type OrderType = "MARKET" | "LIMIT" | "SL" | "SL-M";
export type OrderProduct = "CNC" | "MIS" | "NRML";
export type OrderStatus = "PENDING" | "OPEN" | "COMPLETE" | "CANCELLED" | "REJECTED";

export interface OrderRequest {
  ticker: string;
  exchange: string; // "NSE" | "BSE" | "NFO"
  side: OrderSide;
  order_type: OrderType;
  product: OrderProduct;
  quantity: number;
  price?: number;
  trigger_price?: number;
  stop_loss?: number;
  target?: number;
  tag?: string;
}

export interface OrderResponse {
  order_id: string;
  status: string;
  message?: string;
}

export interface Order {
  order_id: string;
  ticker: string;
  exchange: string;
  side: OrderSide;
  order_type: OrderType;
  product: OrderProduct;
  quantity: number;
  filled_qty: number;
  pending_qty: number;
  price: number;
  avg_price: number;
  trigger_price?: number;
  status: OrderStatus;
  status_message?: string;
  placed_at: string;
  updated_at: string;
  tag?: string;
}

export interface Position {
  ticker: string;
  exchange: string;
  product: OrderProduct;
  quantity: number; // positive = long, negative = short
  avg_price: number;
  ltp: number;
  pnl: number;
  pnl_pct: number;
  day_pnl: number;
  value: number;
  multiplier: number;
}

export interface Margins {
  available_cash: number;
  used_margin: number;
  available_margin: number;
  collateral: number;
  opening_balance: number;
}

// --- Portfolio ---

export interface Holding {
  ticker: string;
  name: string;
  exchange: string;
  isin: string;
  quantity: number;
  avgPrice: number;
  currentPrice: number;
  value: number;
  pnl: number;
  pnlPercent: number;
  dayChange: number;
  dayChangePct: number;
  allocationPercent: number;
}

export interface PortfolioSummary {
  margins: Margins;
  positions: Position[];
  holdings: Holding[];
  orders: Order[];
}

export interface PortfolioOverview {
  totalValue: number;
  totalInvested: number;
  totalPnl: number;
  totalPnlPercent: number;
  dayPnl: number;
  dayPnlPercent: number;
  marginUsed: number;
  marginAvailable: number;
}

// --- Watchlist Groups ---

export interface WatchlistGroup {
  id: number;
  name: string;
  tickers: string[];
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

// --- Configuration ---

export interface LLMConfig {
  primary: string;
  ollama_url: string;
  model: string;
  fallback_model: string;
  temperature: number;
  max_tokens: number;
}

export interface ZerodhaConfig {
  // Credentials excluded from API (json:"-"), managed via env/keys endpoint
}

export interface IBKRConfig {
  host: string;
  port: number;
}

export interface BrokerConfig {
  provider: string;
  zerodha: ZerodhaConfig;
  ibkr: IBKRConfig;
}

export interface TradingConfig {
  mode: string;
  max_position_pct: number;
  daily_loss_limit_pct: number;
  max_open_positions: number;
  require_confirmation: boolean;
  confirm_timeout_sec: number;
  initial_capital: number;
}

export interface AnalysisConfig {
  cache_ttl: number;
  concurrent_fetches: number;
}

export interface FinanceQLConfig {
  cache_ttl: number;
  max_range: string;
  alert_check_interval: number;
  repl_history_file: string;
}

export interface APIServerConfig {
  host: string;
  port: number;
  cors_origins: string[];
}

export interface WebUIConfig {
  url: string;
}

export interface LoggingConfig {
  level: string;
  format: string;
}

export interface AppConfig {
  llm: LLMConfig;
  broker: BrokerConfig;
  trading: TradingConfig;
  analysis: AnalysisConfig;
  financeql: FinanceQLConfig;
  api: APIServerConfig;
  web: WebUIConfig;
  logging: LoggingConfig;
}

export interface ConfigResponse {
  config: AppConfig;
  config_file: string;
}

export type APIKeySource = "env" | "config" | "none";

export interface KeyStatus {
  name: string;
  source: APIKeySource;
  is_set: boolean;
  masked?: string;
}
