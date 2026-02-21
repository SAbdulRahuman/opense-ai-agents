// Package cboe implements a CBOE (Chicago Board Options Exchange) data provider.
// CBOE offers free delayed market data via its CDN JSON APIs — no API key required.
// Coverage: US equities, ETFs, indices, options chains, VIX futures curve.
package cboe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/seenimoa/openseai/internal/infra"
	"github.com/seenimoa/openseai/internal/provider"
)

const (
	providerName = "cboe"

	// CDN base URLs.
	baseDelayedQuotes = "https://cdn.cboe.com/api/global/delayed_quotes"
	baseUSIndices     = "https://cdn.cboe.com/api/global/us_indices"
	baseEUIndices     = "https://cdn.cboe.com/api/global/european_indices"

	// Specific endpoints.
	urlCompanyDir   = "https://www.cboe.com/us/options/symboldir/equity_index_options/?download=csv"
	urlAllIndices   = baseUSIndices + "/definitions/all_indices.json"
	urlAllUSSnaps   = baseDelayedQuotes + "/quotes/all_us_indices.json"
	urlAllEUSnaps   = baseEUIndices + "/index_quotes/all-indices.json"
	urlFuturesRoots = baseDelayedQuotes + "/symbol_book/futures-roots.json"
)

// tickerExceptions are symbols that require an underscore prefix in CBOE CDN URLs.
var tickerExceptions = map[string]bool{
	"NDX": true,
	"RUT": true,
}

// Provider is the CBOE data provider.
type Provider struct {
	provider.BaseProvider

	// indexDir caches the CBOE index directory (refreshed every 24 h).
	indexDirMu    sync.RWMutex
	indexDir      []cboeIndexDef
	indexDirTime  time.Time
	indexSymbols  map[string]bool // fast lookup

	// companyDir caches the CBOE company/options directory.
	companyDirMu   sync.RWMutex
	companyDir     []cboeCompanyEntry
	companyDirTime time.Time
	companySymbols map[string]bool
}

// New creates a new CBOE provider and registers all fetchers.
func New() *Provider {
	p := &Provider{
		BaseProvider: provider.NewBaseProvider(
			providerName,
			"CBOE - Chicago Board Options Exchange, free delayed market data",
			"https://www.cboe.com",
			nil, // no credentials required
		),
		indexSymbols:   make(map[string]bool),
		companySymbols: make(map[string]bool),
	}

	// Equity fetchers (pass p for symbol path resolution).
	p.RegisterFetcher(newEquityHistoricalFetcher(p))
	p.RegisterFetcher(newEquityQuoteFetcher(p))
	p.RegisterFetcher(newEquitySearchFetcher(p))

	// ETF uses same logic as equity historical.
	p.RegisterFetcher(newEtfHistoricalFetcher(p))

	// Index fetchers.
	p.RegisterFetcher(newAvailableIndicesFetcher(p))
	p.RegisterFetcher(newIndexHistoricalFetcher(p))
	p.RegisterFetcher(newIndexSearchFetcher(p))
	p.RegisterFetcher(newIndexSnapshotsFetcher())
	p.RegisterFetcher(newIndexConstituentsFetcher())

	// Derivatives fetchers.
	p.RegisterFetcher(newOptionsChainsFetcher(p))
	p.RegisterFetcher(newFuturesCurveFetcher())

	return p
}

// Ping verifies connectivity to CBOE by fetching the index directory.
func (p *Provider) Ping(ctx context.Context) error {
	_, err := p.getIndexDirectory(ctx)
	return err
}

// ---------------------------------------------------------------------------
// Directory helpers — cached for 24 hours.
// ---------------------------------------------------------------------------

const dirCacheTTL = 24 * time.Hour

