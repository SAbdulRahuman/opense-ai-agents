// ============================================================================
// OpeNSE.ai — Chart Drawings Store Slice (Zustand)
// ============================================================================

import type { StateCreator } from "zustand";
import type { Drawing, DrawingToolId, DrawingStyle, ChartPoint } from "@/lib/drawings";
import { createDrawing, DEFAULT_STYLE } from "@/lib/drawings";

export interface ChartDrawingsSlice {
  // State
  drawings: Drawing[];
  activeTool: DrawingToolId;
  activeStyle: DrawingStyle;
  selectedDrawingId: string | null;
  pendingPoints: ChartPoint[];
  isDrawing: boolean;
  showDrawings: boolean;
  magnetMode: boolean;

  // Actions
  setActiveTool: (tool: DrawingToolId) => void;
  setActiveStyle: (style: Partial<DrawingStyle>) => void;
  addDrawing: (drawing: Drawing) => void;
  removeDrawing: (id: string) => void;
  updateDrawing: (id: string, patch: Partial<Drawing>) => void;
  selectDrawing: (id: string | null) => void;
  clearAllDrawings: () => void;
  toggleDrawingVisibility: (id: string) => void;
  toggleLockDrawing: (id: string) => void;
  setShowDrawings: (show: boolean) => void;
  setMagnetMode: (enabled: boolean) => void;

  // Drawing flow
  startDrawing: (point: ChartPoint) => void;
  addPendingPoint: (point: ChartPoint) => void;
  finishDrawing: () => void;
  cancelDrawing: () => void;
}

export const createChartDrawingsSlice: StateCreator<ChartDrawingsSlice> = (set, get) => ({
  drawings: [],
  activeTool: "cursor",
  activeStyle: { ...DEFAULT_STYLE },
  selectedDrawingId: null,
  pendingPoints: [],
  isDrawing: false,
  showDrawings: true,
  magnetMode: false,

  setActiveTool: (tool) => set({ activeTool: tool, selectedDrawingId: null }),
  setActiveStyle: (style) =>
    set((state) => ({ activeStyle: { ...state.activeStyle, ...style } })),

  addDrawing: (drawing) =>
    set((state) => ({ drawings: [...state.drawings, drawing] })),

  removeDrawing: (id) =>
    set((state) => ({
      drawings: state.drawings.filter((d) => d.id !== id),
      selectedDrawingId: state.selectedDrawingId === id ? null : state.selectedDrawingId,
    })),

  updateDrawing: (id, patch) =>
    set((state) => ({
      drawings: state.drawings.map((d) => (d.id === id ? { ...d, ...patch } : d)),
    })),

  selectDrawing: (id) => set({ selectedDrawingId: id }),

  clearAllDrawings: () => set({ drawings: [], selectedDrawingId: null }),

  toggleDrawingVisibility: (id) =>
    set((state) => ({
      drawings: state.drawings.map((d) =>
        d.id === id ? { ...d, visible: !d.visible } : d,
      ),
    })),

  toggleLockDrawing: (id) =>
    set((state) => ({
      drawings: state.drawings.map((d) =>
        d.id === id ? { ...d, locked: !d.locked } : d,
      ),
    })),

  setShowDrawings: (show) => set({ showDrawings: show }),
  setMagnetMode: (enabled) => set({ magnetMode: enabled }),

  // Drawing flow: user clicks to place points, then finishDrawing commits
  startDrawing: (point) => set({ pendingPoints: [point], isDrawing: true }),

  addPendingPoint: (point) =>
    set((state) => ({ pendingPoints: [...state.pendingPoints, point] })),

  finishDrawing: () => {
    const { activeTool, pendingPoints, activeStyle } = get();
    if (pendingPoints.length === 0) {
      set({ isDrawing: false, pendingPoints: [] });
      return;
    }
    const drawing = createDrawing(activeTool, pendingPoints, activeStyle);
    set((state) => ({
      drawings: [...state.drawings, drawing],
      pendingPoints: [],
      isDrawing: false,
    }));
  },

  cancelDrawing: () => set({ pendingPoints: [], isDrawing: false }),
});
