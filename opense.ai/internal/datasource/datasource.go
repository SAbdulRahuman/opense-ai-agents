// Package datasource provides data fetching from multiple financial data sources.
// It defines a common DataSource interface and implements concrete sources for
// Yahoo Finance, NSE India, derivatives, news, Screener.in, and FII/DII data.
package datasource

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/seenimoa/openseai/internal/infra"
	"github.com/seenimoa/openseai/pkg/models"
)

// DataSource defines the common interface that all data sources must implement.
// Each source may support a subset of methods; unsupported methods return ErrNotSupported.
type DataSource interface {
	// Name returns the human-readable name of this data source.
	Name() string

	// GetQuote returns a real-time (or near-real-time) quote for the given ticker.
	GetQuote(ctx context.Context, ticker string) (*models.Quote, error)

	// GetHistoricalData returns OHLCV candles for the given ticker and date range.
	GetHistoricalData(ctx context.Context, ticker string, from, to time.Time, tf models.Timeframe) ([]models.OHLCV, error)

	// GetFinancials returns financial statements for the given ticker.
	GetFinancials(ctx context.Context, ticker string) (*models.FinancialData, error)

	// GetOptionChain returns the option chain for the given ticker and optional expiry.
	GetOptionChain(ctx context.Context, ticker string, expiry string) (*models.OptionChain, error)

	// GetStockProfile returns an aggregated profile assembled from this source.
	GetStockProfile(ctx context.Context, ticker string) (*models.StockProfile, error)
}

// --- Sentinel errors ---

// ErrNotSupported is returned when a data source does not support a method.
var ErrNotSupported = fmt.Errorf("operation not supported by this data source")

// ErrTickerNotFound is returned when a ticker cannot be resolved.
var ErrTickerNotFound = fmt.Errorf("ticker not found")

// ErrRateLimited is returned when a source rate-limits the request.
var ErrRateLimited = fmt.Errorf("rate limited by data source")

// ErrHTTP wraps an HTTP error with status code.
type ErrHTTP struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *ErrHTTP) Error() string {
	return fmt.Sprintf("HTTP %d %s: %s", e.StatusCode, e.Status, e.Body)
}

// --- Shared HTTP client helpers ---

// DefaultUserAgent is the user agent string used for HTTP requests.
const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// HTTPClient is a pre-configured HTTP client with reasonable timeouts.
var HTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// doGet performs a GET request with the given URL and headers, returning the response body.
// The caller is responsible for closing the returned ReadCloser.
func doGet(ctx context.Context, url string, headers map[string]string) (io.ReadCloser, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	// Set default headers.
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "application/json, text/html, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	// Override/add custom headers.
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP GET %s: %w", url, err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, resp.StatusCode, &ErrHTTP{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
		}
	}

	return resp.Body, resp.StatusCode, nil
}

// --- Simple in-memory cache (delegated to infra package) ---

// CacheEntry is an alias for infra.CacheEntry.
type CacheEntry = infra.CacheEntry

// Cache is an alias for infra.Cache.
type Cache = infra.Cache

// NewCache creates a new cache with the given default TTL.
func NewCache(ttl time.Duration) *Cache {
	return infra.NewCache(ttl)
}

// --- Rate limiter (delegated to infra package) ---

// RateLimiter is an alias for infra.RateLimiter.
type RateLimiter = infra.RateLimiter

// NewRateLimiter creates a rate limiter that allows maxTokens requests
// per refillRate duration.
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return infra.NewRateLimiter(maxTokens, refillRate)
}
