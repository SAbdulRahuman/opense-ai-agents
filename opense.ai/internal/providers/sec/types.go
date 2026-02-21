package sec

import "time"

// --- EDGAR Full-Text Search (efts.sec.gov) ---

// edgarSearchResponse is the response from EDGAR full-text search API.
type edgarSearchResponse struct {
	Query   edgarSearchQuery `json:"query"`
	Hits    edgarSearchHits  `json:"hits"`
}

type edgarSearchQuery struct {
	From  int    `json:"from"`
	Size  int    `json:"size"`
	Query string `json:"q"`
}

type edgarSearchHits struct {
	Total edgarTotalHits    `json:"total"`
	Hits  []edgarSearchHit  `json:"hits"`
}

type edgarTotalHits struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

type edgarSearchHit struct {
	ID     string              `json:"_id"`
	Source edgarSearchDocument `json:"_source"`
}

type edgarSearchDocument struct {
	EntityName     string   `json:"entity_name"`
	FileNum        string   `json:"file_num"`
	FormType       string   `json:"form_type"`
	FiledAt        string   `json:"filed_at"`
	Period         string   `json:"period_of_report"`
	FileDate       string   `json:"file_date"`
	Tickers        []string `json:"tickers"`
	EntityType     string   `json:"entity_type"`
	CIKs          []string `json:"ciks"`
	DisplayNames  []string `json:"display_names"`
	FileDescription string  `json:"file_description"`
}

// --- EDGAR Submissions (data.sec.gov/submissions) ---

// edgarSubmissionsResponse is the response from company submissions endpoint.
type edgarSubmissionsResponse struct {
	CIK             string           `json:"cik"`
	EntityType      string           `json:"entityType"`
	SIC             string           `json:"sic"`
	SICDescription  string           `json:"sicDescription"`
	Name            string           `json:"name"`
	Tickers         []string         `json:"tickers"`
	Exchanges       []string         `json:"exchanges"`
	EIN             string           `json:"ein"`
	StateOfIncorp   string           `json:"stateOfIncorporation"`
	FiscalYearEnd   string           `json:"fiscalYearEnd"`
	Filings         edgarFilings     `json:"filings"`
}

type edgarFilings struct {
	Recent edgarFilingSet `json:"recent"`
	Files  []edgarFile    `json:"files"`
}

type edgarFilingSet struct {
	AccessionNumber []string `json:"accessionNumber"`
	FilingDate      []string `json:"filingDate"`
	ReportDate      []string `json:"reportDate"`
	Form            []string `json:"form"`
	PrimaryDocument []string `json:"primaryDocument"`
	IsXBRL          []int    `json:"isXBRL"`
	IsInlineXBRL    []int    `json:"isInlineXBRL"`
	FileNumber      []string `json:"fileNumber"`
	Items           []string `json:"items"`
	Description     []string `json:"primaryDocDescription"`
}

type edgarFile struct {
	Name        string `json:"name"`
	FilingCount int    `json:"filingCount"`
	FilingFrom  string `json:"filingFrom"`
	FilingTo    string `json:"filingTo"`
}

// --- EDGAR Company Facts (XBRL) ---

// edgarCompanyFactsResponse is the response from the company facts endpoint.
type edgarCompanyFactsResponse struct {
	CIK        int                            `json:"cik"`
	EntityName string                         `json:"entityName"`
	Facts      map[string]map[string]edgarFact `json:"facts"` // taxonomy -> concept -> fact
}

type edgarFact struct {
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Units       map[string][]edgarFactUnit `json:"units"` // unit type ("USD", "shares") -> values
}

type edgarFactUnit struct {
	Start    string  `json:"start"`
	End      string  `json:"end"`
	Val      float64 `json:"val"`
	Accn     string  `json:"accn"`
	FY       int     `json:"fy"`
	FP       string  `json:"fp"` // "Q1", "Q2", "Q3", "FY"
	Form     string  `json:"form"`
	Filed    string  `json:"filed"`
	Frame    string  `json:"frame,omitempty"`
}

// --- CIK / Ticker Mapping ---

// edgarTickerEntry is a row from the CIK<->ticker mapping file.
type edgarTickerEntry struct {
	CIKStr    string `json:"cik_str"`
	Ticker    string `json:"ticker"`
	Title     string `json:"title"`
}

// --- Company Tickers (full JSON) ---
// The company_tickers.json endpoint returns a map: {"0": {cik_str, ticker, title}, ...}

// --- Insider Transactions (via EDGAR owner forms 3/4/5) ---

// edgarOwnershipDoc represents an ownership filing document structure.
type edgarOwnershipDoc struct {
	Issuer      edgarIssuer         `json:"issuer"`
	ReportingOwner edgarReportingOwner `json:"reportingOwner"`
	Transactions []edgarTransaction  `json:"nonDerivativeTable"`
}

type edgarIssuer struct {
	CIK            string `json:"issuerCik"`
	Name           string `json:"issuerName"`
	TradingSymbol  string `json:"issuerTradingSymbol"`
}

type edgarReportingOwner struct {
	OwnerCIK  string `json:"rptOwnerCik"`
	OwnerName string `json:"rptOwnerName"`
}

type edgarTransaction struct {
	TransactionDate     string  `json:"transactionDate"`
	TransactionCode     string  `json:"transactionCode"`
	TransactionShares   float64 `json:"transactionShares"`
	TransactionPricePerShare float64 `json:"transactionPricePerShare"`
	SharesOwnedFollowing    float64 `json:"sharesOwnedFollowingTransaction"`
}

// --- SIC codes ---
type sicEntry struct {
	SICCode     string `json:"SIC"`
	Office      string `json:"Office,omitempty"`
	Description string `json:"Title"`
}

// --- RSS Litigation ---

// edgarRssLitigationResponse represents the SEC litigation releases RSS feed (parsed).
type edgarRssLitigation struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	PubDate     string `json:"pubDate"`
}

// --- Fail-to-Deliver ---

// ftdRecord represents a single fail-to-deliver record from SEC data.
type ftdRecord struct {
	SettlementDate string  `json:"SETTLEMENT DATE"`
	CUSIP          string  `json:"CUSIP"`
	Symbol         string  `json:"SYMBOL"`
	Quantity       string  `json:"QUANTITY (FAILS)"`
	Description    string  `json:"DESCRIPTION"`
	Price          string  `json:"PRICE"`
}

// --- 13F-HR (Institutional Holdings) ---

// form13FEntry represents an entry from the 13F-HR information table.
type form13FEntry struct {
	NameOfIssuer  string `json:"nameOfIssuer"`
	TitleOfClass  string `json:"titleOfClass"`
	CUSIP         string `json:"cusip"`
	Value         int64  `json:"value"` // In thousands of dollars
	SshPrnamnt    int64  `json:"sshPrnamnt"` // number of shares/principal amount
	SshPrntype    string `json:"sshPrntype"` // "SH" for shares
	PutCall       string `json:"putCall,omitempty"`
	InvestmentDiscretion string `json:"investmentDiscretion"`
}

// --- Helper for date parsing ---

func parseSECDate(s string) time.Time {
	// Try common SEC date formats.
	for _, layout := range []string{
		"2006-01-02",
		"2006-01-02T15:04:05.000Z",
		"01/02/2006",
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
