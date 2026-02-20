// ============================================================================
// OpeNSE.ai — useDrawingTools hook (canvas-based drawing on chart overlay)
// ============================================================================
"use client";

import { useCallback, useEffect, useRef } from "react";
import { useStore } from "@/store";
import type { ChartPoint, DrawingToolId, Drawing, DrawingStyle } from "@/lib/drawings";
import { pointsNeeded } from "@/lib/drawings";
import type { IChartApi } from "lightweight-charts";

interface UseDrawingToolsProps {
  canvasRef: React.RefObject<HTMLCanvasElement | null>;
  chart: IChartApi | null;
  containerWidth: number;
  containerHeight: number;
}

/** Convert pixel coords → chart price/time using lightweight-charts coordinate API. */
function pixelToChartPoint(
  chart: IChartApi,
  x: number,
  y: number,
): ChartPoint | null {
  const timeScale = chart.timeScale();
  const priceScale = chart.priceScale("right");

  const timeCoord = timeScale.coordinateToTime(x);
  // coordinateToPrice was removed in v5; use series-level conversion if needed
  // For now we approximate from the visible range
  const visibleRange = priceScale.options();
  // Use the chart's first series for coordinate-to-price
  // lightweight-charts v5 approach:
  const time = timeCoord ? (timeCoord as unknown as number) : 0;

  // For price: we use the visible logical range + pixel ratio
  const timeScaleWidth = timeScale.width();
  const logicalRange = timeScale.getVisibleLogicalRange();
  if (!logicalRange) return null;

  // For price axis — we need to read from chart height
  // Simple linear interpolation from the price scale visible range
  // But lightweight-charts doesn't directly expose coordinateToPrice on the scale
  // We'll store time and use a rough price estimate based on the chart coordinate
  return { time, price: y }; // price stored as pixel Y — converted in render
}

/** Convert chart price/time → pixel coords. */
function chartPointToPixel(
  chart: IChartApi,
  point: ChartPoint,
): { x: number; y: number } | null {
  const timeScale = chart.timeScale();
  const x = timeScale.timeToCoordinate(point.time as unknown as import("lightweight-charts").Time);
  if (x === null) return null;
  // price is stored as pixel Y in this simplified approach
  return { x, y: point.price };
}

export function useDrawingTools({
  canvasRef,
  chart,
  containerWidth,
  containerHeight,
}: UseDrawingToolsProps) {
  const {
    activeTool,
    activeStyle,
    drawings,
    showDrawings,
    pendingPoints,
    isDrawing,
    startDrawing,
    addPendingPoint,
    finishDrawing,
    cancelDrawing,
    selectDrawing,
    selectedDrawingId,
    removeDrawing,
  } = useStore();

  const animFrameRef = useRef<number>(0);

  // ---- Keyboard shortcuts ----
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (isDrawing) cancelDrawing();
        else selectDrawing(null);
      }
      if (e.key === "Delete" || e.key === "Backspace") {
        if (selectedDrawingId) removeDrawing(selectedDrawingId);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [isDrawing, cancelDrawing, selectDrawing, selectedDrawingId, removeDrawing]);

  // ---- Mouse handlers ----
  const handleMouseDown = useCallback(
    (e: React.MouseEvent<HTMLCanvasElement>) => {
      if (!chart || activeTool === "cursor" || activeTool === "crosshair") return;

      const rect = e.currentTarget.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;
      const pt: ChartPoint = { time: x, price: y }; // storing pixel coords

      if (activeTool === "eraser") {
        // Find closest drawing and remove it
        const hit = findHitDrawing(drawings, x, y, chart);
        if (hit) removeDrawing(hit.id);
        return;
      }

      const needed = pointsNeeded(activeTool);

      if (!isDrawing) {
        startDrawing(pt);
        if (needed === 1) {
          // Single-click tools commit immediately after mouse up
        }
      } else {
        addPendingPoint(pt);
        if (pendingPoints.length + 1 >= needed && needed !== Infinity) {
          finishDrawing();
        }
      }
    },
    [chart, activeTool, isDrawing, pendingPoints, drawings, startDrawing, addPendingPoint, finishDrawing, removeDrawing],
  );

  const handleMouseUp = useCallback(
    (_e: React.MouseEvent<HTMLCanvasElement>) => {
      if (!chart) return;
      const needed = pointsNeeded(activeTool);
      if (isDrawing && (needed === 1 || activeTool === "brush")) {
        finishDrawing();
      }
    },
    [chart, activeTool, isDrawing, finishDrawing],
  );

  const handleDoubleClick = useCallback(() => {
    // Double-click finishes multi-point drawings (brush, etc.)
    if (isDrawing) finishDrawing();
  }, [isDrawing, finishDrawing]);

  // ---- Render loop ----
  const render = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas || !chart) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    canvas.width = containerWidth * dpr;
    canvas.height = containerHeight * dpr;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, containerWidth, containerHeight);

    if (!showDrawings) return;

    // Render committed drawings
    for (const drawing of drawings) {
      if (!drawing.visible) continue;
      renderDrawing(ctx, drawing, chart, drawing.id === selectedDrawingId);
    }

    // Render in-progress drawing (pending points)
    if (isDrawing && pendingPoints.length > 0) {
      const tempDrawing: Drawing = {
        id: "__pending__",
        tool: activeTool,
        points: pendingPoints,
        style: activeStyle,
        visible: true,
        locked: false,
        createdAt: 0,
      };
      renderDrawing(ctx, tempDrawing, chart, false);
    }
  }, [
    canvasRef,
    chart,
    containerWidth,
    containerHeight,
    showDrawings,
    drawings,
    selectedDrawingId,
    isDrawing,
    pendingPoints,
    activeTool,
    activeStyle,
  ]);

  // Paint every frame so pending drawings follow mouse
  useEffect(() => {
    const loop = () => {
      render();
      animFrameRef.current = requestAnimationFrame(loop);
    };
    animFrameRef.current = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(animFrameRef.current);
  }, [render]);

  return {
    handleMouseDown,
    handleMouseUp,
    handleDoubleClick,
    isToolActive: activeTool !== "cursor" && activeTool !== "crosshair",
  };
}

