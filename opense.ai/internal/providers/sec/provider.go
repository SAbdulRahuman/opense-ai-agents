// Package sec implements the SEC EDGAR data provider.
// SEC EDGAR provides free access to company filings, ownership data,
// insider trading, CIK mappings, and more via REST APIs.
//
// No API key required. Must include a User-Agent header per SEC policy.
// Docs: https://www.sec.gov/edgar/sec-api-documentation
// Rate limit: 10 requests/second per user-agent.
package sec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/seenimoa/openseai/internal/infra"
	"github.com/seenimoa/openseai/internal/provider"
)

const (
	providerName = "sec"

	// SEC EDGAR API endpoints.
	edgarBaseURL    = "https://efts.sec.gov/LATEST"       // Full-text search
	edgarDataURL    = "https://data.sec.gov"               // JSON data API
	edgarSubmissions = "https://data.sec.gov/submissions"  // Company submissions
	edgarCompanyURL = "https://data.sec.gov/api/xbrl/companyfacts" // XBRL company facts

	// SEC requires a User-Agent with company name, email for EDGAR requests.
	secUserAgent = "openseai/1.0 (github.com/seenimoa/openseai)"
)

// Provider implements provider.Provider for SEC EDGAR.
type Provider struct {
	provider.BaseProvider
}

// New creates a new SEC provider and registers all fetchers.
func New() *Provider {
	p := &Provider{
		BaseProvider: provider.NewBaseProvider(
			providerName,
			"SEC EDGAR - US Securities filings, ownership, and regulatory data",
			"https://www.sec.gov/edgar",
			nil, // No credentials required
		),
	}

	// --- Filings ---
	p.RegisterFetcher(newCompanyFilingsFetcher())
	p.RegisterFetcher(newSecFilingFetcher())
	p.RegisterFetcher(newLatestFinancialReportsFetcher())
	p.RegisterFetcher(newRssLitigationFetcher())

	// --- Mappings ---
	p.RegisterFetcher(newCikMapFetcher())
	p.RegisterFetcher(newSymbolMapFetcher())
	p.RegisterFetcher(newSicSearchFetcher())
	p.RegisterFetcher(newInstitutionsSearchFetcher())

	// --- Equity / Search ---
	p.RegisterFetcher(newEquitySearchFetcher())

	// --- Ownership ---
	p.RegisterFetcher(newInsiderTradingFetcher())
	p.RegisterFetcher(newInstitutionalOwnershipFetcher())
	p.RegisterFetcher(newEquityFTDFetcher())

	// --- Facts & Analysis ---
	p.RegisterFetcher(newCompareCompanyFactsFetcher())
	p.RegisterFetcher(newForm13FHRFetcher())

	return p
}

// Ping checks connectivity to SEC EDGAR.
func (p *Provider) Ping(ctx context.Context) error {
	url := edgarDataURL + "/submissions/CIK0000320193.json" // Apple
	body, _, err := infra.DoGet(ctx, url, secHeaders())
	if err != nil {
		return fmt.Errorf("sec ping: %w", err)
	}
	body.Close()
	return nil
}

// --- Shared helpers ---

func secHeaders() map[string]string {
	return map[string]string{
		"User-Agent": secUserAgent,
		"Accept":     "application/json",
	}
}

// fetchSECJSON performs a GET request to the SEC API and decodes JSON.
func fetchSECJSON(ctx context.Context, url string, dest any) error {
	body, _, err := infra.DoGet(ctx, url, secHeaders())
	if err != nil {
		return err
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read SEC response: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("parse SEC JSON: %w", err)
	}
	return nil
}

// fetchSECRaw performs a GET request and returns raw bytes.
func fetchSECRaw(ctx context.Context, url string) ([]byte, error) {
	body, _, err := infra.DoGet(ctx, url, secHeaders())
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return io.ReadAll(body)
}

// padCIK pads a CIK number to 10 digits with leading zeros.
func padCIK(cik string) string {
	for len(cik) < 10 {
		cik = "0" + cik
	}
	return cik
}

func newResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
	}
}

func newCachedResult(data any) *provider.FetchResult {
	return &provider.FetchResult{
		Data:      data,
		FetchedAt: time.Now(),
		Cached:    true,
	}
}
