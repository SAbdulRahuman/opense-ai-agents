"use client";

import { useEffect, useState } from "react";
import { TrendingUp, TrendingDown } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { getTopMovers } from "@/lib/api";
import { useStore } from "@/store";
import { formatPrice, formatPercent, cn } from "@/lib/utils";
import type { TopMover } from "@/lib/types";

export function TopMovers() {
  const [gainers, setGainers] = useState<TopMover[]>([]);
  const [losers, setLosers] = useState<TopMover[]>([]);
  const [loading, setLoading] = useState(true);
  const { setSelectedTicker } = useStore();

  useEffect(() => {
    Promise.all([getTopMovers("gainers"), getTopMovers("losers")])
      .then(([g, l]) => {
        setGainers(g);
        setLosers(l);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  function MoverList({ movers, type }: { movers: TopMover[]; type: "gainer" | "loser" }) {
    if (loading) {
      return (
        <div className="space-y-2">
          {[1, 2, 3, 4, 5].map((i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </div>
      );
    }

    return (
      <div className="divide-y">
        {movers.slice(0, 10).map((m, i) => {
          const isGainer = type === "gainer";
          return (
            <button
              key={m.ticker}
              className="flex items-center justify-between w-full py-2 hover:bg-muted/50 transition-colors px-1 rounded"
              onClick={() => setSelectedTicker(m.ticker)}
            >
              <div className="flex items-center gap-3">
                <span className="text-xs text-muted-foreground w-5 text-right">
                  {i + 1}
                </span>
                <div className="text-left">
                  <div className="text-sm font-mono font-medium">{m.ticker}</div>
                  {m.name && (
                    <div className="text-xs text-muted-foreground truncate max-w-[120px]">
                      {m.name}
                    </div>
                  )}
                </div>
              </div>
              <div className="text-right">
                <div className="text-sm tabular-nums">{formatPrice(m.price)}</div>
                <div
                  className={cn(
                    "text-xs tabular-nums flex items-center justify-end gap-0.5",
                    isGainer
                      ? "text-green-600 dark:text-green-400"
                      : "text-red-600 dark:text-red-400"
                  )}
                >
                  {isGainer ? <TrendingUp size={10} /> : <TrendingDown size={10} />}
                  {formatPercent(Math.abs(m.changePercent))}
                </div>
              </div>
            </button>
          );
        })}
      </div>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">Top Movers</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="gainers">
          <TabsList className="w-full">
            <TabsTrigger value="gainers" className="flex-1">
              <TrendingUp size={14} className="mr-1.5 text-green-500" />
              Gainers
            </TabsTrigger>
            <TabsTrigger value="losers" className="flex-1">
              <TrendingDown size={14} className="mr-1.5 text-red-500" />
              Losers
            </TabsTrigger>
          </TabsList>
          <TabsContent value="gainers">
            <MoverList movers={gainers} type="gainer" />
          </TabsContent>
          <TabsContent value="losers">
            <MoverList movers={losers} type="loser" />
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}
