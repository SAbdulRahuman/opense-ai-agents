import { describe, it, expect } from "vitest";
import {
  cn,
  formatIndianNumber,
  formatPrice,
  formatLargeNumber,
  formatPercent,
  formatVolume,
} from "@/lib/utils";

describe("cn", () => {
  it("merges class names", () => {
    expect(cn("foo", "bar")).toBe("foo bar");
  });

  it("handles conditional classes", () => {
    expect(cn("foo", false && "bar", "baz")).toBe("foo baz");
  });

  it("deduplicates Tailwind classes", () => {
    expect(cn("px-2", "px-4")).toBe("px-4");
  });
});

describe("formatIndianNumber", () => {
  it("formats numbers in Indian style", () => {
    expect(formatIndianNumber(1234567)).toBe("12,34,567.00");
  });

  it("formats small numbers", () => {
    expect(formatIndianNumber(999)).toBe("999.00");
  });

  it("formats with decimals", () => {
    const result = formatIndianNumber(1234.56);
    expect(result).toBe("1,234.56");
  });

  it("handles zero", () => {
    expect(formatIndianNumber(0)).toBe("0.00");
  });
});

describe("formatPrice", () => {
  it("formats price with ₹ and 2 decimals", () => {
    const result = formatPrice(2500.5);
    expect(result).toContain("₹");
    expect(result).toContain("2,500.50");
  });

  it("handles large prices", () => {
    const result = formatPrice(12345.75);
    expect(result).toContain("₹");
    expect(result).toContain("12,345.75");
  });
});

describe("formatLargeNumber", () => {
  it("formats crores", () => {
    expect(formatLargeNumber(10000000)).toBe("1.00 Cr");
  });

  it("formats lakhs", () => {
    expect(formatLargeNumber(500000)).toBe("5.00 L");
  });

  it("formats thousands", () => {
    expect(formatLargeNumber(5000)).toBe("5.00 K");
  });

  it("returns small numbers as-is", () => {
    expect(formatLargeNumber(99)).toBe("99.00");
  });
});

describe("formatPercent", () => {
  it("formats percent with 2 decimals", () => {
    expect(formatPercent(12.345)).toBe("12.35%");
  });

  it("handles negative", () => {
    expect(formatPercent(-3.1)).toBe("-3.10%");
  });

  it("handles zero", () => {
    expect(formatPercent(0)).toBe("0.00%");
  });
});

describe("formatVolume", () => {
  it("formats large volumes", () => {
    expect(formatVolume(1500000)).toBe("15.00 L");
  });

  it("formats small volumes", () => {
    expect(formatVolume(500)).toBe("500");
  });
});
