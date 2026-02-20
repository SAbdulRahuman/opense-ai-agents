// ============================================================================
// OpeNSE.ai — Logging Settings Section
// ============================================================================

"use client";

import { SectionForm } from "./SectionForm";
import { SelectField } from "./FormFields";
import type { LoggingConfig } from "@/lib/types";

interface LoggingSettingsProps {
  config: LoggingConfig;
  onChange: (updates: Partial<LoggingConfig>) => void;
}

export function LoggingSettings({ config, onChange }: LoggingSettingsProps) {
  return (
    <SectionForm
      title="Logging"
      description="Server log level and output format."
    >
      <SelectField
        label="Log Level"
        description="Minimum severity to log."
        value={config.level}
        onChange={(v) => onChange({ level: v })}
        options={[
          { value: "debug", label: "Debug" },
          { value: "info", label: "Info" },
          { value: "warn", label: "Warning" },
          { value: "error", label: "Error" },
        ]}
      />
      <SelectField
        label="Log Format"
        description="Output format for log entries."
        value={config.format}
        onChange={(v) => onChange({ format: v })}
        options={[
          { value: "text", label: "Text (human-readable)" },
          { value: "json", label: "JSON (structured)" },
        ]}
      />
    </SectionForm>
  );
}
