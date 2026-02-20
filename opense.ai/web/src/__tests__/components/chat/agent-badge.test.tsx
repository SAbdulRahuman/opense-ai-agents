import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { AgentBadge } from "@/components/chat/AgentBadge";

describe("AgentBadge", () => {
  it("renders market analyst badge", () => {
    render(<AgentBadge agent="Market Analyst" />);
    expect(screen.getByText("Market Analyst")).toBeInTheDocument();
  });

  it("renders technical analyst badge", () => {
    render(<AgentBadge agent="Technical Analyst" />);
    expect(screen.getByText("Technical Analyst")).toBeInTheDocument();
  });

  it("renders fundamental analyst badge", () => {
    render(<AgentBadge agent="Fundamental Analyst" />);
    expect(screen.getByText("Fundamental Analyst")).toBeInTheDocument();
  });

  it("renders risk manager badge", () => {
    render(<AgentBadge agent="Risk Manager" />);
    expect(screen.getByText("Risk Manager")).toBeInTheDocument();
  });

  it("renders orchestrator badge", () => {
    render(<AgentBadge agent="Orchestrator" />);
    expect(screen.getByText("Orchestrator")).toBeInTheDocument();
  });

  it("renders unknown agents with raw name", () => {
    render(<AgentBadge agent="custom_agent" />);
    expect(screen.getByText("custom_agent")).toBeInTheDocument();
  });
});
