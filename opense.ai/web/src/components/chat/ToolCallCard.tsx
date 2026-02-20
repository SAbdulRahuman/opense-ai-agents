"use client";

import { useState } from "react";
import { ChevronDown, ChevronRight, Wrench, CheckCircle2, XCircle, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ToolCall } from "@/lib/types";

interface ToolCallCardProps {
  toolCall: ToolCall;
}

const statusConfig = {
  pending: { icon: Loader2, color: "text-yellow-500", label: "Pending" },
  running: { icon: Loader2, color: "text-blue-500 animate-spin", label: "Running" },
  completed: { icon: CheckCircle2, color: "text-green-500", label: "Completed" },
  failed: { icon: XCircle, color: "text-red-500", label: "Failed" },
};

export function ToolCallCard({ toolCall }: ToolCallCardProps) {
  const [expanded, setExpanded] = useState(false);
  const status = statusConfig[toolCall.status];
  const StatusIcon = status.icon;

  // Format result summary
  const resultSummary = toolCall.result
    ? toolCall.result.length > 80
      ? toolCall.result.slice(0, 80) + "..."
      : toolCall.result
    : "—";

  return (
    <div className="rounded-md border bg-muted/30 text-sm">
      <button
        className="flex w-full items-center gap-2 p-2 hover:bg-muted/50 transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        {expanded ? (
          <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
        ) : (
          <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
        )}
        <Wrench className="h-3.5 w-3.5 text-muted-foreground" />
        <span className="font-mono text-xs font-medium">{toolCall.name}</span>
        <StatusIcon className={cn("h-3.5 w-3.5 ml-auto", status.color)} />
        {!expanded && (
          <span className="text-xs text-muted-foreground truncate max-w-[200px]">
            {resultSummary}
          </span>
        )}
      </button>

      {expanded && (
        <div className="border-t p-2 space-y-2">
          <div>
            <span className="text-xs font-medium text-muted-foreground">Arguments:</span>
            <pre className="mt-1 rounded bg-muted p-2 text-xs overflow-auto max-h-40">
              {JSON.stringify(toolCall.arguments, null, 2)}
            </pre>
          </div>
          {toolCall.result && (
            <div>
              <span className="text-xs font-medium text-muted-foreground">Result:</span>
              <pre className="mt-1 rounded bg-muted p-2 text-xs overflow-auto max-h-40 whitespace-pre-wrap">
                {toolCall.result}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
