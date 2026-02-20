"use client";

import { useEffect, useCallback, useState } from "react";
import { Star, Plus, X, RefreshCw } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useStore } from "@/store";
import { getQuote } from "@/lib/api";
import { formatPrice, formatPercent, cn } from "@/lib/utils";
import type { Quote } from "@/lib/types";

export function Watchlist() {
  const { watchlist, setWatchlist, quotes, setQuote, setSelectedTicker } = useStore();
  const [quotesData, setQuotesData] = useState<Record<string, Quote>>({});
  const [loading, setLoading] = useState(true);
  const [addMode, setAddMode] = useState(false);
  const [newTicker, setNewTicker] = useState("");

  const fetchQuotes = useCallback(async () => {
    try {
      const results = await Promise.allSettled(
        watchlist.map((t) => getQuote(t))
      );
      const updated: Record<string, Quote> = {};
      results.forEach((r, i) => {
        if (r.status === "fulfilled") {
          updated[watchlist[i]] = r.value;
          setQuote(watchlist[i], r.value);
        }
      });
      setQuotesData((prev) => ({ ...prev, ...updated }));
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, [watchlist, setQuote]);

  useEffect(() => {
    fetchQuotes();
    const interval = setInterval(fetchQuotes, 15000);
    return () => clearInterval(interval);
  }, [fetchQuotes]);

  function addTicker() {
    const ticker = newTicker.trim().toUpperCase();
    if (ticker && !watchlist.includes(ticker)) {
      setWatchlist([...watchlist, ticker]);
    }
    setNewTicker("");
    setAddMode(false);
  }

  function removeTicker(ticker: string) {
    setWatchlist(watchlist.filter((t) => t !== ticker));
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Star size={18} className="text-yellow-500" />
            Watchlist
          </CardTitle>
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="sm" onClick={fetchQuotes}>
              <RefreshCw size={14} />
            </Button>
            <Button variant="ghost" size="sm" onClick={() => setAddMode(!addMode)}>
              <Plus size={14} />
            </Button>
          </div>
        </div>
        {addMode && (
          <div className="flex items-center gap-2 mt-2">
            <Input
              placeholder="e.g. SBIN"
              value={newTicker}
              onChange={(e) => setNewTicker(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && addTicker()}
              className="h-8 text-sm"
              autoFocus
            />
            <Button size="sm" className="h-8" onClick={addTicker}>
              Add
            </Button>
          </div>
        )}
      </CardHeader>
      <CardContent className="p-0">
        {loading && watchlist.length > 0 ? (
          <div className="px-5 pb-4 space-y-3">
            {watchlist.map((t) => (
              <Skeleton key={t} className="h-10 w-full" />
            ))}
          </div>
        ) : (
          <div className="divide-y">
            {watchlist.map((ticker) => {
              const q = quotesData[ticker];
              if (!q) {
                return (
                  <div
                    key={ticker}
                    className="flex items-center justify-between px-5 py-3"
                  >
                    <span className="font-mono text-sm font-medium">{ticker}</span>
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-muted-foreground">Loading…</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0"
                        onClick={() => removeTicker(ticker)}
                      >
                        <X size={12} />
                      </Button>
                    </div>
                  </div>
                );
              }
              const isPositive = q.change >= 0;
              return (
                <button
                  key={ticker}
                  className="flex items-center justify-between w-full px-5 py-3 hover:bg-muted/50 transition-colors text-left"
                  onClick={() => setSelectedTicker(ticker)}
                >
                  <div>
                    <div className="font-mono text-sm font-medium">{ticker}</div>
                    <div className="text-xs text-muted-foreground">{q.name}</div>
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="text-right">
                      <div className="text-sm font-medium tabular-nums">
                        {formatPrice(q.price)}
                      </div>
                      <div
                        className={cn(
                          "text-xs tabular-nums",
                          isPositive
                            ? "text-green-600 dark:text-green-400"
                            : "text-red-600 dark:text-red-400"
                        )}
                      >
                        {isPositive ? "+" : ""}
                        {formatPercent(q.changePercent)}
                      </div>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100"
                      onClick={(e) => {
                        e.stopPropagation();
                        removeTicker(ticker);
                      }}
                    >
                      <X size={12} />
                    </Button>
                  </div>
                </button>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
