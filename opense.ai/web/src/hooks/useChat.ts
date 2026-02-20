"use client";

import { useCallback } from "react";
import { useStore } from "@/store";
import { sendChatMessage } from "@/lib/api";
import type { ChatMessage } from "@/lib/types";

export function useChat() {
  const {
    messages,
    isStreaming,
    mode,
    activeAgents,
    tradeProposal,
    addMessage,
    updateMessage,
    setStreaming,
    setMode,
    setActiveAgents,
    setTradeProposal,
  } = useStore();

  const send = useCallback(
    async (content: string) => {
      const userMsg: ChatMessage = {
        id: `user-${Date.now()}`,
        role: "user",
        content,
        timestamp: new Date().toISOString(),
      };
      addMessage(userMsg);
      setStreaming(true);

      try {
        const response = await sendChatMessage(content, mode, messages);
        addMessage({
          ...response,
          id: response.id || `assistant-${Date.now()}`,
          timestamp: response.timestamp || new Date().toISOString(),
        });
      } catch (error) {
        addMessage({
          id: `error-${Date.now()}`,
          role: "assistant",
          content: `Error: ${error instanceof Error ? error.message : "Failed to get response"}`,
          timestamp: new Date().toISOString(),
        });
      } finally {
        setStreaming(false);
        setActiveAgents([]);
      }
    },
    [messages, mode, addMessage, setStreaming, setActiveAgents],
  );

  return {
    messages,
    isStreaming,
    mode,
    activeAgents,
    tradeProposal,
    send,
    setMode,
    setTradeProposal,
  };
}
