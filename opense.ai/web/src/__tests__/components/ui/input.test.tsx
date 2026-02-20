import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Input } from "@/components/ui/input";

describe("Input", () => {
  it("renders an input element", () => {
    render(<Input placeholder="Enter text" />);
    expect(screen.getByPlaceholderText("Enter text")).toBeInTheDocument();
  });

  it("accepts type prop", () => {
    render(<Input type="number" placeholder="Number" />);
    const input = screen.getByPlaceholderText("Number");
    expect(input).toHaveAttribute("type", "number");
  });

  it("applies custom className", () => {
    render(<Input className="font-mono" placeholder="Test" />);
    expect(screen.getByPlaceholderText("Test")).toHaveClass("font-mono");
  });

  it("can be disabled", () => {
    render(<Input disabled placeholder="Disabled" />);
    expect(screen.getByPlaceholderText("Disabled")).toBeDisabled();
  });
});
