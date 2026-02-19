package prompts

import "fmt"

// ── Indian Market–Specific Formatting & Context ──

// IndianMarketContext provides India-specific market context for agent prompts.
const IndianMarketContext = `
## Indian Market Context
- Exchange: NSE (National Stock Exchange) / BSE (Bombay Stock Exchange)
- Currency: Indian Rupee (₹ / INR)
- Market Hours: 9:15 AM – 3:30 PM IST (Pre-open: 9:00–9:15 AM)
- Settlement: T+1 rolling settlement
- Circuit Limits: 5%, 10%, 20% circuit breakers on stocks; index-wide halts at 10%, 15%, 20%
- Tick Size: ₹0.05 for stocks priced > ₹1
- F&O Lot Sizes: Vary by stock (e.g., NIFTY 25, BANKNIFTY 15, RELIANCE 250)
- Expiry: Weekly (Thu) for Nifty/Bank Nifty; Monthly (last Thu) for stock options
- Margin: SPAN + Exposure margin for F&O; VAR + ELM for cash
- Key Indices: NIFTY 50, NIFTY Bank, NIFTY IT, NIFTY Midcap 150, India VIX
- Taxation: 15% STCG (<1 yr), 10% LTCG (>1 yr, above ₹1L), STT on delivery & F&O
- Brokerage: STT + Exchange charges + GST (18%) + Stamp duty + SEBI turnover charge
`

// IndianNumberFormat describes Indian number formatting rules for agents.
const IndianNumberFormat = `
## Number Formatting Rules (Indian Convention)
- Use ₹ prefix for all monetary values: ₹2,847.50
- Indian comma grouping: ₹12,34,567 (not ₹1,234,567)
- Large numbers:
  - Thousands: ₹50,000 or ₹50K
  - Lakhs: ₹1,00,000 or ₹1L (= 100K)
  - Crores: ₹1,00,00,000 or ₹1Cr (= 10M)
  - Examples: Market Cap ₹19,27,345 Cr, Revenue ₹2.15L Cr
- Percentages: Always include % symbol: RSI 62.4%, PE 19.8x, ROE 48.2%
- Dates: DD-MMM-YYYY format (e.g., 19-Feb-2026)
- Time: IST (Indian Standard Time, UTC+5:30)
`

// IndianMarketPromptSuffix returns a prompt suffix with Indian market context.
// Append this to any agent's system prompt for India-specific awareness.
func IndianMarketPromptSuffix() string {
	return IndianMarketContext + IndianNumberFormat
}

// NSESectors lists key NSE sector classifications.
var NSESectors = map[string][]string{
	"IT":            {"TCS", "INFY", "WIPRO", "HCLTECH", "TECHM", "LTIM", "MPHASIS", "COFORGE", "PERSISTENT"},
	"Banking":       {"HDFCBANK", "ICICIBANK", "KOTAKBANK", "SBIN", "AXISBANK", "INDUSINDBK", "BANDHANBNK", "FEDERALBNK"},
	"NBFC":          {"BAJFINANCE", "BAJAJFINSV", "CHOLAFIN", "MUTHOOTFIN", "M&MFIN", "LICSG"},
	"Pharma":        {"SUNPHARMA", "DRREDDY", "CIPLA", "DIVISLAB", "BIOCON", "AUROPHARMA", "LUPIN", "TORNTPHARM"},
	"Auto":          {"MARUTI", "TATAMOTORS", "M&M", "BAJAJ-AUTO", "HEROMOTOCO", "ASHOKLEY", "EICHERMOT"},
	"Oil & Gas":     {"RELIANCE", "ONGC", "IOC", "BPCL", "HINDPETRO", "GAIL", "PETRONET"},
	"Metal":         {"TATASTEEL", "HINDALCO", "JSWSTEEL", "VEDL", "NATIONALUM", "COALINDIA", "NMDC"},
	"FMCG":          {"HINDUNILVR", "ITC", "NESTLEIND", "BRITANNIA", "DABUR", "GODREJCP", "MARICO", "COLPAL"},
	"Cement":        {"ULTRACEMCO", "GRASIM", "SHREECEM", "AMBUJACEM", "ACC", "DALMIA", "RAMCOCEM"},
	"Telecom":       {"BHARTIARTL", "IDEA", "TATACOMM"},
	"Power":         {"NTPC", "POWERGRID", "TATAPOWER", "ADANIPOWER", "NHPC", "SJVN"},
	"Infra":         {"LARSENTOUBR", "ADANIENT", "ADANIPORTS", "IRB", "NBCC", "KEC"},
	"Insurance":     {"SBILIFE", "HDFCLIFE", "ICICIPRULI", "STARHEALTH", "NIACL"},
	"Realty":        {"DLF", "GODREJPROP", "OBEROIRLTY", "PHOENIXLTD", "PRESTIGE", "BRIGADE"},
	"Capital Goods": {"ABB", "SIEMENS", "HAL", "BEL", "BHEL", "CUMMINSIND"},
}

