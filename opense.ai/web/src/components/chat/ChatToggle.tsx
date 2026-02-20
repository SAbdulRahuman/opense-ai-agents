// ============================================================================
// OpeNSE.ai — ChatToggle (floating button to open/close the global chat)
// ============================================================================
"use client";

import { useStore } from "@/store";
import { cn } from "@/lib/utils";
import { MessageSquare, X } from "lucide-react";

export function ChatToggle() {
  const { isChatOpen, setChatOpen, messages } = useStore();
  const unread = !isChatOpen && messages.length > 0;

  return (
    <button
      onClick={() => setChatOpen(!isChatOpen)}
      className={cn(
        "fixed bottom-5 right-5 z-40 flex h-12 w-12 items-center justify-center rounded-full shadow-lg transition-all duration-200",
        "hover:scale-105 active:scale-95",
        isChatOpen
          ? "bg-muted text-muted-foreground hover:bg-muted/80"
          : "bg-primary text-primary-foreground hover:bg-primary/90",
      )}
      title={isChatOpen ? "Close chat (Ctrl+Shift+L)" : "Open chat (Ctrl+Shift+L)"}
    >
      {isChatOpen ? (
        <X className="h-5 w-5" />
      ) : (
        <div className="relative">
          <MessageSquare className="h-5 w-5" />
          {unread && (
            <span className="absolute -right-1 -top-1 h-2.5 w-2.5 rounded-full bg-red-500 ring-2 ring-background" />
          )}
        </div>
      )}
    </button>
  );
}
