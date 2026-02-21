package fred

import "time"

// --- FRED Series Search ---

type fredSearchResponse struct {
	RealtimeStart  string       `json:"realtime_start"`
	RealtimeEnd    string       `json:"realtime_end"`
	OrderBy        string       `json:"order_by"`
	SortOrder      string       `json:"sort_order"`
	Count          int          `json:"count"`
	Offset         int          `json:"offset"`
	Limit          int          `json:"limit"`
	Seriess        []fredSeries `json:"seriess"`
}

type fredSeries struct {
	ID                   string `json:"id"`
	RealtimeStart        string `json:"realtime_start"`
	RealtimeEnd          string `json:"realtime_end"`
	Title                string `json:"title"`
	ObservationStart     string `json:"observation_start"`
	ObservationEnd       string `json:"observation_end"`
	Frequency            string `json:"frequency"`
	FrequencyShort       string `json:"frequency_short"`
	Units                string `json:"units"`
	UnitsShort           string `json:"units_short"`
	SeasonalAdjustment   string `json:"seasonal_adjustment"`
	SeasonalAdjustmentShort string `json:"seasonal_adjustment_short"`
	LastUpdated          string `json:"last_updated"`
	Popularity           int    `json:"popularity"`
	Notes                string `json:"notes"`
}

// --- FRED Observations ---

type fredObservationsResponse struct {
	RealtimeStart  string             `json:"realtime_start"`
	RealtimeEnd    string             `json:"realtime_end"`
	ObservationStart string           `json:"observation_start"`
	ObservationEnd   string           `json:"observation_end"`
	Units          string             `json:"units"`
	OutputType     int                `json:"output_type"`
	FileType       string             `json:"file_type"`
	OrderBy        string             `json:"order_by"`
	SortOrder      string             `json:"sort_order"`
	Count          int                `json:"count"`
	Offset         int                `json:"offset"`
	Limit          int                `json:"limit"`
	Observations   []fredObservation  `json:"observations"`
}

type fredObservation struct {
	RealtimeStart string `json:"realtime_start"`
	RealtimeEnd   string `json:"realtime_end"`
	Date          string `json:"date"`
	Value         string `json:"value"`
}

// --- FRED Release ---

type fredReleaseResponse struct {
	RealtimeStart string        `json:"realtime_start"`
	RealtimeEnd   string        `json:"realtime_end"`
	Releases      []fredRelease `json:"releases"`
}

type fredRelease struct {
	ID            int    `json:"id"`
	RealtimeStart string `json:"realtime_start"`
	RealtimeEnd   string `json:"realtime_end"`
	Name          string `json:"name"`
	PressRelease  bool   `json:"press_release"`
	Link          string `json:"link"`
}

type fredReleaseTableResponse struct {
	Name     string                 `json:"name"`
	Elements map[string]fredElement `json:"elements"`
}

type fredElement struct {
	ElementID int    `json:"element_id"`
	SeriesID  string `json:"series_id"`
	Name      string `json:"name"`
	Units     string `json:"units"`
	Type      string `json:"type"`
}

// --- FRED Regional (GeoFRED) ---

type fredRegionalResponse struct {
	Title       string             `json:"title"`
	Region      string             `json:"region_type"`
	Frequency   string             `json:"frequency"`
	Units       string             `json:"units"`
	Data        map[string]fredRegionalValue `json:"data"` // region code -> value
}

type fredRegionalValue struct {
	Value string `json:"value"`
	Date  string `json:"date"`
}

// --- Mapping of well-known FRED series IDs ---

