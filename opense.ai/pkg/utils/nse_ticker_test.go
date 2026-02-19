package utils

import "testing"

func TestNormalizeTicker(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"RELIANCE", "RELIANCE"},
		{"reliance", "RELIANCE"},
		{" reliance ", "RELIANCE"},
		{"RIL", "RELIANCE"},
		{"$TCS", "TCS"},
		{"INFOSYS", "INFY"},
		{"HUL", "HINDUNILVR"},
		{"SBI", "SBIN"},
		{"AIRTEL", "BHARTIARTL"},
		{"NIFTY", "NIFTY 50"},
		{"BANKNIFTY", "NIFTY BANK"},
		{"UNKNOWNSTOCK", "UNKNOWNSTOCK"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeTicker(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeTicker(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToYFinanceTicker(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"RELIANCE", "RELIANCE.NS"},
		{"RIL", "RELIANCE.NS"},
		{"NIFTY", "^NSEI"},
		{"BANKNIFTY", "^NSEBANK"},
		{"SENSEX", "^BSESN"},
		{"TCS.NS", "TCS.NS"},
		{"UNKNOWN", "UNKNOWN.NS"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToYFinanceTicker(tt.input)
			if result != tt.expected {
				t.Errorf("ToYFinanceTicker(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFromYFinanceTicker(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"RELIANCE.NS", "RELIANCE"},
		{"TCS.NS", "TCS"},
		{"RELIANCE.BO", "RELIANCE"},
		{"RELIANCE", "RELIANCE"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FromYFinanceTicker(tt.input)
			if result != tt.expected {
				t.Errorf("FromYFinanceTicker(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsIndex(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"NIFTY", true},
		{"NIFTY 50", true},
		{"BANKNIFTY", true},
		{"SENSEX", true},
		{"RELIANCE", false},
		{"TCS", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsIndex(tt.input)
			if result != tt.expected {
				t.Errorf("IsIndex(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
