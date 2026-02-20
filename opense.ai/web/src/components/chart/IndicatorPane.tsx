// ============================================================================
// OpeNSE.ai — IndicatorPane (resizable sub-chart pane for RSI/MACD/etc.)
// ============================================================================
"use client";

import { useRef, useEffect, useState, useCallback } from "react";
import { useTheme } from "next-themes";
import {
  createChart,
  LineSeries,
  HistogramSeries,
  type IChartApi,
  type LineData,
  type HistogramData,
  type Time,
  ColorType,
} from "lightweight-charts";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { X, GripHorizontal } from "lucide-react";
import type { OHLCV } from "@/lib/types";

type IndicatorType = "RSI" | "MACD" | "Stochastic" | "ATR" | "OBV";

interface IndicatorPaneProps {
  data: OHLCV[];
  type: IndicatorType;
  onRemove: () => void;
  className?: string;
  syncTimeScale?: IChartApi | null;
}

const INITIAL_HEIGHT = 150;
const MIN_HEIGHT = 80;
const MAX_HEIGHT = 300;

export function IndicatorPane({ data, type, onRemove, className, syncTimeScale }: IndicatorPaneProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  const [height, setHeight] = useState(INITIAL_HEIGHT);
  const isDragging = useRef(false);

  useEffect(() => setMounted(true), []);

  // Create & update chart
  useEffect(() => {
    if (!containerRef.current || !mounted || data.length === 0) return;
    const isDark = resolvedTheme === "dark";

    const chart = createChart(containerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: isDark ? "#0a0a0a" : "#ffffff" },
        textColor: isDark ? "#9ca3af" : "#6b7280",
      },
      grid: {
        vertLines: { color: isDark ? "#1e293b" : "#f3f4f6" },
        horzLines: { color: isDark ? "#1e293b" : "#f3f4f6" },
      },
      width: containerRef.current.clientWidth,
      height,
      rightPriceScale: {
        borderColor: isDark ? "#1e293b" : "#e5e7eb",
        scaleMargins: { top: 0.1, bottom: 0.1 },
      },
      timeScale: {
        borderColor: isDark ? "#1e293b" : "#e5e7eb",
        visible: false,
      },
      handleScroll: { vertTouchDrag: false },
    });

    chartRef.current = chart;

    switch (type) {
      case "RSI":
        renderRSI(chart, data, isDark);
        break;
      case "MACD":
        renderMACD(chart, data, isDark);
        break;
      case "Stochastic":
        renderStochastic(chart, data, isDark);
        break;
      case "ATR":
        renderATR(chart, data, isDark);
        break;
      case "OBV":
        renderOBV(chart, data, isDark);
        break;
    }

    chart.timeScale().fitContent();

    // Sync time scale with main chart
    if (syncTimeScale) {
      syncTimeScale.timeScale().subscribeVisibleLogicalRangeChange((range) => {
        if (range) chart.timeScale().setVisibleLogicalRange(range);
      });
    }

    const observer = new ResizeObserver((entries) => {
      chart.applyOptions({ width: entries[0].contentRect.width });
    });
    observer.observe(containerRef.current);

    return () => {
      observer.disconnect();
      chart.remove();
      chartRef.current = null;
    };
  }, [data, type, resolvedTheme, mounted, height, syncTimeScale]);

  // Drag to resize
  const handleResizeStart = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    isDragging.current = true;
    const startY = e.clientY;
    const startH = height;

    const onMove = (ev: MouseEvent) => {
      if (!isDragging.current) return;
      const delta = ev.clientY - startY;
      setHeight(Math.max(MIN_HEIGHT, Math.min(MAX_HEIGHT, startH + delta)));
    };
    const onUp = () => {
      isDragging.current = false;
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
    };
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  }, [height]);

  if (!mounted) return null;

  return (
    <div className={cn("relative border-t", className)}>
      {/* Header */}
      <div className="flex items-center justify-between bg-card/80 px-2 py-0.5 text-xs">
        <span className="font-medium text-muted-foreground">{type}</span>
        <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={onRemove}>
          <X className="h-3 w-3" />
        </Button>
      </div>

      {/* Chart container */}
      <div ref={containerRef} style={{ height }} />

      {/* Resize handle */}
      <div
        className="group flex h-2 cursor-row-resize items-center justify-center hover:bg-primary/10"
        onMouseDown={handleResizeStart}
      >
        <GripHorizontal className="h-3 w-3 text-muted-foreground/50 group-hover:text-muted-foreground" />
      </div>
    </div>
  );
}

