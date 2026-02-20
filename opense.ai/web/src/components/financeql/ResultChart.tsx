"use client";

import { useRef, useEffect, useState } from "react";
import { useTheme } from "next-themes";
import {
  createChart,
  LineSeries,
  type IChartApi,
  type LineData,
  type Time,
  ColorType,
} from "lightweight-charts";
import type { MatrixResult } from "@/lib/types";

const lineColors = ["#3b82f6", "#22c55e", "#f97316", "#a78bfa", "#f43f5e", "#14b8a6"];

interface ResultChartProps {
  data: MatrixResult;
}

export function ResultChart({ data }: ResultChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  useEffect(() => {
    if (!containerRef.current || !mounted || data.series.length === 0) return;
    const isDark = resolvedTheme === "dark";

    const chart = createChart(containerRef.current, {
      layout: {
        background: {
          type: ColorType.Solid,
          color: isDark ? "#0a0a0a" : "#ffffff",
        },
        textColor: isDark ? "#d1d5db" : "#374151",
      },
      grid: {
        vertLines: { color: isDark ? "#1e293b" : "#e5e7eb" },
        horzLines: { color: isDark ? "#1e293b" : "#e5e7eb" },
      },
      width: containerRef.current.clientWidth,
      height: 350,
      rightPriceScale: { borderColor: isDark ? "#1e293b" : "#e5e7eb" },
      timeScale: {
        borderColor: isDark ? "#1e293b" : "#e5e7eb",
        timeVisible: true,
      },
    });

    data.series.forEach((s, i) => {
      const series = chart.addSeries(LineSeries, {
        color: lineColors[i % lineColors.length],
        lineWidth: 2,
        title: s.label,
      });
      const lineData: LineData[] = s.data.map((d) => ({
        time: d.time as Time,
        value: d.value,
      }));
      series.setData(lineData);
    });

    chart.timeScale().fitContent();

    const observer = new ResizeObserver((entries) => {
      chart.applyOptions({ width: entries[0].contentRect.width });
    });
    observer.observe(containerRef.current);

    return () => {
      observer.disconnect();
      chart.remove();
    };
  }, [data, resolvedTheme, mounted]);

  if (!mounted) return null;

  return (
    <div>
      {/* Legend */}
      <div className="flex flex-wrap gap-3 mb-2 px-1">
        {data.series.map((s, i) => (
          <div key={s.label} className="flex items-center gap-1.5 text-xs">
            <span
              className="h-2.5 w-2.5 rounded-full"
              style={{ backgroundColor: lineColors[i % lineColors.length] }}
            />
            {s.label}
          </div>
        ))}
      </div>
      <div ref={containerRef} className="rounded-md border" />
    </div>
  );
}
