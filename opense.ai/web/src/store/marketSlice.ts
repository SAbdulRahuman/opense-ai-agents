// ============================================================================
// OpeNSE.ai — Market Store Slice (Zustand)
// ============================================================================

import type { StateCreator } from "zustand";
import type { Quote, MarketIndex, OHLCV } from "@/lib/types";

export interface MarketSlice {
  quotes: Record<string, Quote>;
  indices: MarketIndex[];
  watchlist: string[];
  selectedTicker: string;
  chartData: OHLCV[];
  timeframe: string;
  isMarketOpen: boolean;

  setQuote: (ticker: string, quote: Quote) => void;
  setIndices: (indices: MarketIndex[]) => void;
  setWatchlist: (watchlist: string[]) => void;
  addToWatchlist: (ticker: string) => void;
  removeFromWatchlist: (ticker: string) => void;
  setSelectedTicker: (ticker: string) => void;
  setChartData: (data: OHLCV[]) => void;
  setTimeframe: (tf: string) => void;
  setMarketOpen: (open: boolean) => void;
}

export const createMarketSlice: StateCreator<MarketSlice> = (set) => ({
  quotes: {},
  indices: [],
  watchlist: ["RELIANCE", "TCS", "INFY", "HDFCBANK", "ICICIBANK"],
  selectedTicker: "RELIANCE",
  chartData: [],
  timeframe: "1D",
  isMarketOpen: false,

  setQuote: (ticker, quote) =>
    set((state) => ({
      quotes: { ...state.quotes, [ticker]: quote },
    })),

  setIndices: (indices) => set({ indices }),
  setWatchlist: (watchlist) => set({ watchlist }),

  addToWatchlist: (ticker) =>
    set((state) => ({
      watchlist: state.watchlist.includes(ticker)
        ? state.watchlist
        : [...state.watchlist, ticker],
    })),

  removeFromWatchlist: (ticker) =>
    set((state) => ({
      watchlist: state.watchlist.filter((t) => t !== ticker),
    })),

  setSelectedTicker: (selectedTicker) => set({ selectedTicker }),
  setChartData: (chartData) => set({ chartData }),
  setTimeframe: (timeframe) => set({ timeframe }),
  setMarketOpen: (isMarketOpen) => set({ isMarketOpen }),
});
