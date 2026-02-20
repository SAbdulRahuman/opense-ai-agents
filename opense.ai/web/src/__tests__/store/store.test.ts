import { describe, it, expect, beforeEach } from "vitest";
import { useStore } from "@/store";

describe("Store", () => {
  beforeEach(() => {
    // Reset store to initial state
    const { getState } = useStore;
    const state = getState();
    state.clearMessages();
    state.setSelectedTicker("RELIANCE");
    state.setTimeframe("1D");
  });

  describe("ChatSlice", () => {
    it("has initial empty messages", () => {
      const state = useStore.getState();
      expect(state.messages).toEqual([]);
    });

    it("adds a message", () => {
      const state = useStore.getState();
      state.addMessage({
        id: "1",
        role: "user",
        content: "Hello",
        timestamp: new Date().toISOString(),
      });
      expect(useStore.getState().messages).toHaveLength(1);
      expect(useStore.getState().messages[0].content).toBe("Hello");
    });

    it("clears messages", () => {
      const state = useStore.getState();
      state.addMessage({
        id: "1",
        role: "user",
        content: "Hello",
        timestamp: new Date().toISOString(),
      });
      state.clearMessages();
      expect(useStore.getState().messages).toHaveLength(0);
    });

    it("sets streaming state", () => {
      const state = useStore.getState();
      state.setStreaming(true);
      expect(useStore.getState().isStreaming).toBe(true);
      state.setStreaming(false);
      expect(useStore.getState().isStreaming).toBe(false);
    });

    it("sets mode", () => {
      const state = useStore.getState();
      state.setMode("deep");
      expect(useStore.getState().mode).toBe("deep");
      state.setMode("quick");
      expect(useStore.getState().mode).toBe("quick");
    });
  });

  describe("MarketSlice", () => {
    it("sets selected ticker", () => {
      const state = useStore.getState();
      state.setSelectedTicker("TCS");
      expect(useStore.getState().selectedTicker).toBe("TCS");
    });

    it("sets timeframe", () => {
      const state = useStore.getState();
      state.setTimeframe("1W");
      expect(useStore.getState().timeframe).toBe("1W");
    });

    it("manages watchlist", () => {
      const state = useStore.getState();
      state.setWatchlist(["RELIANCE", "TCS", "INFY"]);
      expect(useStore.getState().watchlist).toEqual(["RELIANCE", "TCS", "INFY"]);
    });

    it("sets quotes", () => {
      const state = useStore.getState();
      const quote = {
        ticker: "RELIANCE",
        name: "Reliance Industries",
        price: 2500,
        change: 25,
        changePercent: 1.01,
        volume: 1000000,
        high: 2520,
        low: 2480,
        open: 2490,
        prevClose: 2475,
        timestamp: new Date().toISOString(),
      };
      state.setQuote("RELIANCE", quote);
      expect(useStore.getState().quotes["RELIANCE"]).toEqual(quote);
    });

    it("sets indices", () => {
      const state = useStore.getState();
      state.setIndices([
        { name: "NIFTY 50", value: 22500, change: 100, changePercent: 0.45 },
      ]);
      expect(useStore.getState().indices).toHaveLength(1);
    });
  });

  describe("QuerySlice", () => {
    it("sets current query", () => {
      const state = useStore.getState();
      state.setCurrentQuery("close(RELIANCE)");
      expect(useStore.getState().currentQuery).toBe("close(RELIANCE)");
    });

    it("sets executing state", () => {
      const state = useStore.getState();
      state.setExecuting(true);
      expect(useStore.getState().isExecuting).toBe(true);
    });

    it("manages query history", () => {
      const state = useStore.getState();
      state.addToHistory({
        id: "1",
        query: "close(RELIANCE)",
        resultType: "scalar",
        timestamp: new Date().toISOString(),
        duration: 42,
        starred: false,
      });
      expect(useStore.getState().queryHistory).toHaveLength(1);
    });

    it("toggles natural language mode", () => {
      const state = useStore.getState();
      state.setNaturalLanguageMode(true);
      expect(useStore.getState().naturalLanguageMode).toBe(true);
      state.setNaturalLanguageMode(false);
      expect(useStore.getState().naturalLanguageMode).toBe(false);
    });

    it("sets result tab", () => {
      const state = useStore.getState();
      state.setResultTab("table");
      expect(useStore.getState().resultTab).toBe("table");
    });
  });
});
