import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Format a number in the Indian numbering system (lakhs/crores).
 * e.g., 1234567 → "12,34,567"
 */
export function formatIndianNumber(num: number): string {
  const isNegative = num < 0;
  const absNum = Math.abs(num);
  const parts = absNum.toFixed(2).split(".");
  let intPart = parts[0];
  const decPart = parts[1];

  // Indian grouping: last 3, then groups of 2
  if (intPart.length > 3) {
    const last3 = intPart.slice(-3);
    const remaining = intPart.slice(0, -3);
    const groups: string[] = [];
    for (let i = remaining.length; i > 0; i -= 2) {
      groups.unshift(remaining.slice(Math.max(0, i - 2), i));
    }
    intPart = groups.join(",") + "," + last3;
  }

  const formatted = decPart ? `${intPart}.${decPart}` : intPart;
  return isNegative ? `-${formatted}` : formatted;
}

/**
 * Format a price with ₹ symbol and Indian number formatting.
 */
export function formatPrice(price: number): string {
  return `₹${formatIndianNumber(price)}`;
}

/**
 * Format large numbers in lakhs/crores.
 * e.g., 10000000 → "1.00 Cr", 100000 → "1.00 L"
 */
export function formatLargeNumber(num: number): string {
  const absNum = Math.abs(num);
  const sign = num < 0 ? "-" : "";

  if (absNum >= 1e7) {
    return `${sign}${(absNum / 1e7).toFixed(2)} Cr`;
  }
  if (absNum >= 1e5) {
    return `${sign}${(absNum / 1e5).toFixed(2)} L`;
  }
  if (absNum >= 1e3) {
    return `${sign}${(absNum / 1e3).toFixed(2)} K`;
  }
  return `${sign}${absNum.toFixed(2)}`;
}

/**
 * Format percentage with 2 decimal places and % suffix.
 */
export function formatPercent(value: number): string {
  return `${value.toFixed(2)}%`;
}

/**
 * Format a volume number in a compact form.
 */
export function formatVolume(vol: number): string {
  if (vol >= 1e7) return `${(vol / 1e7).toFixed(2)} Cr`;
  if (vol >= 1e5) return `${(vol / 1e5).toFixed(2)} L`;
  if (vol >= 1e3) return `${(vol / 1e3).toFixed(1)} K`;
  return vol.toString();
}
