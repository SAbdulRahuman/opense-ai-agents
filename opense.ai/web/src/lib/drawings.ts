// ============================================================================
// OpeNSE.ai — Drawing Types & Utilities for TradingView-style Charts
// ============================================================================

// --- Drawing Tool IDs ---
export type DrawingToolId =
  | "cursor"
  | "crosshair"
  | "trendline"
  | "ray"
  | "horizontal"
  | "vertical"
  | "parallel"
  | "fibRetracement"
  | "fibExtension"
  | "rectangle"
  | "circle"
  | "triangle"
  | "text"
  | "callout"
  | "measure"
  | "brush"
  | "eraser";

// --- Point on chart (price + time) ---
export interface ChartPoint {
  time: number; // Unix timestamp
  price: number;
}

// --- Drawing object persisted in store ---
export interface Drawing {
  id: string;
  tool: DrawingToolId;
  points: ChartPoint[];
  style: DrawingStyle;
  visible: boolean;
  locked: boolean;
  label?: string;
  /** Fibonacci levels (0–1 ratios) for fib tools */
  fibLevels?: number[];
  createdAt: number;
}

export interface DrawingStyle {
  color: string;
  lineWidth: number;
  lineStyle: "solid" | "dashed" | "dotted";
  fillColor?: string;
  fillOpacity?: number;
  fontSize?: number;
}

// --- Toolbar category grouping ---
export interface ToolGroup {
  label: string;
  tools: ToolDef[];
}

export interface ToolDef {
  id: DrawingToolId;
  label: string;
  icon: string; // lucide icon name
  shortcut?: string;
}

// --- Default tool groups (TradingView-like) ---
export const TOOL_GROUPS: ToolGroup[] = [
  {
    label: "Cursors",
    tools: [
      { id: "cursor", label: "Pointer", icon: "MousePointer2", shortcut: "V" },
      { id: "crosshair", label: "Crosshair", icon: "Crosshair", shortcut: "C" },
    ],
  },
  {
    label: "Lines",
    tools: [
      { id: "trendline", label: "Trend Line", icon: "TrendingUp", shortcut: "T" },
      { id: "ray", label: "Ray", icon: "MoveUpRight" },
      { id: "horizontal", label: "Horizontal Line", icon: "Minus", shortcut: "H" },
      { id: "vertical", label: "Vertical Line", icon: "Grip" },
      { id: "parallel", label: "Parallel Channel", icon: "Columns3" },
    ],
  },
  {
    label: "Fibonacci",
    tools: [
      { id: "fibRetracement", label: "Fib Retracement", icon: "GitBranch" },
      { id: "fibExtension", label: "Fib Extension", icon: "GitMerge" },
    ],
  },
  {
    label: "Shapes",
    tools: [
      { id: "rectangle", label: "Rectangle", icon: "Square", shortcut: "R" },
      { id: "circle", label: "Circle", icon: "Circle" },
      { id: "triangle", label: "Triangle", icon: "Triangle" },
    ],
  },
  {
    label: "Annotation",
    tools: [
      { id: "text", label: "Text", icon: "Type" },
      { id: "callout", label: "Callout", icon: "MessageSquare" },
      { id: "brush", label: "Brush", icon: "Paintbrush", shortcut: "B" },
    ],
  },
  {
    label: "Measure",
    tools: [
      { id: "measure", label: "Price Range", icon: "Ruler" },
      { id: "eraser", label: "Eraser", icon: "Eraser", shortcut: "E" },
    ],
  },
];

// --- Default styles per tool ---
export const DEFAULT_STYLE: DrawingStyle = {
  color: "#2962ff",
  lineWidth: 2,
  lineStyle: "solid",
  fillColor: "#2962ff",
  fillOpacity: 0.1,
  fontSize: 12,
};

export const DEFAULT_FIB_LEVELS = [0, 0.236, 0.382, 0.5, 0.618, 0.786, 1];

// --- Factory ---
export function createDrawing(
  tool: DrawingToolId,
  points: ChartPoint[],
  style?: Partial<DrawingStyle>,
): Drawing {
  return {
    id: crypto.randomUUID(),
    tool,
    points,
    style: { ...DEFAULT_STYLE, ...style },
    visible: true,
    locked: false,
    fibLevels: tool === "fibRetracement" || tool === "fibExtension" ? DEFAULT_FIB_LEVELS : undefined,
    createdAt: Date.now(),
  };
}

// --- Helpers for coordinate conversion ---
export function pointsNeeded(tool: DrawingToolId): number {
  switch (tool) {
    case "cursor":
    case "crosshair":
    case "eraser":
      return 0;
    case "horizontal":
    case "vertical":
    case "text":
    case "callout":
      return 1;
    case "trendline":
    case "ray":
    case "fibRetracement":
    case "fibExtension":
    case "rectangle":
    case "circle":
    case "measure":
      return 2;
    case "parallel":
    case "triangle":
      return 3;
    case "brush":
      return Infinity; // freeform — finish on mouse up
    default:
      return 2;
  }
}

// --- Color presets (TradingView palette) ---
export const COLOR_PRESETS = [
  "#2962ff", // blue
  "#e91e63", // pink
  "#ff9800", // orange
  "#4caf50", // green
  "#00bcd4", // cyan
  "#9c27b0", // purple
  "#f44336", // red
  "#ffeb3b", // yellow
  "#ffffff", // white
  "#787b86", // gray
];
