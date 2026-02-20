import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { ToolCallCard } from "@/components/chat/ToolCallCard";
import type { ToolCall } from "@/lib/types";

describe("ToolCallCard", () => {
  const baseTool: ToolCall = {
    id: "tc-1",
    name: "getQuote",
    status: "completed",
    arguments: { ticker: "RELIANCE" },
  };

  it("renders tool name", () => {
    render(<ToolCallCard toolCall={baseTool} />);
    expect(screen.getByText("getQuote")).toBeInTheDocument();
  });

  it("shows completed status icon", () => {
    render(<ToolCallCard toolCall={{ ...baseTool, status: "completed" }} />);
    expect(screen.getByText("getQuote")).toBeInTheDocument();
  });

  it("shows running status", () => {
    render(<ToolCallCard toolCall={{ ...baseTool, status: "running" }} />);
    expect(screen.getByText("getQuote")).toBeInTheDocument();
  });

  it("shows failed status", () => {
    render(<ToolCallCard toolCall={{ ...baseTool, status: "failed" }} />);
    expect(screen.getByText("getQuote")).toBeInTheDocument();
  });

  it("shows pending status", () => {
    render(<ToolCallCard toolCall={{ ...baseTool, status: "pending" }} />);
    expect(screen.getByText("getQuote")).toBeInTheDocument();
  });
});