// fredSeriesMap maps common financial data types to their FRED series IDs.
var fredSeriesMap = map[string]string{
	// Overnight / Short-Term Rates
	"SOFR":                     "SOFR",
	"EFFR":                     "EFFR",       // Effective Federal Funds Rate
	"DFF":                      "DFF",        // Federal Funds Effective Rate (daily)
	"FEDFUNDS":                 "FEDFUNDS",   // Federal Funds Rate
	"AMERIBOR":                 "AMERIBOR",
	"OBFR":                     "OBFR",       // Overnight Bank Funding Rate
	"IORB":                     "IORB",       // Interest on Reserve Balances
	"DPCREDIT":                 "DPCREDIT",   // Discount Window Primary Credit Rate
	"ECBESTRVOLWGTTRMDMNRT":    "ECBESTRVOLWGTTRMDMNRT", // Euro Short-Term Rate

	// Treasury Rates
	"DGS1MO": "DGS1MO", // 1-Month Treasury
	"DGS3MO": "DGS3MO", // 3-Month Treasury
	"DGS6MO": "DGS6MO", // 6-Month Treasury
	"DGS1":   "DGS1",   // 1-Year Treasury
	"DGS2":   "DGS2",   // 2-Year Treasury
	"DGS3":   "DGS3",   // 3-Year Treasury
	"DGS5":   "DGS5",   // 5-Year Treasury
	"DGS7":   "DGS7",   // 7-Year Treasury
	"DGS10":  "DGS10",  // 10-Year Treasury
	"DGS20":  "DGS20",  // 20-Year Treasury
	"DGS30":  "DGS30",  // 30-Year Treasury

	// Treasury Bills
	"DTB3":  "DTB3",  // 3-Month Treasury Bill
	"DTB6":  "DTB6",  // 6-Month Treasury Bill
	"DTB1YR": "DTB1YR", // 1-Year Treasury Bill
	"TB3MS":  "TB3MS", // 3-Month Treasury Bill Secondary Market Rate

	// TIPS (Treasury Inflation-Protected Securities)
	"DFII5":  "DFII5",  // 5-Year TIPS
	"DFII7":  "DFII7",  // 7-Year TIPS
	"DFII10": "DFII10", // 10-Year TIPS
	"DFII20": "DFII20", // 20-Year TIPS
	"DFII30": "DFII30", // 30-Year TIPS

	// Corporate Bonds & Spreads
	"DAAA":     "DAAA",     // Moody's AAA Corporate Bond Yield
	"DBAA":     "DBAA",     // Moody's BAA Corporate Bond Yield
	"BAMLC0A0CM":  "BAMLC0A0CM",  // ICE BofA US Corporate Index OAS
	"BAMLH0A0HYM2": "BAMLH0A0HYM2", // ICE BofA US High Yield Index OAS

	// Mortgage Rates
	"MORTGAGE30US": "MORTGAGE30US", // 30-Year Fixed Rate Mortgage
	"MORTGAGE15US": "MORTGAGE15US", // 15-Year Fixed Rate Mortgage
	"MORTGAGE5US":  "MORTGAGE5US",  // 5/1-Year ARM

	// Commercial Paper
	"DCPF1M": "DCPF1M", // 1-Month Financial Commercial Paper Rate
	"DCPF3M": "DCPF3M", // 3-Month Financial Commercial Paper Rate
	"DCPN30": "DCPN30", // 30-Day Nonfinancial Commercial Paper Rate

	// Economy
	"CPIAUCSL":      "CPIAUCSL",      // CPI (All Urban Consumers)
	"CPILFESL":      "CPILFESL",      // Core CPI (Less Food and Energy)
	"PAYEMS":        "PAYEMS",        // Total Nonfarm Payrolls
	"UNRATE":        "UNRATE",        // Unemployment Rate
	"GDPC1":         "GDPC1",         // Real GDP
	"PCE":           "PCE",           // Personal Consumption Expenditures
	"PCEPILFE":      "PCEPILFE",      // Core PCE Price Index
	"UMCSENT":       "UMCSENT",       // University of Michigan Consumer Sentiment
	"GACDISA066MSFRBNY": "GACDISA066MSFRBNY", // NY Empire State Manufacturing Index
	"DRTSCILM":     "DRTSCILM",      // Senior Loan Officer Survey

	// Commodity & Retail
	"DCOILWTICO": "DCOILWTICO", // WTI Crude Oil Price
	"DCOILBRENTEU": "DCOILBRENTEU", // Brent Crude Oil Price
	"GOLDAMGBD228NLBM": "GOLDAMGBD228NLBM", // Gold Fixing Price

	// UK Rates
	"IUDSOIA": "IUDSOIA", // SONIA (Sterling Overnight Index Average)

	// ECB Rates
	"ECBMLFR": "ECBMLFR", // ECB Marginal Lending Facility Rate
	"ECBDFR":  "ECBDFR",  // ECB Deposit Facility Rate
	"ECBMRRFR": "ECBMRRFR", // ECB Main Refinancing Rate
}

// parseFredDate parses common FRED date formats.
func parseFredDate(s string) time.Time {
	for _, layout := range []string{
		"2006-01-02",
		"2006-01-02T15:04:05",
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
