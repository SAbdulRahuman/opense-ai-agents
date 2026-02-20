// ============================================================================
// OpeNSE.ai — Config Slice (settings state management)
// ============================================================================

import type { StateCreator } from "zustand";
import type { AppConfig, KeyStatus } from "@/lib/types";
import * as api from "@/lib/api";

export interface ConfigSlice {
  // State
  config: AppConfig | null;
  configFile: string;
  configKeys: KeyStatus[];
  configLoading: boolean;
  configSaving: boolean;
  configError: string | null;
  configDirty: boolean;
  activeSection: string;

  // Actions
  loadConfig: () => Promise<void>;
  saveConfig: (partial: Partial<AppConfig>) => Promise<void>;
  loadConfigKeys: () => Promise<void>;
  setActiveSection: (section: string) => void;
  setConfigDirty: (dirty: boolean) => void;
  clearConfigError: () => void;
}

export const createConfigSlice: StateCreator<ConfigSlice, [], [], ConfigSlice> = (set) => ({
  // Initial state
  config: null,
  configFile: "",
  configKeys: [],
  configLoading: false,
  configSaving: false,
  configError: null,
  configDirty: false,
  activeSection: "llm",

  // Actions
  loadConfig: async () => {
    set({ configLoading: true, configError: null });
    try {
      const resp = await api.getConfig();
      // The Go API wraps in { success, data: ConfigResponse }
      const data = (resp as unknown as { data: { config: AppConfig; config_file: string } }).data ?? resp;
      set({
        config: data.config,
        configFile: data.config_file,
        configLoading: false,
        configDirty: false,
      });
    } catch (err) {
      set({
        configLoading: false,
        configError: err instanceof Error ? err.message : "Failed to load config",
      });
    }
  },

  saveConfig: async (partial: Partial<AppConfig>) => {
    set({ configSaving: true, configError: null });
    try {
      const resp = await api.updateConfig(partial);
      const data = (resp as unknown as { data: { config: AppConfig; config_file: string } }).data ?? resp;
      set({
        config: data.config,
        configFile: data.config_file,
        configSaving: false,
        configDirty: false,
      });
    } catch (err) {
      set({
        configSaving: false,
        configError: err instanceof Error ? err.message : "Failed to save config",
      });
    }
  },

  loadConfigKeys: async () => {
    try {
      const resp = await api.getConfigKeys();
      const keys = (resp as unknown as { data: KeyStatus[] }).data ?? resp;
      set({ configKeys: Array.isArray(keys) ? keys : [] });
    } catch {
      // Non-critical — silently ignore
    }
  },

  setActiveSection: (section: string) => set({ activeSection: section }),
  setConfigDirty: (dirty: boolean) => set({ configDirty: dirty }),
  clearConfigError: () => set({ configError: null }),
});