// getIndexDirectory returns the cached CBOE index directory, refreshing if stale.
func (p *Provider) getIndexDirectory(ctx context.Context) ([]cboeIndexDef, error) {
	p.indexDirMu.RLock()
	if len(p.indexDir) > 0 && time.Since(p.indexDirTime) < dirCacheTTL {
		defer p.indexDirMu.RUnlock()
		return p.indexDir, nil
	}
	p.indexDirMu.RUnlock()

	p.indexDirMu.Lock()
	defer p.indexDirMu.Unlock()

	// Double-check after acquiring write lock.
	if len(p.indexDir) > 0 && time.Since(p.indexDirTime) < dirCacheTTL {
		return p.indexDir, nil
	}

	var indices []cboeIndexDef
	if err := fetchCBOEJSON(ctx, urlAllIndices, &indices); err != nil {
		return nil, fmt.Errorf("cboe index directory: %w", err)
	}

	// Build fast lookup map.
	syms := make(map[string]bool, len(indices))
	for _, idx := range indices {
		syms[idx.IndexSymbol] = true
	}

	p.indexDir = indices
	p.indexDirTime = time.Now()
	p.indexSymbols = syms
	return indices, nil
}

// isIndexSymbol checks if symbol is a CBOE index.
func (p *Provider) isIndexSymbol(sym string) bool {
	p.indexDirMu.RLock()
	defer p.indexDirMu.RUnlock()
	return p.indexSymbols[sym]
}

// ---------------------------------------------------------------------------
// URL builders.
// ---------------------------------------------------------------------------

// needsUnderscore returns true if the symbol needs an underscore prefix in CBOE URLs.
func (p *Provider) needsUnderscore(sym string) bool {
	return tickerExceptions[sym] || p.isIndexSymbol(sym)
}

// symbolPath returns the symbol with optional underscore prefix for CBOE CDN URLs.
func (p *Provider) symbolPath(sym string) string {
	sym = strings.ReplaceAll(sym, "^", "")
	if p.needsUnderscore(sym) {
		return "_" + sym
	}
	return sym
}

// quotesURL returns the delayed quotes URL for a symbol.
func quotesURL(symPath string) string {
	return baseDelayedQuotes + "/quotes/" + symPath + ".json"
}

// chartURL returns the historical or intraday chart URL.
func chartURL(symPath, interval string) string {
	mode := "historical"
	if interval == "1m" {
		mode = "intraday"
	}
	return baseDelayedQuotes + "/charts/" + mode + "/" + symPath + ".json"
}

// optionsURL returns the options chain URL for a symbol.
func optionsURL(symPath string) string {
	return baseDelayedQuotes + "/options/" + symPath + ".json"
}

// ---------------------------------------------------------------------------
// HTTP helpers.
// ---------------------------------------------------------------------------

var cboeHeaders = map[string]string{
	"Accept":          "application/json",
	"Accept-Language": "en-US,en;q=0.9",
}

// fetchCBOEJSON fetches a CBOE CDN JSON endpoint and decodes into dst.
func fetchCBOEJSON(ctx context.Context, url string, dst any) error {
	body, status, err := infra.DoGet(ctx, url, cboeHeaders)
	if err != nil {
		return err
	}
	defer body.Close()
	if status >= 400 {
		b, _ := io.ReadAll(body)
		return fmt.Errorf("cboe HTTP %d: %s", status, string(b))
	}
	return json.NewDecoder(body).Decode(dst)
}

// fetchCBOERaw fetches a CBOE endpoint and returns the raw bytes.
func fetchCBOERaw(ctx context.Context, url string) ([]byte, error) {
	body, status, err := infra.DoGet(ctx, url, cboeHeaders)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	if status >= 400 {
		b, _ := io.ReadAll(body)
		return nil, fmt.Errorf("cboe HTTP %d: %s", status, string(b))
	}
	return io.ReadAll(body)
}

// ---------------------------------------------------------------------------
// Tiny parse helpers.
// ---------------------------------------------------------------------------

func parseFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

func parseInt(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	}
	return 0
}

// newResult creates a FetchResult with the current timestamp.
func newResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
	}
}
