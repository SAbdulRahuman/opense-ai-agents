"use client";

import { useCallback, useState } from "react";
import { useStore } from "@/store";
import {
  QueryEditor,
  QueryHistory,
  ResultScalar,
  ResultTable,
  AlertManager,
  ExpressionTree,
  type ASTNode,
} from "@/components/financeql";
import { GraphPanel } from "@/components/financeql/GraphPanel";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Globe, HelpCircle } from "lucide-react";
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

  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [explaining, setExplaining] = useState(false);
  const [explanation, setExplanation] = useState<string | null>(null);
  const [exprTreeAST, setExprTreeAST] = useState<ASTNode | null>(null);

  const handleExecute = useCallback(async () => {
    if (!currentQuery.trim()) return;
    setExecuting(true);
    setErrorMsg(null);
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
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Query execution failed");
    } finally {
      setExecuting(false);
    }
  }, [currentQuery, timeRange, setExecuting, setQueryResult, addToHistory]);

  const handleExplain = useCallback(async () => {
    if (!currentQuery.trim()) return;
    setExplaining(true);
    setExplanation(null);
    setExprTreeAST(null);
    try {
      const result = await explainQuery(currentQuery);
      setExplanation(result.explanation);
      if (result.ast) {
        setExprTreeAST(result.ast as ASTNode);
      }
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Explain failed");
    } finally {
      setExplaining(false);
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

      {/* Error Banner */}
      {errorMsg && (
        <div className="px-4 py-2 rounded-md bg-destructive/10 border border-destructive/20 text-destructive text-sm flex items-center gap-2">
          <Globe className="h-4 w-4 shrink-0" />
          <span className="font-mono text-xs">{errorMsg}</span>
          <button
            className="ml-auto text-xs underline"
            onClick={() => setErrorMsg(null)}
          >
            dismiss
          </button>
        </div>
      )}

      <div className="grid grid-cols-1 xl:grid-cols-4 gap-4">
        {/* Left: Editor + Results */}
        <div className="xl:col-span-3 space-y-4">
          <QueryEditor
            value={currentQuery}
            onChange={setCurrentQuery}
            onExecute={handleExecute}
            onExplain={handleExplain}
            isExecuting={isExecuting || explaining}
            naturalLanguageMode={naturalLanguageMode}
            onToggleNL={setNaturalLanguageMode}
            timeRange={timeRange}
            onTimeRangeChange={setTimeRange}
          />

          {/* Explanation banner */}
          {explanation && (
            <Card>
              <CardContent className="py-3">
                <div className="flex items-start gap-2">
                  <HelpCircle className="h-4 w-4 text-blue-500 shrink-0 mt-0.5" />
                  <p className="text-sm text-foreground leading-relaxed">{explanation}</p>
                  <button
                    className="ml-auto text-xs text-muted-foreground hover:text-foreground underline shrink-0"
                    onClick={() => setExplanation(null)}
                  >
                    dismiss
                  </button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Expression Tree */}
          {exprTreeAST && (
            <ExpressionTree ast={exprTreeAST} />
          )}

          {/* Result Card — always visible */}
          <Card>
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">Result</CardTitle>
                {queryResult && !isExecuting && (
                  <div className="flex items-center gap-2">
                    <Badge variant="outline">{queryResult.type}</Badge>
                    {queryResult.duration != null && (
                      <span className="text-xs text-muted-foreground">
                        {queryResult.duration}ms
                      </span>
                    )}
                  </div>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {/* Tabs — always rendered so Graph/Table toggle is visible */}
              <Tabs value={resultTab} onValueChange={(v) => setResultTab(v as "table" | "graph")}>
                <TabsList>
                  <TabsTrigger value="graph">Graph</TabsTrigger>
                  <TabsTrigger value="table">Table</TabsTrigger>
                </TabsList>

                <TabsContent value="graph">
                  {isExecuting ? (
                    <div className="flex items-center justify-center py-12">
                      <div className="animate-spin h-8 w-8 border-2 border-primary border-t-transparent rounded-full mr-3" />
                      <span className="text-sm text-muted-foreground">Executing query…</span>
                    </div>
                  ) : queryResult ? (
                    <GraphPanel queryResult={queryResult} isExecuting={false} />
                  ) : (
                    <GraphPanel queryResult={null} isExecuting={false} />
                  )}
                </TabsContent>

                <TabsContent value="table">
                  {isExecuting ? (
                    <div className="flex items-center justify-center py-12">
                      <div className="animate-spin h-8 w-8 border-2 border-primary border-t-transparent rounded-full mr-3" />
                      <span className="text-sm text-muted-foreground">Executing query…</span>
                    </div>
                  ) : queryResult ? (
                    <>
                      {queryResult.type === "scalar" && (
                        <ResultScalar data={queryResult.data as any} />
                      )}
                      {queryResult.type === "table" && (
                        <ResultTable data={queryResult.data as any} />
                      )}
                      {queryResult.type === "vector" && (
                        <ResultTable data={vectorToTable(queryResult.data as any)} />
                      )}
                      {queryResult.type === "matrix" && (
                        <ResultTable data={matrixToTable(queryResult.data as any)} />
                      )}
                    </>
                  ) : (
                    <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                      <p className="text-sm">No query results yet</p>
                      <p className="text-xs mt-1 opacity-60">
                        Enter a FinanceQL expression and click Execute
                      </p>
                    </div>
                  )}
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>
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

/**
 * Convert VectorResult to TableResult for Table tab display.
 */
function vectorToTable(data: import("@/lib/types").VectorResult): import("@/lib/types").TableResult {
  if (!data.items || data.items.length === 0) {
    return { columns: [], rows: [] };
  }
  const labelKeys = new Set<string>();
  data.items.forEach((item) => {
    Object.keys(item.labels || {}).forEach((k) => labelKeys.add(k));
  });
  const columns = ["ticker", ...Array.from(labelKeys), "value"];
  const rows = data.items.map((item) => {
    const row: Record<string, unknown> = { ticker: item.ticker, value: item.value };
    labelKeys.forEach((k) => {
      row[k] = item.labels?.[k] ?? "";
    });
    return row;
  });
  return { columns, rows };
}

/**
 * Convert MatrixResult to TableResult for Table tab display.
 */
function matrixToTable(data: import("@/lib/types").MatrixResult): import("@/lib/types").TableResult {
  if (!data.series || data.series.length === 0) {
    return { columns: [], rows: [] };
  }
  const columns = ["time", ...data.series.map((s) => s.label)];
  // Collect all unique timestamps
  const timeSet = new Set<number>();
  data.series.forEach((s) => s.data.forEach((d) => timeSet.add(d.time)));
  const times = Array.from(timeSet).sort((a, b) => a - b);

  const rows = times.map((t) => {
    const row: Record<string, unknown> = {
      time: new Date(t * 1000).toLocaleString("en-IN"),
    };
    data.series.forEach((s) => {
      const point = s.data.find((d) => d.time === t);
      row[s.label] = point?.value ?? null;
    });
    return row;
  });
  return { columns, rows };
}
