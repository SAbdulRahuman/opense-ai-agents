"use client";

import { useCallback } from "react";
import { useStore } from "@/store";
import {
  QueryEditor,
  QueryHistory,
  ResultScalar,
  ResultTable,
  ResultChart,
  AlertManager,
} from "@/components/financeql";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { executeQuery, explainQuery } from "@/lib/api";

export default function FinanceQLPage() {
  const {
    currentQuery,
    setCurrentQuery,
    queryResult,
    setQueryResult,
    resultTab,
    setResultTab,
    isExecuting,
    setExecuting,
    naturalLanguageMode,
    setNaturalLanguageMode,
    timeRange,
    setTimeRange,
    queryHistory,
    addToHistory,
    toggleHistoryStar,
  } = useStore();

  const handleExecute = useCallback(async () => {
    if (!currentQuery.trim()) return;
    setExecuting(true);
    try {
      const result = await executeQuery(currentQuery, timeRange);
      setQueryResult(result);
      addToHistory({
        id: crypto.randomUUID(),
        query: currentQuery,
        resultType: result.type,
        duration: result.duration ?? 0,
        timestamp: new Date().toISOString(),
        starred: false,
      });
    } catch {
      // TODO: toast
    } finally {
      setExecuting(false);
    }
  }, [currentQuery, timeRange, setExecuting, setQueryResult, addToHistory]);

  const handleExplain = useCallback(async () => {
    if (!currentQuery.trim()) return;
    try {
      await explainQuery(currentQuery);
    } catch {
      // TODO: toast
    }
  }, [currentQuery]);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">FinanceQL Explorer</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Query Indian market data with the FinanceQL language or natural language
        </p>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-4 gap-4">
        {/* Left: Editor + Results */}
        <div className="xl:col-span-3 space-y-4">
          <QueryEditor
            value={currentQuery}
            onChange={setCurrentQuery}
            onExecute={handleExecute}
            onExplain={handleExplain}
            isExecuting={isExecuting}
            naturalLanguageMode={naturalLanguageMode}
            onToggleNL={setNaturalLanguageMode}
            timeRange={timeRange}
            onTimeRangeChange={setTimeRange}
          />

          {/* Results */}
          {isExecuting && (
            <Card>
              <CardContent className="py-8 text-center">
                <div className="animate-spin h-8 w-8 border-2 border-primary border-t-transparent rounded-full mx-auto mb-3" />
                <p className="text-sm text-muted-foreground">Executing query…</p>
              </CardContent>
            </Card>
          )}

          {queryResult && !isExecuting && (
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">Result</CardTitle>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline">{queryResult.type}</Badge>
                    {queryResult.duration && (
                      <span className="text-xs text-muted-foreground">
                        {queryResult.duration}ms
                      </span>
                    )}
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                {queryResult.type === "scalar" && (
                  <ResultScalar data={queryResult.data as any} />
                )}
                {queryResult.type === "table" && (
                  <ResultTable data={queryResult.data as any} />
                )}
                {(queryResult.type === "vector" || queryResult.type === "matrix") && (
                  <Tabs value={resultTab} onValueChange={(v) => setResultTab(v as "table" | "graph")}>
                    <TabsList>
                      <TabsTrigger value="chart">Chart</TabsTrigger>
                      <TabsTrigger value="table">Table</TabsTrigger>
                    </TabsList>
                    <TabsContent value="chart">
                      <ResultChart data={queryResult.data as any} />
                    </TabsContent>
                    <TabsContent value="table">
                      <ResultTable data={queryResult.data as any} />
                    </TabsContent>
                  </Tabs>
                )}
              </CardContent>
            </Card>
          )}
        </div>

        {/* Right: History + Alerts */}
        <div className="space-y-4">
          <QueryHistory
            entries={queryHistory}
            onSelect={setCurrentQuery}
            onToggleStar={toggleHistoryStar}
          />
          <AlertManager />
        </div>
      </div>
    </div>
  );
}
