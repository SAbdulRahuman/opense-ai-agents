import { describe, it, expect } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { ResultTable } from "@/components/financeql/ResultTable";
import type { TableResult } from "@/lib/types";

const mockData: TableResult = {
  columns: ["ticker", "close", "volume"],
  rows: [
    { ticker: "RELIANCE", close: 2500.5, volume: 1000000 },
    { ticker: "TCS", close: 3400.75, volume: 500000 },
    { ticker: "INFY", close: 1500.25, volume: 750000 },
  ],
};

describe("ResultTable", () => {
  it("renders table headers", () => {
    render(<ResultTable data={mockData} />);
    expect(screen.getByText("ticker")).toBeInTheDocument();
    expect(screen.getByText("close")).toBeInTheDocument();
    expect(screen.getByText("volume")).toBeInTheDocument();
  });

  it("renders table rows", () => {
    render(<ResultTable data={mockData} />);
    expect(screen.getByText("RELIANCE")).toBeInTheDocument();
    expect(screen.getByText("TCS")).toBeInTheDocument();
    expect(screen.getByText("INFY")).toBeInTheDocument();
  });

  it("shows row count", () => {
    render(<ResultTable data={mockData} />);
    expect(screen.getByText(/3 rows/)).toBeInTheDocument();
  });

  it("supports CSV export button", () => {
    render(<ResultTable data={mockData} />);
    expect(screen.getByText("CSV")).toBeInTheDocument();
  });
});
