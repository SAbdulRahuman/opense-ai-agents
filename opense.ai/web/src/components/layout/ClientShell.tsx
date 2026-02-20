// ============================================================================
// OpeNSE.ai — ClientShell (client layout wrapper)
// Embeds the chat panel inline as a flex sibling so it pushes content
// instead of overlaying it.
// ============================================================================
"use client";

import { ChatDrawer } from "@/components/chat";

export function ClientShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-full min-h-0 flex-1 overflow-hidden">
      {/* Main page content (shrinks when chat is open) */}
      <div className="flex-1 overflow-auto">{children}</div>
      {/* Inline chat panel */}
      <ChatDrawer />
    </div>
  );
}
