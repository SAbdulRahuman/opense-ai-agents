"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { getFIIDII } from "@/lib/api";
import { formatIndianNumber, cn } from "@/lib/utils";
import type { FIIDIIData } from "@/lib/types";

export function FIIDIIBar() {
  const [data, setData] = useState<FIIDIIData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getFIIDII()
      .then(setData)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">FII / DII Activity</CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-10 w-full mb-3" />
          <Skeleton className="h-10 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (!data) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">FII / DII Activity</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">Data unavailable</p>
        </CardContent>
      </Card>
    );
  }

  const maxVal = Math.max(
    Math.abs(data.fiiBuy),
    Math.abs(data.fiiSell),
    Math.abs(data.diiBuy),
    Math.abs(data.diiSell),
    1
  );

  function Bar({ label, buy, sell, net }: { label: string; buy: number; sell: number; net: number }) {
    const buyWidth = (buy / maxVal) * 100;
    const sellWidth = (sell / maxVal) * 100;
    const isPositive = net >= 0;

    return (
      <div className="space-y-1.5">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium">{label}</span>
          <span
            className={cn(
              "text-sm font-semibold tabular-nums",
              isPositive ? "text-green-600 dark:text-green-400" : "text-red-600 dark:text-red-400"
            )}
          >
            Net: {isPositive ? "+" : ""}₹{formatIndianNumber(Math.abs(net))} Cr
          </span>
        </div>
        <div className="flex gap-1 h-6">
          <div className="flex-1 bg-muted rounded overflow-hidden relative">
            <div
              className="h-full bg-green-500/80 rounded-l"
              style={{ width: `${buyWidth}%` }}
            />
            <span className="absolute inset-0 flex items-center justify-center text-[10px] font-medium">
              Buy: ₹{formatIndianNumber(buy)} Cr
            </span>
          </div>
          <div className="flex-1 bg-muted rounded overflow-hidden relative">
            <div
              className="h-full bg-red-500/80 rounded-r ml-auto"
              style={{ width: `${sellWidth}%` }}
            />
            <span className="absolute inset-0 flex items-center justify-center text-[10px] font-medium">
              Sell: ₹{formatIndianNumber(sell)} Cr
            </span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base">FII / DII Activity</CardTitle>
          <span className="text-xs text-muted-foreground">{data.date}</span>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <Bar
          label="FII (Foreign)"
          buy={data.fiiBuy}
          sell={data.fiiSell}
          net={data.fiiNet}
        />
        <Bar
          label="DII (Domestic)"
          buy={data.diiBuy}
          sell={data.diiSell}
          net={data.diiNet}
        />
      </CardContent>
    </Card>
  );
}
