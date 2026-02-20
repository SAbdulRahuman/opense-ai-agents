"use client";

import { useState } from "react";
import { ArrowUpDown, Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import type { TableResult } from "@/lib/types";

interface ResultTableProps {
  data: TableResult;
}

export function ResultTable({ data }: ResultTableProps) {
  const [sortCol, setSortCol] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<"asc" | "desc">("asc");
  const [page, setPage] = useState(0);
  const pageSize = 20;

  const handleSort = (col: string) => {
    if (sortCol === col) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortCol(col);
      setSortDir("asc");
    }
  };

  const sorted = [...data.rows].sort((a, b) => {
    if (!sortCol) return 0;
    const av = a[sortCol];
    const bv = b[sortCol];
    if (typeof av === "number" && typeof bv === "number") {
      return sortDir === "asc" ? av - bv : bv - av;
    }
    return sortDir === "asc"
      ? String(av).localeCompare(String(bv))
      : String(bv).localeCompare(String(av));
  });

  const paged = sorted.slice(page * pageSize, (page + 1) * pageSize);
  const totalPages = Math.ceil(sorted.length / pageSize);

  const exportCSV = () => {
    const header = data.columns.join(",");
    const rows = data.rows.map((r) =>
      data.columns.map((c) => JSON.stringify(r[c] ?? "")).join(","),
    );
    const csv = [header, ...rows].join("\n");
    const blob = new Blob([csv], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "financeql-result.csv";
    a.click();
    URL.revokeObjectURL(url);
  };

  const formatCell = (value: unknown): { text: string; className: string } => {
    if (typeof value === "number") {
      const isPercent = value > -100 && value < 100 && value !== Math.round(value);
      return {
        text: isPercent ? `${value.toFixed(2)}%` : value.toLocaleString("en-IN"),
        className: value > 0 ? "text-green-500" : value < 0 ? "text-red-500" : "",
      };
    }
    return { text: String(value ?? "—"), className: "" };
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-xs text-muted-foreground">
          {data.rows.length} rows
        </span>
        <Button variant="ghost" size="sm" onClick={exportCSV} className="gap-1 h-7">
          <Download className="h-3.5 w-3.5" />
          CSV
        </Button>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            {data.columns.map((col) => (
              <TableHead
                key={col}
                className="cursor-pointer hover:bg-muted/50"
                onClick={() => handleSort(col)}
              >
                <div className="flex items-center gap-1">
                  {col}
                  <ArrowUpDown className="h-3 w-3 text-muted-foreground" />
                </div>
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {paged.map((row, i) => (
            <TableRow key={i}>
              {data.columns.map((col) => {
                const cell = formatCell(row[col]);
                return (
                  <TableCell key={col} className={cn("font-mono text-xs", cell.className)}>
                    {cell.text}
                  </TableCell>
                );
              })}
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page === 0}
            onClick={() => setPage((p) => p - 1)}
          >
            Previous
          </Button>
          <span className="text-xs text-muted-foreground">
            Page {page + 1} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages - 1}
            onClick={() => setPage((p) => p + 1)}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}
