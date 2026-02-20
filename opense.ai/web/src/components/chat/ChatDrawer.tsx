// ============================================================================
// OpeNSE.ai — ChatDrawer (inline embedded chat panel)
// Renders as a flex sibling inside the layout — pushes content instead of
// overlaying it, so the page remains visible and interactive.
// ============================================================================
"use client";

import { useEffect, useCallback } from "react";
import { useStore } from "@/store";
import { ChatPanel } from "@/components/chat";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { X, Maximize2, Minimize2 } from "lucide-react";

export function ChatDrawer() {
  const {
    isChatOpen,
    setChatOpen,
    chatDrawerSize,
    setChatDrawerSize,
  } = useStore();

  // Keyboard shortcut: Ctrl+Shift+L to toggle
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === "l") {
        e.preventDefault();
        setChatOpen(!isChatOpen);
      }
      if (e.key === "Escape" && isChatOpen) {
        setChatOpen(false);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [isChatOpen, setChatOpen]);

  const toggleSize = useCallback(() => {
    setChatDrawerSize(chatDrawerSize === "normal" ? "large" : "normal");
  }, [chatDrawerSize, setChatDrawerSize]);

  if (!isChatOpen) return null;

  const widthClass = chatDrawerSize === "large" ? "w-[560px]" : "w-[380px]";

  return (
    <aside
      className={cn(
        "flex h-full shrink-0 flex-col border-l bg-background",
        widthClass,
      )}
    >
      {/* Header */}
      <div className="flex h-10 items-center justify-between border-b bg-card px-3">
        <div className="flex items-center gap-2">
          <div className="h-2 w-2 rounded-full bg-primary animate-pulse" />
          <span className="text-xs font-semibold">AI Chat</span>
          <kbd className="hidden md:inline-flex h-4 items-center rounded border bg-muted px-1 text-[9px] font-mono text-muted-foreground">
            ⌘⇧L
          </kbd>
        </div>
        <div className="flex items-center gap-0.5">
          <Button
            variant="ghost"
            size="sm"
            className="h-6 w-6 p-0"
            onClick={toggleSize}
            title={chatDrawerSize === "normal" ? "Expand" : "Shrink"}
          >
            {chatDrawerSize === "normal" ? (
              <Maximize2 className="h-3 w-3" />
            ) : (
              <Minimize2 className="h-3 w-3" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-6 w-6 p-0"
            onClick={() => setChatOpen(false)}
          >
            <X className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      {/* Chat content */}
      <div className="flex-1 overflow-hidden">
        <ChatPanel />
      </div>
    </aside>
  );
}
