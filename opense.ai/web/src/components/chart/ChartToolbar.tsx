"use client";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { ChartConfig } from "@/hooks/useTradingView";

const timeframes = ["1m", "5m", "15m", "1h", "1D", "1W", "1M"];
const indicators = [
  { id: "SMA20", label: "SMA 20" },
  { id: "SMA50", label: "SMA 50" },
  { id: "SMA200", label: "SMA 200" },
  { id: "EMA20", label: "EMA 20" },
  { id: "BB", label: "Bollinger" },
];

interface ChartToolbarProps {
  config: ChartConfig;
  onConfigChange: (config: ChartConfig) => void;
}

export function ChartToolbar({ config, onConfigChange }: ChartToolbarProps) {
  const toggleIndicator = (id: string) => {
    const current = config.indicators;
    const next = current.includes(id)
      ? current.filter((i) => i !== id)
      : [...current, id];
    onConfigChange({ ...config, indicators: next });
  };

  return (
    <div className="flex flex-wrap items-center gap-2 border-b p-2">
      {/* Timeframe selector */}
      <div className="flex items-center gap-1 border-r pr-2">
        {timeframes.map((tf) => (
          <Button
            key={tf}
            variant={config.timeframe === tf ? "default" : "ghost"}
            size="sm"
            className="h-7 px-2 text-xs"
            onClick={() => onConfigChange({ ...config, timeframe: tf })}
          >
            {tf}
          </Button>
        ))}
      </div>

      {/* Indicator toggles */}
      <div className="flex items-center gap-1 border-r pr-2">
        {indicators.map((ind) => (
          <Button
            key={ind.id}
            variant={config.indicators.includes(ind.id) ? "secondary" : "ghost"}
            size="sm"
            className={cn(
              "h-7 px-2 text-xs",
              config.indicators.includes(ind.id) && "border border-primary/30",
            )}
            onClick={() => toggleIndicator(ind.id)}
          >
            {ind.label}
          </Button>
        ))}
      </div>

      {/* Volume toggle */}
      <Button
        variant={config.showVolume ? "secondary" : "ghost"}
        size="sm"
        className="h-7 px-2 text-xs"
        onClick={() => onConfigChange({ ...config, showVolume: !config.showVolume })}
      >
        Volume
      </Button>
    </div>
  );
}
