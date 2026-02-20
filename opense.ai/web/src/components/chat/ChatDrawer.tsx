// ============================================================================
// OpeNSE.ai — ChatDrawer (global slide-in chat panel, like VS Code Copilot)
// ============================================================================
"use client";

import { useEffect, useCallback } from "react";
import { useStore } from "@/store";
import { ChatPanel } from "@/components/chat";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { X, Minus, Maximize2, Minimize2 } from "lucide-react";

export function ChatDrawer() {
  const {
    isChatOpen,
    setChatOpen,
    chatDrawerSize,
    setChatDrawerSize,
  } = useStore();

  // Keyboard shortcut: Ctrl+Shift+L to toggle (same as VS Code Copilot)
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === "l") {
        e.preventDefault();
        setChatOpen(!isChatOpen);
      }
      // Escape closes when chat is open
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

  const widthClass = chatDrawerSize === "large" ? "w-[560px]" : "w-[380px]";

  return (
    <>
      {/* Backdrop (only in mobile or when large) */}
      {isChatOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/20 backdrop-blur-[1px] md:hidden"
          onClick={() => setChatOpen(false)}
        />
      )}

      {/* Drawer panel */}
      <div
        className={cn(
          "fixed right-0 top-0 z-50 flex h-full flex-col border-l bg-background shadow-xl transition-transform duration-300 ease-in-out",
          widthClass,
          isChatOpen ? "translate-x-0" : "translate-x-full",
        )}
      >
        {/* Drawer header */}
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

        {/* Chat content — reuse the existing ChatPanel */}
        <div className="flex-1 overflow-hidden">
          <ChatPanel />
        </div>
      </div>
    </>
  );
}
