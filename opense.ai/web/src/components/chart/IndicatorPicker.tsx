// ============================================================================
// OpeNSE.ai — IndicatorPicker (dialog for adding indicator panes)
// ============================================================================
"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Activity, X } from "lucide-react";

type IndicatorType = "RSI" | "MACD" | "Stochastic" | "ATR" | "OBV";

interface IndicatorDef {
  id: IndicatorType;
  label: string;
  description: string;
  category: string;
}

const INDICATORS: IndicatorDef[] = [
  { id: "RSI", label: "RSI", description: "Relative Strength Index (14)", category: "Momentum" },
  { id: "MACD", label: "MACD", description: "Moving Average Convergence Divergence (12, 26, 9)", category: "Momentum" },
  { id: "Stochastic", label: "Stochastic", description: "Stochastic Oscillator (%K 14, %D 3)", category: "Momentum" },
  { id: "ATR", label: "ATR", description: "Average True Range (14)", category: "Volatility" },
  { id: "OBV", label: "OBV", description: "On-Balance Volume", category: "Volume" },
];

interface IndicatorPickerProps {
  activeIndicators: IndicatorType[];
  onToggle: (id: IndicatorType) => void;
}

export function IndicatorPicker({ activeIndicators, onToggle }: IndicatorPickerProps) {
  const [isOpen, setIsOpen] = useState(false);

  if (!isOpen) {
    return (
      <Button
        variant="ghost"
        size="sm"
        className="h-7 gap-1 px-2 text-xs"
        onClick={() => setIsOpen(true)}
      >
        <Activity className="h-3.5 w-3.5" />
        Indicators
        {activeIndicators.length > 0 && (
          <span className="ml-1 rounded-full bg-primary/20 px-1.5 text-[10px] font-bold text-primary">
            {activeIndicators.length}
          </span>
        )}
      </Button>
    );
  }

  const categories = [...new Set(INDICATORS.map((i) => i.category))];

  return (
    <div className="absolute top-8 left-0 z-50 w-72 rounded-lg border bg-popover p-3 shadow-xl">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-semibold">Indicator Panes</span>
        <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={() => setIsOpen(false)}>
          <X className="h-3.5 w-3.5" />
        </Button>
      </div>

      {categories.map((cat) => (
        <div key={cat} className="mb-2">
          <div className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">{cat}</div>
          <div className="space-y-1">
            {INDICATORS.filter((i) => i.category === cat).map((ind) => {
              const isActive = activeIndicators.includes(ind.id);
              return (
                <button
                  key={ind.id}
                  className={cn(
                    "flex w-full items-center justify-between rounded-md px-2 py-1.5 text-left text-xs hover:bg-muted/50",
                    isActive && "bg-primary/10",
                  )}
                  onClick={() => onToggle(ind.id)}
                >
                  <div>
                    <div className="font-medium">{ind.label}</div>
                    <div className="text-[10px] text-muted-foreground">{ind.description}</div>
                  </div>
                  {isActive && (
                    <span className="text-[10px] font-bold text-primary">ON</span>
                  )}
                </button>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}

export type { IndicatorType };
