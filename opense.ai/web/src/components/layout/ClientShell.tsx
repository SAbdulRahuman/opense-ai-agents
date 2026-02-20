// ============================================================================
// OpeNSE.ai — ClientShell (client wrapper for global overlays)
// ============================================================================
"use client";

import { ChatDrawer, ChatToggle } from "@/components/chat";

export function ClientShell({ children }: { children: React.ReactNode }) {
  return (
    <>
      {children}
      <ChatDrawer />
      <ChatToggle />
    </>
  );
}
