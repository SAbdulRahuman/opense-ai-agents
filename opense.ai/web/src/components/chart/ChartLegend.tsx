"use client";

import type { CrosshairData } from "@/hooks/useTradingView";
import { formatPrice, formatVolume } from "@/lib/utils";
import { cn } from "@/lib/utils";

interface ChartLegendProps {
  data: CrosshairData;
}

export function ChartLegend({ data }: ChartLegendProps) {
  const change = data.close - data.open;
  const isPositive = change >= 0;

  return (
    <div className="absolute left-2 top-2 z-10 flex flex-wrap gap-3 rounded-md bg-card/80 px-3 py-1.5 text-xs backdrop-blur-sm">
      <span className="text-muted-foreground">
        O: <span className={cn(isPositive ? "text-green-500" : "text-red-500")}>{formatPrice(data.open)}</span>
      </span>
      <span className="text-muted-foreground">
        H: <span className={cn(isPositive ? "text-green-500" : "text-red-500")}>{formatPrice(data.high)}</span>
      </span>
      <span className="text-muted-foreground">
        L: <span className={cn(isPositive ? "text-green-500" : "text-red-500")}>{formatPrice(data.low)}</span>
      </span>
      <span className="text-muted-foreground">
        C: <span className={cn(isPositive ? "text-green-500" : "text-red-500")}>{formatPrice(data.close)}</span>
      </span>
      <span className="text-muted-foreground">
        Vol: <span className="text-foreground">{formatVolume(data.volume)}</span>
      </span>
    </div>
  );
}
