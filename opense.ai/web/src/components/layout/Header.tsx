"use client";

import { useState, useCallback } from "react";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ThemeToggle } from "@/components/layout/ThemeToggle";
import { useStore } from "@/store";
import { cn } from "@/lib/utils";

export function Header() {
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<Array<{ ticker: string; name: string }>>([]);
  const [showResults, setShowResults] = useState(false);
  const { isMarketOpen, setSelectedTicker } = useStore();

  const handleSearch = useCallback(
    async (q: string) => {
      setSearchQuery(q);
      if (q.length < 2) {
        setSearchResults([]);
        setShowResults(false);
        return;
      }

      try {
        const { searchTickers } = await import("@/lib/api");
        const results = await searchTickers(q);
        setSearchResults(results);
        setShowResults(true);
      } catch {
        // Search failed — show empty results
        setSearchResults([]);
      }
    },
    [],
  );

  const selectTicker = (ticker: string) => {
    setSelectedTicker(ticker);
    setSearchQuery("");
    setShowResults(false);
  };

  return (
    <header className="flex h-14 items-center justify-between border-b bg-card px-4">
      {/* Market Status */}
      <div className="flex items-center gap-3">
        <Badge
          variant={isMarketOpen ? "default" : "secondary"}
          className={cn(
            "gap-1.5",
            isMarketOpen ? "bg-green-500/10 text-green-600 border-green-500/20" : "",
          )}
        >
          <span
            className={cn(
              "h-2 w-2 rounded-full",
              isMarketOpen ? "bg-green-500 animate-pulse" : "bg-gray-400",
            )}
          />
          {isMarketOpen ? "Market Open" : "Market Closed"}
        </Badge>
      </div>

      {/* Global Ticker Search */}
      <div className="relative w-80">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search ticker (e.g., RELIANCE, TCS)..."
          value={searchQuery}
          onChange={(e) => handleSearch(e.target.value)}
          onFocus={() => searchResults.length > 0 && setShowResults(true)}
          onBlur={() => setTimeout(() => setShowResults(false), 200)}
          className="pl-9"
        />
        {showResults && searchResults.length > 0 && (
          <div className="absolute top-full z-50 mt-1 w-full rounded-md border bg-popover p-1 shadow-md">
            {searchResults.map((r) => (
              <button
                key={r.ticker}
                className="flex w-full items-center justify-between rounded-sm px-3 py-2 text-sm hover:bg-accent"
                onMouseDown={() => selectTicker(r.ticker)}
              >
                <span className="font-mono font-semibold">{r.ticker}</span>
                <span className="text-muted-foreground text-xs truncate ml-2">
                  {r.name}
                </span>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Right side actions */}
      <div className="flex items-center gap-2">
        <ThemeToggle />
      </div>
    </header>
  );
}
