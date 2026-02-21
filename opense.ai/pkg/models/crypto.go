package models

// CryptoSearchResult represents a cryptocurrency search result.
type CryptoSearchResult struct {
	Symbol     string  `json:"symbol"`
	Name       string  `json:"name"`
	Exchange   string  `json:"exchange,omitempty"`
	MarketCap  float64 `json:"market_cap,omitempty"`
	Rank       int     `json:"rank,omitempty"`
	Currency   string  `json:"currency,omitempty"`
}

// CryptoQuote represents a real-time cryptocurrency quote.
type CryptoQuote struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Change24h     float64 `json:"change_24h"`
	ChangePct24h  float64 `json:"change_pct_24h"`
	MarketCap     float64 `json:"market_cap"`
	Volume24h     float64 `json:"volume_24h"`
	High24h       float64 `json:"high_24h"`
	Low24h        float64 `json:"low_24h"`
	CirculatingSupply float64 `json:"circulating_supply,omitempty"`
	TotalSupply       float64 `json:"total_supply,omitempty"`
	MaxSupply         float64 `json:"max_supply,omitempty"`
}
