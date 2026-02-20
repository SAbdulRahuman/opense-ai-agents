"use client";

import { useState, useRef, type KeyboardEvent } from "react";
import { Send, Slash } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const slashCommands = [
  { cmd: "/analyze", description: "Analyze a stock" },
  { cmd: "/technical", description: "Technical analysis" },
  { cmd: "/fno", description: "F&O analysis" },
  { cmd: "/report", description: "Generate report" },
  { cmd: "/trade", description: "Execute trade" },
];

interface ChatInputProps {
  onSend: (message: string) => void;
  disabled?: boolean;
}

export function ChatInput({ onSend, disabled }: ChatInputProps) {
  const [value, setValue] = useState("");
  const [showCommands, setShowCommands] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSend = () => {
    const trimmed = value.trim();
    if (!trimmed || disabled) return;
    onSend(trimmed);
    setValue("");
    setShowCommands(false);
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setValue(e.target.value);
    setShowCommands(e.target.value.startsWith("/") && e.target.value.length < 15);

    // Auto-resize
    const textarea = e.target;
    textarea.style.height = "auto";
    textarea.style.height = Math.min(textarea.scrollHeight, 120) + "px";
  };

  const insertCommand = (cmd: string) => {
    setValue(cmd + " ");
    setShowCommands(false);
    textareaRef.current?.focus();
  };

  return (
    <div className="relative border-t bg-card p-3">
      {/* Slash commands popup */}
      {showCommands && (
        <div className="absolute bottom-full left-3 mb-1 rounded-md border bg-popover p-1 shadow-md w-64">
          {slashCommands
            .filter((c) => c.cmd.startsWith(value))
            .map((c) => (
              <button
                key={c.cmd}
                className="flex w-full items-center gap-2 rounded-sm px-3 py-1.5 text-sm hover:bg-accent"
                onClick={() => insertCommand(c.cmd)}
              >
                <Slash className="h-3 w-3 text-muted-foreground" />
                <span className="font-mono font-medium">{c.cmd}</span>
                <span className="text-xs text-muted-foreground">{c.description}</span>
              </button>
            ))}
        </div>
      )}

      <div className="flex items-end gap-2">
        <textarea
          ref={textareaRef}
          value={value}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder="Ask about any stock, use /commands, or type $ for ticker autocomplete..."
          className={cn(
            "flex-1 resize-none rounded-md border bg-transparent px-3 py-2 text-sm",
            "placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring",
            "min-h-[40px] max-h-[120px]",
          )}
          rows={1}
          disabled={disabled}
        />
        <Button
          size="icon"
          onClick={handleSend}
          disabled={!value.trim() || disabled}
        >
          <Send className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
