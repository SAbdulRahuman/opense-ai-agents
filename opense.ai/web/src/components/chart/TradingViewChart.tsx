"use client";

import { useRef } from "react";
import { useTheme } from "next-themes";
import { useTradingView, type ChartConfig } from "@/hooks/useTradingView";
import { ChartLegend } from "./ChartLegend";
import { DrawingCanvas } from "./DrawingCanvas";
import type { OHLCV } from "@/lib/types";
import type { IChartApi } from "lightweight-charts";

interface TradingViewChartProps {
  data: OHLCV[];
  config: ChartConfig;
  className?: string;
  height?: string;
  /** Callback to expose the chart instance to parent */
  onChartReady?: (chart: IChartApi | null) => void;
}

export function TradingViewChart({ data, config, className, height = "500px", onChartReady }: TradingViewChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { resolvedTheme } = useTheme();
  const theme = (resolvedTheme || "dark") as "light" | "dark";

  const { chart, crosshairData } = useTradingView({
    containerRef,
    data,
    config,
    theme,
  });

  // Expose chart ref to parent
  if (onChartReady) {
    onChartReady(chart);
  }

  return (
    <div className={className}>
      <div className="relative">
        {crosshairData && <ChartLegend data={crosshairData} />}
        <DrawingCanvas chart={chart} />
        <div ref={containerRef} className="w-full" style={{ height }} />
      </div>
    </div>
  );
}
