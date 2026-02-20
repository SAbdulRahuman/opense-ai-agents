// ============================================================================
// OpeNSE.ai — DrawingToolbar (TradingView-style left sidebar)
// ============================================================================
"use client";

import { useState } from "react";
import { useStore } from "@/store";
import { TOOL_GROUPS, COLOR_PRESETS, type DrawingToolId } from "@/lib/drawings";
import { Button } from "@/components/ui/button";
import { Tooltip } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import {
  MousePointer2,
  Crosshair,
  TrendingUp,
  MoveUpRight,
  Minus,
  GripVertical,
  Columns3,
  GitBranch,
  GitMerge,
  Square,
  Circle,
  Triangle,
  Type,
  MessageSquare,
  Paintbrush,
  Ruler,
  Eraser,
  Eye,
  EyeOff,
  Magnet,
  Trash2,
  Lock,
  Unlock,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

const ICON_MAP: Record<string, LucideIcon> = {
  MousePointer2,
  Crosshair,
  TrendingUp,
  MoveUpRight,
  Minus,
  Grip: GripVertical,
  Columns3,
  GitBranch,
  GitMerge,
  Square,
  Circle,
  Triangle,
  Type,
  MessageSquare,
  Paintbrush,
  Ruler,
  Eraser,
};

export function DrawingToolbar() {
  const {
    activeTool,
    setActiveTool,
    activeStyle,
    setActiveStyle,
    showDrawings,
    setShowDrawings,
    magnetMode,
    setMagnetMode,
    clearAllDrawings,
    drawings,
    selectedDrawingId,
    toggleLockDrawing,
  } = useStore();

  const [expandedGroup, setExpandedGroup] = useState<string | null>(null);

  const selectedDrawing = drawings.find((d) => d.id === selectedDrawingId);

  return (
    <div className="flex h-full w-11 flex-col items-center gap-0.5 border-r bg-card py-2 overflow-y-auto">
      {/* Tool groups */}
      {TOOL_GROUPS.map((group) => (
        <div key={group.label} className="relative">
          {group.tools.map((tool) => {
            const Icon = ICON_MAP[tool.icon] || MousePointer2;
            const isActive = activeTool === tool.id;

            return (
              <Tooltip key={tool.id} content={`${tool.label}${tool.shortcut ? ` (${tool.shortcut})` : ""}`} side="right">
                <Button
                  variant={isActive ? "default" : "ghost"}
                  size="sm"
                  className={cn(
                    "h-8 w-8 p-0",
                    isActive && "bg-primary text-primary-foreground",
                  )}
                  onClick={() => setActiveTool(tool.id)}
                >
                  <Icon className="h-4 w-4" />
                </Button>
              </Tooltip>
            );
          })}
          {/* Divider between groups */}
          <div className="mx-auto my-1 h-px w-6 bg-border" />
        </div>
      ))}

      {/* Spacer */}
      <div className="flex-1" />

      {/* Quick actions at bottom */}
      <div className="flex flex-col items-center gap-1 border-t pt-2">
        {/* Color quick-pick */}
        <Tooltip content="Drawing Color" side="right">
          <button
            className="h-5 w-5 rounded-full border-2 border-border"
            style={{ backgroundColor: activeStyle.color }}
            onClick={() => {
              const idx = COLOR_PRESETS.indexOf(activeStyle.color);
              const next = COLOR_PRESETS[(idx + 1) % COLOR_PRESETS.length];
              setActiveStyle({ color: next, fillColor: next });
            }}
          />
        </Tooltip>

        {/* Magnet mode */}
        <Tooltip content={`Magnet ${magnetMode ? "On" : "Off"}`} side="right">
          <Button
            variant={magnetMode ? "secondary" : "ghost"}
            size="sm"
            className="h-8 w-8 p-0"
            onClick={() => setMagnetMode(!magnetMode)}
          >
            <Magnet className={cn("h-4 w-4", magnetMode && "text-primary")} />
          </Button>
        </Tooltip>

        {/* Toggle visibility */}
        <Tooltip content={showDrawings ? "Hide Drawings" : "Show Drawings"} side="right">
          <Button
            variant="ghost"
            size="sm"
            className="h-8 w-8 p-0"
            onClick={() => setShowDrawings(!showDrawings)}
          >
            {showDrawings ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
          </Button>
        </Tooltip>

        {/* Lock selected */}
        {selectedDrawing && (
          <Tooltip content={selectedDrawing.locked ? "Unlock" : "Lock"} side="right">
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0"
              onClick={() => toggleLockDrawing(selectedDrawing.id)}
            >
              {selectedDrawing.locked ? (
                <Lock className="h-4 w-4 text-yellow-500" />
              ) : (
                <Unlock className="h-4 w-4" />
              )}
            </Button>
          </Tooltip>
        )}

        {/* Clear all */}
        {drawings.length > 0 && (
          <Tooltip content="Clear All Drawings" side="right">
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0 text-destructive hover:text-destructive"
              onClick={clearAllDrawings}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </Tooltip>
        )}
      </div>
    </div>
  );
}
