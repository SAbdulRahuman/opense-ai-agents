"use client";

import { useEffect, useRef, useCallback, useState } from "react";
import {
  createChart,
  CandlestickSeries,
  HistogramSeries,
  LineSeries,
  type IChartApi,
  type ISeriesApi,
  type CandlestickData,
  type HistogramData,
  type LineData,
  type Time,
  ColorType,
  CrosshairMode,
} from "lightweight-charts";
import type { OHLCV } from "@/lib/types";

export interface ChartConfig {
  showVolume: boolean;
  indicators: string[];
  timeframe: string;
}

interface UseTradingViewProps {
  containerRef: React.RefObject<HTMLDivElement | null>;
  data: OHLCV[];
  config: ChartConfig;
  theme: "light" | "dark";
}

interface UseTradingViewReturn {
  chart: IChartApi | null;
  candleSeries: ISeriesApi<"Candlestick"> | null;
  updateData: (candle: OHLCV) => void;
  crosshairData: CrosshairData | null;
}

export interface CrosshairData {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export function useTradingView({
  containerRef,
  data,
  config,
  theme,
}: UseTradingViewProps): UseTradingViewReturn {
  const chartRef = useRef<IChartApi | null>(null);
  const candleSeriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null);
  const volumeSeriesRef = useRef<ISeriesApi<"Histogram"> | null>(null);
  const indicatorSeriesRef = useRef<Map<string, ISeriesApi<"Line">>>(new Map());
  const [crosshairData, setCrosshairData] = useState<CrosshairData | null>(null);

  const colors = theme === "dark"
    ? {
        background: "#0a0a0a",
        text: "#d1d5db",
        grid: "#1e293b",
        upColor: "#22c55e",
        downColor: "#ef4444",
        volumeUp: "rgba(34, 197, 94, 0.3)",
        volumeDown: "rgba(239, 68, 68, 0.3)",
      }
    : {
        background: "#ffffff",
        text: "#374151",
        grid: "#e5e7eb",
        upColor: "#16a34a",
        downColor: "#dc2626",
        volumeUp: "rgba(22, 163, 74, 0.3)",
        volumeDown: "rgba(220, 38, 38, 0.3)",
      };

