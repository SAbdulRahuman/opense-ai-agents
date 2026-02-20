"use client";

import { useEffect, useState } from "react";
import { useStore } from "@/store";
import { getOHLCV } from "@/lib/api";
import { TradingViewChart, ChartToolbar, IndicatorOverlay } from "@/components/chart";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { formatPrice, formatIndianNumber, formatPercent, cn } from "@/lib/utils";
import type { ChartConfig } from "@/hooks/useTradingView";

const defaultConfig: ChartConfig = {
  showVolume: true,
  indicators: [],
  timeframe: "1D",
};

export default function ChartsPage() {
  const {
    selectedTicker,
    setSelectedTicker,
    chartData,
    setChartData,
    timeframe,
    quotes,
  } = useStore();

  const [config, setConfig] = useState<ChartConfig>({ ...defaultConfig, timeframe });

  useEffect(() => {
    if (!selectedTicker) return;
    getOHLCV(selectedTicker, config.timeframe)
      .then(setChartData)
      .catch(() => {});
  }, [selectedTicker, config.timeframe, setChartData]);

  const quote = quotes[selectedTicker];

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold">Charts</h1>
          <p className="text-sm text-muted-foreground">
            Interactive TradingView-style charts with technical indicators
          </p>
        </div>
        <div className="ml-auto">
          <Input
            placeholder="Enter ticker (e.g. RELIANCE)"
            value={selectedTicker}
            onChange={(e) => setSelectedTicker(e.target.value.toUpperCase())}
            className="w-64 font-mono"
          />
        </div>
      </div>

      {/* Quote Summary */}
      {quote && (
        <Card>
          <CardContent className="py-3 flex items-center gap-6">
            <div>
              <span className="text-lg font-bold font-mono">{selectedTicker}</span>
              <span className="text-sm text-muted-foreground ml-2">{quote.name}</span>
            </div>
            <div className="text-lg font-bold tabular-nums">
              {formatPrice(quote.price)}
            </div>
            <div
              className={cn(
                "text-sm tabular-nums font-medium",
                quote.change >= 0
                  ? "text-green-600 dark:text-green-400"
                  : "text-red-600 dark:text-red-400"
              )}
            >
              {quote.change >= 0 ? "+" : ""}
              {formatIndianNumber(quote.change)} ({formatPercent(quote.changePercent)})
            </div>
            <div className="text-sm text-muted-foreground">
              Vol: {formatIndianNumber(quote.volume)}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Toolbar */}
      <ChartToolbar config={config} onConfigChange={setConfig} />

      {/* Main Chart */}
      <Card>
        <CardContent className="p-2">
          <TradingViewChart data={chartData} config={config} />
        </CardContent>
      </Card>

      {/* RSI Indicator */}
      {chartData.length > 0 && (
        <IndicatorOverlay data={chartData} type="RSI" />
      )}
    </div>
  );
}
