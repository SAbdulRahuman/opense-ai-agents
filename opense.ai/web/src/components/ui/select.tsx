"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { ChevronDown } from "lucide-react";

/* ---------- leaf components (used for composition) ---------- */

interface SelectItemProps {
  value: string;
  children: React.ReactNode;
}

function SelectItem({ children }: SelectItemProps) {
  return <>{children}</>;
}

function SelectValue({ placeholder }: { placeholder?: string }) {
  // Rendering handled by Select – this is a marker component
  return <>{placeholder ?? ""}</>;
}

function SelectTrigger({ children, className }: { children?: React.ReactNode; className?: string }) {
  // Rendering handled by Select – this is a marker component
  void className;
  return <>{children}</>;
}

function SelectContent({ children }: { children?: React.ReactNode }) {
  // Rendering handled by Select – this is a marker component
  return <>{children}</>;
}

/* ---------- helpers ---------- */

function collectItems(children: React.ReactNode): React.ReactElement<SelectItemProps>[] {
  const items: React.ReactElement<SelectItemProps>[] = [];
  React.Children.forEach(children, (child) => {
    if (!React.isValidElement(child)) return;
    if ((child.type as unknown) === SelectItem) {
      items.push(child as React.ReactElement<SelectItemProps>);
    } else if (child.props && (child.props as { children?: React.ReactNode }).children) {
      items.push(...collectItems((child.props as { children?: React.ReactNode }).children));
    }
  });
  return items;
}

/* ---------- Select ---------- */

interface SelectProps {
  value?: string;
  onValueChange?: (value: string) => void;
  children: React.ReactNode;
  className?: string;
  placeholder?: string;
}

function Select({ value, onValueChange, children, className, placeholder = "Select..." }: SelectProps) {
  const [isOpen, setIsOpen] = React.useState(false);
  const ref = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const options = collectItems(children);
  const selectedLabel = options.find((o) => o.props.value === value)?.props.children || placeholder;

  return (
    <div ref={ref} className={cn("relative", className)}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex h-9 w-full items-center justify-between rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
      >
        <span className="truncate">{selectedLabel}</span>
        <ChevronDown className="h-4 w-4 opacity-50" />
      </button>
      {isOpen && (
        <div className="absolute z-50 mt-1 w-full rounded-md border bg-popover p-1 text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95">
          {options.map((option) => (
            <button
              key={option.props.value}
              type="button"
              className={cn(
                "relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 px-2 text-sm outline-none hover:bg-accent hover:text-accent-foreground",
                value === option.props.value && "bg-accent text-accent-foreground",
              )}
              onClick={() => {
                onValueChange?.(option.props.value);
                setIsOpen(false);
              }}
            >
              {option.props.children}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

export { Select, SelectItem, SelectTrigger, SelectContent, SelectValue };
