// ============================================================================
// OpeNSE.ai — API Server Settings Section
// ============================================================================

"use client";

import { useState } from "react";
import { SectionForm } from "./SectionForm";
import { TextField, NumberField } from "./FormFields";
import { Button } from "@/components/ui/button";
import { Plus, X } from "lucide-react";
import type { APIServerConfig } from "@/lib/types";

interface APISettingsProps {
  config: APIServerConfig;
  onChange: (updates: Partial<APIServerConfig>) => void;
}

export function APISettings({ config, onChange }: APISettingsProps) {
  const [newOrigin, setNewOrigin] = useState("");

  const addOrigin = () => {
    const trimmed = newOrigin.trim();
    if (trimmed && !config.cors_origins.includes(trimmed)) {
      onChange({ cors_origins: [...config.cors_origins, trimmed] });
      setNewOrigin("");
    }
  };

  const removeOrigin = (idx: number) => {
    onChange({ cors_origins: config.cors_origins.filter((_, i) => i !== idx) });
  };

  return (
    <SectionForm
      title="API Server"
      description="HTTP server host, port, and CORS settings. Changes may require a restart."
    >
      <div className="grid grid-cols-2 gap-4">
        <TextField
          label="Host"
          description="Bind address (0.0.0.0 = all interfaces)."
          value={config.host}
          onChange={(v) => onChange({ host: v })}
          placeholder="0.0.0.0"
        />
        <NumberField
          label="Port"
          value={config.port}
          onChange={(v) => onChange({ port: v })}
          min={1}
          max={65535}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">CORS Origins</label>
        <div className="space-y-1">
          {config.cors_origins.map((origin, idx) => (
            <div key={idx} className="flex items-center gap-2">
              <code className="flex-1 rounded border bg-muted px-2 py-1 text-xs">
                {origin}
              </code>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0"
                onClick={() => removeOrigin(idx)}
              >
                <X className="h-3 w-3" />
              </Button>
            </div>
          ))}
        </div>
        <div className="flex gap-2">
          <TextField
            label=""
            value={newOrigin}
            onChange={setNewOrigin}
            placeholder="https://example.com"
          />
          <Button
            variant="outline"
            size="sm"
            className="mt-auto"
            onClick={addOrigin}
            disabled={!newOrigin.trim()}
          >
            <Plus className="mr-1 h-3 w-3" /> Add
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          Allowed origins for cross-origin requests.
        </p>
      </div>
    </SectionForm>
  );
}
