// ============================================================================
// OpeNSE.ai — Web UI Settings Section
// ============================================================================

"use client";

import { SectionForm } from "./SectionForm";
import { TextField } from "./FormFields";
import type { WebUIConfig } from "@/lib/types";

interface WebSettingsProps {
  config: WebUIConfig;
  onChange: (updates: Partial<WebUIConfig>) => void;
}

export function WebSettings({ config, onChange }: WebSettingsProps) {
  return (
    <SectionForm
      title="Web UI"
      description="Frontend URL used by the backend for CORS and redirects."
    >
      <TextField
        label="Web UI URL"
        description="The URL where the Next.js frontend is served."
        value={config.url}
        onChange={(v) => onChange({ url: v })}
        placeholder="http://localhost:3000"
      />
    </SectionForm>
  );
}
