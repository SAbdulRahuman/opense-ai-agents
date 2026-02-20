"use client";

import { useRef, useEffect, useState, useCallback } from "react";
import { useTheme } from "next-themes";
import {
  createChart,
  LineSeries,
  AreaSeries,
  type IChartApi,
  type LineData,
  type Time,
  ColorType,
} from "lightweight-charts";
import {
  Minus,
  Plus,
  ChevronLeft,
  ChevronRight,
  RefreshCw,
  Layers,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import type { MatrixResult, TableResult, VectorResult, ScalarResult, QueryResult } from "@/lib/types";

// Prometheus-style colors
const GRAPH_COLORS = [
  "#73bf69", // green
  "#f2cc0c", // yellow
  "#ff9830", // orange
  "#e02f44", // red
  "#8ab8ff", // light blue
  "#ca95e5", // purple
  "#ff6eb4", // pink
  "#4ecdc4", // teal
  "#f77b72", // salmon
  "#b877d9", // lavender
];

const RANGE_OPTIONS = [
  { label: "5m", seconds: 5 * 60 },
  { label: "15m", seconds: 15 * 60 },
  { label: "30m", seconds: 30 * 60 },
  { label: "1h", seconds: 60 * 60 },
  { label: "3h", seconds: 3 * 60 * 60 },
  { label: "6h", seconds: 6 * 60 * 60 },
  { label: "12h", seconds: 12 * 60 * 60 },
  { label: "1d", seconds: 24 * 60 * 60 },
  { label: "2d", seconds: 2 * 24 * 60 * 60 },
  { label: "7d", seconds: 7 * 24 * 60 * 60 },
  { label: "30d", seconds: 30 * 24 * 60 * 60 },
  { label: "90d", seconds: 90 * 24 * 60 * 60 },
  { label: "1y", seconds: 365 * 24 * 60 * 60 },
];

interface GraphPanelProps {
  queryResult: QueryResult | null;
  isExecuting: boolean;
}

function formatEndTime(date: Date): string {
  return date.toISOString().slice(0, 19).replace("T", " ");
}

/**
 * Convert query result data into a common series format for the graph.
 */
function resultToSeries(
  result: QueryResult | null,
): Array<{ label: string; data: Array<{ time: number; value: number }> }> {
  if (!result) return [];
  const { type, data } = result;

  if (type === "matrix") {
    return (data as MatrixResult).series;
  }

  if (type === "vector") {
    const vec = data as VectorResult;
    // Convert vector items into a single "instant" series for bar-like display
    // We'll create a synthetic time series from vector data
    const now = Math.floor(Date.now() / 1000);
    return vec.items.map((item, i) => ({
      label: item.ticker || `series-${i}`,
      data: [{ time: now, value: item.value }],
    }));
  }

  if (type === "scalar") {
    const scalar = data as ScalarResult;
    const now = Math.floor(Date.now() / 1000);
    return [
      {
        label: scalar.label || "result",
        data: [{ time: now, value: scalar.value }],
      },
    ];
  }

  if (type === "table") {
    const table = data as TableResult;
    // Try to find time + numeric columns for graphing
    const timeCol = table.columns.find(
      (c) =>
        c.toLowerCase().includes("time") ||
        c.toLowerCase().includes("date") ||
        c.toLowerCase() === "timestamp",
    );
    const numCols = table.columns.filter(
      (c) => c !== timeCol && table.rows.length > 0 && typeof table.rows[0][c] === "number",
    );
    if (timeCol && numCols.length > 0) {
      return numCols.map((col) => ({
        label: col,
        data: table.rows
          .map((row) => ({
            time: typeof row[timeCol] === "number" ? (row[timeCol] as number) : new Date(row[timeCol] as string).getTime() / 1000,
            value: row[col] as number,
          }))
          .filter((d) => !isNaN(d.time) && !isNaN(d.value))
          .sort((a, b) => a.time - b.time),
      }));
    }
    return [];
  }

  return [];
}

export function GraphPanel({ queryResult, isExecuting }: GraphPanelProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  const [selectedRange, setSelectedRange] = useState("1h");
  const [endTime, setEndTime] = useState(() => formatEndTime(new Date()));
  const [resolution, setResolution] = useState("");
  const [stacked, setStacked] = useState(false);
  const [showLegend, setShowLegend] = useState(true);

  useEffect(() => setMounted(true), []);

  const series = resultToSeries(queryResult);
  const hasSeries = series.length > 0 && series.some((s) => s.data.length > 0);

  // Build chart
  useEffect(() => {
    if (!containerRef.current || !mounted) return;

    const isDark = resolvedTheme === "dark";

    // Remove old chart
    if (chartRef.current) {
      chartRef.current.remove();
      chartRef.current = null;
    }

    const chart = createChart(containerRef.current, {
      layout: {
        background: {
          type: ColorType.Solid,
          color: isDark ? "#181b1f" : "#ffffff",
        },
        textColor: isDark ? "#9ca3af" : "#6b7280",
        fontFamily: "system-ui, -apple-system, sans-serif",
        fontSize: 11,
      },
      grid: {
        vertLines: { color: isDark ? "#1f2937" : "#f3f4f6", style: 1 },
        horzLines: { color: isDark ? "#1f2937" : "#f3f4f6", style: 1 },
      },
      width: containerRef.current.clientWidth,
      height: 380,
      rightPriceScale: {
        borderColor: isDark ? "#374151" : "#e5e7eb",
        scaleMargins: { top: 0.1, bottom: 0.1 },
      },
      timeScale: {
        borderColor: isDark ? "#374151" : "#e5e7eb",
        timeVisible: true,
        secondsVisible: false,
        rightOffset: 5,
      },
      crosshair: {
        vertLine: {
          color: isDark ? "#4b5563" : "#9ca3af",
          width: 1,
          style: 2,
          labelBackgroundColor: isDark ? "#374151" : "#6b7280",
        },
        horzLine: {
          color: isDark ? "#4b5563" : "#9ca3af",
          width: 1,
          style: 2,
          labelBackgroundColor: isDark ? "#374151" : "#6b7280",
        },
      },
    });

    chartRef.current = chart;

    if (hasSeries) {
      series.forEach((s, i) => {
        const color = GRAPH_COLORS[i % GRAPH_COLORS.length];
        if (stacked) {
          const areaSeries = chart.addSeries(AreaSeries, {
            lineColor: color,
            topColor: color + "40",
            bottomColor: color + "05",
            lineWidth: 1,
            title: s.label,
          });
          const lineData: LineData[] = s.data.map((d) => ({
            time: d.time as Time,
            value: d.value,
          }));
          areaSeries.setData(lineData);
        } else {
          const lineSeries = chart.addSeries(LineSeries, {
            color: color,
            lineWidth: 2,
            title: s.label,
            crosshairMarkerRadius: 3,
          });
          const lineData: LineData[] = s.data.map((d) => ({
            time: d.time as Time,
            value: d.value,
          }));
          lineSeries.setData(lineData);
        }
      });

      chart.timeScale().fitContent();
    }

    const observer = new ResizeObserver((entries) => {
      if (chartRef.current) {
        chartRef.current.applyOptions({ width: entries[0].contentRect.width });
      }
    });
    observer.observe(containerRef.current);

    return () => {
      observer.disconnect();
      if (chartRef.current) {
        chartRef.current.remove();
        chartRef.current = null;
      }
    };
  }, [series, resolvedTheme, mounted, stacked, hasSeries]);

  const handleRangeChange = useCallback(
    (range: string) => {
      setSelectedRange(range);
    },
    [],
  );

  const handlePanLeft = useCallback(() => {
    const rangeObj = RANGE_OPTIONS.find((r) => r.label === selectedRange);
    if (!rangeObj) return;
    const current = new Date(endTime.replace(" ", "T"));
    current.setSeconds(current.getSeconds() - rangeObj.seconds / 2);
    setEndTime(formatEndTime(current));
  }, [endTime, selectedRange]);

  const handlePanRight = useCallback(() => {
    const rangeObj = RANGE_OPTIONS.find((r) => r.label === selectedRange);
    if (!rangeObj) return;
    const current = new Date(endTime.replace(" ", "T"));
    current.setSeconds(current.getSeconds() + rangeObj.seconds / 2);
    setEndTime(formatEndTime(current));
  }, [endTime, selectedRange]);

  const handleZoomOut = useCallback(() => {
    const idx = RANGE_OPTIONS.findIndex((r) => r.label === selectedRange);
    if (idx < RANGE_OPTIONS.length - 1) {
      setSelectedRange(RANGE_OPTIONS[idx + 1].label);
    }
  }, [selectedRange]);

  const handleZoomIn = useCallback(() => {
    const idx = RANGE_OPTIONS.findIndex((r) => r.label === selectedRange);
    if (idx > 0) {
      setSelectedRange(RANGE_OPTIONS[idx - 1].label);
    }
  }, [selectedRange]);

  const handleResetTime = useCallback(() => {
    setEndTime(formatEndTime(new Date()));
  }, []);

  if (!mounted) return null;

  return (
    <div className="prometheus-graph-panel">
      {/* Prometheus-style graph controls bar */}
      <div className="flex flex-wrap items-center gap-1 px-3 py-2 border-b bg-muted/30">
        {/* Range duration buttons */}
        <div className="flex items-center rounded border bg-background overflow-hidden">
          {RANGE_OPTIONS.map((opt) => (
            <button
              key={opt.label}
              onClick={() => handleRangeChange(opt.label)}
              className={cn(
                "px-2 py-1 text-xs font-medium transition-colors border-r last:border-r-0",
                selectedRange === opt.label
                  ? "bg-primary text-primary-foreground"
                  : "hover:bg-muted text-muted-foreground",
              )}
            >
              {opt.label}
            </button>
          ))}
        </div>

        <div className="flex items-center gap-1 ml-2">
          {/* Pan & Zoom controls */}
          <Button
            variant="outline"
            size="icon"
            className="h-7 w-7"
            onClick={handlePanLeft}
            title="Pan left"
          >
            <ChevronLeft className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-7 w-7"
            onClick={handlePanRight}
            title="Pan right"
          >
            <ChevronRight className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-7 w-7"
            onClick={handleZoomIn}
            title="Zoom in"
          >
            <Plus className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-7 w-7"
            onClick={handleZoomOut}
            title="Zoom out"
          >
            <Minus className="h-3.5 w-3.5" />
          </Button>
        </div>

        {/* End time input */}
        <div className="flex items-center gap-1 ml-3">
          <span className="text-xs text-muted-foreground">End time:</span>
          <Input
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
            className="h-7 w-44 text-xs font-mono"
          />
          <Button
            variant="outline"
            size="icon"
            className="h-7 w-7"
            onClick={handleResetTime}
            title="Reset to now"
          >
            <RefreshCw className="h-3 w-3" />
          </Button>
        </div>

        {/* Resolution */}
        <div className="flex items-center gap-1 ml-3">
          <span className="text-xs text-muted-foreground">Res:</span>
          <Input
            value={resolution}
            onChange={(e) => setResolution(e.target.value)}
            placeholder="auto"
            className="h-7 w-16 text-xs font-mono"
          />
        </div>

        {/* Stacked toggle */}
        <Button
          variant={stacked ? "secondary" : "outline"}
          size="sm"
          className="h-7 ml-auto gap-1 text-xs"
          onClick={() => setStacked(!stacked)}
        >
          <Layers className="h-3 w-3" />
          {stacked ? "Stacked" : "Lines"}
        </Button>
      </div>

      {/* Graph area */}
      <div className="relative min-h-[380px]">
        {isExecuting && (
          <div className="absolute inset-0 flex items-center justify-center bg-background/60 z-10">
            <div className="flex items-center gap-2">
              <div className="animate-spin h-5 w-5 border-2 border-primary border-t-transparent rounded-full" />
              <span className="text-sm text-muted-foreground">Loading…</span>
            </div>
          </div>
        )}

        {!hasSeries && !isExecuting && (
          <div className="absolute inset-0 flex flex-col items-center justify-center text-muted-foreground">
            <svg
              className="h-12 w-12 opacity-20 mb-3"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"
              />
            </svg>
            <p className="text-sm">No data to graph</p>
            <p className="text-xs mt-1 opacity-60">
              Enter a FinanceQL expression and press Execute
            </p>
          </div>
        )}

        <div ref={containerRef} />
      </div>

      {/* Legend (Prometheus-style) */}
      {hasSeries && showLegend && (
        <div className="px-3 py-2 border-t bg-muted/20">
          <div className="flex flex-wrap gap-x-4 gap-y-1">
            {series
              .filter((s) => s.data.length > 0)
              .map((s, i) => {
                const color = GRAPH_COLORS[i % GRAPH_COLORS.length];
                const lastVal = s.data[s.data.length - 1]?.value;
                return (
                  <div
                    key={s.label}
                    className="flex items-center gap-1.5 text-xs cursor-default group"
                  >
                    <span
                      className="inline-block h-[3px] w-3 rounded-full"
                      style={{ backgroundColor: color }}
                    />
                    <span className="text-foreground font-medium">
                      {s.label}
                    </span>
                    {lastVal !== undefined && (
                      <span className="text-muted-foreground">
                        {typeof lastVal === "number"
                          ? lastVal.toLocaleString("en-IN", {
                              maximumFractionDigits: 2,
                            })
                          : lastVal}
                      </span>
                    )}
                  </div>
                );
              })}
          </div>
        </div>
      )}
    </div>
  );
}
