// ============================================================================
// OpeNSE.ai — Settings Page
// ============================================================================

"use client";

import { useEffect, useCallback, useMemo } from "react";
import { useStore } from "@/store";
import { Button } from "@/components/ui/button";
import {
  LLMSettings,
  BrokerSettings,
  TradingSettings,
  AnalysisSettings,
  FinanceQLSettings,
  APISettings,
  WebSettings,
  LoggingSettings,
} from "@/components/settings";
import type { AppConfig } from "@/lib/types";
import { cn } from "@/lib/utils";
import {
  Brain,
  Landmark,
  ShieldCheck,
  BarChart3,
  Terminal,
  Server,
  Globe,
  ScrollText,
  Save,
  Loader2,
  RefreshCw,
  FileText,
  AlertCircle,
} from "lucide-react";

// Sidebar navigation sections
const SECTIONS = [
  { id: "llm", label: "LLM Provider", icon: Brain },
  { id: "broker", label: "Broker", icon: Landmark },
  { id: "trading", label: "Trading & Risk", icon: ShieldCheck },
  { id: "analysis", label: "Analysis", icon: BarChart3 },
  { id: "financeql", label: "FinanceQL", icon: Terminal },
  { id: "api", label: "API Server", icon: Server },
  { id: "web", label: "Web UI", icon: Globe },
  { id: "logging", label: "Logging", icon: ScrollText },
] as const;

export default function SettingsPage() {
  const {
    config,
    configFile,
    configKeys,
    configLoading,
    configSaving,
    configError,
    configDirty,
    activeSection,
    loadConfig,
    saveConfig,
    loadConfigKeys,
    setActiveSection,
    setConfigDirty,
    clearConfigError,
  } = useStore();

  // Load config on mount
  useEffect(() => {
    loadConfig();
    loadConfigKeys();
  }, [loadConfig, loadConfigKeys]);

  // Accumulate partial changes locally; will be sent on save.
  // We merge changes into the store config directly for live preview.
  const handleChange = useCallback(
    (section: keyof AppConfig, updates: Record<string, unknown>) => {
      if (!config) return;
      // Merge into the store config for immediate UI feedback
      const merged = { ...config, [section]: { ...config[section], ...updates } };
      // We use a trick: update the store config in-place via saveConfig
      // But we don't persist yet — just mark dirty
      useStore.setState({ config: merged as AppConfig, configDirty: true });
    },
    [config],
  );

  const handleSave = useCallback(() => {
    if (!config) return;
    saveConfig(config);
  }, [config, saveConfig]);

  const handleReset = useCallback(() => {
    clearConfigError();
    loadConfig();
  }, [clearConfigError, loadConfig]);

  // Render active section
  const renderSection = useMemo(() => {
    if (!config) return null;

    switch (activeSection) {
      case "llm":
        return (
          <LLMSettings
            config={config.llm}
            keys={configKeys}
            onChange={(u) => handleChange("llm", u as Record<string, unknown>)}
          />
        );
      case "broker":
        return (
          <BrokerSettings
            config={config.broker}
            keys={configKeys}
            onChange={(u) => handleChange("broker", u as Record<string, unknown>)}
          />
        );
      case "trading":
        return (
          <TradingSettings
            config={config.trading}
            onChange={(u) => handleChange("trading", u as Record<string, unknown>)}
          />
        );
      case "analysis":
        return (
          <AnalysisSettings
            config={config.analysis}
            onChange={(u) => handleChange("analysis", u as Record<string, unknown>)}
          />
        );
      case "financeql":
        return (
          <FinanceQLSettings
            config={config.financeql}
            onChange={(u) => handleChange("financeql", u as Record<string, unknown>)}
          />
        );
      case "api":
        return (
          <APISettings
            config={config.api}
            onChange={(u) => handleChange("api", u as Record<string, unknown>)}
          />
        );
      case "web":
        return (
          <WebSettings
            config={config.web}
            onChange={(u) => handleChange("web", u as Record<string, unknown>)}
          />
        );
      case "logging":
        return (
          <LoggingSettings
            config={config.logging}
            onChange={(u) => handleChange("logging", u as Record<string, unknown>)}
          />
        );
      default:
        return null;
    }
  }, [activeSection, config, configKeys, handleChange]);

  // Loading state
  if (configLoading && !config) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="flex h-full">
      {/* Section Navigation (left sidebar) */}
      <nav className="w-52 shrink-0 space-y-1 border-r p-3">
        <h2 className="mb-3 px-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Settings
        </h2>
        {SECTIONS.map((s) => (
          <button
            key={s.id}
            onClick={() => setActiveSection(s.id)}
            className={cn(
              "flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-sm font-medium transition-colors",
              "hover:bg-accent hover:text-accent-foreground",
              activeSection === s.id && "bg-accent text-accent-foreground",
            )}
          >
            <s.icon className="h-4 w-4 shrink-0" />
            <span className="truncate">{s.label}</span>
          </button>
        ))}

        {/* Config file info */}
        {configFile && (
          <div className="mt-6 border-t pt-3">
            <div className="flex items-start gap-1.5 px-2">
              <FileText className="mt-0.5 h-3 w-3 shrink-0 text-muted-foreground" />
              <p className="break-all text-xs text-muted-foreground" title={configFile}>
                {configFile}
              </p>
            </div>
          </div>
        )}
      </nav>

      {/* Content area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Top bar with save / reset */}
        <div className="flex items-center justify-between border-b px-6 py-3">
          <div className="flex items-center gap-2">
            <h1 className="text-lg font-semibold">
              {SECTIONS.find((s) => s.id === activeSection)?.label ?? "Settings"}
            </h1>
            {configDirty && (
              <span className="rounded bg-yellow-500/10 px-1.5 py-0.5 text-xs font-medium text-yellow-600 dark:text-yellow-400">
                Unsaved changes
              </span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleReset}
              disabled={configLoading}
            >
              <RefreshCw className="mr-1 h-3 w-3" />
              Reset
            </Button>
            <Button
              size="sm"
              onClick={handleSave}
              disabled={configSaving || !configDirty}
            >
              {configSaving ? (
                <Loader2 className="mr-1 h-3 w-3 animate-spin" />
              ) : (
                <Save className="mr-1 h-3 w-3" />
              )}
              Save Changes
            </Button>
          </div>
        </div>

        {/* Error banner */}
        {configError && (
          <div className="mx-6 mt-3 flex items-center gap-2 rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-2 text-sm text-destructive">
            <AlertCircle className="h-4 w-4 shrink-0" />
            {configError}
            <button
              className="ml-auto text-xs underline"
              onClick={clearConfigError}
            >
              Dismiss
            </button>
          </div>
        )}

        {/* Section content */}
        <div className="flex-1 overflow-y-auto p-6">
          <div className="mx-auto max-w-2xl">{renderSection}</div>
        </div>
      </div>
    </div>
  );
}
