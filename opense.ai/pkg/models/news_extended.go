package models

import "time"

// --- Extended News Models ---
// Complements the existing NewsArticle in analysis.go with richer structures.

// CompanyNewsArticle represents a news article associated with a specific company.
type CompanyNewsArticle struct {
	Symbol      string    `json:"symbol"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	Author      string    `json:"author,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	Content     string    `json:"content,omitempty"`
	Category    string    `json:"category,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	Language    string    `json:"language,omitempty"`
	Sentiment   string    `json:"sentiment,omitempty"` // "positive", "negative", "neutral"
	PublishedAt time.Time `json:"published_at"`
}

// WorldNewsArticle represents a world/market news article.
type WorldNewsArticle struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	Author      string    `json:"author,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	Content     string    `json:"content,omitempty"`
	Category    string    `json:"category,omitempty"`
	Region      string    `json:"region,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	Language    string    `json:"language,omitempty"`
	Tickers     []string  `json:"tickers,omitempty"`
	PublishedAt time.Time `json:"published_at"`
}
