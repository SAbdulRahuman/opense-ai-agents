// Package datasource provides data fetching from multiple financial data sources.
// It defines a common DataSource interface and implements concrete sources for
// Yahoo Finance, NSE India, derivatives, news, Screener.in, and FII/DII data.
package datasource

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

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

// --- Simple in-memory cache ---

// CacheEntry holds a cached value with expiration.
type CacheEntry struct {
	Value     any
	ExpiresAt time.Time
}

// Cache is a simple thread-safe in-memory cache with TTL.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
	ttl     time.Duration
}

// NewCache creates a new cache with the given default TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a value from the cache. Returns nil, false if not found or expired.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Value, true
}

// Set stores a value in the cache with the default TTL.
func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	c.entries[key] = CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// SetWithTTL stores a value in the cache with a custom TTL.
func (c *Cache) SetWithTTL(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	c.entries[key] = CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

// Invalidate removes a key from the cache.
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Flush removes all entries from the cache.
func (c *Cache) Flush() {
	c.mu.Lock()
	c.entries = make(map[string]CacheEntry)
	c.mu.Unlock()
}

// Cleanup removes expired entries. Can be called periodically.
func (c *Cache) Cleanup() {
	c.mu.Lock()
	now := time.Now()
	for k, v := range c.entries {
		if now.After(v.ExpiresAt) {
			delete(c.entries, k)
		}
	}
	c.mu.Unlock()
}

// --- Rate limiter ---

// RateLimiter provides simple token-bucket rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   int
	maxTokens int
	refillRate time.Duration
	lastRefill time.Time
}

// NewRateLimiter creates a rate limiter that allows maxTokens requests
// per refillRate duration.
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		rl.mu.Lock()
		rl.refill()
		if rl.tokens > 0 {
			rl.tokens--
			rl.mu.Unlock()
			return nil
		}
		rl.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Check again after a short sleep.
		}
	}
}

// refill adds tokens based on elapsed time. Must be called with mu held.
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	if elapsed >= rl.refillRate {
		periods := int(elapsed / rl.refillRate)
		rl.tokens += periods
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = rl.lastRefill.Add(time.Duration(periods) * rl.refillRate)
	}
}
