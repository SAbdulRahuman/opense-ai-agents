// ============================================================================
// OpeNSE.ai — DrawingCanvas (transparent overlay on top of chart)
// ============================================================================
"use client";

import { useRef, useEffect, useState } from "react";
import { useDrawingTools } from "@/hooks/useDrawingTools";
import { useStore } from "@/store";
import { cn } from "@/lib/utils";
import type { IChartApi } from "lightweight-charts";

interface DrawingCanvasProps {
  chart: IChartApi | null;
  className?: string;
}

export function DrawingCanvas({ chart, className }: DrawingCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [size, setSize] = useState({ width: 0, height: 0 });
  const { activeTool } = useStore();

  // Track container size
  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;

    const update = () => {
      setSize({ width: el.clientWidth, height: el.clientHeight });
    };
    update();

    const observer = new ResizeObserver(update);
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  const { handleMouseDown, handleMouseUp, handleDoubleClick, isToolActive } =
    useDrawingTools({
      canvasRef,
      chart,
      containerWidth: size.width,
      containerHeight: size.height,
    });

  // Track mouse for brush tool (add pending points on move while drawing)
  const { isDrawing, addPendingPoint, activeTool: tool } = useStore();
  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isDrawing || tool !== "brush") return;
    const rect = e.currentTarget.getBoundingClientRect();
    addPendingPoint({ time: e.clientX - rect.left, price: e.clientY - rect.top });
  };

  const cursorClass = isToolActive
    ? "cursor-crosshair"
    : "cursor-default pointer-events-none";

  return (
    <div
      ref={containerRef}
      className={cn("absolute inset-0 z-10", className)}
      style={{ pointerEvents: isToolActive ? "auto" : "none" }}
    >
      <canvas
        ref={canvasRef}
        className={cn("absolute inset-0 h-full w-full", cursorClass)}
        style={{
          width: size.width,
          height: size.height,
        }}
        onMouseDown={handleMouseDown}
        onMouseUp={handleMouseUp}
        onMouseMove={handleMouseMove}
        onDoubleClick={handleDoubleClick}
      />
    </div>
  );
}