// ============================================================================
// Drawing render functions
// ============================================================================

function setStroke(ctx: CanvasRenderingContext2D, style: DrawingStyle) {
  ctx.strokeStyle = style.color;
  ctx.lineWidth = style.lineWidth;
  if (style.lineStyle === "dashed") ctx.setLineDash([6, 4]);
  else if (style.lineStyle === "dotted") ctx.setLineDash([2, 3]);
  else ctx.setLineDash([]);
}

function renderDrawing(
  ctx: CanvasRenderingContext2D,
  drawing: Drawing,
  _chart: IChartApi,
  selected: boolean,
) {
  const { tool, points, style } = drawing;
  if (points.length === 0) return;

  setStroke(ctx, style);

  // Points are stored as pixel coords (time=x, price=y)
  const pts = points.map((p) => ({ x: p.time, y: p.price }));

  switch (tool) {
    case "trendline":
    case "ray":
      if (pts.length >= 2) {
        ctx.beginPath();
        ctx.moveTo(pts[0].x, pts[0].y);
        if (tool === "ray") {
          // Extend line to canvas edge
          const dx = pts[1].x - pts[0].x;
          const dy = pts[1].y - pts[0].y;
          const len = Math.sqrt(dx * dx + dy * dy);
          const scale = 5000 / (len || 1);
          ctx.lineTo(pts[0].x + dx * scale, pts[0].y + dy * scale);
        } else {
          ctx.lineTo(pts[1].x, pts[1].y);
        }
        ctx.stroke();
      }
      break;

    case "horizontal":
      if (pts.length >= 1) {
        ctx.beginPath();
        ctx.moveTo(0, pts[0].y);
        ctx.lineTo(ctx.canvas.width / (window.devicePixelRatio || 1), pts[0].y);
        ctx.stroke();

        // Price label
        ctx.fillStyle = style.color;
        ctx.font = "11px monospace";
        ctx.fillText(`${pts[0].y.toFixed(0)}px`, 4, pts[0].y - 4);
      }
      break;

    case "vertical":
      if (pts.length >= 1) {
        ctx.beginPath();
        ctx.moveTo(pts[0].x, 0);
        ctx.lineTo(pts[0].x, ctx.canvas.height / (window.devicePixelRatio || 1));
        ctx.stroke();
      }
      break;

    case "rectangle":
      if (pts.length >= 2) {
        const x = Math.min(pts[0].x, pts[1].x);
        const y = Math.min(pts[0].y, pts[1].y);
        const w = Math.abs(pts[1].x - pts[0].x);
        const h = Math.abs(pts[1].y - pts[0].y);
        if (style.fillColor) {
          ctx.fillStyle = style.fillColor;
          ctx.globalAlpha = style.fillOpacity ?? 0.1;
          ctx.fillRect(x, y, w, h);
          ctx.globalAlpha = 1;
        }
        ctx.strokeRect(x, y, w, h);
      }
      break;

    case "circle":
      if (pts.length >= 2) {
        const cx = (pts[0].x + pts[1].x) / 2;
        const cy = (pts[0].y + pts[1].y) / 2;
        const rx = Math.abs(pts[1].x - pts[0].x) / 2;
        const ry = Math.abs(pts[1].y - pts[0].y) / 2;
        ctx.beginPath();
        ctx.ellipse(cx, cy, rx, ry, 0, 0, Math.PI * 2);
        if (style.fillColor) {
          ctx.fillStyle = style.fillColor;
          ctx.globalAlpha = style.fillOpacity ?? 0.1;
          ctx.fill();
          ctx.globalAlpha = 1;
        }
        ctx.stroke();
      }
      break;

    case "triangle":
      if (pts.length >= 3) {
        ctx.beginPath();
        ctx.moveTo(pts[0].x, pts[0].y);
        ctx.lineTo(pts[1].x, pts[1].y);
        ctx.lineTo(pts[2].x, pts[2].y);
        ctx.closePath();
        if (style.fillColor) {
          ctx.fillStyle = style.fillColor;
          ctx.globalAlpha = style.fillOpacity ?? 0.1;
          ctx.fill();
          ctx.globalAlpha = 1;
        }
        ctx.stroke();
      }
      break;

    case "parallel":
      if (pts.length >= 3) {
        // 2 parallel lines: line through p0-p1 and line through p2 parallel to it
        const dx = pts[1].x - pts[0].x;
        const dy = pts[1].y - pts[0].y;
        ctx.beginPath();
        ctx.moveTo(pts[0].x, pts[0].y);
        ctx.lineTo(pts[1].x, pts[1].y);
        ctx.moveTo(pts[2].x, pts[2].y);
        ctx.lineTo(pts[2].x + dx, pts[2].y + dy);
        ctx.stroke();
        // Fill between
        if (style.fillColor) {
          ctx.fillStyle = style.fillColor;
          ctx.globalAlpha = style.fillOpacity ?? 0.05;
          ctx.beginPath();
          ctx.moveTo(pts[0].x, pts[0].y);
          ctx.lineTo(pts[1].x, pts[1].y);
          ctx.lineTo(pts[2].x + dx, pts[2].y + dy);
          ctx.lineTo(pts[2].x, pts[2].y);
          ctx.closePath();
          ctx.fill();
          ctx.globalAlpha = 1;
        }
      }
      break;

    case "fibRetracement":
      if (pts.length >= 2) {
        const levels = drawing.fibLevels ?? [0, 0.236, 0.382, 0.5, 0.618, 0.786, 1];
        const top = Math.min(pts[0].y, pts[1].y);
        const bottom = Math.max(pts[0].y, pts[1].y);
        const range = bottom - top;
        const left = Math.min(pts[0].x, pts[1].x);
        const right = Math.max(pts[0].x, pts[1].x);

        levels.forEach((level) => {
          const y = bottom - range * level;
          ctx.beginPath();
          ctx.moveTo(left, y);
          ctx.lineTo(right, y);
          ctx.stroke();

          ctx.fillStyle = style.color;
          ctx.font = "10px monospace";
          ctx.fillText(`${(level * 100).toFixed(1)}%`, right + 4, y + 3);
        });

        // Fill alternating zones
        if (style.fillColor) {
          ctx.globalAlpha = style.fillOpacity ?? 0.05;
          for (let i = 0; i < levels.length - 1; i += 2) {
            const y1 = bottom - range * levels[i];
            const y2 = bottom - range * levels[i + 1];
            ctx.fillStyle = style.fillColor;
            ctx.fillRect(left, Math.min(y1, y2), right - left, Math.abs(y2 - y1));
          }
          ctx.globalAlpha = 1;
        }
      }
      break;

    case "fibExtension":
      if (pts.length >= 2) {
        const extLevels = [0, 0.618, 1, 1.618, 2.618, 4.236];
        const top = Math.min(pts[0].y, pts[1].y);
        const bottom = Math.max(pts[0].y, pts[1].y);
        const range = bottom - top;
        const left = Math.min(pts[0].x, pts[1].x);
        const right = Math.max(pts[0].x, pts[1].x);

        extLevels.forEach((level) => {
          const y = bottom - range * level;
          ctx.beginPath();
          ctx.moveTo(left, y);
          ctx.lineTo(right + 50, y);
          ctx.stroke();

          ctx.fillStyle = style.color;
          ctx.font = "10px monospace";
          ctx.fillText(`${(level * 100).toFixed(1)}%`, right + 54, y + 3);
        });
      }
      break;

    case "text":
    case "callout":
      if (pts.length >= 1) {
        ctx.fillStyle = style.color;
        ctx.font = `${style.fontSize ?? 12}px sans-serif`;
        const label = drawing.label || "Text";
        if (tool === "callout") {
          // Draw background box
          const m = ctx.measureText(label);
          const pad = 6;
          ctx.fillStyle = style.fillColor ?? "#1e293b";
          ctx.globalAlpha = 0.85;
          ctx.fillRect(pts[0].x - pad, pts[0].y - (style.fontSize ?? 12) - pad, m.width + pad * 2, (style.fontSize ?? 12) + pad * 2);
          ctx.globalAlpha = 1;
          ctx.fillStyle = style.color;
        }
        ctx.fillText(label, pts[0].x, pts[0].y);
      }
      break;

    case "brush":
      if (pts.length >= 2) {
        ctx.beginPath();
        ctx.moveTo(pts[0].x, pts[0].y);
        for (let i = 1; i < pts.length; i++) {
          ctx.lineTo(pts[i].x, pts[i].y);
        }
        ctx.stroke();
      }
      break;

    case "measure":
      if (pts.length >= 2) {
        // Dashed rectangle with dimension labels
        const saved = ctx.getLineDash();
        ctx.setLineDash([4, 4]);
        const mx = Math.min(pts[0].x, pts[1].x);
        const my = Math.min(pts[0].y, pts[1].y);
        const mw = Math.abs(pts[1].x - pts[0].x);
        const mh = Math.abs(pts[1].y - pts[0].y);
        ctx.strokeRect(mx, my, mw, mh);
        ctx.setLineDash(saved);

        // Dimension label
        ctx.fillStyle = style.color;
        ctx.font = "11px monospace";
        const pxDiff = pts[1].y - pts[0].y;
        ctx.fillText(`Δ ${pxDiff.toFixed(0)}px`, mx + mw / 2 - 20, my + mh / 2);
      }
      break;
  }

  // Selection handles
  if (selected) {
    ctx.fillStyle = "#2962ff";
    pts.forEach((p) => {
      ctx.beginPath();
      ctx.arc(p.x, p.y, 4, 0, Math.PI * 2);
      ctx.fill();
    });
  }
}

