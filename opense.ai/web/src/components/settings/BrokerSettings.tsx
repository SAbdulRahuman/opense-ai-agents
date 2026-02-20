// ============================================================================
// OpeNSE.ai — Broker Settings Section
// ============================================================================

"use client";

import { SectionForm } from "./SectionForm";
import { TextField, NumberField, SelectField } from "./FormFields";
import { KeyStatusList } from "./KeyStatusBadge";
import type { BrokerConfig, KeyStatus } from "@/lib/types";

interface BrokerSettingsProps {
  config: BrokerConfig;
  keys: KeyStatus[];
  onChange: (updates: Partial<BrokerConfig>) => void;
}

export function BrokerSettings({ config, keys, onChange }: BrokerSettingsProps) {
  const brokerKeys = keys.filter((k) => k.name.includes("Zerodha"));

  return (
    <div className="space-y-6">
      <SectionForm
        title="Broker Integration"
        description="Configure your broker for paper or live trading."
      >
        <SelectField
          label="Broker Provider"
          description="Choose your broker or use paper trading for simulation."
          value={config.provider}
          onChange={(v) => onChange({ provider: v })}
          options={[
            { value: "paper", label: "Paper Trading (Simulated)" },
            { value: "zerodha", label: "Zerodha Kite" },
            { value: "ibkr", label: "Interactive Brokers (IBKR)" },
          ]}
        />

        {config.provider === "ibkr" && (
          <div className="grid grid-cols-2 gap-4">
            <TextField
              label="IBKR Host"
              value={config.ibkr.host}
              onChange={(v) => onChange({ ibkr: { ...config.ibkr, host: v } })}
              placeholder="127.0.0.1"
            />
            <NumberField
              label="IBKR Port"
              value={config.ibkr.port}
              onChange={(v) => onChange({ ibkr: { ...config.ibkr, port: v } })}
              min={1}
              max={65535}
            />
          </div>
        )}
      </SectionForm>

      {brokerKeys.length > 0 && <KeyStatusList keys={brokerKeys} />}
    </div>
  );
}
