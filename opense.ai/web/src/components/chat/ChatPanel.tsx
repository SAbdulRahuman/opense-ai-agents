"use client";

import { useRef, useEffect } from "react";
import { MessageSquare, Sparkles, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { MessageBubble } from "./MessageBubble";
import { ChatInput } from "./ChatInput";
import { TradeConfirm } from "./TradeConfirm";
import { useChat } from "@/hooks/useChat";
import { confirmTrade } from "@/lib/api";
import { cn } from "@/lib/utils";

export function ChatPanel() {
  const {
    messages,
    isStreaming,
    mode,
    activeAgents,
    tradeProposal,
    send,
    setMode,
    setTradeProposal,
  } = useChat();

  const scrollRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const handleTradeConfirm = async (action: "approve" | "reject") => {
    if (!tradeProposal) return;
    try {
      await confirmTrade(tradeProposal.id, action);
    } catch {
      // Handle error silently
    }
    setTradeProposal(null);
  };

  return (
    <div className="flex h-full flex-col">
      {/* Chat Header */}
      <div className="flex items-center justify-between border-b px-4 py-2">
        <div className="flex items-center gap-2">
          <MessageSquare className="h-4 w-4" />
          <span className="font-semibold text-sm">AI Assistant</span>

          {/* Active agents indicator */}
          {activeAgents.length > 0 && (
            <div className="flex items-center gap-1 ml-2">
              <Loader2 className="h-3 w-3 animate-spin text-primary" />
              <span className="text-xs text-muted-foreground">
                {activeAgents.join(", ")}
              </span>
            </div>
          )}
        </div>

        {/* Mode toggle */}
        <div className="flex items-center gap-1">
          <Button
            variant={mode === "quick" ? "default" : "ghost"}
            size="sm"
            className="h-7 text-xs"
            onClick={() => setMode("quick")}
          >
            Quick
          </Button>
          <Button
            variant={mode === "deep" ? "default" : "ghost"}
            size="sm"
            className="h-7 text-xs gap-1"
            onClick={() => setMode("deep")}
          >
            <Sparkles className="h-3 w-3" />
            Deep Analysis
          </Button>
        </div>
      </div>

      {/* Messages */}
      <div ref={scrollRef} className="flex-1 overflow-auto p-4 space-y-4">
        {messages.length === 0 ? (
          <div className="flex h-full flex-col items-center justify-center text-center text-muted-foreground">
            <MessageSquare className="h-12 w-12 mb-3 opacity-20" />
            <p className="text-sm font-medium">Start a conversation</p>
            <p className="text-xs mt-1 max-w-sm">
              Ask about any Indian stock, use slash commands like <code>/analyze RELIANCE</code>,
              or type naturally.
            </p>
          </div>
        ) : (
          messages.map((msg) => <MessageBubble key={msg.id} message={msg} />)
        )}

        {/* Streaming indicator */}
        {isStreaming && (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Thinking...</span>
          </div>
        )}

        {/* Trade confirmation */}
        {tradeProposal && (
          <TradeConfirm proposal={tradeProposal} onConfirm={handleTradeConfirm} />
        )}
      </div>

      {/* Input */}
      <ChatInput onSend={send} disabled={isStreaming} />
    </div>
  );
}
