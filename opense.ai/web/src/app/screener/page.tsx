"use client";

import { useState, useCallback } from "react";
import { Search, SlidersHorizontal, Download, RotateCcw } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Select, SelectTrigger, SelectContent, SelectItem, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { runScreener } from "@/lib/api";
import { formatPrice, formatPercent, formatIndianNumber, cn } from "@/lib/utils";
import type { ScreenerResult } from "@/lib/types";

interface FilterState {
  marketCapMin: string;
  marketCapMax: string;
  peMin: string;
  peMax: string;
  roeMin: string;
  sectorFilter: string;
  volumeMin: string;
  priceMin: string;
  priceMax: string;
  sortBy: string;
  sortOrder: "asc" | "desc";
}

const defaultFilters: FilterState = {
  marketCapMin: "",
  marketCapMax: "",
  peMin: "",
  peMax: "",
  roeMin: "",
  sectorFilter: "",
  volumeMin: "",
  priceMin: "",
  priceMax: "",
  sortBy: "marketCap",
  sortOrder: "desc",
};

export default function ScreenerPage() {
  const [filters, setFilters] = useState<FilterState>(defaultFilters);
  const [results, setResults] = useState<ScreenerResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [showFilters, setShowFilters] = useState(true);

  const handleScreen = useCallback(async () => {
    setLoading(true);
    try {
      const parts: string[] = [];
      Object.entries(filters).forEach(([k, v]) => {
        if (v) parts.push(`${k}=${encodeURIComponent(v)}`);
      });
      const data = await runScreener(parts.join("&"));
      setResults(data);
    } catch {
      // TODO: toast
    } finally {
      setLoading(false);
    }
  }, [filters]);

  function updateFilter(key: keyof FilterState, value: string) {
    setFilters((prev) => ({ ...prev, [key]: value }));
  }

  function downloadCSV() {
    if (results.length === 0) return;
    const headers = ["Ticker", "Name", "Price", "Change%", "MarketCap", "PE", "RSI", "Signal", "Sector"];
    const rows = results.map((r) =>
      [r.ticker, r.name, r.price, r.changePercent, r.marketCap, r.pe, r.rsi, r.signal, r.sector].join(",")
    );
    const csv = [headers.join(","), ...rows].join("\n");
    const blob = new Blob([csv], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "screener_results.csv";
    a.click();
    URL.revokeObjectURL(url);
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Stock Screener</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Filter NSE stocks by fundamental and technical criteria
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setShowFilters(!showFilters)}
        >
          <SlidersHorizontal size={14} className="mr-1.5" />
          {showFilters ? "Hide" : "Show"} Filters
        </Button>
      </div>

      {/* Filters */}
      {showFilters && (
        <Card>
          <CardContent className="py-4">
            <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-3">
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  Market Cap Min (Cr)
                </label>
                <Input
                  type="number"
                  placeholder="e.g. 10000"
                  value={filters.marketCapMin}
                  onChange={(e) => updateFilter("marketCapMin", e.target.value)}
                  className="h-8 text-sm"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  Market Cap Max (Cr)
                </label>
                <Input
                  type="number"
                  placeholder="e.g. 500000"
                  value={filters.marketCapMax}
                  onChange={(e) => updateFilter("marketCapMax", e.target.value)}
                  className="h-8 text-sm"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  PE Min
                </label>
                <Input
                  type="number"
                  placeholder="e.g. 5"
                  value={filters.peMin}
                  onChange={(e) => updateFilter("peMin", e.target.value)}
                  className="h-8 text-sm"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  PE Max
                </label>
                <Input
                  type="number"
                  placeholder="e.g. 30"
                  value={filters.peMax}
                  onChange={(e) => updateFilter("peMax", e.target.value)}
                  className="h-8 text-sm"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  ROE Min (%)
                </label>
                <Input
                  type="number"
                  placeholder="e.g. 15"
                  value={filters.roeMin}
                  onChange={(e) => updateFilter("roeMin", e.target.value)}
                  className="h-8 text-sm"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  Price Min
                </label>
                <Input
                  type="number"
                  placeholder="e.g. 100"
                  value={filters.priceMin}
                  onChange={(e) => updateFilter("priceMin", e.target.value)}
                  className="h-8 text-sm"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  Price Max
                </label>
                <Input
                  type="number"
                  placeholder="e.g. 5000"
                  value={filters.priceMax}
                  onChange={(e) => updateFilter("priceMax", e.target.value)}
                  className="h-8 text-sm"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  Min Volume
                </label>
                <Input
                  type="number"
                  placeholder="e.g. 100000"
                  value={filters.volumeMin}
                  onChange={(e) => updateFilter("volumeMin", e.target.value)}
                  className="h-8 text-sm"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  Sector
                </label>
                <Input
                  placeholder="e.g. IT"
                  value={filters.sectorFilter}
                  onChange={(e) => updateFilter("sectorFilter", e.target.value)}
                  className="h-8 text-sm"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">
                  Sort By
                </label>
                <Select
                  value={filters.sortBy}
                  onValueChange={(v) => updateFilter("sortBy", v)}
                >
                  <SelectTrigger className="h-8 text-sm">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="marketCap">Market Cap</SelectItem>
                    <SelectItem value="price">Price</SelectItem>
                    <SelectItem value="changePercent">Change %</SelectItem>
                    <SelectItem value="volume">Volume</SelectItem>
                    <SelectItem value="pe">P/E Ratio</SelectItem>
                    <SelectItem value="roe">ROE</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="flex items-center gap-2 mt-4">
              <Button onClick={handleScreen} disabled={loading}>
                <Search size={14} className="mr-1.5" />
                {loading ? "Screening…" : "Screen"}
              </Button>
              <Button
                variant="outline"
                onClick={() => setFilters(defaultFilters)}
              >
                <RotateCcw size={14} className="mr-1.5" />
                Reset
              </Button>
              {results.length > 0 && (
                <Button variant="outline" onClick={downloadCSV}>
                  <Download size={14} className="mr-1.5" />
                  Export CSV
                </Button>
              )}
              {results.length > 0 && (
                <span className="text-sm text-muted-foreground ml-2">
                  {results.length} results
                </span>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Results */}
      {loading ? (
        <Card>
          <CardContent className="py-4">
            <div className="space-y-2">
              {[1, 2, 3, 4, 5].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          </CardContent>
        </Card>
      ) : results.length > 0 ? (
        <Card>
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left bg-muted/30">
                    <th className="p-3 font-medium">#</th>
                    <th className="p-3 font-medium">Stock</th>
                    <th className="p-3 font-medium text-right">Price</th>
                    <th className="p-3 font-medium text-right">Change</th>
                    <th className="p-3 font-medium text-right">Mkt Cap (Cr)</th>
                    <th className="p-3 font-medium text-right">P/E</th>
                    <th className="p-3 font-medium text-right">RSI</th>
                    <th className="p-3 font-medium">Signal</th>
                    <th className="p-3 font-medium">Sector</th>
                  </tr>
                </thead>
                <tbody>
                  {results.map((r, i) => {
                    const isPositive = r.changePercent >= 0;
                    return (
                      <tr key={r.ticker} className="border-b last:border-0 hover:bg-muted/30">
                        <td className="p-3 text-muted-foreground">{i + 1}</td>
                        <td className="p-3">
                          <div className="font-mono font-medium">{r.ticker}</div>
                          <div className="text-xs text-muted-foreground truncate max-w-[150px]">
                            {r.name}
                          </div>
                        </td>
                        <td className="p-3 text-right tabular-nums">
                          {formatPrice(r.price)}
                        </td>
                        <td className="p-3 text-right">
                          <Badge
                            variant={isPositive ? "default" : "destructive"}
                            className="text-xs"
                          >
                            {isPositive ? "+" : ""}
                            {formatPercent(r.changePercent)}
                          </Badge>
                        </td>
                        <td className="p-3 text-right tabular-nums">
                          {formatIndianNumber(r.marketCap)}
                        </td>
                        <td className="p-3 text-right tabular-nums">
                          {r.pe?.toFixed(1) ?? "-"}
                        </td>
                        <td className="p-3 text-right tabular-nums">
                          {r.rsi?.toFixed(1) ?? "-"}
                        </td>
                        <td className="p-3">
                          <Badge variant={r.signal === "BUY" ? "default" : r.signal === "SELL" ? "destructive" : "outline"} className="text-xs">
                            {r.signal}
                          </Badge>
                        </td>
                        <td className="p-3">
                          <Badge variant="outline" className="text-xs">
                            {r.sector}
                          </Badge>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}
