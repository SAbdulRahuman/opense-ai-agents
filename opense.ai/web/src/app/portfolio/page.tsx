"use client";

import { useEffect, useState } from "react";
import {
  Wallet,
  TrendingUp,
  TrendingDown,
  PieChart,
  IndianRupee,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getPortfolio } from "@/lib/api";
import { formatPrice, formatPercent, formatIndianNumber, cn } from "@/lib/utils";
import type { Holding, PortfolioSummary } from "@/lib/types";

export default function PortfolioPage() {
  const [holdings, setHoldings] = useState<Holding[]>([]);
  const [summary, setSummary] = useState<PortfolioSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getPortfolio()
      .then((data) => {
        setHoldings(data.holdings);
        setSummary(data);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Portfolio</h1>
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
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Portfolio</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Track your holdings and portfolio performance
        </p>
      </div>

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-5">
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <Wallet size={14} />
                Invested Value
              </div>
              <div className="text-2xl font-bold tabular-nums">
                ₹{formatIndianNumber(summary.totalInvested)}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-5">
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <IndianRupee size={14} />
                Current Value
              </div>
              <div className="text-2xl font-bold tabular-nums">
                ₹{formatIndianNumber(summary.totalValue)}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-5">
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                {summary.totalPnl >= 0 ? <TrendingUp size={14} /> : <TrendingDown size={14} />}
                Total P&L
              </div>
              <div
                className={cn(
                  "text-2xl font-bold tabular-nums",
                  summary.totalPnl >= 0
                    ? "text-green-600 dark:text-green-400"
                    : "text-red-600 dark:text-red-400"
                )}
              >
                {summary.totalPnl >= 0 ? "+" : ""}₹{formatIndianNumber(Math.abs(summary.totalPnl))}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-5">
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <PieChart size={14} />
                Day P&L
              </div>
              <div
                className={cn(
                  "text-2xl font-bold tabular-nums",
                  summary.dayPnl >= 0
                    ? "text-green-600 dark:text-green-400"
                    : "text-red-600 dark:text-red-400"
                )}
              >
                {summary.dayPnl >= 0 ? "+" : ""}₹{formatIndianNumber(Math.abs(summary.dayPnl))}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Holdings Table */}
      <Card>
        <CardHeader>
          <CardTitle>Holdings</CardTitle>
        </CardHeader>
        <CardContent>
          {holdings.length === 0 ? (
            <p className="text-center py-8 text-muted-foreground">
              No holdings found. Connect your broker account to import holdings.
            </p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left">
                    <th className="pb-3 font-medium">Stock</th>
                    <th className="pb-3 font-medium text-right">Qty</th>
                    <th className="pb-3 font-medium text-right">Avg Cost</th>
                    <th className="pb-3 font-medium text-right">LTP</th>
                    <th className="pb-3 font-medium text-right">Current Value</th>
                    <th className="pb-3 font-medium text-right">P&L</th>
                    <th className="pb-3 font-medium text-right">P&L %</th>
                  </tr>
                </thead>
                <tbody>
                  {holdings.map((h) => {
                    const isPositive = h.pnl >= 0;
                    return (
                      <tr key={h.ticker} className="border-b last:border-0">
                        <td className="py-3">
                          <div className="font-mono font-medium">{h.ticker}</div>
                          {h.name && (
                            <div className="text-xs text-muted-foreground">{h.name}</div>
                          )}
                        </td>
                        <td className="py-3 text-right tabular-nums">{h.quantity}</td>
                        <td className="py-3 text-right tabular-nums">
                          {formatPrice(h.avgPrice)}
                        </td>
                        <td className="py-3 text-right tabular-nums">
                          {formatPrice(h.currentPrice)}
                        </td>
                        <td className="py-3 text-right tabular-nums">
                          ₹{formatIndianNumber(h.value)}
                        </td>
                        <td
                          className={cn(
                            "py-3 text-right tabular-nums font-medium",
                            isPositive
                              ? "text-green-600 dark:text-green-400"
                              : "text-red-600 dark:text-red-400"
                          )}
                        >
                          {isPositive ? "+" : ""}₹{formatIndianNumber(Math.abs(h.pnl))}
                        </td>
                        <td className="py-3 text-right">
                          <Badge
                            variant={isPositive ? "default" : "destructive"}
                            className="text-xs"
                          >
                            {isPositive ? "+" : ""}
                            {formatPercent(Math.abs(h.pnlPercent))}
                          </Badge>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