// ============================================================================
// Indicator render functions
// ============================================================================

function renderRSI(chart: IChartApi, data: OHLCV[], isDark: boolean) {
  const period = 14;
  const closes = data.map((d) => d.close);
  const times = data.map((d) => d.time as Time);
  const rsiData: LineData[] = [];

  for (let i = period; i < closes.length; i++) {
    let gains = 0, losses = 0;
    for (let j = i - period + 1; j <= i; j++) {
      const change = closes[j] - closes[j - 1];
      if (change > 0) gains += change;
      else losses -= change;
    }
    const avgGain = gains / period;
    const avgLoss = losses / period;
    const rs = avgLoss === 0 ? 100 : avgGain / avgLoss;
    rsiData.push({ time: times[i], value: 100 - 100 / (1 + rs) });
  }

  chart.addSeries(LineSeries, {
    color: isDark ? "#a78bfa" : "#7c3aed",
    lineWidth: 1,
    priceLineVisible: false,
  }).setData(rsiData);

  // Overbought / oversold reference lines
  const refData = rsiData.map((d) => d.time);
  chart.addSeries(LineSeries, {
    color: "rgba(239, 68, 68, 0.25)",
    lineWidth: 1,
    lineStyle: 2,
    priceLineVisible: false,
    lastValueVisible: false,
  }).setData(refData.map((t) => ({ time: t, value: 70 })));

  chart.addSeries(LineSeries, {
    color: "rgba(34, 197, 94, 0.25)",
    lineWidth: 1,
    lineStyle: 2,
    priceLineVisible: false,
    lastValueVisible: false,
  }).setData(refData.map((t) => ({ time: t, value: 30 })));
}

function renderMACD(chart: IChartApi, data: OHLCV[], isDark: boolean) {
  const closes = data.map((d) => d.close);
  const times = data.map((d) => d.time as Time);
  const ema12 = emaValues(closes, 12);
  const ema26 = emaValues(closes, 26);

  const macdLine: number[] = closes.map((_, i) =>
    ema12[i] != null && ema26[i] != null ? ema12[i]! - ema26[i]! : 0,
  );
  const signalLine = emaValues(macdLine, 9);
  const startIdx = 33;

  const macdData: LineData[] = [];
  const signalData: LineData[] = [];
  const histData: HistogramData[] = [];

  for (let i = startIdx; i < closes.length; i++) {
    macdData.push({ time: times[i], value: macdLine[i] });
    if (signalLine[i] != null) {
      signalData.push({ time: times[i], value: signalLine[i]! });
      const h = macdLine[i] - signalLine[i]!;
      histData.push({
        time: times[i],
        value: h,
        color: h >= 0
          ? (isDark ? "rgba(34,197,94,0.5)" : "rgba(22,163,74,0.5)")
          : (isDark ? "rgba(239,68,68,0.5)" : "rgba(220,38,38,0.5)"),
      });
    }
  }

  chart.addSeries(HistogramSeries, { priceLineVisible: false, lastValueVisible: false }).setData(histData);
  chart.addSeries(LineSeries, { color: isDark ? "#3b82f6" : "#2563eb", lineWidth: 1, priceLineVisible: false }).setData(macdData);
  chart.addSeries(LineSeries, { color: isDark ? "#f97316" : "#ea580c", lineWidth: 1, priceLineVisible: false }).setData(signalData);
}