  // Initialize chart
  useEffect(() => {
    if (!containerRef.current) return;

    const chart = createChart(containerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: colors.background },
        textColor: colors.text,
      },
      grid: {
        vertLines: { color: colors.grid },
        horzLines: { color: colors.grid },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
      },
      rightPriceScale: {
        borderColor: colors.grid,
      },
      timeScale: {
        borderColor: colors.grid,
        timeVisible: true,
        secondsVisible: false,
      },
      width: containerRef.current.clientWidth,
      height: containerRef.current.clientHeight,
    });

    // Candlestick series
    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: colors.upColor,
      downColor: colors.downColor,
      borderDownColor: colors.downColor,
      borderUpColor: colors.upColor,
      wickDownColor: colors.downColor,
      wickUpColor: colors.upColor,
    });

    chartRef.current = chart;
    candleSeriesRef.current = candleSeries;

    // Handle resize
    const observer = new ResizeObserver((entries) => {
      const { width, height } = entries[0].contentRect;
      chart.applyOptions({ width, height });
    });
    observer.observe(containerRef.current);

    // Crosshair data
    chart.subscribeCrosshairMove((param) => {
      if (!param.time || !param.seriesData) {
        setCrosshairData(null);
        return;
      }
      const candleData = param.seriesData.get(candleSeries) as CandlestickData | undefined;
      const volumeData = param.seriesData.get(volumeSeriesRef.current!) as HistogramData | undefined;
      if (candleData) {
        setCrosshairData({
          time: candleData.time as number,
          open: candleData.open,
          high: candleData.high,
          low: candleData.low,
          close: candleData.close,
          volume: volumeData?.value || 0,
        });
      }
    });

    return () => {
      observer.disconnect();
      chart.remove();
      chartRef.current = null;
      candleSeriesRef.current = null;
      volumeSeriesRef.current = null;
      indicatorSeriesRef.current.clear();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [theme]);

  // Update data
  useEffect(() => {
    if (!candleSeriesRef.current || data.length === 0) return;

    const candles: CandlestickData[] = data.map((d) => ({
      time: d.time as Time,
      open: d.open,
      high: d.high,
      low: d.low,
      close: d.close,
    }));

    candleSeriesRef.current.setData(candles);

    // Volume
    if (config.showVolume && chartRef.current) {
      if (!volumeSeriesRef.current) {
        volumeSeriesRef.current = chartRef.current.addSeries(HistogramSeries, {
          priceFormat: { type: "volume" },
          priceScaleId: "volume",
        });
        chartRef.current.priceScale("volume").applyOptions({
          scaleMargins: { top: 0.8, bottom: 0 },
        });
      }

      const volumes: HistogramData[] = data.map((d) => ({
        time: d.time as Time,
        value: d.volume,
        color: d.close >= d.open ? colors.volumeUp : colors.volumeDown,
      }));

      volumeSeriesRef.current.setData(volumes);
    }

    // Indicators
    updateIndicators(data, config.indicators);

    chartRef.current?.timeScale().fitContent();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data, config.showVolume, config.indicators]);

  const updateIndicators = useCallback(
    (ohlcv: OHLCV[], indicators: string[]) => {
      if (!chartRef.current) return;

      // Clear old indicators that are no longer in list
      indicatorSeriesRef.current.forEach((series, key) => {
        if (!indicators.includes(key)) {
          chartRef.current!.removeSeries(series);
          indicatorSeriesRef.current.delete(key);
        }
      });

      const closes = ohlcv.map((d) => d.close);
      const times = ohlcv.map((d) => d.time as Time);

      indicators.forEach((indicator) => {
        let lineData: LineData[] = [];
        let color = "#fbbf24";

        switch (indicator) {
          case "SMA20":
            lineData = computeSMA(closes, times, 20);
            color = "#fbbf24";
            break;
          case "SMA50":
            lineData = computeSMA(closes, times, 50);
            color = "#a78bfa";
            break;
          case "SMA200":
            lineData = computeSMA(closes, times, 200);
            color = "#f97316";
            break;
          case "EMA20":
            lineData = computeEMA(closes, times, 20);
            color = "#38bdf8";
            break;
          case "BB":
            // Bollinger Bands — upper + lower
            addBollingerBands(ohlcv);
            return;
          default:
            return;
        }

        if (!indicatorSeriesRef.current.has(indicator)) {
          const series = chartRef.current!.addSeries(LineSeries, {
            color,
            lineWidth: 1,
            priceLineVisible: false,
            lastValueVisible: false,
          });
          indicatorSeriesRef.current.set(indicator, series);
        }
        indicatorSeriesRef.current.get(indicator)!.setData(lineData);
      });
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  const addBollingerBands = useCallback((ohlcv: OHLCV[]) => {
    if (!chartRef.current) return;
    const closes = ohlcv.map((d) => d.close);
    const times = ohlcv.map((d) => d.time as Time);
    const period = 20;
    const stdDev = 2;

    const upper: LineData[] = [];
    const lower: LineData[] = [];

    for (let i = period - 1; i < closes.length; i++) {
      const slice = closes.slice(i - period + 1, i + 1);
      const mean = slice.reduce((a, b) => a + b, 0) / period;
      const variance = slice.reduce((a, b) => a + (b - mean) ** 2, 0) / period;
      const sd = Math.sqrt(variance);
      upper.push({ time: times[i], value: mean + stdDev * sd });
      lower.push({ time: times[i], value: mean - stdDev * sd });
    }

    if (!indicatorSeriesRef.current.has("BB_upper")) {
      const upperSeries = chartRef.current!.addSeries(LineSeries, {
        color: "rgba(59, 130, 246, 0.5)",
        lineWidth: 1,
        priceLineVisible: false,
        lastValueVisible: false,
      });
      indicatorSeriesRef.current.set("BB_upper", upperSeries);
    }
    indicatorSeriesRef.current.get("BB_upper")!.setData(upper);

    if (!indicatorSeriesRef.current.has("BB_lower")) {
      const lowerSeries = chartRef.current!.addSeries(LineSeries, {
        color: "rgba(59, 130, 246, 0.5)",
        lineWidth: 1,
        priceLineVisible: false,
        lastValueVisible: false,
      });
      indicatorSeriesRef.current.set("BB_lower", lowerSeries);
    }
    indicatorSeriesRef.current.get("BB_lower")!.setData(lower);
  }, []);

  const updateData = useCallback((candle: OHLCV) => {
    if (!candleSeriesRef.current) return;
    candleSeriesRef.current.update({
      time: candle.time as Time,
      open: candle.open,
      high: candle.high,
      low: candle.low,
      close: candle.close,
    });
    if (volumeSeriesRef.current) {
      volumeSeriesRef.current.update({
        time: candle.time as Time,
        value: candle.volume,
        color: candle.close >= candle.open ? colors.volumeUp : colors.volumeDown,
      });
    }
  }, [colors.volumeUp, colors.volumeDown]);

  return {
    chart: chartRef.current,
    candleSeries: candleSeriesRef.current,
    updateData,
    crosshairData,
  };
}

// --- Helper: Simple Moving Average ---
function computeSMA(closes: number[], times: Time[], period: number): LineData[] {
  const result: LineData[] = [];
  for (let i = period - 1; i < closes.length; i++) {
    const sum = closes.slice(i - period + 1, i + 1).reduce((a, b) => a + b, 0);
    result.push({ time: times[i], value: sum / period });
  }
  return result;
}

// --- Helper: Exponential Moving Average ---
function computeEMA(closes: number[], times: Time[], period: number): LineData[] {
  const result: LineData[] = [];
  const k = 2 / (period + 1);
  let ema = closes.slice(0, period).reduce((a, b) => a + b, 0) / period;

  for (let i = period - 1; i < closes.length; i++) {
    if (i === period - 1) {
      ema = closes.slice(0, period).reduce((a, b) => a + b, 0) / period;
    } else {
      ema = closes[i] * k + ema * (1 - k);
    }
    result.push({ time: times[i], value: ema });
  }
  return result;
}
