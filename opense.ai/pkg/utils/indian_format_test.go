package utils

import "testing"

func TestFormatINR(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "₹0.00"},
		{100, "₹100.00"},
		{1000, "₹1,000.00"},
		{12345, "₹12,345.00"},
		{123456, "₹1,23,456.00"},
		{1234567, "₹12,34,567.00"},
		{12345678, "₹1,23,45,678.00"},
		{123456789, "₹12,34,56,789.00"},
		{2847.50, "₹2,847.50"},
		{-1234.56, "-₹1,234.56"},
		{100000, "₹1,00,000.00"},
		{10000000, "₹1,00,00,000.00"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatINR(tt.input)
			if result != tt.expected {
				t.Errorf("FormatINR(%f) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatINRCompact(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{500, "₹500.00"},
		{100000, "₹1 L"},
		{1500000, "₹15 L"},
		{10000000, "₹1 Cr"},
		{192734500000, "₹19273.45 Cr"},
		{1000000000000, "₹1 L Cr"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatINRCompact(tt.input)
			if result != tt.expected {
				t.Errorf("FormatINRCompact(%f) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToLakhsCrores(t *testing.T) {
	if got := ToLakhs(100000); got != 1.0 {
		t.Errorf("ToLakhs(100000) = %f, want 1.0", got)
	}
	if got := ToCrores(10000000); got != 1.0 {
		t.Errorf("ToCrores(10000000) = %f, want 1.0", got)
	}
	if got := FromLakhs(1.0); got != 100000 {
		t.Errorf("FromLakhs(1.0) = %f, want 100000", got)
	}
	if got := FromCrores(1.0); got != 10000000 {
		t.Errorf("FromCrores(1.0) = %f, want 10000000", got)
	}
}

func TestFormatPct(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{2.45, "+2.45%"},
		{-1.23, "-1.23%"},
		{0.0, "+0.00%"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatPct(tt.input)
			if result != tt.expected {
				t.Errorf("FormatPct(%f) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatVolume(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{500, "500"},
		{1500, "1.50 K"},
		{150000, "1.50 L"},
		{15000000, "1.50 Cr"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatVolume(tt.input)
			if result != tt.expected {
				t.Errorf("FormatVolume(%d) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}
