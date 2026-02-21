package cboe

import "time"

// ---------------------------------------------------------------------------
// Index directory definitions (from all_indices.json).
// ---------------------------------------------------------------------------

// cboeIndexDef is a single entry from the CBOE index directory.
type cboeIndexDef struct {
	IndexSymbol   string `json:"index_symbol"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Currency      string `json:"currency,omitempty"`
	Source        string `json:"source,omitempty"`
	TimeZone      string `json:"time_zone,omitempty"`
	TickDays      string `json:"tick_days,omitempty"`
	TickFrequency string `json:"tick_frequency,omitempty"`
	TickPeriod    string `json:"tick_period,omitempty"`
	MktDataDelay  int    `json:"mkt_data_delay,omitempty"`
	CalcStartTime string `json:"calc_start_time,omitempty"`
	CalcEndTime   string `json:"calc_end_time,omitempty"`
}

// cboeCompanyEntry is a single entry from the CBOE company directory CSV.
type cboeCompanyEntry struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	DPMName     string `json:"dpm_name,omitempty"`
	PostStation string `json:"post_station,omitempty"`
}

// ---------------------------------------------------------------------------
// Delayed quotes response (for quotes/{SYMBOL}.json).
// ---------------------------------------------------------------------------

// cboeQuoteResponse is the top-level JSON response for a delayed quote.
type cboeQuoteResponse struct {
	Data cboeQuoteData `json:"data"`
}

// cboeQuoteData is the quote data payload.
type cboeQuoteData struct {
	Symbol             string  `json:"symbol"`
	SecurityType       string  `json:"security_type,omitempty"`
	CurrentPrice       float64 `json:"current_price"`
	PriceChange        float64 `json:"price_change"`
	PriceChangePct     float64 `json:"price_change_percent"`
	Bid                float64 `json:"bid"`
	BidSize            int64   `json:"bid_size"`
	Ask                float64 `json:"ask"`
	AskSize            int64   `json:"ask_size"`
	Open               float64 `json:"open"`
	High               float64 `json:"high"`
	Low                float64 `json:"low"`
	Close              float64 `json:"close"`
	Volume             int64   `json:"volume"`
	PrevDayClose       float64 `json:"prev_day_close"`
	AnnualHigh         float64 `json:"annual_high"`
	AnnualLow          float64 `json:"annual_low"`
	Tick               string  `json:"tick,omitempty"`
	IV30               float64 `json:"iv30,omitempty"`
	IV30Change         float64 `json:"iv30_change,omitempty"`
	IV30ChangePct      float64 `json:"iv30_change_percent,omitempty"`
	LastTradeTime      string  `json:"last_trade_time,omitempty"`
}

// ---------------------------------------------------------------------------
// Historical / intraday chart response.
// ---------------------------------------------------------------------------

// cboeChartResponse wraps the chart data (both daily and intraday).
type cboeChartResponse struct {
	Symbol string        `json:"symbol"`
	Data   []interface{} `json:"data"` // either dailyBar or intradayBar depending on interval
}

// cboeDailyBar represents a single daily OHLCV bar.
type cboeDailyBar struct {
	Date        string  `json:"date"`
	Open        float64 `json:"open"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	Close       float64 `json:"close"`
	StockVolume int64   `json:"stock_volume"`
}

// cboeIntradayBar represents a single intraday bar.
type cboeIntradayBar struct {
	Datetime string `json:"datetime"`
	Price    struct {
		Open  float64 `json:"open"`
		High  float64 `json:"high"`
		Low   float64 `json:"low"`
		Close float64 `json:"close"`
	} `json:"price"`
	Volume struct {
		StockVolume        int64 `json:"stock_volume"`
		CallsVolume        int64 `json:"calls_volume"`
		PutsVolume         int64 `json:"puts_volume"`
		TotalOptionsVolume int64 `json:"total_options_volume"`
	} `json:"volume"`
}

// ---------------------------------------------------------------------------
// Options chain response (from options/{SYMBOL}.json).
// ---------------------------------------------------------------------------

// cboeOptionsResponse wraps the options chain.
type cboeOptionsResponse struct {
	Data cboeOptionsPayload `json:"data"`
}

// cboeOptionsPayload is the inner payload.
type cboeOptionsPayload struct {
	Symbol          string             `json:"symbol"`
	SecurityType    string             `json:"security_type,omitempty"`
	CurrentPrice    float64            `json:"current_price"`
	Bid             float64            `json:"bid"`
	Ask             float64            `json:"ask"`
	Open            float64            `json:"open"`
	High            float64            `json:"high"`
	Low             float64            `json:"low"`
	Close           float64            `json:"close"`
	Volume          int64              `json:"volume"`
	PrevDayClose    float64            `json:"prev_day_close"`
	PriceChange     float64            `json:"price_change"`
	PriceChangePct  float64            `json:"percent_change"`
	IV30            float64            `json:"iv30,omitempty"`
	IV30Change      float64            `json:"iv30_change,omitempty"`
	IV30ChangePct   float64            `json:"iv30_change_percent,omitempty"`
	LastTradeTime   string             `json:"last_trade_time,omitempty"`
	Options         []cboeOptionRecord `json:"options"`
}

