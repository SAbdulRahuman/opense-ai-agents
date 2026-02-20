import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { ExpressionTree, type ASTNode } from "@/components/financeql/ExpressionTree";

const mockAST: ASTNode = {
  type: "function",
  name: "sma",
  children: [
    {
      type: "function",
      name: "close",
      children: [
        { type: "ticker", name: "RELIANCE" },
      ],
    },
    { type: "number", name: "20" },
  ],
};

describe("ExpressionTree", () => {
  it("renders root node", () => {
    render(<ExpressionTree ast={mockAST} />);
    expect(screen.getByText("sma")).toBeInTheDocument();
  });

  it("renders child nodes", () => {
    render(<ExpressionTree ast={mockAST} />);
    expect(screen.getByText("close")).toBeInTheDocument();
    expect(screen.getByText("RELIANCE")).toBeInTheDocument();
    expect(screen.getByText("20")).toBeInTheDocument();
  });

  it("shows Expression Tree header", () => {
    render(<ExpressionTree ast={mockAST} />);
    expect(screen.getByText("Expression Tree")).toBeInTheDocument();
  });

  it("collapses nodes on click", () => {
    render(<ExpressionTree ast={mockAST} />);
    // Initially expanded (depth < 3)
    expect(screen.getByText("RELIANCE")).toBeInTheDocument();
    // Click to collapse sma node
    fireEvent.click(screen.getByText("sma"));
    // Children should be hidden
    expect(screen.queryByText("close")).not.toBeInTheDocument();
  });
});