// SectorForTicker returns the sector classification for a given NSE ticker.
// Returns empty string if the ticker is not in any known sector.
func SectorForTicker(ticker string) string {
	for sector, tickers := range NSESectors {
		for _, t := range tickers {
			if t == ticker {
				return sector
			}
		}
	}
	return ""
}

// SectorPeers returns peer tickers for the given ticker's sector.
// Excludes the ticker itself from the list.
func SectorPeers(ticker string) []string {
	sector := SectorForTicker(ticker)
	if sector == "" {
		return nil
	}
	tickers := NSESectors[sector]
	peers := make([]string, 0, len(tickers)-1)
	for _, t := range tickers {
		if t != ticker {
			peers = append(peers, t)
		}
	}
	return peers
}

// FormatTickerPrompt creates a ticker-specific prompt section with sector context.
func FormatTickerPrompt(ticker string) string {
	sector := SectorForTicker(ticker)
	if sector == "" {
		return fmt.Sprintf("Stock: %s (NSE)\nSector: Unknown — use tools to determine sector classification.\n", ticker)
	}
	peers := SectorPeers(ticker)
	peerStr := ""
	if len(peers) > 5 {
		peers = peers[:5]
	}
	for i, p := range peers {
		if i > 0 {
			peerStr += ", "
		}
		peerStr += p
	}
	return fmt.Sprintf("Stock: %s (NSE)\nSector: %s\nKey Peers: %s\n", ticker, sector, peerStr)
}

// IndianBrokerage calculates approximate brokerage costs for an NSE trade.
// Returns a formatted string showing the cost breakdown.
func IndianBrokerageEstimate(buyPrice, sellPrice float64, qty int, isDelivery bool) string {
	turnover := (buyPrice + sellPrice) * float64(qty)

	// STT (Securities Transaction Tax)
	var stt float64
	if isDelivery {
		stt = turnover * 0.001 // 0.1% on both buy + sell
	} else {
		stt = sellPrice * float64(qty) * 0.00025 // 0.025% on sell side only (intraday)
	}

	// Exchange charges: ~0.00345%
	exchangeChg := turnover * 0.0000345

	// SEBI charges: ₹10 per crore
	sebiChg := turnover * 0.000001

	// Stamp duty: 0.015% (buy side)
	stampDuty := buyPrice * float64(qty) * 0.00015

	// GST on brokerage + exchange charges (18%)
	gst := exchangeChg * 0.18

	total := stt + exchangeChg + sebiChg + stampDuty + gst

	return fmt.Sprintf("Brokerage Estimate:\n"+
		"  Turnover: ₹%.2f\n"+
		"  STT: ₹%.2f\n"+
		"  Exchange: ₹%.2f\n"+
		"  SEBI: ₹%.2f\n"+
		"  Stamp Duty: ₹%.2f\n"+
		"  GST: ₹%.2f\n"+
		"  Total: ₹%.2f",
		turnover, stt, exchangeChg, sebiChg, stampDuty, gst, total)
}
