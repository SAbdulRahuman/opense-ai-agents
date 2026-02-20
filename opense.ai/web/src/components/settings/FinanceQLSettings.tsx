// ============================================================================
// OpeNSE.ai — FinanceQL Settings Section
// ============================================================================

"use client";

import { SectionForm } from "./SectionForm";
import { TextField, NumberField } from "./FormFields";
import type { FinanceQLConfig } from "@/lib/types";

interface FinanceQLSettingsProps {
  config: FinanceQLConfig;
  onChange: (updates: Partial<FinanceQLConfig>) => void;
}

export function FinanceQLSettings({ config, onChange }: FinanceQLSettingsProps) {
  return (
    <SectionForm
      title="FinanceQL"
      description="Query language cache, alerting, and REPL settings."
    >
      <NumberField
        label="Cache TTL (seconds)"
        description="Duration to cache FinanceQL query results."
        value={config.cache_ttl}
        onChange={(v) => onChange({ cache_ttl: v })}
        min={0}
        max={3600}
        step={10}
      />
      <TextField
        label="Max Range"
        description='Maximum date range for queries (e.g., "365d", "730d").'
        value={config.max_range}
        onChange={(v) => onChange({ max_range: v })}
        placeholder="365d"
      />
      <NumberField
        label="Alert Check Interval (seconds)"
        description="How often to re-evaluate alert conditions."
        value={config.alert_check_interval}
        onChange={(v) => onChange({ alert_check_interval: v })}
        min={5}
        max={300}
      />
      <TextField
        label="REPL History File"
        description="Path to store FinanceQL REPL history."
        value={config.repl_history_file}
        onChange={(v) => onChange({ repl_history_file: v })}
        placeholder="~/.openseai/financeql_history"
      />
    </SectionForm>
  );
}
