"use client";

import { Badge } from "@/components/ui/badge";
import { formatPrice, formatIndianNumber } from "@/lib/utils";
import type { ScalarResult } from "@/lib/types";

interface ResultScalarProps {
  data: ScalarResult;
}

export function ResultScalar({ data }: ResultScalarProps) {
  const formattedValue = data.label.toLowerCase().includes("price") || data.label.includes("₹")
    ? formatPrice(data.value)
    : formatIndianNumber(data.value);

  return (
    <div className="flex flex-col items-center justify-center py-8">
      <div className="text-4xl font-bold tracking-tight">{formattedValue}</div>
      <div className="mt-2 flex items-center gap-2">
        <span className="text-sm text-muted-foreground">{data.label}</span>
        {data.ticker && (
          <Badge variant="outline" className="text-xs">
            {data.ticker}
          </Badge>
        )}
        {data.metric && (
          <Badge variant="secondary" className="text-xs">
            {data.metric}
          </Badge>
        )}
      </div>
    </div>
  );
}
