// ============================================================================
// OpeNSE.ai — Trading Safety Settings Section
// ============================================================================

"use client";

import { SectionForm } from "./SectionForm";
import { NumberField, SelectField, ToggleField } from "./FormFields";
import type { TradingConfig } from "@/lib/types";

interface TradingSettingsProps {
  config: TradingConfig;
  onChange: (updates: Partial<TradingConfig>) => void;
}

export function TradingSettings({ config, onChange }: TradingSettingsProps) {
  return (
    <SectionForm
      title="Trading & Risk Management"
      description="Safety-first settings for position sizing, loss limits, and trade confirmation."
    >
      <SelectField
        label="Trading Mode"
        description="Paper mode simulates trades without real money."
        value={config.mode}
        onChange={(v) => onChange({ mode: v })}
        options={[
          { value: "paper", label: "Paper (Simulated)" },
          { value: "live", label: "Live Trading" },
        ]}
      />

      {config.mode === "live" && (
        <div className="rounded-lg border border-yellow-500/30 bg-yellow-500/5 p-3 text-sm text-yellow-600 dark:text-yellow-400">
          <strong>Warning:</strong> Live trading uses real money. Ensure risk limits are properly configured.
        </div>
      )}

      <NumberField
        label="Initial Capital (₹)"
        description="Starting capital for portfolio tracking and position sizing."
        value={config.initial_capital}
        onChange={(v) => onChange({ initial_capital: v })}
        min={0}
        step={100000}
      />

      <div className="grid grid-cols-2 gap-4">
        <NumberField
          label="Max Position (%)"
          description="Max percentage of capital per single trade."
          value={config.max_position_pct}
          onChange={(v) => onChange({ max_position_pct: v })}
          min={0.1}
          max={100}
          step={0.5}
        />
        <NumberField
          label="Daily Loss Limit (%)"
          description="Trading halts if daily loss exceeds this."
          value={config.daily_loss_limit_pct}
          onChange={(v) => onChange({ daily_loss_limit_pct: v })}
          min={0.1}
          max={100}
          step={0.5}
        />
      </div>

      <NumberField
        label="Max Open Positions"
        description="Maximum number of simultaneous open positions."
        value={config.max_open_positions}
        onChange={(v) => onChange({ max_open_positions: v })}
        min={1}
        max={100}
      />

      <ToggleField
        label="Require Confirmation"
        description="Require manual approval before executing live trades (human-in-the-loop)."
        checked={config.require_confirmation}
        onChange={(v) => onChange({ require_confirmation: v })}
      />

      {config.require_confirmation && (
        <NumberField
          label="Confirmation Timeout (seconds)"
          description="Auto-reject if not confirmed within this time."
          value={config.confirm_timeout_sec}
          onChange={(v) => onChange({ confirm_timeout_sec: v })}
          min={10}
          max={600}
        />
      )}
    </SectionForm>
  );
}
