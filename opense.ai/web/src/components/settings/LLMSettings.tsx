// ============================================================================
// OpeNSE.ai — LLM Settings Section
// ============================================================================

"use client";

import { SectionForm } from "./SectionForm";
import { TextField, NumberField, SelectField } from "./FormFields";
import { KeyStatusList } from "./KeyStatusBadge";
import type { LLMConfig, KeyStatus } from "@/lib/types";

interface LLMSettingsProps {
  config: LLMConfig;
  keys: KeyStatus[];
  onChange: (updates: Partial<LLMConfig>) => void;
}

export function LLMSettings({ config, keys, onChange }: LLMSettingsProps) {
  const llmKeys = keys.filter(
    (k) =>
      k.name.includes("OpenAI") ||
      k.name.includes("Gemini") ||
      k.name.includes("Anthropic"),
  );

  return (
    <div className="space-y-6">
      <SectionForm
        title="LLM Provider"
        description="Configure your AI model provider and parameters."
      >
        <SelectField
          label="Primary Provider"
          description="The default LLM provider for analysis and chat."
          value={config.primary}
          onChange={(v) => onChange({ primary: v })}
          options={[
            { value: "openai", label: "OpenAI" },
            { value: "ollama", label: "Ollama (Local)" },
            { value: "gemini", label: "Google Gemini" },
            { value: "anthropic", label: "Anthropic Claude" },
          ]}
        />

        <TextField
          label="Model"
          description="Model name (e.g., gpt-4o, qwen2.5:32b, gemini-pro)."
          value={config.model}
          onChange={(v) => onChange({ model: v })}
          placeholder="gpt-4o"
        />

        <TextField
          label="Fallback Model"
          description="Used when the primary model fails or is unavailable."
          value={config.fallback_model}
          onChange={(v) => onChange({ fallback_model: v })}
          placeholder="gpt-4o-mini"
        />

        {config.primary === "ollama" && (
          <TextField
            label="Ollama URL"
            description="Ollama server endpoint."
            value={config.ollama_url}
            onChange={(v) => onChange({ ollama_url: v })}
            placeholder="http://localhost:11434"
          />
        )}

        <div className="grid grid-cols-2 gap-4">
          <NumberField
            label="Temperature"
            description="0 = deterministic, 1 = creative"
            value={config.temperature}
            onChange={(v) => onChange({ temperature: v })}
            min={0}
            max={2}
            step={0.1}
          />
          <NumberField
            label="Max Tokens"
            description="Maximum response length."
            value={config.max_tokens}
            onChange={(v) => onChange({ max_tokens: v })}
            min={256}
            max={128000}
            step={256}
          />
        </div>
      </SectionForm>

      {llmKeys.length > 0 && <KeyStatusList keys={llmKeys} />}
    </div>
  );
}
