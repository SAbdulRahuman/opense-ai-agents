"use client";

import { useEffect } from "react";
import {
  ArrowUpDown,
  TrendingUp,
  TrendingDown,
  RefreshCw,
  ShoppingCart,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useStore } from "@/store";
import { formatPrice, formatPercent, formatIndianNumber, cn } from "@/lib/utils";

export default function PositionsPage() {
  const {
    positions,
    positionsLoading,
    fetchPositions,
    openOrderWindow,
  } = useStore();

  useEffect(() => {
    fetchPositions();
    const interval = setInterval(fetchPositions, 5000);
    return () => clearInterval(interval);
  }, [fetchPositions]);

  const totalPnl = positions.reduce((sum, p) => sum + p.pnl, 0);
  const totalDayPnl = positions.reduce((sum, p) => sum + p.day_pnl, 0);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Positions</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Open intraday and delivery positions
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={fetchPositions}
          disabled={positionsLoading}
        >
          <RefreshCw size={14} className={cn(positionsLoading && "animate-spin")} />
          <span className="ml-1.5">Refresh</span>
        </Button>
      </div>

      {/* P&L Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="p-5">
            <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
              <ArrowUpDown size={14} />
              Open Positions
            </div>
            <div className="text-2xl font-bold tabular-nums">{positions.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-5">
            <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
              {totalPnl >= 0 ? <TrendingUp size={14} /> : <TrendingDown size={14} />}
              Total P&L
            </div>
            <div
              className={cn(
                "text-2xl font-bold tabular-nums",
                totalPnl >= 0
                  ? "text-green-600 dark:text-green-400"
                  : "text-red-600 dark:text-red-400",
              )}
            >
              {totalPnl >= 0 ? "+" : ""}₹{formatIndianNumber(Math.abs(totalPnl))}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-5">
            <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
              {totalDayPnl >= 0 ? <TrendingUp size={14} /> : <TrendingDown size={14} />}
              Day P&L
            </div>
            <div
              className={cn(
                "text-2xl font-bold tabular-nums",
                totalDayPnl >= 0
                  ? "text-green-600 dark:text-green-400"
                  : "text-red-600 dark:text-red-400",
              )}
            >
              {totalDayPnl >= 0 ? "+" : ""}₹{formatIndianNumber(Math.abs(totalDayPnl))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Positions Table */}
      <Card>
        <CardHeader>
          <CardTitle>Open Positions</CardTitle>
        </CardHeader>
        <CardContent>
          {positionsLoading && positions.length === 0 ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : positions.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
              <ArrowUpDown size={48} className="mb-4 opacity-50" />
              <p className="text-lg font-medium">No open positions</p>
              <p className="text-sm mt-1">
                Your positions will appear here once you place orders
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left">
                    <th className="pb-3 font-medium">Product</th>
                    <th className="pb-3 font-medium">Instrument</th>
                    <th className="pb-3 font-medium text-right">Qty</th>
                    <th className="pb-3 font-medium text-right">Avg Price</th>
                    <th className="pb-3 font-medium text-right">LTP</th>
                    <th className="pb-3 font-medium text-right">P&L</th>
                    <th className="pb-3 font-medium text-right">Change</th>
                    <th className="pb-3 font-medium text-right">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {positions.map((p) => {
                    const isPositive = p.pnl >= 0;
                    return (
                      <tr key={`${p.ticker}-${p.product}`} className="border-b last:border-0">
                        <td className="py-3">
                          <Badge variant="outline" className="text-xs">
                            {p.product}
                          </Badge>
                        </td>
                        <td className="py-3">
                          <div className="font-mono font-medium">{p.ticker}</div>
                          <div className="text-xs text-muted-foreground">{p.exchange}</div>
                        </td>
                        <td
                          className={cn(
                            "py-3 text-right tabular-nums font-medium",
                            p.quantity > 0
                              ? "text-green-600 dark:text-green-400"
                              : "text-red-600 dark:text-red-400",
                          )}
                        >
                          {p.quantity}
                        </td>
                        <td className="py-3 text-right tabular-nums">
                          {formatPrice(p.avg_price)}
                        </td>
                        <td className="py-3 text-right tabular-nums">{formatPrice(p.ltp)}</td>
                        <td
                          className={cn(
                            "py-3 text-right tabular-nums font-medium",
                            isPositive
                              ? "text-green-600 dark:text-green-400"
                              : "text-red-600 dark:text-red-400",
                          )}
                        >
                          {isPositive ? "+" : ""}₹{formatIndianNumber(Math.abs(p.pnl))}
                        </td>
                        <td className="py-3 text-right">
                          <Badge
                            variant={isPositive ? "default" : "destructive"}
                            className="text-xs"
                          >
                            {isPositive ? "+" : ""}
                            {formatPercent(Math.abs(p.pnl_pct))}
                          </Badge>
                        </td>
                        <td className="py-3 text-right">
                          <div className="flex gap-1 justify-end">
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-7 text-xs text-blue-500 hover:text-blue-600"
                              onClick={() => openOrderWindow(p.ticker, "BUY")}
                            >
                              Buy
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-7 text-xs text-red-500 hover:text-red-600"
                              onClick={() => openOrderWindow(p.ticker, "SELL")}
                            >
                              Sell
                            </Button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
                <tfoot>
                  <tr className="border-t font-medium">
                    <td colSpan={5} className="py-3 text-right text-xs text-muted-foreground">
                      Total
                    </td>
                    <td
                      className={cn(
                        "py-3 text-right tabular-nums",
                        totalPnl >= 0
                          ? "text-green-600 dark:text-green-400"
                          : "text-red-600 dark:text-red-400",
                      )}
                    >
                      {totalPnl >= 0 ? "+" : ""}₹{formatIndianNumber(Math.abs(totalPnl))}
                    </td>
                    <td colSpan={2} />
                  </tr>
                </tfoot>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
