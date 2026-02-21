"use client";

import { useEffect, useCallback, useState } from "react";
import { Star, Plus, X, RefreshCw, ShoppingCart, TrendingDown } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useStore } from "@/store";
import { getQuote } from "@/lib/api";
import { formatPrice, formatPercent, cn } from "@/lib/utils";
import type { Quote } from "@/lib/types";

const WATCHLIST_TABS = [1, 2, 3, 4, 5];

export function Watchlist() {
  const {
    watchlist,
    setWatchlist,
    quotes,
    setQuote,
    setSelectedTicker,
    activeWatchlistTab,
    setActiveWatchlistTab,
    openOrderWindow,
  } = useStore();
  const [quotesData, setQuotesData] = useState<Record<string, Quote>>({});
  const [loading, setLoading] = useState(true);
  const [addMode, setAddMode] = useState(false);
  const [newTicker, setNewTicker] = useState("");
  const [hoveredTicker, setHoveredTicker] = useState<string | null>(null);

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

        {/* Watchlist Tabs */}
        <div className="flex gap-1 mt-2">
          {WATCHLIST_TABS.map((tab) => (
            <button
              key={tab}
              className={cn(
                "h-7 w-7 rounded text-xs font-medium transition-colors",
                activeWatchlistTab === tab - 1
                  ? "bg-primary text-primary-foreground"
                  : "bg-muted text-muted-foreground hover:bg-accent",
              )}
              onClick={() => setActiveWatchlistTab(tab - 1)}
            >
              {tab}
            </button>
          ))}
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
              const isHovered = hoveredTicker === ticker;
              return (
                <div
                  key={ticker}
                  className="relative flex items-center justify-between w-full px-5 py-3 hover:bg-muted/50 transition-colors text-left"
                  onMouseEnter={() => setHoveredTicker(ticker)}
                  onMouseLeave={() => setHoveredTicker(null)}
                >
                  <button
                    className="flex-1 text-left"
                    onClick={() => setSelectedTicker(ticker)}
                  >
                    <div>
                      <div className="font-mono text-sm font-medium">{ticker}</div>
                      <div className="text-xs text-muted-foreground">{q.name}</div>
                    </div>
                  </button>
                  <div className="flex items-center gap-2">
                    {/* Price info */}
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

                    {/* Action buttons on hover */}
                    {isHovered && (
                      <div className="flex gap-1 ml-1">
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-6 px-2 text-xs text-blue-500 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-950"
                          onClick={(e) => {
                            e.stopPropagation();
                            openOrderWindow(ticker, "BUY");
                          }}
                        >
                          B
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-6 px-2 text-xs text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950"
                          onClick={(e) => {
                            e.stopPropagation();
                            openOrderWindow(ticker, "SELL");
                          }}
                        >
                          S
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 w-6 p-0"
                          onClick={(e) => {
                            e.stopPropagation();
                            removeTicker(ticker);
                          }}
                        >
                          <X size={12} />
                        </Button>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
