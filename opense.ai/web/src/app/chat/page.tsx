import { ChatPanel } from "@/components/chat";

export const metadata = {
  title: "AI Chat | OpeNSE.ai",
};

export default function ChatPage() {
  return (
    <div className="h-[calc(100vh-8rem)]">
      <ChatPanel />
    </div>
  );
}
