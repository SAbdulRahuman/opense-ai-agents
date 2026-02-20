// ============================================================================
// OpeNSE.ai — StatusBar (bottom bar showing chart info)
// ============================================================================
"use client";

import { useStore } from "@/store";
import { formatPrice, formatVolume, formatPercent } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { Wifi, WifiOff, Clock } from "lucide-react";

export function StatusBar() {
  const { selectedTicker, quotes, isMarketOpen, chartData, timeframe } = useStore();
  const quote = quotes[selectedTicker];
  const lastCandle = chartData.length > 0 ? chartData[chartData.length - 1] : null;

  return (
    <div className="flex h-6 items-center gap-4 border-t bg-card px-3 text-[10px] text-muted-foreground">
      {/* Market status */}
      <div className="flex items-center gap-1">
        {isMarketOpen ? (
          <>
            <Wifi className="h-2.5 w-2.5 text-green-500" />
            <span className="text-green-600 dark:text-green-400">Live</span>
          </>
        ) : (
          <>
            <WifiOff className="h-2.5 w-2.5" />
            <span>Closed</span>
          </>
        )}
      </div>

      {/* Ticker + Exchange */}
      <div className="font-mono font-bold">{selectedTicker}</div>
      <div>NSE</div>

      {/* Timeframe */}
      <div className="flex items-center gap-1">
        <Clock className="h-2.5 w-2.5" />
        {timeframe}
      </div>

      {/* OHLCV of last candle */}
      {lastCandle && (
        <>
          <div className="border-l pl-3 tabular-nums">
            O {formatPrice(lastCandle.open)}
          </div>
          <div className="tabular-nums">H {formatPrice(lastCandle.high)}</div>
          <div className="tabular-nums">L {formatPrice(lastCandle.low)}</div>
          <div className="tabular-nums">C {formatPrice(lastCandle.close)}</div>
          <div className="tabular-nums">V {formatVolume(lastCandle.volume)}</div>
        </>
      )}

      {/* Spacer */}
      <div className="flex-1" />

      {/* Change */}
      {quote && (
        <div
          className={cn(
            "tabular-nums font-mono",
            quote.change >= 0
              ? "text-green-600 dark:text-green-400"
              : "text-red-600 dark:text-red-400",
          )}
        >
          {quote.change >= 0 ? "+" : ""}
          {formatPercent(quote.changePercent)}
        </div>
      )}

      {/* Data count */}
      <div className="tabular-nums">{chartData.length} bars</div>
    </div>
  );
}
