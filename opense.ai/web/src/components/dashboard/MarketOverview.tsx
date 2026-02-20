"use client";

import { useEffect } from "react";
import { TrendingUp, TrendingDown, Activity } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useStore } from "@/store";
import { getMarketIndices } from "@/lib/api";
import { formatIndianNumber, formatPercent, cn } from "@/lib/utils";

export function MarketOverview() {
  const { indices, setIndices } = useStore();

  useEffect(() => {
    getMarketIndices()
      .then(setIndices)
      .catch(() => {});
    const interval = setInterval(() => {
      getMarketIndices()
        .then(setIndices)
        .catch(() => {});
    }, 10000);
    return () => clearInterval(interval);
  }, [setIndices]);

  if (indices.length === 0) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <Card key={i}>
            <CardContent className="p-5">
              <Skeleton className="h-4 w-24 mb-3" />
              <Skeleton className="h-8 w-32 mb-2" />
              <Skeleton className="h-4 w-20" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {indices.map((idx) => {
        const isPositive = idx.change >= 0;
        return (
          <Card key={idx.name} className="relative overflow-hidden">
            <CardContent className="p-5">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-muted-foreground">
                  {idx.name}
                </span>
                <Badge variant={isPositive ? "default" : "destructive"} className="text-xs">
                  {isPositive ? <TrendingUp size={12} className="mr-1" /> : <TrendingDown size={12} className="mr-1" />}
                  {formatPercent(idx.changePercent)}
                </Badge>
              </div>
              <div className="text-2xl font-bold tabular-nums">
                {formatIndianNumber(idx.value)}
              </div>
              <div
                className={cn(
                  "text-sm mt-1 tabular-nums",
                  isPositive ? "text-green-600 dark:text-green-400" : "text-red-600 dark:text-red-400"
                )}
              >
                {isPositive ? "+" : ""}
                {formatIndianNumber(idx.change)} ({formatPercent(idx.changePercent)})
              </div>
              {/* Subtle background icon */}
              <Activity
                size={80}
                className="absolute -bottom-2 -right-2 text-muted-foreground/5"
              />
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
