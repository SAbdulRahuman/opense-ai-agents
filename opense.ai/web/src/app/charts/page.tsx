"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useStore } from "@/store";
import { getOHLCV } from "@/lib/api";
import {
  TradingViewChart,
  ChartToolbar,
  DrawingToolbar,
  IndicatorPane,
  IndicatorPicker,
  WatchlistPanel,
  TickerSearch,
  StatusBar,
} from "@/components/chart";
import { Button } from "@/components/ui/button";
import { Tooltip } from "@/components/ui/tooltip";
import { cn, formatPrice, formatIndianNumber, formatPercent } from "@/lib/utils";
import type { ChartConfig } from "@/hooks/useTradingView";
import type { IndicatorType } from "@/components/chart/IndicatorPicker";
import type { IChartApi } from "lightweight-charts";
import {
  PanelRightOpen,
  PanelRightClose,
  Maximize2,
  Minimize2,
  Camera,
} from "lucide-react";

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
    setTimeframe,
    quotes,
  } = useStore();

  const [config, setConfig] = useState<ChartConfig>({ ...defaultConfig, timeframe });
  const [showWatchlist, setShowWatchlist] = useState(true);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [indicatorPanes, setIndicatorPanes] = useState<IndicatorType[]>(["RSI"]);
  const chartInstanceRef = useRef<IChartApi | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Fetch OHLCV data when ticker or timeframe changes
  useEffect(() => {
    if (!selectedTicker) return;
    getOHLCV(selectedTicker, config.timeframe)
      .then(setChartData)
      .catch(() => {});
  }, [selectedTicker, config.timeframe, setChartData]);

  // Sync timeframe from config → store
  useEffect(() => {
    setTimeframe(config.timeframe);
  }, [config.timeframe, setTimeframe]);

  const quote = quotes[selectedTicker];

  const handleConfigChange = useCallback((newConfig: ChartConfig) => {
    setConfig(newConfig);
  }, []);

  const toggleIndicatorPane = useCallback((id: IndicatorType) => {
    setIndicatorPanes((prev) =>
      prev.includes(id) ? prev.filter((p) => p !== id) : [...prev, id],
    );
  }, []);

  const handleFullscreen = useCallback(() => {
    if (!containerRef.current) return;
    if (!isFullscreen) {
      containerRef.current.requestFullscreen?.();
    } else {
      document.exitFullscreen?.();
    }
    setIsFullscreen(!isFullscreen);
  }, [isFullscreen]);

  const handleScreenshot = useCallback(() => {
    if (!chartInstanceRef.current) return;
    const canvas = chartInstanceRef.current.takeScreenshot();
    const link = document.createElement("a");
    link.download = `${selectedTicker}_${config.timeframe}_${Date.now()}.png`;
    link.href = canvas.toDataURL();
    link.click();
  }, [selectedTicker, config.timeframe]);

  return (
    <div
      ref={containerRef}
      className={cn(
        "flex h-[calc(100vh-4rem)] flex-col overflow-hidden rounded-lg border bg-background",
        isFullscreen && "h-screen rounded-none border-none",
      )}
    >
      {/* ══════════ Top Header Bar ══════════ */}
      <div className="flex items-center gap-2 border-b px-2 py-1">
        {/* Ticker search */}
        <TickerSearch />

        {/* Ticker info */}
        {quote && (
          <div className="flex items-center gap-3 border-l pl-3">
            <span className="text-sm font-bold tabular-nums">
              {formatPrice(quote.price)}
            </span>
            <span
              className={cn(
                "text-xs tabular-nums font-medium",
                quote.change >= 0
                  ? "text-green-600 dark:text-green-400"
                  : "text-red-600 dark:text-red-400",
              )}
            >
              {quote.change >= 0 ? "+" : ""}
              {formatIndianNumber(quote.change)} ({formatPercent(quote.changePercent)})
            </span>
          </div>
        )}

        {/* Toolbar: timeframes + overlay indicators */}
        <div className="ml-2 flex-1">
          <ChartToolbar config={config} onConfigChange={handleConfigChange} />
        </div>

        {/* Indicator pane picker */}
        <div className="relative">
          <IndicatorPicker
            activeIndicators={indicatorPanes}
            onToggle={toggleIndicatorPane}
          />
        </div>

        {/* Right actions */}
        <div className="flex items-center gap-1 border-l pl-2">
          <Tooltip content="Screenshot" side="bottom">
            <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={handleScreenshot}>
              <Camera className="h-3.5 w-3.5" />
            </Button>
          </Tooltip>
          <Tooltip content={isFullscreen ? "Exit Fullscreen" : "Fullscreen"} side="bottom">
            <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={handleFullscreen}>
              {isFullscreen ? <Minimize2 className="h-3.5 w-3.5" /> : <Maximize2 className="h-3.5 w-3.5" />}
            </Button>
          </Tooltip>
          <Tooltip content={showWatchlist ? "Hide Watchlist" : "Show Watchlist"} side="bottom">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0"
              onClick={() => setShowWatchlist(!showWatchlist)}
            >
              {showWatchlist ? (
                <PanelRightClose className="h-3.5 w-3.5" />
              ) : (
                <PanelRightOpen className="h-3.5 w-3.5" />
              )}
            </Button>
          </Tooltip>
        </div>
      </div>

      {/* ══════════ Main Content Area ══════════ */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left: Drawing Toolbar */}
        <DrawingToolbar />

        {/* Center: Chart + Indicator Panes */}
        <div className="flex flex-1 flex-col overflow-hidden">
          {/* Main candlestick chart */}
          <div className="flex-1 min-h-0">
            <TradingViewChart
              data={chartData}
              config={config}
              height="100%"
              onChartReady={(c) => { chartInstanceRef.current = c; }}
            />
          </div>

          {/* Indicator sub-panes (RSI, MACD, etc.) */}
          {indicatorPanes.map((type) => (
            <IndicatorPane
              key={type}
              data={chartData}
              type={type}
              onRemove={() => toggleIndicatorPane(type)}
              syncTimeScale={chartInstanceRef.current}
            />
          ))}
        </div>

        {/* Right: Watchlist */}
        {showWatchlist && <WatchlistPanel />}
      </div>

      {/* ══════════ Bottom Status Bar ══════════ */}
      <StatusBar />
    </div>
  );
}
