"use client";

import { useRef, useEffect, useState } from "react";
import { useTheme } from "next-themes";
import {
  createChart,
  LineSeries,
  HistogramSeries,
  type IChartApi,
  type ISeriesApi,
  type LineData,
  type HistogramData,
  type Time,
  ColorType,
} from "lightweight-charts";
import type { OHLCV } from "@/lib/types";

interface IndicatorOverlayProps {
  data: OHLCV[];
  type: "RSI" | "MACD";
  className?: string;
}

export function IndicatorOverlay({ data, type, className }: IndicatorOverlayProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  useEffect(() => {
    if (!containerRef.current || !mounted || data.length === 0) return;
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
      height: 150,
      rightPriceScale: { borderColor: isDark ? "#1e293b" : "#e5e7eb" },
      timeScale: { borderColor: isDark ? "#1e293b" : "#e5e7eb", visible: false },
    });

    if (type === "RSI") {
      renderRSI(chart, data, isDark);
    } else if (type === "MACD") {
      renderMACD(chart, data, isDark);
    }

    const observer = new ResizeObserver((entries) => {
      chart.applyOptions({ width: entries[0].contentRect.width });
    });
    observer.observe(containerRef.current);

    chart.timeScale().fitContent();

    return () => {
      observer.disconnect();
      chart.remove();
    };
  }, [data, type, resolvedTheme, mounted]);

  if (!mounted) return null;

  return (
    <div className={className}>
      <div className="flex items-center gap-2 px-2 py-1 text-xs text-muted-foreground font-medium">
        {type}
      </div>
      <div ref={containerRef} />
    </div>
  );
}

function renderRSI(chart: IChartApi, data: OHLCV[], isDark: boolean) {
  const period = 14;
  const closes = data.map((d) => d.close);
  const times = data.map((d) => d.time as Time);
  const rsiData: LineData[] = [];

  for (let i = period; i < closes.length; i++) {
    let gains = 0;
    let losses = 0;
    for (let j = i - period + 1; j <= i; j++) {
      const change = closes[j] - closes[j - 1];
      if (change > 0) gains += change;
      else losses -= change;
    }
    const avgGain = gains / period;
    const avgLoss = losses / period;
    const rs = avgLoss === 0 ? 100 : avgGain / avgLoss;
    const rsi = 100 - 100 / (1 + rs);
    rsiData.push({ time: times[i], value: rsi });
  }

  const series = chart.addSeries(LineSeries, {
    color: isDark ? "#a78bfa" : "#7c3aed",
    lineWidth: 1,
    priceLineVisible: false,
  });
  series.setData(rsiData);

  // Overbought/oversold lines
  const obLine = chart.addSeries(LineSeries, {
    color: "rgba(239, 68, 68, 0.3)",
    lineWidth: 1,
    lineStyle: 2,
    priceLineVisible: false,
    lastValueVisible: false,
  });
  obLine.setData(rsiData.map((d) => ({ time: d.time, value: 70 })));

  const osLine = chart.addSeries(LineSeries, {
    color: "rgba(34, 197, 94, 0.3)",
    lineWidth: 1,
    lineStyle: 2,
    priceLineVisible: false,
    lastValueVisible: false,
  });
  osLine.setData(rsiData.map((d) => ({ time: d.time, value: 30 })));
}

function renderMACD(chart: IChartApi, data: OHLCV[], isDark: boolean) {
  const closes = data.map((d) => d.close);
  const times = data.map((d) => d.time as Time);

  // Compute EMAs
  const ema12 = computeEMAValues(closes, 12);
  const ema26 = computeEMAValues(closes, 26);

  // MACD line = EMA12 - EMA26
  const macdLine: number[] = [];
  for (let i = 0; i < closes.length; i++) {
    if (ema12[i] !== null && ema26[i] !== null) {
      macdLine.push(ema12[i]! - ema26[i]!);
    } else {
      macdLine.push(0);
    }
  }

  // Signal line = 9-period EMA of MACD
  const signalLine = computeEMAValues(macdLine, 9);
  const startIdx = 33; // ~26 + 9 - 2

  const macdData: LineData[] = [];
  const signalData: LineData[] = [];
  const histData: HistogramData[] = [];

  for (let i = startIdx; i < closes.length; i++) {
    macdData.push({ time: times[i], value: macdLine[i] });
    if (signalLine[i] !== null) {
      signalData.push({ time: times[i], value: signalLine[i]! });
      const hist = macdLine[i] - signalLine[i]!;
      histData.push({
        time: times[i],
        value: hist,
        color: hist >= 0
          ? (isDark ? "rgba(34, 197, 94, 0.5)" : "rgba(22, 163, 74, 0.5)")
          : (isDark ? "rgba(239, 68, 68, 0.5)" : "rgba(220, 38, 38, 0.5)"),
      });
    }
  }

  // Histogram
  const histSeries = chart.addSeries(HistogramSeries, {
    priceLineVisible: false,
    lastValueVisible: false,
  });
  histSeries.setData(histData);

  // MACD line
  const macdSeries = chart.addSeries(LineSeries, {
    color: isDark ? "#3b82f6" : "#2563eb",
    lineWidth: 1,
    priceLineVisible: false,
  });
  macdSeries.setData(macdData);

  // Signal line 
  const signalSeries = chart.addSeries(LineSeries, {
    color: isDark ? "#f97316" : "#ea580c",
    lineWidth: 1,
    priceLineVisible: false,
  });
  signalSeries.setData(signalData);
}

function computeEMAValues(values: number[], period: number): (number | null)[] {
  const result: (number | null)[] = new Array(values.length).fill(null);
  const k = 2 / (period + 1);

  let sum = 0;
  for (let i = 0; i < period && i < values.length; i++) {
    sum += values[i];
  }

  if (values.length >= period) {
    result[period - 1] = sum / period;
    for (let i = period; i < values.length; i++) {
      result[i] = values[i] * k + (result[i - 1] ?? 0) * (1 - k);
    }
  }

  return result;
}
