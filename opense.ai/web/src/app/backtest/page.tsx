"use client";

import { useState } from "react";
import { Play, TrendingUp, TrendingDown, BarChart3, Clock } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Select, SelectTrigger, SelectContent, SelectItem, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { runBacktest } from "@/lib/api";
import { formatPrice, formatPercent, formatIndianNumber, cn } from "@/lib/utils";
import type { BacktestResult, BacktestTrade } from "@/lib/types";

export default function BacktestPage() {
  const [params, setParams] = useState({
    ticker: "RELIANCE",
    strategy: "sma_crossover",
    startDate: "2023-01-01",
    endDate: new Date().toISOString().split("T")[0],
    initialCapital: "1000000",
    shortPeriod: "20",
    longPeriod: "50",
  });
  const [result, setResult] = useState<BacktestResult | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleBacktest() {
    setLoading(true);
    try {
      const data = await runBacktest({
        tickers: [params.ticker],
        strategy: params.strategy,
        startDate: params.startDate,
        endDate: params.endDate,
        capital: parseFloat(params.initialCapital),
        params: {
          shortPeriod: parseInt(params.shortPeriod),
          longPeriod: parseInt(params.longPeriod),
        },
      });
      setResult(data);
    } catch {
      // TODO: toast
    } finally {
      setLoading(false);
    }
  }

  function updateParam(key: string, value: string) {
    setParams((prev) => ({ ...prev, [key]: value }));
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">Backtesting</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Test trading strategies against historical NSE data
        </p>
      </div>

      {/* Configuration */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Strategy Configuration</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">
                Ticker
              </label>
              <Input
                value={params.ticker}
                onChange={(e) => updateParam("ticker", e.target.value.toUpperCase())}
                className="h-8 text-sm font-mono"
              />
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">
                Strategy
              </label>
              <Select
                value={params.strategy}
                onValueChange={(v) => updateParam("strategy", v)}
              >
                <SelectTrigger className="h-8 text-sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="sma_crossover">SMA Crossover</SelectItem>
                  <SelectItem value="ema_crossover">EMA Crossover</SelectItem>
                  <SelectItem value="rsi_oversold">RSI Oversold</SelectItem>
                  <SelectItem value="macd_signal">MACD Signal</SelectItem>
                  <SelectItem value="bollinger_bounce">Bollinger Bounce</SelectItem>
                  <SelectItem value="mean_reversion">Mean Reversion</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">
                Start Date
              </label>
              <Input
                type="date"
                value={params.startDate}
                onChange={(e) => updateParam("startDate", e.target.value)}
                className="h-8 text-sm"
              />
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">
                End Date
              </label>
              <Input
                type="date"
                value={params.endDate}
                onChange={(e) => updateParam("endDate", e.target.value)}
                className="h-8 text-sm"
              />
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">
                Initial Capital (₹)
              </label>
              <Input
                type="number"
                value={params.initialCapital}
                onChange={(e) => updateParam("initialCapital", e.target.value)}
                className="h-8 text-sm"
              />
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">
                Short Period
              </label>
              <Input
                type="number"
                value={params.shortPeriod}
                onChange={(e) => updateParam("shortPeriod", e.target.value)}
                className="h-8 text-sm"
              />
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">
                Long Period
              </label>
              <Input
                type="number"
                value={params.longPeriod}
                onChange={(e) => updateParam("longPeriod", e.target.value)}
                className="h-8 text-sm"
              />
            </div>
          </div>
          <Button onClick={handleBacktest} disabled={loading} className="mt-4">
            <Play size={14} className="mr-1.5" />
            {loading ? "Running Backtest…" : "Run Backtest"}
          </Button>
        </CardContent>
      </Card>

      {loading && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {[1, 2, 3, 4].map((i) => (
              <Card key={i}>
                <CardContent className="p-5">
                  <Skeleton className="h-4 w-20 mb-3" />
                  <Skeleton className="h-8 w-28" />
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {result && !loading && (
        <>
          {/* Metrics Summary */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Card>
              <CardContent className="p-4">
                <div className="text-xs text-muted-foreground mb-1">Total Return</div>
                <div
                  className={cn(
                    "text-xl font-bold tabular-nums",
                    result.metrics.totalReturn >= 0
                      ? "text-green-600 dark:text-green-400"
                      : "text-red-600 dark:text-red-400"
                  )}
                >
                  {result.metrics.totalReturn >= 0 ? "+" : ""}
                  {formatPercent(result.metrics.totalReturn)}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="text-xs text-muted-foreground mb-1">CAGR</div>
                <div className="text-xl font-bold tabular-nums">
                  {formatPercent(result.metrics.cagr)}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="text-xs text-muted-foreground mb-1">Sharpe Ratio</div>
                <div className="text-xl font-bold tabular-nums">
                  {result.metrics.sharpeRatio.toFixed(2)}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="text-xs text-muted-foreground mb-1">Max Drawdown</div>
                <div className="text-xl font-bold tabular-nums text-red-600 dark:text-red-400">
                  {formatPercent(result.metrics.maxDrawdown)}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="text-xs text-muted-foreground mb-1">Win Rate</div>
                <div className="text-xl font-bold tabular-nums">
                  {formatPercent(result.metrics.winRate)}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="text-xs text-muted-foreground mb-1">Total Trades</div>
                <div className="text-xl font-bold tabular-nums">
                  {result.metrics.totalTrades}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="text-xs text-muted-foreground mb-1">Profit Factor</div>
                <div className="text-xl font-bold tabular-nums">
                  {result.metrics.profitFactor.toFixed(2)}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="text-xs text-muted-foreground mb-1">Avg Win / Loss</div>
                <div className="text-xl font-bold tabular-nums flex items-center gap-1">
                  <Clock size={16} className="text-muted-foreground" />
                  {formatPercent(result.metrics.avgWin)} / {formatPercent(result.metrics.avgLoss)}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Trades */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-base">Trade Log</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b text-left bg-muted/30">
                      <th className="p-3 font-medium">#</th>
                      <th className="p-3 font-medium">Type</th>
                      <th className="p-3 font-medium">Entry Date</th>
                      <th className="p-3 font-medium text-right">Entry Price</th>
                      <th className="p-3 font-medium">Exit Date</th>
                      <th className="p-3 font-medium text-right">Exit Price</th>
                      <th className="p-3 font-medium text-right">Quantity</th>
                      <th className="p-3 font-medium text-right">P&L</th>
                      <th className="p-3 font-medium text-right">Return</th>
                    </tr>
                  </thead>
                  <tbody>
                    {result.trades.map((t, i) => {
                      const pnl = t.pnl;
                      const ret = t.pnlPercent;
                      const isPositive = pnl >= 0;
                      return (
                        <tr key={i} className="border-b last:border-0 hover:bg-muted/30">
                          <td className="p-3 text-muted-foreground">{i + 1}</td>
                          <td className="p-3">
                            <Badge
                              variant={t.side === "LONG" ? "default" : "destructive"}
                              className="text-xs"
                            >
                              {t.side === "LONG" ? (
                                <TrendingUp size={10} className="mr-1" />
                              ) : (
                                <TrendingDown size={10} className="mr-1" />
                              )}
                              {t.side}
                            </Badge>
                          </td>
                          <td className="p-3 text-xs">{t.entryDate}</td>
                          <td className="p-3 text-right tabular-nums">{formatPrice(t.entryPrice)}</td>
                          <td className="p-3 text-xs">{t.exitDate}</td>
                          <td className="p-3 text-right tabular-nums">{formatPrice(t.exitPrice)}</td>
                          <td className="p-3 text-right tabular-nums">{t.quantity}</td>
                          <td
                            className={cn(
                              "p-3 text-right tabular-nums font-medium",
                              isPositive
                                ? "text-green-600 dark:text-green-400"
                                : "text-red-600 dark:text-red-400"
                            )}
                          >
                            {isPositive ? "+" : ""}₹{formatIndianNumber(Math.abs(pnl))}
                          </td>
                          <td className="p-3 text-right">
                            <Badge
                              variant={isPositive ? "default" : "destructive"}
                              className="text-xs"
                            >
                              {isPositive ? "+" : ""}
                              {ret.toFixed(1)}%
                            </Badge>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
