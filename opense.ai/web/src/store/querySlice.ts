// ============================================================================
// OpeNSE.ai — Query Store Slice (Zustand) — FinanceQL State
// ============================================================================

import type { StateCreator } from "zustand";
import type { QueryResult, QueryHistoryEntry, Alert } from "@/lib/types";

export interface QuerySlice {
  currentQuery: string;
  queryResult: QueryResult | null;
  isExecuting: boolean;
  queryHistory: QueryHistoryEntry[];
  alerts: Alert[];
  resultTab: "table" | "graph";
  naturalLanguageMode: boolean;
  timeRange: { start?: string; end?: string; relative?: string };

  setCurrentQuery: (query: string) => void;
  setQueryResult: (result: QueryResult | null) => void;
  setExecuting: (executing: boolean) => void;
  addToHistory: (entry: QueryHistoryEntry) => void;
  toggleHistoryStar: (id: string) => void;
  setAlerts: (alerts: Alert[]) => void;
  addAlert: (alert: Alert) => void;
  removeAlert: (id: string) => void;
  setResultTab: (tab: "table" | "graph") => void;
  setNaturalLanguageMode: (mode: boolean) => void;
  setTimeRange: (range: { start?: string; end?: string; relative?: string }) => void;
}

export const createQuerySlice: StateCreator<QuerySlice> = (set) => ({
  currentQuery: "",
  queryResult: null,
  isExecuting: false,
  queryHistory: [],
  alerts: [],
  resultTab: "table",
  naturalLanguageMode: false,
  timeRange: { relative: "30d" },

  setCurrentQuery: (currentQuery) => set({ currentQuery }),
  setQueryResult: (queryResult) => set({ queryResult }),
  setExecuting: (isExecuting) => set({ isExecuting }),

  addToHistory: (entry) =>
    set((state) => ({
      queryHistory: [entry, ...state.queryHistory].slice(0, 100),
    })),

  toggleHistoryStar: (id) =>
    set((state) => ({
      queryHistory: state.queryHistory.map((e) =>
        e.id === id ? { ...e, starred: !e.starred } : e,
      ),
    })),

  setAlerts: (alerts) => set({ alerts }),

  addAlert: (alert) =>
    set((state) => ({ alerts: [...state.alerts, alert] })),

  removeAlert: (id) =>
    set((state) => ({
      alerts: state.alerts.filter((a) => a.id !== id),
    })),

  setResultTab: (resultTab) => set({ resultTab }),
  setNaturalLanguageMode: (naturalLanguageMode) => set({ naturalLanguageMode }),
  setTimeRange: (timeRange) => set({ timeRange }),
});
