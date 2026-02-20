"use client";

import { useRef } from "react";
import { useTheme } from "next-themes";
import { useTradingView, type ChartConfig } from "@/hooks/useTradingView";
import { ChartLegend } from "./ChartLegend";
import type { OHLCV } from "@/lib/types";

interface TradingViewChartProps {
  data: OHLCV[];
  config: ChartConfig;
  className?: string;
}

export function TradingViewChart({ data, config, className }: TradingViewChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { resolvedTheme } = useTheme();
  const theme = (resolvedTheme || "dark") as "light" | "dark";

  const { crosshairData } = useTradingView({
    containerRef,
    data,
    config,
    theme,
  });

  return (
    <div className={className}>
      <div className="relative">
        {crosshairData && <ChartLegend data={crosshairData} />}
        <div ref={containerRef} className="h-[500px] w-full" />
      </div>
    </div>
  );
}