function renderStochastic(chart: IChartApi, data: OHLCV[], isDark: boolean) {
  const period = 14;
  const kSmooth = 3;
  const times = data.map((d) => d.time as Time);
  const kValues: LineData[] = [];

  for (let i = period - 1; i < data.length; i++) {
    const slice = data.slice(i - period + 1, i + 1);
    const high = Math.max(...slice.map((d) => d.high));
    const low = Math.min(...slice.map((d) => d.low));
    const k = high === low ? 50 : ((data[i].close - low) / (high - low)) * 100;
    kValues.push({ time: times[i], value: k });
  }

  // %D = 3-period SMA of %K
  const dValues: LineData[] = [];
  for (let i = kSmooth - 1; i < kValues.length; i++) {
    const sum = kValues.slice(i - kSmooth + 1, i + 1).reduce((a, b) => a + b.value, 0);
    dValues.push({ time: kValues[i].time, value: sum / kSmooth });
  }

  chart.addSeries(LineSeries, { color: isDark ? "#3b82f6" : "#2563eb", lineWidth: 1, priceLineVisible: false }).setData(kValues);
  chart.addSeries(LineSeries, { color: isDark ? "#f97316" : "#ea580c", lineWidth: 1, priceLineVisible: false }).setData(dValues);

  // Reference lines 80/20
  const refTimes = kValues.map((d) => d.time);
  chart.addSeries(LineSeries, { color: "rgba(239,68,68,0.25)", lineWidth: 1, lineStyle: 2, priceLineVisible: false, lastValueVisible: false })
    .setData(refTimes.map((t) => ({ time: t, value: 80 })));
  chart.addSeries(LineSeries, { color: "rgba(34,197,94,0.25)", lineWidth: 1, lineStyle: 2, priceLineVisible: false, lastValueVisible: false })
    .setData(refTimes.map((t) => ({ time: t, value: 20 })));
}

function renderATR(chart: IChartApi, data: OHLCV[], isDark: boolean) {
  const period = 14;
  const times = data.map((d) => d.time as Time);
  const atrData: LineData[] = [];

  for (let i = 1; i < data.length; i++) {
    const tr = Math.max(
      data[i].high - data[i].low,
      Math.abs(data[i].high - data[i - 1].close),
      Math.abs(data[i].low - data[i - 1].close),
    );
    if (i >= period) {
      const prevATR = atrData.length > 0 ? atrData[atrData.length - 1].value : tr;
      atrData.push({ time: times[i], value: (prevATR * (period - 1) + tr) / period });
    } else if (i === period - 1) {
      // First ATR = average of first `period` TRs
      let sum = tr;
      for (let j = 1; j < i; j++) {
        sum += Math.max(
          data[j].high - data[j].low,
          Math.abs(data[j].high - data[j - 1].close),
          Math.abs(data[j].low - data[j - 1].close),
        );
      }
      atrData.push({ time: times[i], value: sum / period });
    }
  }

  chart.addSeries(LineSeries, {
    color: isDark ? "#fbbf24" : "#d97706",
    lineWidth: 1,
    priceLineVisible: false,
  }).setData(atrData);
}

function renderOBV(chart: IChartApi, data: OHLCV[], isDark: boolean) {
  const times = data.map((d) => d.time as Time);
  const obvData: LineData[] = [{ time: times[0], value: 0 }];

  for (let i = 1; i < data.length; i++) {
    const prev = obvData[i - 1].value;
    const dir = data[i].close > data[i - 1].close ? 1 : data[i].close < data[i - 1].close ? -1 : 0;
    obvData.push({ time: times[i], value: prev + dir * data[i].volume });
  }

  chart.addSeries(LineSeries, {
    color: isDark ? "#38bdf8" : "#0ea5e9",
    lineWidth: 1,
    priceLineVisible: false,
  }).setData(obvData);
}

// EMA helper
function emaValues(values: number[], period: number): (number | null)[] {
  const result: (number | null)[] = new Array(values.length).fill(null);
  const k = 2 / (period + 1);
  let sum = 0;
  for (let i = 0; i < period && i < values.length; i++) sum += values[i];
  if (values.length >= period) {
    result[period - 1] = sum / period;
    for (let i = period; i < values.length; i++) {
      result[i] = values[i] * k + (result[i - 1] ?? 0) * (1 - k);
    }
  }
  return result;
}
