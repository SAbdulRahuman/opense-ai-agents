import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ResultScalar } from "@/components/financeql/ResultScalar";

describe("ResultScalar", () => {
  it("renders scalar value with ticker and metric", () => {
    render(<ResultScalar data={{ value: 2500.75, label: "close price", ticker: "RELIANCE", metric: "close" }} />);
    expect(screen.getByText("RELIANCE")).toBeInTheDocument();
    expect(screen.getByText("close")).toBeInTheDocument();
  });

  it("renders with label only", () => {
    render(<ResultScalar data={{ value: 42, label: "count" }} />);
    expect(screen.getByText(/42/)).toBeInTheDocument();
    expect(screen.getByText("count")).toBeInTheDocument();
  });

  it("formats price-like values with ₹", () => {
    render(<ResultScalar data={{ value: 2500, label: "price" }} />);
    expect(screen.getByText(/₹/)).toBeInTheDocument();
  });
});
