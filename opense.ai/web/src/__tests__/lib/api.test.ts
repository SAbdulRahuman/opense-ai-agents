import { describe, it, expect, vi, beforeEach } from "vitest";
import * as api from "@/lib/api";

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe("API Client", () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it("getHealth returns health data", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ status: "ok" }),
    });
    const result = await api.getHealth();
    expect(result).toEqual({ status: "ok" });
  });

  it("getQuote fetches quote by ticker", async () => {
    const mockQuote = {
      ticker: "RELIANCE",
      price: 2500,
      change: 25,
      changePercent: 1.01,
    };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockQuote),
    });
    const result = await api.getQuote("RELIANCE");
    expect(result).toEqual(mockQuote);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/quote/RELIANCE"),
      expect.any(Object)
    );
  });

  it("throws APIError on non-ok response", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: () => Promise.resolve({ error: "Not found" }),
    });
    await expect(api.getQuote("INVALID")).rejects.toThrow();
  });

  it("analyze sends correct params", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ ticker: "RELIANCE", analysis: {} }),
    });
    await api.analyze("RELIANCE", "deep");
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/analyze"),
      expect.objectContaining({
        method: "POST",
      })
    );
  });

  it("executeQuery sends query", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ type: "scalar", data: { value: 42 } }),
    });
    await api.executeQuery("close(RELIANCE)");
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/query"),
      expect.objectContaining({
        method: "POST",
      })
    );
  });

  it("searchTickers returns results", async () => {
    const mockResults = [
      { ticker: "RELIANCE", name: "Reliance Industries" },
    ];
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockResults),
    });
    const result = await api.searchTickers("REL");
    expect(result).toEqual(mockResults);
  });

  it("runBacktest sends params", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ metrics: {}, trades: [] }),
    });
    await api.runBacktest({
      tickers: ["RELIANCE"],
      strategy: "sma_crossover",
      startDate: "2023-01-01",
      endDate: "2024-01-01",
      capital: 1000000,
      params: {},
    });
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/backtest"),
      expect.objectContaining({ method: "POST" })
    );
  });
});