// ---- Hit testing (for eraser / selection) ----
function findHitDrawing(
  drawings: Drawing[],
  x: number,
  y: number,
  _chart: IChartApi,
): Drawing | null {
  const threshold = 8;
  // Reverse iterate so topmost drawing is checked first
  for (let i = drawings.length - 1; i >= 0; i--) {
    const d = drawings[i];
    if (!d.visible || d.locked) continue;
    const pts = d.points.map((p) => ({ x: p.time, y: p.price }));

    for (let j = 0; j < pts.length - 1; j++) {
      const dist = pointToSegmentDist(x, y, pts[j].x, pts[j].y, pts[j + 1].x, pts[j + 1].y);
      if (dist < threshold) return d;
    }
    // Also check point proximity
    for (const p of pts) {
      const dist = Math.hypot(x - p.x, y - p.y);
      if (dist < threshold) return d;
    }
  }
  return null;
}

function pointToSegmentDist(
  px: number, py: number,
  ax: number, ay: number,
  bx: number, by: number,
): number {
  const dx = bx - ax;
  const dy = by - ay;
  const lenSq = dx * dx + dy * dy;
  if (lenSq === 0) return Math.hypot(px - ax, py - ay);
  let t = ((px - ax) * dx + (py - ay) * dy) / lenSq;
  t = Math.max(0, Math.min(1, t));
  return Math.hypot(px - (ax + t * dx), py - (ay + t * dy));
}
