"use client";

import { useState, useEffect } from "react";
import { AlertTriangle, Check, X, Edit3 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { formatPrice, formatPercent } from "@/lib/utils";
import type { TradeProposal } from "@/lib/types";

interface TradeConfirmProps {
  proposal: TradeProposal;
  onConfirm: (action: "approve" | "reject") => void;
}

export function TradeConfirm({ proposal, onConfirm }: TradeConfirmProps) {
  const [timeLeft, setTimeLeft] = useState(proposal.timeout);

  useEffect(() => {
    if (timeLeft <= 0) {
      onConfirm("reject");
      return;
    }
    const timer = setInterval(() => setTimeLeft((t) => t - 1), 1000);
    return () => clearInterval(timer);
  }, [timeLeft, onConfirm]);

  const changeText = formatPercent(proposal.positionPct);
  const changeClass = proposal.positionPct > 10
    ? "text-red-600 dark:text-red-400"
    : proposal.positionPct > 5
      ? "text-yellow-600 dark:text-yellow-400"
      : "text-green-600 dark:text-green-400";
  const isBuy = proposal.action === "BUY";

  return (
    <Card className="border-yellow-500/30 bg-yellow-500/5">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <AlertTriangle className="h-4 w-4 text-yellow-500" />
            Trade Confirmation Required
          </CardTitle>
          <Badge variant="outline" className="text-xs">
            {timeLeft}s remaining
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-muted-foreground">Ticker:</span>
            <span className="ml-2 font-mono font-bold">{proposal.ticker}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Action:</span>
            <Badge
              variant={isBuy ? "default" : "destructive"}
              className="ml-2"
            >
              {proposal.action}
            </Badge>
          </div>
          <div>
            <span className="text-muted-foreground">Quantity:</span>
            <span className="ml-2 font-semibold">{proposal.quantity}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Price:</span>
            <span className="ml-2 font-semibold">{formatPrice(proposal.price)}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Order Type:</span>
            <span className="ml-2">{proposal.orderType}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Est. Cost:</span>
            <span className="ml-2 font-semibold">{formatPrice(proposal.estimatedCost)}</span>
          </div>
        </div>

        {/* Risk summary */}
        <div className="rounded-md bg-muted/50 p-2 text-xs">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Position Size:</span>
            <span className={changeClass}>{changeText} of capital</span>
          </div>
          <div className="flex justify-between mt-1">
            <span className="text-muted-foreground">Current Exposure:</span>
            <span>{proposal.currentExposure.toFixed(1)}%</span>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-2">
          <Button
            size="sm"
            className="flex-1"
            onClick={() => onConfirm("approve")}
          >
            <Check className="h-4 w-4 mr-1" />
            Approve
          </Button>
          <Button
            size="sm"
            variant="destructive"
            className="flex-1"
            onClick={() => onConfirm("reject")}
          >
            <X className="h-4 w-4 mr-1" />
            Reject
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
