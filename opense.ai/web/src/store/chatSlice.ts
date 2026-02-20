// ============================================================================
// OpeNSE.ai — Chat Store Slice (Zustand)
// ============================================================================

import type { StateCreator } from "zustand";
import type { ChatMessage, TradeProposal } from "@/lib/types";

export interface ChatSlice {
  messages: ChatMessage[];
  isStreaming: boolean;
  mode: "quick" | "deep";
  activeAgents: string[];
  tradeProposal: TradeProposal | null;

  addMessage: (msg: ChatMessage) => void;
  updateMessage: (id: string, update: Partial<ChatMessage>) => void;
  clearMessages: () => void;
  setStreaming: (streaming: boolean) => void;
  setMode: (mode: "quick" | "deep") => void;
  setActiveAgents: (agents: string[]) => void;
  setTradeProposal: (proposal: TradeProposal | null) => void;
}

export const createChatSlice: StateCreator<ChatSlice> = (set) => ({
  messages: [],
  isStreaming: false,
  mode: "quick",
  activeAgents: [],
  tradeProposal: null,

  addMessage: (msg) =>
    set((state) => ({ messages: [...state.messages, msg] })),

  updateMessage: (id, update) =>
    set((state) => ({
      messages: state.messages.map((m) =>
        m.id === id ? { ...m, ...update } : m,
      ),
    })),

  clearMessages: () => set({ messages: [] }),
  setStreaming: (isStreaming) => set({ isStreaming }),
  setMode: (mode) => set({ mode }),
  setActiveAgents: (activeAgents) => set({ activeAgents }),
  setTradeProposal: (tradeProposal) => set({ tradeProposal }),
});
