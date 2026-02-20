// ============================================================================
// OpeNSE.ai — API Key Status Badge
// ============================================================================

"use client";

import { Badge } from "@/components/ui/badge";
import { CheckCircle, XCircle, MinusCircle } from "lucide-react";
import type { KeyStatus } from "@/lib/types";

export function KeyStatusBadge({ keyInfo }: { keyInfo: KeyStatus }) {
  if (!keyInfo.is_set) {
    return (
      <div className="flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/5 p-3">
        <XCircle className="h-4 w-4 text-destructive" />
        <div className="flex-1">
          <p className="text-sm font-medium">{keyInfo.name}</p>
          <p className="text-xs text-muted-foreground">Not configured</p>
        </div>
        <Badge variant="destructive">Missing</Badge>
      </div>
    );
  }

  const isEnv = keyInfo.source === "env";

  return (
    <div className="flex items-center gap-2 rounded-lg border p-3">
      <CheckCircle className="h-4 w-4 text-green-500" />
      <div className="flex-1">
        <p className="text-sm font-medium">{keyInfo.name}</p>
        <p className="font-mono text-xs text-muted-foreground">{keyInfo.masked}</p>
      </div>
      <Badge variant={isEnv ? "secondary" : "outline"}>
        {isEnv ? "ENV" : "Config"}
      </Badge>
    </div>
  );
}

export function KeyStatusList({ keys }: { keys: KeyStatus[] }) {
  if (keys.length === 0) return null;

  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium text-muted-foreground">API Key Status</h4>
      {keys.map((k) => (
        <KeyStatusBadge key={k.name} keyInfo={k} />
      ))}
      <p className="text-xs text-muted-foreground">
        API keys are managed via environment variables or config file. They are not editable from the UI for security.
      </p>
    </div>
  );
}