// cboeOptionRecord is a single option contract from CBOE.
type cboeOptionRecord struct {
	Option           string  `json:"option"` // encoded contract symbol e.g. "AAPL240119C00150000"
	Bid              float64 `json:"bid"`
	BidSize          int64   `json:"bid_size"`
	Ask              float64 `json:"ask"`
	AskSize          int64   `json:"ask_size"`
	IV               float64 `json:"iv"`
	OpenInterest     int64   `json:"open_interest"`
	Volume           int64   `json:"volume"`
	Delta            float64 `json:"delta"`
	Gamma            float64 `json:"gamma"`
	Theta            float64 `json:"theta"`
	Vega             float64 `json:"vega"`
	Rho              float64 `json:"rho"`
	Theo             float64 `json:"theo"`
	Change           float64 `json:"change"`
	PctChange        float64 `json:"percent_change"`
	PrevDayClose     float64 `json:"prev_day_close"`
	LastTradePrice   float64 `json:"last_trade_price"`
	LastTradeTime    string  `json:"last_trade_time,omitempty"`
}

// ---------------------------------------------------------------------------
// All-index snapshot response (from all_us_indices.json / all-indices.json).
// ---------------------------------------------------------------------------

// cboeSnapshotResponse wraps the list of index snapshots.
type cboeSnapshotResponse struct {
	Data []cboeSnapshotEntry `json:"data"`
}

// cboeSnapshotEntry represents a single index snapshot.
type cboeSnapshotEntry struct {
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name,omitempty"`
	SecurityType    string  `json:"security_type,omitempty"`
	CurrentPrice    float64 `json:"current_price"`
	PriceChange     float64 `json:"price_change"`
	PriceChangePct  float64 `json:"price_change_percent"`
	Bid             float64 `json:"bid,omitempty"`
	Ask             float64 `json:"ask,omitempty"`
	Open            float64 `json:"open,omitempty"`
	High            float64 `json:"high,omitempty"`
	Low             float64 `json:"low,omitempty"`
	Close           float64 `json:"close,omitempty"`
	Volume          int64   `json:"volume,omitempty"`
	PrevDayClose    float64 `json:"prev_day_close,omitempty"`
	LastTradeTime   string  `json:"last_trade_time,omitempty"`
}

// ---------------------------------------------------------------------------
// EU index constituents response.
// ---------------------------------------------------------------------------

// cboeConstituentResponse wraps the constituents.
type cboeConstituentResponse struct {
	Data []cboeConstituentEntry `json:"data"`
}

// cboeConstituentEntry represents a single constituent.
type cboeConstituentEntry struct {
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name,omitempty"`
	Type            string  `json:"type,omitempty"`
	SecurityType    string  `json:"security_type,omitempty"`
	CurrentPrice    float64 `json:"current_price"`
	PriceChange     float64 `json:"price_change"`
	PriceChangePct  float64 `json:"price_change_percent"`
	Open            float64 `json:"open,omitempty"`
	High            float64 `json:"high,omitempty"`
	Low             float64 `json:"low,omitempty"`
	Close           float64 `json:"close,omitempty"`
	Volume          int64   `json:"volume,omitempty"`
	PrevDayClose    float64 `json:"prev_day_close,omitempty"`
	Tick            string  `json:"tick,omitempty"`
	LastTradeTime   string  `json:"last_trade_time,omitempty"`
}

// ---------------------------------------------------------------------------
// Futures roots response.
// ---------------------------------------------------------------------------

// cboeFuturesRootsResponse wraps the futures roots list.
type cboeFuturesRootsResponse struct {
	Data []cboeFuturesRoot `json:"data"`
}

// cboeFuturesRoot describes a futures root symbol.
type cboeFuturesRoot struct {
	Symbol      string `json:"symbol"`
	Description string `json:"description,omitempty"`
	RootSymbol  string `json:"root_symbol,omitempty"`
}

// ---------------------------------------------------------------------------
// Parse helpers for timestamps.
// ---------------------------------------------------------------------------

// parseCBOETime parses CBOE's "2024-01-19T15:45:00" timestamp format.
func parseCBOETime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		// Try date-only fallback.
		t, _ = time.Parse("2006-01-02", s)
	}
	return t
}

// parseCBOEDate parses CBOE's "2024-01-19" date format.
func parseCBOEDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}
