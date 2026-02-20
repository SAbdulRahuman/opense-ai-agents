// ============================================================================
// OpeNSE.ai — FinanceQL Language Definition for Monaco Editor
// ============================================================================

import type { languages } from "monaco-editor";

export const FINANCEQL_LANGUAGE_ID = "financeql";

export const financeqlLanguage: languages.IMonarchLanguage = {
  defaultToken: "",
  tokenPostfix: ".financeql",

  keywords: ["AND", "OR", "NOT", "WHERE", "BY", "GROUP", "ORDER", "ASC", "DESC", "LIMIT", "OFFSET"],

  functions: [
    "price", "sma", "ema", "rsi", "macd", "bbands", "supertrend",
    "volume", "atr", "adx", "obv", "vwap",
    "pe", "pb", "eps", "roe", "debt_to_equity", "market_cap", "dividend_yield",
    "screener", "alert", "rank", "compare",
    "avg", "sum", "min", "max", "count", "stddev",
    "change", "pct_change", "returns",
    "sector", "industry", "nifty50", "banknifty",
    "fii_dii", "sentiment", "news",
  ],

  operators: [
    ">", "<", ">=", "<=", "==", "!=",
    "+", "-", "*", "/", "%",
    "|", "&",
  ],

  symbols: /[=><!~?:&|+\-*\/\^%]+/,
  escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,

  tokenizer: {
    root: [
      // Identifiers and keywords
      [/[a-zA-Z_]\w*/, {
        cases: {
          "@keywords": "keyword",
          "@functions": "type.identifier",
          "@default": "identifier",
        },
      }],

      // Tickers (uppercase identifiers)
      [/[A-Z][A-Z0-9_]{1,20}/, "variable.name"],

      // Whitespace
      { include: "@whitespace" },

      // Delimiters and operators
      [/[{}()\[\]]/, "@brackets"],
      [/[<>](?!@symbols)/, "@brackets"],
      [/@symbols/, {
        cases: {
          "@operators": "operator",
          "@default": "",
        },
      }],

      // Numbers
      [/\d*\.\d+([eE][\-+]?\d+)?/, "number.float"],
      [/\d+/, "number"],

      // Durations (e.g., 30d, 1y, 90d)
      [/\d+[dhms]/, "number"],
      [/\d+[yY]/, "number"],

      // Strings
      [/"([^"\\]|\\.)*$/, "string.invalid"],
      [/"/, { token: "string.quote", bracket: "@open", next: "@string" }],

      // Pipe operator
      [/\|>?/, "operator"],

      // Range selector
      [/\[.*?\]/, "annotation"],

      // Comments
      [/#.*$/, "comment"],
    ],

    string: [
      [/[^\\"]+/, "string"],
      [/@escapes/, "string.escape"],
      [/\\./, "string.escape.invalid"],
      [/"/, { token: "string.quote", bracket: "@close", next: "@pop" }],
    ],

    whitespace: [
      [/[ \t\r\n]+/, "white"],
      [/#.*$/, "comment"],
    ],
  },
};

export const financeqlCompletionItems = [
  // Functions
  ...["price", "sma", "ema", "rsi", "macd", "bbands", "supertrend", "volume", "atr", "adx", "obv", "vwap",
    "pe", "pb", "eps", "roe", "debt_to_equity", "market_cap", "dividend_yield",
    "screener", "alert", "rank", "compare",
    "avg", "sum", "min", "max", "count", "stddev",
    "change", "pct_change", "returns",
    "sector", "industry", "nifty50", "banknifty",
    "fii_dii", "sentiment", "news",
  ].map((fn) => ({
    label: fn,
    kind: 1, // Function
    insertText: `${fn}($0)`,
    insertTextRules: 4, // InsertAsSnippet
    detail: "FinanceQL function",
  })),

  // Keywords
  ...["AND", "OR", "NOT", "WHERE", "BY", "GROUP", "ORDER"].map((kw) => ({
    label: kw,
    kind: 14, // Keyword
    insertText: kw,
    detail: "Keyword",
  })),

  // Common tickers
  ...["RELIANCE", "TCS", "INFY", "HDFCBANK", "ICICIBANK", "BHARTIARTL", "ITC", "SBIN", "LT", "KOTAKBANK",
    "HINDUNILVR", "AXISBANK", "BAJFINANCE", "MARUTI", "TITAN", "ASIANPAINT", "WIPRO", "HCLTECH",
    "SUNPHARMA", "ULTRACEMCO",
  ].map((ticker) => ({
    label: ticker,
    kind: 5, // Variable
    insertText: ticker,
    detail: "NSE Ticker",
  })),
];

export const financeqlTheme = {
  base: "vs-dark" as const,
  inherit: true,
  rules: [
    { token: "type.identifier", foreground: "3b82f6", fontStyle: "bold" },
    { token: "variable.name", foreground: "22c55e" },
    { token: "keyword", foreground: "a78bfa" },
    { token: "operator", foreground: "ef4444" },
    { token: "number", foreground: "f97316" },
    { token: "number.float", foreground: "f97316" },
    { token: "string", foreground: "fbbf24" },
    { token: "comment", foreground: "6b7280", fontStyle: "italic" },
    { token: "annotation", foreground: "38bdf8" },
  ],
  colors: {
    "editor.background": "#0a0a0a",
    "editor.foreground": "#d1d5db",
    "editor.lineHighlightBackground": "#1e293b",
  },
};
