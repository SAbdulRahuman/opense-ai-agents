"use client";

import { useCallback } from "react";
import { useStore } from "@/store";
import { executeQuery, explainQuery, naturalLanguageQuery } from "@/lib/api";
import type { QueryHistoryEntry } from "@/lib/types";

export function useFinanceQL() {
  const {
    currentQuery,
    queryResult,
    isExecuting,
    queryHistory,
    alerts,
    resultTab,
    naturalLanguageMode,
    timeRange,
    setCurrentQuery,
    setQueryResult,
    setExecuting,
    addToHistory,
    toggleHistoryStar,
    setResultTab,
    setNaturalLanguageMode,
    setTimeRange,
  } = useStore();

  const execute = useCallback(
    async (query?: string) => {
      const q = query || currentQuery;
      if (!q.trim()) return;

      setExecuting(true);
      const startTime = Date.now();

      try {
        let result;
        if (naturalLanguageMode) {
          const nlResult = await naturalLanguageQuery(q);
          result = nlResult.result;
          // Also update the editor with the generated FinanceQL
          setCurrentQuery(nlResult.financeql);
        } else {
          result = await executeQuery(q, timeRange);
        }

        setQueryResult(result);

        const entry: QueryHistoryEntry = {
          id: `q-${Date.now()}`,
          query: q,
          resultType: result.type,
          duration: Date.now() - startTime,
          timestamp: new Date().toISOString(),
          starred: false,
        };
        addToHistory(entry);
      } catch (error) {
        setQueryResult({
          type: "scalar",
          data: {
            value: 0,
            label: `Error: ${error instanceof Error ? error.message : "Query failed"}`,
          },
          query: q,
          duration: Date.now() - startTime,
          timestamp: new Date().toISOString(),
        });
      } finally {
        setExecuting(false);
      }
    },
    [currentQuery, naturalLanguageMode, timeRange, setExecuting, setQueryResult, addToHistory, setCurrentQuery],
  );

  const explain = useCallback(
    async (query?: string) => {
      const q = query || currentQuery;
      if (!q.trim()) return null;

      try {
        const result = await explainQuery(q);
        return result;
      } catch {
        return null;
      }
    },
    [currentQuery],
  );

  return {
    currentQuery,
    queryResult,
    isExecuting,
    queryHistory,
    alerts,
    resultTab,
    naturalLanguageMode,
    timeRange,
    setCurrentQuery,
    execute,
    explain,
    toggleHistoryStar,
    setResultTab,
    setNaturalLanguageMode,
    setTimeRange,
  };
}
