package models

import "time"

// CurrencyPair represents a currency pair.
type CurrencyPair struct {
	Symbol       string `json:"symbol"`        // e.g., "USDINR"
	BaseCurrency string `json:"base_currency"` // e.g., "USD"
	QuoteCurrency string `json:"quote_currency"` // e.g., "INR"
	Exchange     string `json:"exchange,omitempty"`
	Name         string `json:"name,omitempty"`
}

// CurrencySnapshot represents current currency exchange rate snapshot.
type CurrencySnapshot struct {
	Symbol       string    `json:"symbol"`
	BaseCurrency string    `json:"base_currency"`
	QuoteCurrency string   `json:"quote_currency"`
	Rate         float64   `json:"rate"`
	Change       float64   `json:"change"`
	ChangePct    float64   `json:"change_pct"`
	Open         float64   `json:"open"`
	High         float64   `json:"high"`
	Low          float64   `json:"low"`
	PrevClose    float64   `json:"prev_close"`
	Volume       int64     `json:"volume,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// CurrencyReferenceRate represents a central bank reference rate.
type CurrencyReferenceRate struct {
	Currency  string    `json:"currency"`
	Rate      float64   `json:"rate"`
	Source    string    `json:"source"` // e.g., "ECB", "BOE", "RBI"
	Date      time.Time `json:"date"`
}
