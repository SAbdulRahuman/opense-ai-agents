"use client";

import { Star, Clock, Play } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import type { QueryHistoryEntry } from "@/lib/types";

interface QueryHistoryProps {
  entries: QueryHistoryEntry[];
  onSelect: (query: string) => void;
  onToggleStar: (id: string) => void;
}

export function QueryHistory({ entries, onSelect, onToggleStar }: QueryHistoryProps) {
  if (entries.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-8 text-center text-muted-foreground">
        <Clock className="h-8 w-8 mb-2 opacity-20" />
        <p className="text-xs">No query history yet</p>
      </div>
    );
  }

  return (
    <div className="space-y-1">
      {entries.map((entry) => (
        <div
          key={entry.id}
          className="group flex items-start gap-2 rounded-md px-2 py-1.5 hover:bg-muted/50 cursor-pointer"
          onClick={() => onSelect(entry.query)}
        >
          <button
            onClick={(e) => {
              e.stopPropagation();
              onToggleStar(entry.id);
            }}
            className="mt-0.5 shrink-0"
          >
            <Star
              className={cn(
                "h-3.5 w-3.5",
                entry.starred
                  ? "fill-yellow-500 text-yellow-500"
                  : "text-muted-foreground opacity-0 group-hover:opacity-100",
              )}
            />
          </button>

          <div className="flex-1 min-w-0">
            <p className="font-mono text-xs truncate">{entry.query}</p>
            <div className="flex items-center gap-2 mt-0.5">
              <Badge variant="outline" className="text-[10px] px-1 py-0">
                {entry.resultType}
              </Badge>
              <span className="text-[10px] text-muted-foreground">
                {entry.duration}ms
              </span>
              <span className="text-[10px] text-muted-foreground">
                {new Date(entry.timestamp).toLocaleTimeString("en-IN", {
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </span>
            </div>
          </div>

          <Play className="h-3.5 w-3.5 mt-0.5 text-muted-foreground opacity-0 group-hover:opacity-100 shrink-0" />
        </div>
      ))}
    </div>
  );
}
