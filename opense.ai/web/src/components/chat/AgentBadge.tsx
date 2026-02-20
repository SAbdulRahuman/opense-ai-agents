"use client";

import { Badge } from "@/components/ui/badge";

const agentColors: Record<string, string> = {
  "Market Analyst": "bg-blue-500/10 text-blue-600 border-blue-500/20",
  "Technical Analyst": "bg-purple-500/10 text-purple-600 border-purple-500/20",
  "Fundamental Analyst": "bg-green-500/10 text-green-600 border-green-500/20",
  "Sentiment Analyst": "bg-yellow-500/10 text-yellow-600 border-yellow-500/20",
  "Risk Manager": "bg-red-500/10 text-red-600 border-red-500/20",
  "Trade Executor": "bg-orange-500/10 text-orange-600 border-orange-500/20",
  "Orchestrator": "bg-cyan-500/10 text-cyan-600 border-cyan-500/20",
};

interface AgentBadgeProps {
  agent: string;
}

export function AgentBadge({ agent }: AgentBadgeProps) {
  const colorClass = agentColors[agent] || "bg-gray-500/10 text-gray-600 border-gray-500/20";

  return (
    <Badge variant="outline" className={colorClass}>
      {agent}
    </Badge>
  );
}
