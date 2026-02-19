// Package utils provides common utility functions for OpeNSE.ai.
package utils

import (
	"fmt"
	"math"
	"strings"
)

// FormatINR formats a number in Indian Rupee format (₹12,34,567.89).
// Uses the Indian numbering system: last 3 digits, then groups of 2.
func FormatINR(amount float64) string {
	negative := amount < 0
	amount = math.Abs(amount)

	intPart := int64(amount)
	decPart := amount - float64(intPart)

	formatted := formatIndianNumber(intPart)

	if decPart > 0 {
		decStr := fmt.Sprintf("%.2f", decPart)
		formatted += decStr[1:] // skip the leading "0"
	} else {
		formatted += ".00"
	}

	if negative {
		return "-₹" + formatted
	}
	return "₹" + formatted
}

// FormatINRCompact formats a number in compact Indian notation.
// e.g., 1927345 → "₹19.27 L", 192734500000 → "₹1,92,734.50 Cr"
func FormatINRCompact(amount float64) string {
	negative := amount < 0
	amount = math.Abs(amount)

	prefix := "₹"
	if negative {
		prefix = "-₹"
	}

	switch {
	case amount >= 1e12:
		// Lakh crores
		return fmt.Sprintf("%s%s L Cr", prefix, formatWithDecimals(amount/1e12))
	case amount >= 1e7:
		// Crores
		return fmt.Sprintf("%s%s Cr", prefix, formatWithDecimals(amount/1e7))
	case amount >= 1e5:
		// Lakhs
		return fmt.Sprintf("%s%s L", prefix, formatWithDecimals(amount/1e5))
	case amount >= 1e3:
		// Thousands
		return fmt.Sprintf("%s%s K", prefix, formatWithDecimals(amount/1e3))
	default:
		return fmt.Sprintf("%s%.2f", prefix, amount)
	}
}

// ToLakhs converts a raw number to lakhs.
func ToLakhs(amount float64) float64 {
	return amount / 1e5
}

// ToCrores converts a raw number to crores.
func ToCrores(amount float64) float64 {
	return amount / 1e7
}

// FromLakhs converts lakhs to raw number.
func FromLakhs(lakhs float64) float64 {
	return lakhs * 1e5
}

// FromCrores converts crores to raw number.
func FromCrores(crores float64) float64 {
	return crores * 1e7
}

// FormatPct formats a percentage value with sign and suffix.
// e.g., 2.45 → "+2.45%", -1.23 → "-1.23%"
func FormatPct(pct float64) string {
	if pct >= 0 {
		return fmt.Sprintf("+%.2f%%", pct)
	}
	return fmt.Sprintf("%.2f%%", pct)
}

// FormatVolume formats volume in human-readable Indian format.
// e.g., 1500000 → "15.00 L", 25000000 → "2.50 Cr"
func FormatVolume(volume int64) string {
	v := float64(volume)
	switch {
	case v >= 1e7:
		return fmt.Sprintf("%.2f Cr", v/1e7)
	case v >= 1e5:
		return fmt.Sprintf("%.2f L", v/1e5)
	case v >= 1e3:
		return fmt.Sprintf("%.2f K", v/1e3)
	default:
		return fmt.Sprintf("%d", volume)
	}
}

// formatIndianNumber formats an integer with Indian grouping (last 3, then 2s).
func formatIndianNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	s := fmt.Sprintf("%d", n)
	length := len(s)

	// Take the last 3 digits
	result := s[length-3:]
	remaining := s[:length-3]

	// Group remaining digits in pairs from right
	for len(remaining) > 0 {
		if len(remaining) > 2 {
			result = remaining[len(remaining)-2:] + "," + result
			remaining = remaining[:len(remaining)-2]
		} else {
			result = remaining + "," + result
			remaining = ""
		}
	}

	return result
}

// formatWithDecimals formats a number with up to 2 decimal places,
// removing trailing zeros.
func formatWithDecimals(n float64) string {
	s := fmt.Sprintf("%.2f", n)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
