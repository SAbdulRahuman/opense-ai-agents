package utils

import (
	"strings"
)

// Common NSE ticker aliases and normalizations.
var tickerAliases = map[string]string{
	"RELIANCE":    "RELIANCE",
	"RIL":         "RELIANCE",
	"TCS":         "TCS",
	"INFOSYS":     "INFY",
	"INFY":        "INFY",
	"HDFCBANK":    "HDFCBANK",
	"HDFC BANK":   "HDFCBANK",
	"ICICIBANK":   "ICICIBANK",
	"ICICI BANK":  "ICICIBANK",
	"SBIN":        "SBIN",
	"SBI":         "SBIN",
	"BHARTIARTL":  "BHARTIARTL",
	"AIRTEL":      "BHARTIARTL",
	"BAJFINANCE":  "BAJFINANCE",
	"BAJAJ FIN":   "BAJFINANCE",
	"ITC":         "ITC",
	"LT":          "LT",
	"L&T":         "LT",
	"TATAMOTORS":  "TATAMOTORS",
	"TATA MOTORS": "TATAMOTORS",
	"TATASTEEL":   "TATASTEEL",
	"TATA STEEL":  "TATASTEEL",
	"WIPRO":       "WIPRO",
	"HCLTECH":     "HCLTECH",
	"HCL TECH":    "HCLTECH",
	"MARUTI":      "MARUTI",
	"KOTAKBANK":   "KOTAKBANK",
	"KOTAK":       "KOTAKBANK",
	"AXISBANK":    "AXISBANK",
	"AXIS BANK":   "AXISBANK",
	"SUNPHARMA":   "SUNPHARMA",
	"SUN PHARMA":  "SUNPHARMA",
	"ASIANPAINT":  "ASIANPAINT",
	"ASIAN PAINTS":"ASIANPAINT",
	"TITAN":       "TITAN",
	"NESTLEIND":   "NESTLEIND",
	"NESTLE":      "NESTLEIND",
	"ULTRACEMCO":  "ULTRACEMCO",
	"ULTRATECH":   "ULTRACEMCO",
	"POWERGRID":   "POWERGRID",
	"NTPC":        "NTPC",
	"TECHM":       "TECHM",
	"TECH MAHINDRA": "TECHM",
	"M&M":         "M&M",
	"MAHINDRA":    "M&M",
	"ADANIENT":    "ADANIENT",
	"ADANI":       "ADANIENT",
	"HINDUNILVR":  "HINDUNILVR",
	"HUL":         "HINDUNILVR",
	"DRREDDY":     "DRREDDY",
	"CIPLA":       "CIPLA",
	"COALINDIA":   "COALINDIA",
	"COAL INDIA":  "COALINDIA",
	"ONGC":        "ONGC",
	"IOC":         "IOC",
	"BPCL":        "BPCL",
}

// NSE index tickers.
var indexTickers = map[string]string{
	"NIFTY":        "NIFTY 50",
	"NIFTY50":      "NIFTY 50",
	"NIFTY 50":     "NIFTY 50",
	"BANKNIFTY":    "NIFTY BANK",
	"NIFTYBANK":    "NIFTY BANK",
	"NIFTY BANK":   "NIFTY BANK",
	"FINNIFTY":     "NIFTY FIN SERVICE",
	"NIFTYIT":      "NIFTY IT",
	"NIFTY IT":     "NIFTY IT",
	"NIFTYMIDCAP":  "NIFTY MIDCAP 50",
	"SENSEX":       "SENSEX",
}

// NormalizeTicker normalizes a user-input ticker to the canonical NSE format.
// It handles aliases, uppercasing, and whitespace.
func NormalizeTicker(ticker string) string {
	ticker = strings.TrimSpace(strings.ToUpper(ticker))

	// Remove $ prefix if present (common in chat)
	ticker = strings.TrimPrefix(ticker, "$")

	// Check if it's an index
	if idx, ok := indexTickers[ticker]; ok {
		return idx
	}

	// Check aliases
	if canonical, ok := tickerAliases[ticker]; ok {
		return canonical
	}

	// Already normalized — return as-is
	return ticker
}

// ToYFinanceTicker converts an NSE ticker to Yahoo Finance format by appending .NS.
// Index tickers are converted to their Yahoo Finance format (^NSEI, ^NSEBANK, etc.).
func ToYFinanceTicker(ticker string) string {
	ticker = NormalizeTicker(ticker)

	// Handle index tickers
	switch ticker {
	case "NIFTY 50":
		return "^NSEI"
	case "NIFTY BANK":
		return "^NSEBANK"
	case "SENSEX":
		return "^BSESN"
	case "NIFTY IT":
		return "^CNXIT"
	case "NIFTY FIN SERVICE":
		return "^CNXFIN"
	}

	// Already has .NS suffix
	if strings.HasSuffix(ticker, ".NS") || strings.HasSuffix(ticker, ".BO") {
		return ticker
	}

	return ticker + ".NS"
}

// FromYFinanceTicker strips the .NS or .BO suffix to get the NSE/BSE ticker.
func FromYFinanceTicker(yfTicker string) string {
	yfTicker = strings.TrimSuffix(yfTicker, ".NS")
	yfTicker = strings.TrimSuffix(yfTicker, ".BO")
	return yfTicker
}

// IsIndex checks if the ticker is an index (not a stock).
func IsIndex(ticker string) bool {
	ticker = NormalizeTicker(ticker)
	_, ok := indexTickers[ticker]
	if ok {
		return true
	}
	// Also check if it was already resolved to an index name
	for _, v := range indexTickers {
		if v == ticker {
			return true
		}
	}
	return false
}
