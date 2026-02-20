"use client";

import { useState } from "react";
import { ChevronRight, ChevronDown, Parentheses, Hash, Type, Timer } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ASTNode {
  type: "function" | "operator" | "ticker" | "number" | "string" | "duration" | "range";
  name: string;
  children?: ASTNode[];
}

interface ExpressionTreeProps {
  ast: ASTNode;
}

function NodeIcon({ type }: { type: ASTNode["type"] }) {
  const size = 14;
  switch (type) {
    case "function":
      return <Parentheses size={size} className="text-blue-500" />;
    case "operator":
      return <span className="text-red-500 font-mono text-xs font-bold">op</span>;
    case "ticker":
      return <span className="text-green-500 font-mono text-xs font-bold">$</span>;
    case "number":
      return <Hash size={size} className="text-orange-500" />;
    case "string":
      return <Type size={size} className="text-yellow-500" />;
    case "duration":
      return <Timer size={size} className="text-purple-500" />;
    case "range":
      return <span className="text-cyan-500 font-mono text-xs font-bold">[ ]</span>;
    default:
      return null;
  }
}

function TreeNode({ node, depth = 0 }: { node: ASTNode; depth?: number }) {
  const [expanded, setExpanded] = useState(depth < 3);
  const hasChildren = node.children && node.children.length > 0;

  return (
    <div style={{ paddingLeft: depth * 16 }}>
      <button
        onClick={() => hasChildren && setExpanded(!expanded)}
        className={cn(
          "flex items-center gap-1.5 py-0.5 px-1 rounded text-sm w-full text-left",
          "hover:bg-muted/60 transition-colors",
          hasChildren && "cursor-pointer",
          !hasChildren && "cursor-default"
        )}
      >
        {hasChildren ? (
          expanded ? (
            <ChevronDown size={14} className="text-muted-foreground shrink-0" />
          ) : (
            <ChevronRight size={14} className="text-muted-foreground shrink-0" />
          )
        ) : (
          <span className="w-3.5 shrink-0" />
        )}
        <NodeIcon type={node.type} />
        <span
          className={cn(
            "font-mono",
            node.type === "function" && "text-blue-500 font-semibold",
            node.type === "ticker" && "text-green-500",
            node.type === "operator" && "text-red-500 font-semibold",
            node.type === "number" && "text-orange-500",
            node.type === "string" && "text-yellow-500",
            node.type === "duration" && "text-purple-500",
            node.type === "range" && "text-cyan-500"
          )}
        >
          {node.name}
        </span>
      </button>
      {hasChildren && expanded && (
        <div className="border-l border-muted ml-[7px]">
          {node.children!.map((child, i) => (
            <TreeNode key={`${child.name}-${i}`} node={child} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  );
}

export function ExpressionTree({ ast }: ExpressionTreeProps) {
  return (
    <div className="rounded-md border bg-card p-3 text-sm overflow-auto max-h-[400px]">
      <div className="text-xs text-muted-foreground font-medium mb-2">Expression Tree</div>
      <TreeNode node={ast} />
    </div>
  );
}
