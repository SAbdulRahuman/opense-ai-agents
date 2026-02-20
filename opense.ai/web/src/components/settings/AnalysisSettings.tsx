// ============================================================================
// OpeNSE.ai — Analysis Settings Section
// ============================================================================

"use client";

import { SectionForm } from "./SectionForm";
import { NumberField } from "./FormFields";
import type { AnalysisConfig } from "@/lib/types";

interface AnalysisSettingsProps {
  config: AnalysisConfig;
  onChange: (updates: Partial<AnalysisConfig>) => void;
}

export function AnalysisSettings({ config, onChange }: AnalysisSettingsProps) {
  return (
    <SectionForm
      title="Analysis Engine"
      description="Tune caching and concurrency for data fetching."
    >
      <NumberField
        label="Cache TTL (seconds)"
        description="How long market data is cached before re-fetching."
        value={config.cache_ttl}
        onChange={(v) => onChange({ cache_ttl: v })}
        min={0}
        max={3600}
        step={30}
      />
      <NumberField
        label="Concurrent Fetches"
        description="Number of parallel goroutines for data retrieval."
        value={config.concurrent_fetches}
        onChange={(v) => onChange({ concurrent_fetches: v })}
        min={1}
        max={20}
      />
    </SectionForm>
  );
}
