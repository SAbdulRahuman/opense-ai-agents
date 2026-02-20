// ============================================================================
// OpeNSE.ai — WatchlistPanel (TradingView-style right sidebar)
// ============================================================================
"use client";

import { useStore } from "@/store";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Plus, X, Star, TrendingUp, TrendingDown } from "lucide-react";
import { formatPrice, formatPercent } from "@/lib/utils";
import { useState } from "react";
import { Input } from "@/components/ui/input";

export function WatchlistPanel() {
  const {
    watchlist,
    selectedTicker,
    setSelectedTicker,
    addToWatchlist,
    removeFromWatchlist,
    quotes,
  } = useStore();

  const [isAdding, setIsAdding] = useState(false);
  const [newTicker, setNewTicker] = useState("");

  const handleAdd = () => {
    const ticker = newTicker.trim().toUpperCase();
    if (ticker && !watchlist.includes(ticker)) {
      addToWatchlist(ticker);
    }
    setNewTicker("");
    setIsAdding(false);
  };

  return (
    <div className="flex h-full w-56 flex-col border-l bg-card">
      {/* Header */}
      <div className="flex items-center justify-between border-b px-3 py-2">
        <div className="flex items-center gap-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          <Star className="h-3.5 w-3.5" />
          Watchlist
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="h-6 w-6 p-0"
          onClick={() => setIsAdding(!isAdding)}
        >
          <Plus className="h-3.5 w-3.5" />
        </Button>
      </div>

      {/* Add ticker input */}
      {isAdding && (
        <div className="border-b p-2">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              handleAdd();
            }}
          >
            <Input
              autoFocus
              placeholder="Add ticker..."
              value={newTicker}
              onChange={(e) => setNewTicker(e.target.value.toUpperCase())}
              className="h-7 text-xs font-mono"
            />
          </form>
        </div>
      )}

      {/* Watchlist items */}
      <div className="flex-1 overflow-y-auto">
        {watchlist.map((ticker) => {
          const quote = quotes[ticker];
          const isActive = ticker === selectedTicker;
          const isPositive = quote ? quote.change >= 0 : true;

          return (
            <button
              key={ticker}
              className={cn(
                "group flex w-full items-center justify-between px-3 py-2 text-left transition-colors hover:bg-muted/50",
                isActive && "bg-muted",
              )}
              onClick={() => setSelectedTicker(ticker)}
            >
              <div className="flex flex-col">
                <span className={cn("text-xs font-bold font-mono", isActive && "text-primary")}>
                  {ticker}
                </span>
                {quote && (
                  <span className="text-[10px] text-muted-foreground truncate max-w-[80px]">
                    {quote.name}
                  </span>
                )}
              </div>

              <div className="flex flex-col items-end gap-0.5">
                {quote ? (
                  <>
                    <span className="text-xs font-mono tabular-nums">
                      {formatPrice(quote.price)}
                    </span>
                    <span
                      className={cn(
                        "flex items-center gap-0.5 text-[10px] font-mono tabular-nums",
                        isPositive
                          ? "text-green-600 dark:text-green-400"
                          : "text-red-600 dark:text-red-400",
                      )}
                    >
                      {isPositive ? (
                        <TrendingUp className="h-2.5 w-2.5" />
                      ) : (
                        <TrendingDown className="h-2.5 w-2.5" />
                      )}
                      {formatPercent(quote.changePercent)}
                    </span>
                  </>
                ) : (
                  <span className="text-[10px] text-muted-foreground">—</span>
                )}
              </div>

              {/* Remove button (on hover) */}
              <X
                className="ml-1 h-3 w-3 shrink-0 text-muted-foreground opacity-0 group-hover:opacity-100 hover:text-destructive"
                onClick={(e) => {
                  e.stopPropagation();
                  removeFromWatchlist(ticker);
                }}
              />
            </button>
          );
        })}
      </div>

      {/* Footer */}
      <div className="border-t px-3 py-1.5 text-[10px] text-muted-foreground">
        {watchlist.length} symbols
      </div>
    </div>
  );
}
