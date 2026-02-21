"use client";

import { useEffect } from "react";
import {
  Wallet,
  Banknote,
  ShieldCheck,
  TrendingUp,
  RefreshCw,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { useStore } from "@/store";
import { formatIndianNumber, cn } from "@/lib/utils";

export default function FundsPage() {
  const { margins, fundsLoading, fetchFunds } = useStore();

  useEffect(() => {
    fetchFunds();
  }, [fetchFunds]);

  if (fundsLoading && !margins) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Funds</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[1, 2, 3, 4, 5].map((i) => (
            <Card key={i}>
              <CardContent className="p-5">
                <Skeleton className="h-4 w-24 mb-3" />
                <Skeleton className="h-8 w-32" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  const m = margins ?? {
    available_cash: 0,
    used_margin: 0,
    available_margin: 0,
    collateral: 0,
    opening_balance: 0,
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Funds</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Account margins and fund details
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={fetchFunds} disabled={fundsLoading}>
          <RefreshCw size={14} className={cn(fundsLoading && "animate-spin")} />
          <span className="ml-1.5">Refresh</span>
        </Button>
      </div>

      {/* Equity Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Banknote size={18} />
            Equity
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div>
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <Wallet size={14} />
                Available Margin
              </div>
              <div className="text-2xl font-bold tabular-nums text-green-600 dark:text-green-400">
                ₹{formatIndianNumber(m.available_margin)}
              </div>
            </div>

            <div>
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <TrendingUp size={14} />
                Used Margin
              </div>
              <div className="text-2xl font-bold tabular-nums">
                ₹{formatIndianNumber(m.used_margin)}
              </div>
            </div>

            <div>
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <Banknote size={14} />
                Available Cash
              </div>
              <div className="text-2xl font-bold tabular-nums">
                ₹{formatIndianNumber(m.available_cash)}
              </div>
            </div>

            <div>
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <ShieldCheck size={14} />
                Collateral
              </div>
              <div className="text-2xl font-bold tabular-nums">
                ₹{formatIndianNumber(m.collateral)}
              </div>
            </div>

            <div>
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <Wallet size={14} />
                Opening Balance
              </div>
              <div className="text-2xl font-bold tabular-nums">
                ₹{formatIndianNumber(m.opening_balance)}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Margin Details Table */}
      <Card>
        <CardHeader>
          <CardTitle>Margin Details</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left">
                  <th className="pb-3 font-medium">Particular</th>
                  <th className="pb-3 font-medium text-right">Equity</th>
                </tr>
              </thead>
              <tbody>
                <tr className="border-b">
                  <td className="py-3">Opening Balance</td>
                  <td className="py-3 text-right tabular-nums font-medium">
                    ₹{formatIndianNumber(m.opening_balance)}
                  </td>
                </tr>
                <tr className="border-b">
                  <td className="py-3">Available Cash</td>
                  <td className="py-3 text-right tabular-nums font-medium">
                    ₹{formatIndianNumber(m.available_cash)}
                  </td>
                </tr>
                <tr className="border-b">
                  <td className="py-3">Collateral (Liquid funds, etc.)</td>
                  <td className="py-3 text-right tabular-nums font-medium">
                    ₹{formatIndianNumber(m.collateral)}
                  </td>
                </tr>
                <tr className="border-b">
                  <td className="py-3 font-medium">Total Available Margin</td>
                  <td className="py-3 text-right tabular-nums font-bold text-green-600 dark:text-green-400">
                    ₹{formatIndianNumber(m.available_margin)}
                  </td>
                </tr>
                <tr className="border-b">
                  <td className="py-3">Used Margin</td>
                  <td className="py-3 text-right tabular-nums font-medium text-orange-500">
                    ₹{formatIndianNumber(m.used_margin)}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Commodity Placeholder */}
      <Card>
        <CardHeader>
          <CardTitle className="text-muted-foreground">Commodity</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground text-center py-6">
            You don&apos;t have a commodity account yet.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
