package models

import "time"

// --- SEC Filings ---

// CompanyFiling represents an SEC filing.
type CompanyFiling struct {
	Date        time.Time `json:"date"`
	Symbol      string    `json:"symbol,omitempty"`
	CIK         string    `json:"cik"`
	CompanyName string    `json:"company_name"`
	FormType    string    `json:"form_type"`  // "10-K", "10-Q", "8-K", "S-1", etc.
	AccessionNo string    `json:"accession_no"`
	FilingURL   string    `json:"filing_url,omitempty"`
	Description string    `json:"description,omitempty"`
}

// CIKMapping represents a mapping from ticker/name to CIK number.
type CIKMapping struct {
	CIK     string `json:"cik"`
	Symbol  string `json:"symbol,omitempty"`
	Name    string `json:"name"`
}

// SICEntry represents a Standard Industrial Classification entry.
type SICEntry struct {
	SICCode     string `json:"sic_code"`
	Description string `json:"description"`
	Industry    string `json:"industry,omitempty"`
}

// InstitutionEntry represents a financial institution registered with SEC.
type InstitutionEntry struct {
	CIK  string `json:"cik"`
	Name string `json:"name"`
}

// --- CFTC ---

// COTReport represents Commitments of Traders report data.
type COTReport struct {
	Date                      time.Time `json:"date"`
	Market                    string    `json:"market"`
	Exchange                  string    `json:"exchange,omitempty"`
	// Commercial positions
	CommercialLong            int64     `json:"commercial_long"`
	CommercialShort           int64     `json:"commercial_short"`
	CommercialSpreading       int64     `json:"commercial_spreading,omitempty"`
	// Non-commercial (speculators)
	NonCommercialLong         int64     `json:"non_commercial_long"`
	NonCommercialShort        int64     `json:"non_commercial_short"`
	NonCommercialSpreading    int64     `json:"non_commercial_spreading,omitempty"`
	// Non-reportable
	NonReportableLong         int64     `json:"non_reportable_long,omitempty"`
	NonReportableShort        int64     `json:"non_reportable_short,omitempty"`
	// Open Interest
	OpenInterest              int64     `json:"open_interest"`
	// Changes
	ChangeInOpenInterest      int64     `json:"change_in_open_interest,omitempty"`
	ChangeCommercialLong      int64     `json:"change_commercial_long,omitempty"`
	ChangeCommercialShort     int64     `json:"change_commercial_short,omitempty"`
	ChangeNonCommercialLong   int64     `json:"change_non_commercial_long,omitempty"`
	ChangeNonCommercialShort  int64     `json:"change_non_commercial_short,omitempty"`
}

// COTSearchResult represents a COT report search result (market/instrument).
type COTSearchResult struct {
	Code        string `json:"code"`
	Market      string `json:"market"`
	Exchange    string `json:"exchange,omitempty"`
	Category    string `json:"category,omitempty"` // "futures", "options", "combined"
}

// --- Congress ---

// CongressBill represents a congressional bill.
type CongressBill struct {
	BillID          string    `json:"bill_id"`
	BillType        string    `json:"bill_type,omitempty"` // "HR", "S", etc.
	BillNumber      int       `json:"bill_number,omitempty"`
	Congress        int       `json:"congress,omitempty"`
	Title           string    `json:"title"`
	Sponsor         string    `json:"sponsor,omitempty"`
	SponsorParty    string    `json:"sponsor_party,omitempty"`
	IntroducedDate  time.Time `json:"introduced_date,omitempty"`
	LatestAction    string    `json:"latest_action,omitempty"`
	LatestActionDate time.Time `json:"latest_action_date,omitempty"`
	Status          string    `json:"status,omitempty"`
	URL             string    `json:"url,omitempty"`
}

// --- Factor Models ---

// FamaFrenchFactor represents Fama-French factor data.
type FamaFrenchFactor struct {
	Date     time.Time          `json:"date"`
	Factors  map[string]float64 `json:"factors"` // e.g., "Mkt-RF", "SMB", "HML", "RMW", "CMA", "RF"
}
