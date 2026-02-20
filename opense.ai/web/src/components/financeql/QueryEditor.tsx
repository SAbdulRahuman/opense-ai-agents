"use client";

import { useEffect, useRef, useCallback, useState } from "react";
import { Play, HelpCircle, Languages, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Select, SelectItem } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface QueryEditorProps {
  value: string;
  onChange: (value: string) => void;
  onExecute: () => void;
  onExplain: () => void;
  isExecuting: boolean;
  naturalLanguageMode: boolean;
  onToggleNL: (mode: boolean) => void;
  timeRange: { start?: string; end?: string; relative?: string };
  onTimeRangeChange: (range: { start?: string; end?: string; relative?: string }) => void;
}

const timeRangeOptions = [
  { value: "1d", label: "Last 1 Day" },
  { value: "7d", label: "Last 7 Days" },
  { value: "30d", label: "Last 30 Days" },
  { value: "90d", label: "Last 90 Days" },
  { value: "180d", label: "Last 6 Months" },
  { value: "365d", label: "Last 1 Year" },
];

export function QueryEditor({
  value,
  onChange,
  onExecute,
  onExplain,
  isExecuting,
  naturalLanguageMode,
  onToggleNL,
  timeRange,
  onTimeRangeChange,
}: QueryEditorProps) {
  const editorContainerRef = useRef<HTMLDivElement>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const editorRef = useRef<any>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [Monaco, setMonaco] = useState<any>(null);

  // Dynamically import Monaco Editor (client-only)
  useEffect(() => {
    import("@monaco-editor/react").then(setMonaco);
  }, []);

  const handleEditorMount = useCallback(
    (editor: unknown, monaco: unknown) => {
      // Register FinanceQL language
      const m = monaco as typeof import("monaco-editor");
      if (!m.languages.getLanguages().some((l: { id: string }) => l.id === "financeql")) {
        m.languages.register({ id: "financeql" });

        import("@/lib/financeql-lang").then(
          ({ financeqlLanguage, financeqlCompletionItems, financeqlTheme }) => {
            m.languages.setMonarchTokensProvider("financeql", financeqlLanguage);
            m.editor.defineTheme("financeql-dark", financeqlTheme);
            m.editor.setTheme("financeql-dark");

            // Completion provider
            m.languages.registerCompletionItemProvider("financeql", {
              provideCompletionItems: (model: unknown, position: unknown) => {
                const p = position as { lineNumber: number; column: number };
                const mdl = model as { getWordUntilPosition: (p: unknown) => { startColumn: number; endColumn: number } };
                const word = mdl.getWordUntilPosition(position);
                const range = {
                  startLineNumber: p.lineNumber,
                  endLineNumber: p.lineNumber,
                  startColumn: word.startColumn,
                  endColumn: word.endColumn,
                };
                return {
                  suggestions: financeqlCompletionItems.map((item) => ({
                    ...item,
                    range,
                  })),
                };
              },
            });
          },
        );
      }

      // Keybinding: Ctrl+Enter to execute
      const ed = editor as { addCommand: (keybinding: number, handler: () => void) => void };
      ed.addCommand(m.KeyMod.CtrlCmd | m.KeyCode.Enter, () => {
        onExecute();
      });
    },
    [onExecute],
  );

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
      e.preventDefault();
      onExecute();
    }
  };

  return (
    <div className="space-y-2">
      {/* Toolbar */}
      <div className="flex flex-wrap items-center gap-2">
        <Button
          size="sm"
          onClick={onExecute}
          disabled={isExecuting || !value.trim()}
          className="gap-1"
        >
          {isExecuting ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <Play className="h-3.5 w-3.5" />
          )}
          Execute
        </Button>

        <Button
          variant="outline"
          size="sm"
          onClick={onExplain}
          disabled={isExecuting || !value.trim()}
          className="gap-1"
        >
          <HelpCircle className="h-3.5 w-3.5" />
          Explain
        </Button>

        <Button
          variant={naturalLanguageMode ? "secondary" : "ghost"}
          size="sm"
          onClick={() => onToggleNL(!naturalLanguageMode)}
          className={cn("gap-1", naturalLanguageMode && "border border-primary/30")}
        >
          <Languages className="h-3.5 w-3.5" />
          Natural Language
        </Button>

        <div className="ml-auto flex items-center gap-2">
          <span className="text-xs text-muted-foreground">Range:</span>
          <Select
            value={timeRange.relative || "30d"}
            onValueChange={(v) => onTimeRangeChange({ relative: v })}
            className="w-36"
          >
            {timeRangeOptions.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </Select>
        </div>

        <Badge variant="outline" className="text-xs">
          Ctrl+Enter to execute
        </Badge>
      </div>

      {/* Editor */}
      <div
        ref={editorContainerRef}
        className="rounded-md border overflow-hidden"
        onKeyDown={handleKeyDown}
      >
        {Monaco ? (
          <Monaco.default
            height="120px"
            language={naturalLanguageMode ? "plaintext" : "financeql"}
            theme="financeql-dark"
            value={value}
            onChange={(v: string | undefined) => onChange(v || "")}
            onMount={handleEditorMount}
            options={{
              fontSize: 14,
              fontFamily: "var(--font-geist-mono), monospace",
              minimap: { enabled: false },
              lineNumbers: "off",
              scrollBeyondLastLine: false,
              wordWrap: "on",
              padding: { top: 8, bottom: 8 },
              renderLineHighlight: "none",
              overviewRulerBorder: false,
              scrollbar: {
                vertical: "hidden",
                horizontal: "auto",
              },
              placeholder: naturalLanguageMode
                ? "Describe what you want in plain English..."
                : 'e.g., rsi(RELIANCE, 14) | price(TCS)[90d] | screener("pe < 15 AND roe > 20")',
            }}
          />
        ) : (
          <textarea
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={
              naturalLanguageMode
                ? "Describe what you want in plain English..."
                : 'e.g., rsi(RELIANCE, 14) | price(TCS)[90d]'
            }
            className="w-full h-[120px] bg-[#0a0a0a] text-[#d1d5db] p-3 font-mono text-sm resize-none focus:outline-none"
          />
        )}
      </div>
    </div>
  );
}
