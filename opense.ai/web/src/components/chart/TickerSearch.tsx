// ============================================================================
// OpeNSE.ai — TickerSearch (searchable ticker input with dropdown)
// ============================================================================
"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { useStore } from "@/store";
import { searchTickers } from "@/lib/api";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { Search, X } from "lucide-react";

interface TickerResult {
  ticker: string;
  name: string;
}

export function TickerSearch() {
  const { selectedTicker, setSelectedTicker } = useStore();
  const [query, setQuery] = useState(selectedTicker);
  const [results, setResults] = useState<TickerResult[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [selectedIdx, setSelectedIdx] = useState(-1);
  const inputRef = useRef<HTMLInputElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();

  // Sync external ticker changes
  useEffect(() => {
    setQuery(selectedTicker);
  }, [selectedTicker]);

  const doSearch = useCallback(async (q: string) => {
    if (q.length < 1) {
      setResults([]);
      return;
    }
    try {
      const res = await searchTickers(q);
      setResults(
        (res ?? []).slice(0, 8).map((r: { ticker: string; name: string }) => ({
          ticker: r.ticker,
          name: r.name,
        })),
      );
    } catch {
      // Fallback: show the typed ticker itself
      setResults([{ ticker: q.toUpperCase(), name: "" }]);
    }
  }, []);

  const handleChange = (value: string) => {
    const upper = value.toUpperCase();
    setQuery(upper);
    setSelectedIdx(-1);
    setIsOpen(true);

    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => doSearch(upper), 200);
  };

  const handleSelect = (ticker: string) => {
    setSelectedTicker(ticker);
    setQuery(ticker);
    setIsOpen(false);
    inputRef.current?.blur();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIdx((i) => Math.min(i + 1, results.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIdx((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (selectedIdx >= 0 && results[selectedIdx]) {
        handleSelect(results[selectedIdx].ticker);
      } else {
        handleSelect(query);
      }
    } else if (e.key === "Escape") {
      setIsOpen(false);
      setQuery(selectedTicker);
    }
  };

  return (
    <div className="relative">
      <div className="relative flex items-center">
        <Search className="absolute left-2 h-3.5 w-3.5 text-muted-foreground" />
        <Input
          ref={inputRef}
          value={query}
          onChange={(e) => handleChange(e.target.value)}
          onFocus={() => query && setIsOpen(true)}
          onBlur={() => setTimeout(() => setIsOpen(false), 150)}
          onKeyDown={handleKeyDown}
          placeholder="Search ticker..."
          className="h-8 w-48 pl-7 pr-7 text-xs font-mono"
        />
        {query && query !== selectedTicker && (
          <button
            className="absolute right-2"
            onClick={() => {
              setQuery(selectedTicker);
              setIsOpen(false);
            }}
          >
            <X className="h-3 w-3 text-muted-foreground" />
          </button>
        )}
      </div>

      {/* Dropdown */}
      {isOpen && results.length > 0 && (
        <div className="absolute top-full left-0 z-50 mt-1 w-64 rounded-md border bg-popover shadow-lg">
          {results.map((r, idx) => (
            <button
              key={r.ticker}
              className={cn(
                "flex w-full items-center gap-3 px-3 py-2 text-left text-xs hover:bg-muted/50",
                idx === selectedIdx && "bg-muted",
              )}
              onMouseDown={() => handleSelect(r.ticker)}
            >
              <span className="font-mono font-bold">{r.ticker}</span>
              <span className="text-muted-foreground truncate">{r.name}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
